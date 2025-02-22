<script lang="ts">
  import VirtualizedGrid from "@rilldata/web-common/components/VirtualizedGrid.svelte";
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { createShowHideDimensionsStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import {
    createQueryServiceMetricsViewTotals,
    MetricsViewDimension,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy, onMount } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import {
    metricsExplorerStore,
    useDashboardStore,
    useFetchTimeRange,
  } from "../dashboard-stores";
  import { FormatPreset } from "../humanize-numbers";
  import Leaderboard from "./Leaderboard.svelte";
  import LeaderboardControls from "./LeaderboardControls.svelte";

  export let metricViewName: string;

  const queryClient = useQueryClient();

  $: dashboardStore = useDashboardStore(metricViewName);
  $: fetchTimeStore = useFetchTimeRange(metricViewName);

  // query the `/meta` endpoint to get the metric's measures and dimensions
  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);
  let dimensions: Array<MetricsViewDimension>;
  $: dimensions = $metaQuery.data?.dimensions;
  $: measures = $metaQuery.data?.measures;

  $: selectedMeasureNames = $dashboardStore?.selectedMeasureNames;

  $: activeMeasure =
    measures &&
    measures.find(
      (measure) => measure.name === $dashboardStore?.leaderboardMeasureName
    );

  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  $: timeStart = $fetchTimeStore?.start?.toISOString();
  $: timeEnd = $fetchTimeStore?.end?.toISOString();
  $: totalsQuery = createQueryServiceMetricsViewTotals(
    $runtime.instanceId,
    metricViewName,
    {
      measureNames: selectedMeasureNames,
      timeStart: hasTimeSeries ? timeStart : undefined,
      timeEnd: hasTimeSeries ? timeEnd : undefined,
      filter: $dashboardStore?.filters,
    },
    {
      query: {
        enabled:
          selectedMeasureNames?.length > 0 &&
          (hasTimeSeries ? !!timeStart && !!timeEnd : true) &&
          !!$dashboardStore?.filters,
      },
    }
  );

  $: formatPreset =
    (activeMeasure?.format as FormatPreset) ?? FormatPreset.HUMANIZE;

  let referenceValue: number;
  $: if (activeMeasure?.name && $totalsQuery?.data?.data) {
    referenceValue = $totalsQuery.data.data?.[activeMeasure.name];
  }

  $: unfilteredTotalsQuery = createQueryServiceMetricsViewTotals(
    $runtime.instanceId,
    metricViewName,
    {
      measureNames: selectedMeasureNames,
      timeStart: hasTimeSeries ? timeStart : undefined,
      timeEnd: hasTimeSeries ? timeEnd : undefined,
    },
    {
      query: {
        enabled: hasTimeSeries ? !!timeStart && !!timeEnd : true,
      },
    }
  );

  let unfilteredTotal: number;
  $: if (activeMeasure?.name) {
    unfilteredTotal = $unfilteredTotalsQuery.data?.data?.[activeMeasure.name];
  }

  let leaderboardExpanded;

  function onSelectItem(event, item: MetricsViewDimension) {
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilter(
      metricViewName,
      item.name,
      event.detail.label
    );
  }

  /** Functionality for resizing the virtual leaderboard */
  let columns = 3;
  let availableWidth = 0;
  let leaderboardContainer: HTMLElement;
  let observer: ResizeObserver;

  function onResize() {
    if (!leaderboardContainer) return;
    availableWidth = leaderboardContainer.offsetWidth;
    columns = Math.max(1, Math.floor(availableWidth / (315 + 20)));
  }

  onMount(() => {
    onResize();
    const observer = new ResizeObserver(() => {
      onResize();
    });
    observer.observe(leaderboardContainer);
  });

  onDestroy(() => {
    observer?.disconnect();
  });

  $: showHideDimensions = createShowHideDimensionsStore(
    metricViewName,
    metaQuery
  );

  $: dimensionsShown =
    dimensions?.filter((_, i) => $showHideDimensions.selectedItems[i]) ?? [];
</script>

<svelte:window on:resize={onResize} />
<!-- container for the metrics leaderboard components and controls -->
<div
  bind:this={leaderboardContainer}
  class="flex flex-col overflow-hidden"
  style:height="calc(100vh - 130px - 4rem)"
  style:min-width="365px"
>
  <div
    class="grid grid-auto-cols justify-between grid-flow-col items-center pl-1 pb-3 flex-grow-0"
  >
    <LeaderboardControls {metricViewName} />
  </div>
  <div class="grow overflow-hidden">
    {#if $dashboardStore}
      <VirtualizedGrid {columns} height="100%" items={dimensionsShown} let:item>
        <!-- the single virtual element -->
        <Leaderboard
          {formatPreset}
          isSummableMeasure={activeMeasure?.expression
            .toLowerCase()
            ?.includes("count(") ||
            activeMeasure?.expression?.toLowerCase()?.includes("sum(")}
          {metricViewName}
          dimensionName={item.name}
          on:expand={() => {
            if (leaderboardExpanded === item.name) {
              leaderboardExpanded = undefined;
            } else {
              leaderboardExpanded = item.name;
            }
          }}
          on:select-item={(event) => onSelectItem(event, item)}
          referenceValue={referenceValue || 0}
          {unfilteredTotal}
        />
      </VirtualizedGrid>
    {/if}
  </div>
</div>
