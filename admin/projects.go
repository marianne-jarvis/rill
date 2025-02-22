package admin

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// TODO: The functions in this file are not truly fault tolerant. They should be refactored to run as idempotent, retryable background tasks.

// CreateProject creates a new project and provisions and reconciles a prod deployment for it.
func (s *Service) CreateProject(ctx context.Context, org *database.Organization, userID string, opts *database.InsertProjectOptions) (*database.Project, error) {
	// Check Github info is set (presently required to make a deployment)
	if opts.GithubURL == nil || opts.GithubInstallationID == nil || opts.ProdBranch == "" {
		return nil, fmt.Errorf("cannot create project without github info")
	}

	// Get roles for initial setup
	adminRole, err := s.DB.FindProjectRole(ctx, database.ProjectRoleNameAdmin)
	if err != nil {
		panic(err)
	}
	viewerRole, err := s.DB.FindProjectRole(ctx, database.ProjectRoleNameViewer)
	if err != nil {
		panic(err)
	}

	// Create the project and add initial members using a transaction.
	// The transaction is not used for provisioning and deployments, since they involve external services.
	txCtx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	proj, err := s.DB.InsertProject(txCtx, opts)
	if err != nil {
		return nil, err
	}

	// The creating user becomes project admin
	err = s.DB.InsertProjectMemberUser(txCtx, proj.ID, userID, adminRole.ID)
	if err != nil {
		return nil, err
	}

	// All org members as a group get viewer role
	err = s.DB.InsertProjectMemberUsergroup(txCtx, *org.AllUsergroupID, proj.ID, viewerRole.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// Provision prod deployment.
	// Start using original context again since transaction in txCtx is done.
	depl, err := s.createDeployment(ctx, &createDeploymentOptions{
		ProjectID:            proj.ID,
		Region:               proj.Region,
		GithubURL:            proj.GithubURL,
		GithubInstallationID: proj.GithubInstallationID,
		Subpath:              proj.Subpath,
		ProdBranch:           proj.ProdBranch,
		ProdVariables:        proj.ProdVariables,
		ProdOLAPDriver:       proj.ProdOLAPDriver,
		ProdOLAPDSN:          proj.ProdOLAPDSN,
		ProdSlots:            proj.ProdSlots,
		Annotations:          newDeploymentAnnotations(org, proj),
	})
	if err != nil {
		err2 := s.DB.DeleteProject(ctx, proj.ID)
		return nil, multierr.Combine(err, err2)
	}

	// Update prod deployment on project
	res, err := s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		GithubURL:            proj.GithubURL,
		GithubInstallationID: proj.GithubInstallationID,
		ProdBranch:           proj.ProdBranch,
		ProdVariables:        proj.ProdVariables,
		ProdSlots:            proj.ProdSlots,
		Region:               proj.Region,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		ProdDeploymentID:     &depl.ID,
	})
	if err != nil {
		err2 := s.teardownDeployment(ctx, proj, depl)
		err3 := s.DB.DeleteProject(ctx, proj.ID)
		return nil, multierr.Combine(err, err2, err3)
	}

	// Log project creation
	s.logger.Info("created project", zap.String("id", proj.ID), zap.String("name", proj.Name), zap.String("org", org.Name), zap.String("user_id", userID))

	// Trigger reconcile
	err = s.TriggerReconcile(ctx, depl)
	if err != nil {
		// This error is weird. But it's safe not to teardown the rest.
		return nil, err
	}

	return res, nil
}

// TeardownProject tears down a project and all its deployments.
func (s *Service) TeardownProject(ctx context.Context, p *database.Project) error {
	ds, err := s.DB.FindDeploymentsForProject(ctx, p.ID)
	if err != nil {
		return err
	}

	for _, d := range ds {
		err := s.teardownDeployment(ctx, p, d)
		if err != nil {
			return err
		}
	}

	err = s.DB.DeleteProject(ctx, p.ID)
	if err != nil {
		return err
	}

	return nil
}

