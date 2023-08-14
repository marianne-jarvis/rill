package server

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestMetricsViewAggregation_Toplist(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewAggregation(testCtx(), &runtimev1.MetricsViewAggregationRequest{
		InstanceId:  instanceId,
		MetricsView: "ad_bids_metrics",
		Dimensions:  []string{"domain"},
		Measures:    []string{"measure_2", "__count"},
		InlineMeasureDefinitions: []*runtimev1.InlineMeasure{
			{
				Name:       "__count",
				Expression: "count(*)",
			},
		},
		Sort: []*runtimev1.MetricsViewSort{
			{Name: "measure_2"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))

	require.Equal(t, 3, len(tr.Data[0].Fields))
	require.Equal(t, 3, len(tr.Data[1].Fields))

	require.Equal(t, "msn.com", tr.Data[0].Fields["domain"].GetStringValue())
	require.Equal(t, 2.0, tr.Data[0].Fields["measure_2"].GetNumberValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["__count"].GetNumberValue())

	require.Equal(t, "yahoo.com", tr.Data[1].Fields["domain"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[1].Fields["measure_2"].GetNumberValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["__count"].GetNumberValue())
}

func TestMetricsViewAggregation_Totals(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewAggregation(testCtx(), &runtimev1.MetricsViewAggregationRequest{
		InstanceId:  instanceId,
		MetricsView: "ad_bids_metrics",
		Measures:    []string{"measure_0"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data))
	require.Equal(t, 2.0, tr.Data[0].Fields["measure_0"].GetNumberValue())
}

func TestMetricsViewAggregation_Distinct(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewAggregation(testCtx(), &runtimev1.MetricsViewAggregationRequest{
		InstanceId:  instanceId,
		MetricsView: "ad_bids_metrics",
		Dimensions:  []string{"domain"},
		Sort: []*runtimev1.MetricsViewSort{
			{Name: "domain", Ascending: false},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))
	require.Equal(t, 1, len(tr.Data[0].Fields))
	require.Equal(t, "yahoo.com", tr.Data[0].Fields["domain"].GetStringValue())
	require.Equal(t, "msn.com", tr.Data[1].Fields["domain"].GetStringValue())
}

func TestMetricsViewAggregation_Timeseries(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewAggregation(testCtx(), &runtimev1.MetricsViewAggregationRequest{
		InstanceId:      instanceId,
		MetricsView:     "ad_bids_metrics",
		Dimensions:      []string{"timestamp"},
		Measures:        []string{"measure_0", "measure_2"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		Sort: []*runtimev1.MetricsViewSort{
			{Name: "timestamp", Ascending: true},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))
	require.Equal(t, 3, len(tr.Data[0].Fields))

	require.Equal(t, "2022-01-01T14:00:00Z", tr.Data[0].Fields["timestamp"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["measure_0"].GetNumberValue())
	require.Equal(t, 2.0, tr.Data[0].Fields["measure_2"].GetNumberValue())

	require.Equal(t, "2022-01-02T11:00:00Z", tr.Data[1].Fields["timestamp"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[1].Fields["measure_0"].GetNumberValue())
	require.Equal(t, 1.0, tr.Data[1].Fields["measure_2"].GetNumberValue())
}
