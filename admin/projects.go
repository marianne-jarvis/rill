package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/client"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *Service) CreateProject(ctx context.Context, opts *database.InsertProjectOptions) (*database.Project, error) {
	// TODO: Make this actually fault tolerant.

	// Create the project
	proj, err := s.DB.InsertProject(ctx, opts)
	if err != nil {
		return nil, err
	}

	if proj.GithubURL == nil || proj.GithubInstallationID == nil {
		return nil, fmt.Errorf("cannot create project without github info")
	}

	// Provision it
	provOpts := &provisioner.ProvisionOptions{
		Region:               proj.Region,
		Slots:                proj.ProductionSlots,
		GithubURL:            *proj.GithubURL,
		GitBranch:            proj.ProductionBranch,
		GithubInstallationID: *proj.GithubInstallationID,
		Variables:            proj.ProductionVariables,
	}
	inst, err := s.provisioner.Provision(ctx, provOpts)
	if err != nil {
		err = fmt.Errorf("provisioner: %w", err)
		err2 := s.DB.DeleteProject(ctx, proj.ID)
		return nil, multierr.Combine(err, err2)
	}

	// Store deployment
	depl, err := s.DB.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         proj.ID,
		Branch:            proj.ProductionBranch,
		Slots:             proj.ProductionSlots,
		RuntimeHost:       inst.Host,
		RuntimeInstanceID: inst.InstanceID,
		RuntimeAudience:   inst.Audience,
		Status:            database.DeploymentStatusPending,
		Logs:              "",
	})
	if err != nil {
		err2 := s.provisioner.Teardown(ctx, inst.Host, inst.InstanceID)
		err3 := s.DB.DeleteProject(ctx, proj.ID)
		return nil, multierr.Combine(err, err2, err3)
	}

	// Update prod deployment on project
	res, err := s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Description:            proj.Description,
		Public:                 proj.Public,
		ProductionBranch:       proj.ProductionBranch,
		ProductionVariables:    proj.ProductionVariables,
		GithubURL:              proj.GithubURL,
		GithubInstallationID:   proj.GithubInstallationID,
		ProductionDeploymentID: &depl.ID,
	})
	if err != nil {
		err2 := s.DB.DeleteDeployment(ctx, depl.ID)
		err3 := s.provisioner.Teardown(ctx, inst.Host, inst.InstanceID)
		err4 := s.DB.DeleteProject(ctx, proj.ID)
		return nil, multierr.Combine(err, err2, err3, err4)
	}

	// Trigger reconcile
	err = s.TriggerReconcile(ctx, depl.ID)
	if err != nil {
		// This error is weird. But it's safe not to teardown the rest.
		return nil, err
	}

	return res, nil
}