// UpdateProject updates a project and any impacted deployments.
// It runs a reconcile if deployment parameters (like branch or variables) have been changed and reconcileDeployment is set.
func (s *Service) UpdateProject(ctx context.Context, proj *database.Project, opts *database.UpdateProjectOptions) (*database.Project, error) {
	if proj.Region != opts.Region || proj.ProdSlots != opts.ProdSlots { // require new deployments
		s.logger.Info("recreating deployment", observability.ZapCtx(ctx))
		var oldDepl *database.Deployment
		var err error
		if proj.ProdDeploymentID != nil {
			oldDepl, err = s.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
			if err != nil && !errors.Is(err, database.ErrNotFound) {
				return nil, err
			}
		}

		org, err := s.DB.FindOrganization(ctx, proj.OrganizationID)
		if err != nil {
			return nil, err
		}

		depl, err := s.createDeployment(ctx, &createDeploymentOptions{
			ProjectID:            proj.ID,
			Subpath:              proj.Subpath,
			ProdOLAPDriver:       proj.ProdOLAPDriver,
			ProdOLAPDSN:          proj.ProdOLAPDSN,
			Region:               opts.Region,
			GithubURL:            opts.GithubURL,
			GithubInstallationID: opts.GithubInstallationID,
			ProdBranch:           opts.ProdBranch,
			ProdVariables:        opts.ProdVariables,
			ProdSlots:            opts.ProdSlots,
			Annotations:          newDeploymentAnnotations(org, proj),
		})
		if err != nil {
			return nil, err
		}

		opts.ProdDeploymentID = &depl.ID
		res, err := s.DB.UpdateProject(ctx, proj.ID, opts)
		if err != nil {
			return nil, multierr.Combine(err, s.teardownDeployment(ctx, proj, depl))
		}

		if oldDepl != nil {
			if err := s.teardownDeployment(context.Background(), proj, oldDepl); err != nil {
				s.logger.Error("could not delete old deployment", zap.Error(err), observability.ZapCtx(ctx))
			}
		}

		if err := s.TriggerReconcile(ctx, depl); err != nil {
			return nil, fmt.Errorf("reconcile failed with error %w", err)
		}

		return res, nil
	}

	projNameChanged := proj.Name != opts.Name
	var deplAnnotations *deploymentAnnotations

	if projNameChanged {
		// Regenerate deployment annotations
		org, err := s.DB.FindOrganization(ctx, proj.OrganizationID)
		if err != nil {
			return nil, err
		}

		deplAnnotations = &deploymentAnnotations{
			orgID:    org.ID,
			orgName:  org.Name,
			projID:   proj.ID,
			projName: opts.Name, // Project name changed
		}
	}

	impactsDeployments := (proj.ProdBranch != opts.ProdBranch ||
		!reflect.DeepEqual(proj.GithubURL, opts.GithubURL) ||
		!reflect.DeepEqual(proj.GithubInstallationID, opts.GithubInstallationID))

	if impactsDeployments {
		s.logger.Info("updating deployments", observability.ZapCtx(ctx))
		ds, err := s.DB.FindDeploymentsForProject(ctx, proj.ID)
		if err != nil {
			return nil, err
		}

		// NOTE: This assumes every deployment (almost always, there's just one) deploys the prod branch.
		// It needs to be refactored when implementing preview deploys.
		for _, d := range ds {
			err := s.updateDeployment(ctx, d, &updateDeploymentOptions{
				GithubURL:            opts.GithubURL,
				GithubInstallationID: opts.GithubInstallationID,
				Subpath:              proj.Subpath,
				Branch:               opts.ProdBranch,
				Variables:            opts.ProdVariables,
				Annotations:          deplAnnotations,
			})
			if err != nil {
				// TODO: This may leave things in an inconsistent state. (Although presently, there's almost never multiple deployments.)
				return nil, err
			}
		}
	} else if projNameChanged {
		err := s.UpdateDeploymentAnnotations(ctx, proj.ID, *deplAnnotations)
		if err != nil {
			return nil, err
		}
	}

	proj, err := s.DB.UpdateProject(ctx, proj.ID, opts)
	if err != nil {
		return nil, err
	}

	return proj, nil
}

// UpdateProjectVariables updates project's variables and pushes the updated variables to deployments.
// NOTE : this does not trigger reconcile.
func (s *Service) UpdateProjectVariables(ctx context.Context, proj *database.Project, variables map[string]string) (*database.Project, error) {
	ds, err := s.DB.FindDeploymentsForProject(ctx, proj.ID)
	if err != nil {
		return nil, err
	}

	// NOTE: This assumes every deployment (almost always, there's just one) deploys the prod branch.
	// It needs to be refactored when implementing preview deploys.
	for _, d := range ds {
		err := s.updateDeplVariables(ctx, d, variables)
		if err != nil {
			// TODO: This may leave things in an inconsistent state. (Although presently, there's almost never multiple deployments.)
			return nil, err
		}
	}

	return s.DB.UpdateProjectVariables(ctx, proj.ID, variables)
}

