package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"text/template"

	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
)

// NOTE: The queries in here are generally not vetted or fully implemented. Use it as guidelines for the real implementation
// once the metrics view artifact representation is ready.

// MetricsViewToplist implements QueryService.
func (s *Server) MetricsViewToplist(ctx context.Context, req *runtimev1.MetricsViewToplistRequest) (*runtimev1.MetricsViewToplistResponse, error) {
	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	cat, err := s.runtime.GetCatalogEntry(ctx, req.InstanceId, req.MetricsViewName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("failed with error %w", err)
	}

	mv := cat.GetMetricsView()
	policy, err := s.resolvedPolicy(mv.Policies, claims.GetEmail(), "")
	if err != nil {
		return nil, err
	}
	q := &queries.MetricsViewToplist{
		MetricsViewName: req.MetricsViewName,
		DimensionName:   req.DimensionName,
		MeasureNames:    req.MeasureNames,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Limit:           req.Limit,
		Offset:          req.Offset,
		Sort:            req.Sort,
		Filter:          req.Filter,
		Policy:          policy,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewTimeSeries implements QueryService.
func (s *Server) MetricsViewTimeSeries(ctx context.Context, req *runtimev1.MetricsViewTimeSeriesRequest) (*runtimev1.MetricsViewTimeSeriesResponse, error) {
	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	cat, err := s.runtime.GetCatalogEntry(ctx, req.InstanceId, req.MetricsViewName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("failed with error %w", err)
	}

	mv := cat.GetMetricsView()
	policy, err := s.resolvedPolicy(mv.Policies, claims.GetEmail(), "")
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName: req.MetricsViewName,
		MeasureNames:    req.MeasureNames,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		TimeGranularity: req.TimeGranularity,
		Filter:          req.Filter,
		Policy:          policy,
	}

	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return q.Result, nil
}

// MetricsViewTotals implements QueryService.
func (s *Server) MetricsViewTotals(ctx context.Context, req *runtimev1.MetricsViewTotalsRequest) (*runtimev1.MetricsViewTotalsResponse, error) {
	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	cat, err := s.runtime.GetCatalogEntry(ctx, req.InstanceId, req.MetricsViewName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("failed with error %w", err)
	}

	mv := cat.GetMetricsView()
	policy, err := s.resolvedPolicy(mv.Policies, claims.GetEmail(), "")
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTotals{
		MetricsViewName: req.MetricsViewName,
		MeasureNames:    req.MeasureNames,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Filter:          req.Filter,
		Policy:          policy,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return q.Result, nil
}

func (s *Server) resolvedPolicy(policy, email, group string) (string, error) {
	// this is required in order to be able to use env.KEY and not .KEY in template placeholders
	env := map[string]map[string]string{"user": {"email": email, "group": group}}

	// convert templatised artifact
	t, err := template.New("source").Option("missingkey=error").Parse(policy)
	if err != nil {
		return "", err
	}

	bw := new(bytes.Buffer)
	if err := t.Execute(bw, env); err != nil {
		return "", err
	}
	return bw.String(), nil
}

// Commenting as its unused

// func (s *Server) lookupMetricsView(ctx context.Context, instanceID, name string) (*runtimev1.MetricsView, error) {
// 	obj, err := s.runtime.GetCatalogEntry(ctx, instanceID, name)
// 	if err != nil {
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}

// 	if obj.GetMetricsView() == nil {
// 		return nil, status.Errorf(codes.NotFound, "object named '%s' is not a metrics view", name)
// 	}

// 	return obj.GetMetricsView(), nil
// }
