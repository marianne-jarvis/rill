package server

import (
	"context"
	"net/http"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) TriggerReconcile(ctx context.Context, req *adminv1.TriggerReconcileRequest) (*adminv1.TriggerReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.deployment_id", req.DeploymentId),
	)

	depl, err := s.admin.DB.FindDeployment(ctx, req.DeploymentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, depl.ProjectID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	err = s.admin.TriggerReconcile(ctx, depl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerReconcileResponse{}, nil
}

func (s *Server) TriggerRefreshSources(ctx context.Context, req *adminv1.TriggerRefreshSourcesRequest) (*adminv1.TriggerRefreshSourcesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.deployment_id", req.DeploymentId),
		attribute.StringSlice("args.sources", req.Sources),
	)

	depl, err := s.admin.DB.FindDeployment(ctx, req.DeploymentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, depl.ProjectID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	err = s.admin.TriggerRefreshSources(ctx, depl, req.Sources)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerRefreshSourcesResponse{}, nil
}

func (s *Server) triggerRefreshSourcesInternal(w http.ResponseWriter, r *http.Request) {
	orgName := r.URL.Query().Get("organization")
	projectName := r.URL.Query().Get("project")
	if orgName == "" || projectName == "" {
		http.Error(w, "organization or project not specified", http.StatusBadRequest)
		return
	}

	proj, err := s.admin.DB.FindProjectByName(r.Context(), orgName, projectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if proj.ProdDeploymentID == nil {
		http.Error(w, "project does not have a deployment", http.StatusBadRequest)
		return
	}

	depl, err := s.admin.DB.FindDeployment(r.Context(), *proj.ProdDeploymentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.admin.TriggerRefreshSources(r.Context(), depl, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) TriggerRedeploy(ctx context.Context, req *adminv1.TriggerRedeployRequest) (*adminv1.TriggerRedeployResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.deployment_id", req.DeploymentId),
	)

	// For backwards compatibility, this RPC supports passing either DeploymentId or Organization+Project names
	var proj *database.Project
	var depl *database.Deployment
	if req.DeploymentId != "" {
		var err error
		depl, err = s.admin.DB.FindDeployment(ctx, req.DeploymentId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		proj, err = s.admin.DB.FindProject(ctx, depl.ProjectID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	} else {
		var err error
		proj, err = s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if proj.ProdDeploymentID != nil {
			depl, err = s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	err := s.admin.TriggerRedeploy(ctx, proj, depl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerRedeployResponse{}, nil
}