// UpdateOrgDeploymentAnnotations iterates over projects of the given org and
// updates annotations of corresponding deployments with the new organization name
// NOTE : this does not trigger reconcile.
func (s *Service) UpdateOrgDeploymentAnnotations(ctx context.Context, orgID, orgNewName string) error {
	limit := 10
	afterProjectName := ""
	for {
		projects, err := s.DB.FindProjectsForOrganization(ctx, orgID, afterProjectName, limit)
		if err != nil {
			return err
		}

		for _, project := range projects {
			err := s.UpdateDeploymentAnnotations(ctx, project.ID, deploymentAnnotations{
				orgID:    orgID,
				orgName:  orgNewName,
				projID:   project.ID,
				projName: project.Name,
			})
			if err != nil {
				return err
			}

			afterProjectName = project.Name
		}

		if len(projects) < limit {
			break
		}
	}

	return nil
}

// UpdateDeploymentAnnotations pushes the updated annotations to deployments.
// NOTE : this does not trigger reconcile.
func (s *Service) UpdateDeploymentAnnotations(ctx context.Context, projID string, annotations deploymentAnnotations) error {
	ds, err := s.DB.FindDeploymentsForProject(ctx, projID)
	if err != nil {
		return err
	}

	for _, d := range ds {
		err := s.updateDeplAnnotations(ctx, d, annotations)
		if err != nil {
			return err
		}
	}

	return nil
}

// TriggerRedeploy de-provisions and re-provisions a project's prod deployment.
func (s *Service) TriggerRedeploy(ctx context.Context, proj *database.Project, prevDepl *database.Deployment) error {
	org, err := s.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return err
	}

	// Provision new deployment
	newDepl, err := s.createDeployment(ctx, &createDeploymentOptions{
		ProjectID:            proj.ID,
		Region:               proj.Region,
		GithubURL:            proj.GithubURL,
		GithubInstallationID: proj.GithubInstallationID,
		Subpath:              proj.Subpath,
		ProdBranch:           proj.ProdBranch,
		ProdVariables:        proj.ProdVariables,
		ProdOLAPDriver:       proj.ProdOLAPDriver,
		ProdOLAPDSN:          proj.ProdOLAPDSN,
		ProdSlots:            proj.ProdSlots,
		Annotations:          newDeploymentAnnotations(org, proj),
	})
	if err != nil {
		return err
	}

	// Update prod deployment on project
	_, err = s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		GithubURL:            proj.GithubURL,
		GithubInstallationID: proj.GithubInstallationID,
		ProdBranch:           proj.ProdBranch,
		ProdVariables:        proj.ProdVariables,
		ProdDeploymentID:     &newDepl.ID,
		ProdSlots:            proj.ProdSlots,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		Region:               proj.Region,
	})
	if err != nil {
		err2 := s.teardownDeployment(ctx, proj, newDepl)
		return multierr.Combine(err, err2)
	}

	// Delete old prod deployment if exists
	if prevDepl != nil {
		err = s.teardownDeployment(ctx, proj, prevDepl)
		if err != nil {
			s.logger.Error("trigger redeploy: could not teardown old deployment", zap.String("deployment_id", prevDepl.ID), zap.Error(err), observability.ZapCtx(ctx))
		}
	}

	// Trigger reconcile on new deployment
	err = s.TriggerReconcile(ctx, newDepl)
	if err != nil {
		// This error is weird. But it's safe not to teardown the rest.
		return err
	}

	return nil
}

