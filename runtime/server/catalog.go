package server

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListCatalogEntries implements RuntimeService.
func (s *Server) ListCatalogEntries(ctx context.Context, req *runtimev1.ListCatalogEntriesRequest) (*runtimev1.ListCatalogEntriesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.type", req.Type.String()),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	entries, err := s.runtime.ListCatalogEntries(ctx, req.InstanceId, pbToObjectType(req.Type))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	var pbs []*runtimev1.CatalogEntry
	for _, entry := range entries {
		var pb *runtimev1.CatalogEntry
		if req.Type == runtimev1.ObjectType_OBJECT_TYPE_METRICS_VIEW || entry.IsMetricsView() {
			mv := entry.GetMetricsView()
			policy, err := s.runtime.ResolveMetricsViewPolicy(auth.GetClaims(ctx).Attributes(), req.InstanceId, mv, entry.UpdatedOn)
			if err != nil {
				return nil, err
			}

			if policy != nil && !policy.HasAccess {
				continue
			}
			newMv := filterDimensionsAndMeasures(policy, mv)
			pb, err = mvCatalogObjectToPB(entry, newMv)
			if err != nil {
				return nil, status.Error(codes.Unknown, err.Error())
			}
		} else {
			pb, err = catalogObjectToPB(entry)
			if err != nil {
				return nil, status.Error(codes.Unknown, err.Error())
			}
		}
		pbs = append(pbs, pb)
	}

	return &runtimev1.ListCatalogEntriesResponse{Entries: pbs}, nil
}