func (s *Service) TeardownProject(ctx context.Context, p *database.Project) error {
	// TODO: Make this actually fault tolerant.

	ds, err := s.DB.FindDeployments(ctx, p.ID)
	if err != nil {
		return err
	}

	for _, d := range ds {
		err := s.provisioner.Teardown(ctx, d.RuntimeHost, d.RuntimeInstanceID)
		if err != nil {
			return err
		}

		err = s.DB.DeleteDeployment(ctx, d.ID)
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

func (s *Service) TriggerReconcile(ctx context.Context, deploymentID string) error {
	// TODO: Make this actually fault tolerant

	// Run it all in the background
	go func() {
		// Use s.closeCtx to cancel if the service is stopped
		ctx := s.closeCtx

		s.logger.Info("reconcile: starting", zap.String("deployment_id", deploymentID))

		// Get deployment
		depl, err := s.DB.FindDeployment(ctx, deploymentID)
		if err != nil {
			s.logger.Error("reconcile: could not find deployment", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}

		// Check status
		if depl.Status == database.DeploymentStatusReconciling && time.Since(depl.UpdatedOn) < 30*time.Minute {
			s.logger.Error("reconcile: skipping because it is already running", zap.String("deployment_id", deploymentID))
			return
		}

		// Set deployment status to reconciling
		depl, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusReconciling, "")
		if err != nil {
			s.logger.Error("reconcile: could not update status", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}

		// Get superuser token for runtime host
		jwt, err := s.issuer.NewToken(auth.TokenOptions{
			AudienceURL:         depl.RuntimeAudience,
			TTL:                 time.Hour,
			InstancePermissions: map[string][]auth.Permission{depl.RuntimeInstanceID: {auth.EditInstance}},
		})
		if err != nil {
			s.logger.Error("reconcile: could not get token", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}

		// Make runtime client
		rt, err := client.New(depl.RuntimeHost, jwt)
		if err != nil {
			s.logger.Error("reconcile: could not create client", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}

		// Call reconcile
		res, err := rt.Reconcile(ctx, &runtimev1.ReconcileRequest{InstanceId: depl.RuntimeInstanceID})
		if err != nil {
			s.logger.Error("reconcile: rpc error", zap.String("deployment_id", deploymentID), zap.Error(err))
			_, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusError, err.Error())
			if err != nil {
				s.logger.Error("reconcile: could not update logs", zap.String("deployment_id", deploymentID), zap.Error(err))
			}
			return
		}

		// Set status
		if len(res.Errors) > 0 {
			json, err := protojson.Marshal(res)
			if err != nil {
				s.logger.Error("reconcile: json error", zap.String("deployment_id", deploymentID), zap.Error(err))
				return
			}

			_, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusError, string(json))
			if err != nil {
				s.logger.Error("reconcile: could not update logs", zap.String("deployment_id", deploymentID), zap.Error(err))
				return
			}
		} else {
			_, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusOK, "")
			if err != nil {
				s.logger.Error("reconcile: could not clear logs", zap.String("deployment_id", deploymentID), zap.Error(err))
				return
			}
		}

		s.logger.Info("reconcile: completed", zap.String("deployment_id", deploymentID))
	}()
	return nil
}

func (s *Service) UpdateProject(ctx context.Context, projID string, opts *database.UpdateProjectOptions) (*database.Project, error) {
	// TODO: Make this actually fault tolerant.

	// TODO: Handle if ProductionBranch or GithubURL was changed.

	ds, err := s.DB.FindDeployments(ctx, projID)
	if err != nil {
		return nil, err
	}

	for _, d := range ds {
		if err := s.editInstance(ctx, d, opts.ProductionVariables); err != nil {
			return nil, err
		}
	}

	// Update the project
	proj, err := s.DB.UpdateProject(ctx, projID, opts)
	if err != nil {
		return nil, err
	}
	return proj, nil
}

func (s *Service) editInstance(ctx context.Context, d *database.Deployment, variables map[string]string) error {
	jwt, err := s.issuer.NewToken(auth.TokenOptions{
		AudienceURL:       d.RuntimeAudience,
		TTL:               time.Hour,
		SystemPermissions: []auth.Permission{auth.ManageInstances, auth.ReadInstance},
	})
	if err != nil {
		return err
	}

	rt, err := client.New(d.RuntimeHost, jwt)
	if err != nil {
		return err
	}
	defer rt.Close()

	resp, err := rt.GetInstance(ctx, &runtimev1.GetInstanceRequest{
		InstanceId: d.RuntimeInstanceID,
	})
	if err != nil {
		return err
	}

	// Edit the instance
	inst := resp.Instance
	_, err = rt.EditInstance(ctx, &runtimev1.EditInstanceRequest{
		InstanceId:          d.RuntimeInstanceID,
		OlapDriver:          inst.OlapDriver,
		OlapDsn:             inst.OlapDsn,
		RepoDriver:          inst.RepoDriver,
		RepoDsn:             inst.RepoDsn,
		EmbedCatalog:        inst.EmbedCatalog,
		Variables:           variables,
		IngestionLimitBytes: inst.IngestionLimitBytes,
	})
	return err
}