// TriggerReconcile triggers a reconcile for a deployment.
func (s *Service) TriggerReconcile(ctx context.Context, depl *database.Deployment) error {
	// Run reconcile in the background (since it's sync)
	s.reconcileWg.Add(1)
	go func() {
		defer s.reconcileWg.Done()
		s.logger.Info("reconcile: starting", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
		err := s.triggerReconcile(s.closeCtx, depl) // Use s.closeCtx to cancel if the service is stopped
		if err == nil {
			s.logger.Info("reconcile: completed", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
		} else {
			s.logger.Error("reconcile: failed", zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
		}
	}()
	return nil
}

func (s *Service) triggerReconcile(ctx context.Context, depl *database.Deployment) error {
	err := s.startReconcile(ctx, depl)
	if err != nil {
		return err
	}

	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return s.endReconcile(ctx, depl, nil, err)
	}
	defer rt.Close()

	res, err := rt.Reconcile(ctx, &runtimev1.ReconcileRequest{InstanceId: depl.RuntimeInstanceID})
	return s.endReconcile(ctx, depl, res, err)
}

// TriggerRefreshSource triggers refresh of a deployment's sources. If the sources slice is nil, it will refresh all sources.f
func (s *Service) TriggerRefreshSources(ctx context.Context, depl *database.Deployment, sources []string) error {
	// check if provided sources are exists in catalog
	if len(sources) > 0 {
		rt, err := s.openRuntimeClientForDeployment(depl)
		if err != nil {
			return err
		}
		defer rt.Close()

		// Get paths of sources
		res, err := rt.ListCatalogEntries(ctx, &runtimev1.ListCatalogEntriesRequest{InstanceId: depl.RuntimeInstanceID, Type: runtimev1.ObjectType_OBJECT_TYPE_SOURCE})
		if err != nil {
			return err
		}

		for _, source := range sources {
			found := false
			for _, entry := range res.Entries {
				if strings.EqualFold(source, entry.Name) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("source %q not found", source)
			}
		}
	}

	// Run reconcile in the background (since it's sync)
	s.reconcileWg.Add(1)
	go func() {
		defer s.reconcileWg.Done()
		s.logger.Info("refresh sources: starting", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
		err := s.triggerRefreshSources(s.closeCtx, depl, sources) // Use s.closeCtx to cancel if the service is stopped
		if err == nil {
			s.logger.Info("refresh sources: completed", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
		} else {
			s.logger.Error("refresh sources: failed", zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
		}
	}()
	return nil
}

func (s *Service) triggerRefreshSources(ctx context.Context, depl *database.Deployment, sources []string) error {
	err := s.startReconcile(ctx, depl)
	if err != nil {
		return err
	}

	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return s.endReconcile(ctx, depl, nil, err)
	}
	defer rt.Close()

	// Get paths of sources
	res1, err := rt.ListCatalogEntries(ctx, &runtimev1.ListCatalogEntriesRequest{InstanceId: depl.RuntimeInstanceID, Type: runtimev1.ObjectType_OBJECT_TYPE_SOURCE})
	if err != nil {
		return s.endReconcile(ctx, depl, nil, err)
	}
	var paths []string
	for _, entry := range res1.Entries {
		// If sources is nil, refresh all sources
		if len(sources) == 0 {
			paths = append(paths, entry.Path)
			continue
		}
		// Otherwise, only refresh the selected sources
		for _, name := range sources {
			if entry.Name == name {
				paths = append(paths, entry.Path)
			}
		}
	}

	// If paths is empty, there are no sources to refresh
	if len(paths) == 0 {
		return s.endReconcile(ctx, depl, nil, nil)
	}

	res2, err := rt.Reconcile(ctx, &runtimev1.ReconcileRequest{
		InstanceId:   depl.RuntimeInstanceID,
		ChangedPaths: nil, // Forces it to parse all artifacts, ensuring a complete DAG (TODO: pretty hacky)
		ForcedPaths:  paths,
		Dry:          false,
		Strict:       false,
	})
	return s.endReconcile(ctx, depl, res2, err)
}

func (s *Service) startReconcile(ctx context.Context, depl *database.Deployment) error {
	if depl.Status == database.DeploymentStatusReconciling && time.Since(depl.UpdatedOn) < 30*time.Minute {
		return fmt.Errorf("skipping because it is already running")
	}

	updatedDepl, err := s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusReconciling, "")
	if err != nil {
		return fmt.Errorf("could not update status: %w", err)
	}
	depl.Status = updatedDepl.Status
	depl.Logs = updatedDepl.Logs

	return nil
}

func (s *Service) endReconcile(ctx context.Context, depl *database.Deployment, res *runtimev1.ReconcileResponse, err error) error {
	if err != nil {
		updatedDepl, err2 := s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusError, err.Error())
		if err2 != nil {
			return multierr.Combine(err, fmt.Errorf("could not update logs: %w", err2))
		}
		depl.Status = updatedDepl.Status
		depl.Logs = updatedDepl.Logs
		return err
	}

	var updatedDepl *database.Deployment
	if res != nil && len(res.Errors) > 0 {
		json, err := protojson.Marshal(res)
		if err != nil {
			return fmt.Errorf("could not marshal logs: %w", err)
		}

		updatedDepl, err = s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusError, string(json))
		if err != nil {
			return fmt.Errorf("could not update logs: %w", err)
		}
	} else {
		updatedDepl, err = s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusOK, "")
		if err != nil {
			return fmt.Errorf("could not clear logs: %w", err)
		}
	}

	depl.Status = updatedDepl.Status
	depl.Logs = updatedDepl.Logs
	return nil
}