// GetCatalogEntry implements RuntimeService.
func (s *Server) GetCatalogEntry(ctx context.Context, req *runtimev1.GetCatalogEntryRequest) (*runtimev1.GetCatalogEntryResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.name", req.Name),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	entry, err := s.runtime.GetCatalogEntry(ctx, req.InstanceId, req.Name)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	var pb *runtimev1.CatalogEntry
	if entry.Type == drivers.ObjectTypeMetricsView || entry.IsMetricsView() {
		mv := entry.GetMetricsView()
		policy, err := s.runtime.ResolveMetricsViewPolicy(auth.GetClaims(ctx).Attributes(), req.InstanceId, mv, entry.UpdatedOn)
		if err != nil {
			return nil, err
		}

		if policy != nil && !policy.HasAccess {
			return nil, ErrForbidden
		}
		newMv := filterDimensionsAndMeasures(policy, mv)
		pb, err = mvCatalogObjectToPB(entry, newMv)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	} else {
		pb, err = catalogObjectToPB(entry)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return &runtimev1.GetCatalogEntryResponse{Entry: pb}, nil
}

// Filters the dimensions and measures of a metrics view based on the policy, if something is filtered out, it creates a new metrics view
// otherwise it returns the original metrics view
func filterDimensionsAndMeasures(policy *runtime.ResolvedMetricsViewPolicy, mv *runtimev1.MetricsView) *runtimev1.MetricsView {
	if policy == nil {
		return mv
	}
	allowedDims := make([]*runtimev1.MetricsView_Dimension, 0)
	allowedMeasures := make([]*runtimev1.MetricsView_Measure, 0)

	if len(policy.Include) > 0 {
		allowed := make(map[string]bool)
		for _, include := range policy.Include {
			allowed[include] = true
		}
		for _, dim := range mv.Dimensions {
			if allowed[dim.Name] {
				allowedDims = append(allowedDims, dim)
			}
		}
		for _, measure := range mv.Measures {
			if allowed[measure.Name] {
				allowedMeasures = append(allowedMeasures, measure)
			}
		}
		newMv := proto.Clone(mv).(*runtimev1.MetricsView)
		newMv.Dimensions = allowedDims
		newMv.Measures = allowedMeasures
		return newMv
	} else if len(policy.Exclude) > 0 {
		restricted := make(map[string]bool)
		for _, exclude := range policy.Exclude {
			restricted[exclude] = true
		}
		for _, dim := range mv.Dimensions {
			if !restricted[dim.Name] {
				allowedDims = append(allowedDims, dim)
			}
		}
		for _, measure := range mv.Measures {
			if !restricted[measure.Name] {
				allowedMeasures = append(allowedMeasures, measure)
			}
		}
		newMv := proto.Clone(mv).(*runtimev1.MetricsView)
		newMv.Dimensions = allowedDims
		newMv.Measures = allowedMeasures
		return newMv
	}
	return mv
}

// Reconcile implements RuntimeService.
func (s *Server) Reconcile(ctx context.Context, req *runtimev1.ReconcileRequest) (*runtimev1.ReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	res, err := s.runtime.Reconcile(ctx, req.InstanceId, req.ChangedPaths, req.ForcedPaths, req.Dry, req.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.ReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

// PutFileAndReconcile implements RuntimeService.
func (s *Server) PutFileAndReconcile(ctx context.Context, req *runtimev1.PutFileAndReconcileRequest) (*runtimev1.PutFileAndReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.EditRepo) || !claims.CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.PutFile(ctx, req.InstanceId, req.Path, strings.NewReader(req.Blob), req.Create, req.CreateOnly)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	changedPaths := []string{req.Path}
	res, err := s.runtime.Reconcile(ctx, req.InstanceId, changedPaths, nil, req.Dry, req.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.PutFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

// RenameFileAndReconcile implements RuntimeService.
func (s *Server) RenameFileAndReconcile(ctx context.Context, req *runtimev1.RenameFileAndReconcileRequest) (*runtimev1.RenameFileAndReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.EditRepo) || !claims.CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.RenameFile(ctx, req.InstanceId, req.FromPath, req.ToPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	changedPaths := []string{req.FromPath, req.ToPath}
	res, err := s.runtime.Reconcile(ctx, req.InstanceId, changedPaths, nil, req.Dry, req.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.RenameFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

// DeleteFileAndReconcile implements RuntimeService.
func (s *Server) DeleteFileAndReconcile(ctx context.Context, req *runtimev1.DeleteFileAndReconcileRequest) (*runtimev1.DeleteFileAndReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.EditRepo) || !claims.CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.DeleteFile(ctx, req.InstanceId, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	changedPaths := []string{req.Path}
	res, err := s.runtime.Reconcile(ctx, req.InstanceId, changedPaths, nil, req.Dry, req.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.DeleteFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

// RefreshAndReconcile implements RuntimeService.
func (s *Server) RefreshAndReconcile(ctx context.Context, req *runtimev1.RefreshAndReconcileRequest) (*runtimev1.RefreshAndReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	changedPaths := []string{req.Path}
	res, err := s.runtime.Reconcile(ctx, req.InstanceId, changedPaths, changedPaths, req.Dry, req.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.RefreshAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

// TriggerRefresh implements RuntimeService.
func (s *Server) TriggerRefresh(ctx context.Context, req *runtimev1.TriggerRefreshRequest) (*runtimev1.TriggerRefreshResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.RefreshSource(ctx, req.InstanceId, req.Name)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &runtimev1.TriggerRefreshResponse{}, nil
}

// TriggerSync implements RuntimeService.
func (s *Server) TriggerSync(ctx context.Context, req *runtimev1.TriggerSyncRequest) (*runtimev1.TriggerSyncResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	err := s.runtime.SyncExistingTables(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	// Done
	// TODO: This should return stats about synced tables. However, it will be refactored into reconcile, so no need to fix this now.
	return &runtimev1.TriggerSyncResponse{}, nil
}

func pbToObjectType(in runtimev1.ObjectType) drivers.ObjectType {
	switch in {
	case runtimev1.ObjectType_OBJECT_TYPE_UNSPECIFIED:
		return drivers.ObjectTypeUnspecified
	case runtimev1.ObjectType_OBJECT_TYPE_TABLE:
		return drivers.ObjectTypeTable
	case runtimev1.ObjectType_OBJECT_TYPE_SOURCE:
		return drivers.ObjectTypeSource
	case runtimev1.ObjectType_OBJECT_TYPE_MODEL:
		return drivers.ObjectTypeModel
	case runtimev1.ObjectType_OBJECT_TYPE_METRICS_VIEW:
		return drivers.ObjectTypeMetricsView
	}
	panic(fmt.Errorf("unhandled object type %s", in))
}

func catalogObjectToPB(obj *drivers.CatalogEntry) (*runtimev1.CatalogEntry, error) {
	catalog := &runtimev1.CatalogEntry{
		Name:        obj.Name,
		Path:        obj.Path,
		Embedded:    obj.Embedded,
		Parents:     obj.Parents,
		Children:    obj.Children,
		CreatedOn:   timestamppb.New(obj.CreatedOn),
		UpdatedOn:   timestamppb.New(obj.UpdatedOn),
		RefreshedOn: timestamppb.New(obj.RefreshedOn),
	}

	switch obj.Type {
	case drivers.ObjectTypeTable:
		catalog.Object = &runtimev1.CatalogEntry_Table{
			Table: obj.GetTable(),
		}
	case drivers.ObjectTypeSource:
		catalog.Object = &runtimev1.CatalogEntry_Source{
			Source: obj.GetSource(),
		}
	case drivers.ObjectTypeModel:
		catalog.Object = &runtimev1.CatalogEntry_Model{
			Model: obj.GetModel(),
		}
	case drivers.ObjectTypeMetricsView:
		catalog.Object = &runtimev1.CatalogEntry_MetricsView{
			MetricsView: obj.GetMetricsView(),
		}
	default:
		panic("not implemented")
	}

	return catalog, nil
}

func mvCatalogObjectToPB(obj *drivers.CatalogEntry, mv *runtimev1.MetricsView) (*runtimev1.CatalogEntry, error) {
	catalog := &runtimev1.CatalogEntry{
		Name:        obj.Name,
		Path:        obj.Path,
		Embedded:    obj.Embedded,
		Parents:     obj.Parents,
		Children:    obj.Children,
		CreatedOn:   timestamppb.New(obj.CreatedOn),
		UpdatedOn:   timestamppb.New(obj.UpdatedOn),
		RefreshedOn: timestamppb.New(obj.RefreshedOn),
	}

	catalog.Object = &runtimev1.CatalogEntry_MetricsView{
		MetricsView: mv,
	}
	return catalog, nil
}
