<script lang="ts">
  /**
   * Leaderboard.svelte
   * -------------------------
   * This is the "implemented" feature of the leaderboard, meant to be used
   * in the application itself.
   */
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import {
    getFilterForDimension,
    useMetaDimension,
    useMetaMeasure,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    createQueryServiceMetricsViewToplist,
    MetricsViewDimension,
    MetricsViewMeasure,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import {
    metricsExplorerStore,
    useComparisonRange,
    useDashboardStore,
    useFetchTimeRange,
  } from "../dashboard-stores";
  import { getFilterForComparsion } from "../dimension-table/dimension-table-utils";
  import type { FormatPreset } from "../humanize-numbers";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import { prepareLeaderboardItemData } from "./leaderboard-utils";
  import LeaderboardListItem from "./LeaderboardListItem.svelte";

  export let metricViewName: string;
  export let dimensionName: string;
  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */
  export let referenceValue: number;
  export let unfilteredTotal: number;

  export let formatPreset: FormatPreset;
  export let isSummableMeasure = false;

  let slice = 7;

  const queryClient = useQueryClient();

  $: dashboardStore = useDashboardStore(metricViewName);
  $: fetchTimeStore = useFetchTimeRange(metricViewName);
  $: comparisonStore = useComparisonRange(metricViewName);

  let filterExcludeMode: boolean;
  $: filterExcludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;
  let filterKey: "exclude" | "include";
  $: filterKey = filterExcludeMode ? "exclude" : "include";

  $: dimensionQuery = useMetaDimension(
    $runtime.instanceId,
    metricViewName,
    dimensionName
  );
  let dimension: MetricsViewDimension;
  $: dimension = $dimensionQuery?.data;
  $: displayName = dimension?.label || dimension?.name;
  $: dimensionColumn = dimension?.column || dimension?.name;

  $: measureQuery = useMetaMeasure(
    $runtime.instanceId,
    metricViewName,
    $dashboardStore?.leaderboardMeasureName
  );
  let measure: MetricsViewMeasure;
  $: measure = $measureQuery?.data;

  $: filterForDimension = getFilterForDimension(
    $dashboardStore?.filters,
    dimensionName
  );

  let activeValues: Array<unknown>;
  $: activeValues =
    $dashboardStore?.filters[filterKey]?.find((d) => d.name === dimension?.name)
      ?.in ?? [];
  $: atLeastOneActive = !!activeValues?.length;

  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  function toggleFilterMode() {
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilterMode(metricViewName, dimensionName);
  }

  function selectDimension(dimensionName) {
    metricsExplorerStore.setMetricDimensionName(metricViewName, dimensionName);
  }

  $: timeStart = $fetchTimeStore?.start?.toISOString();
  $: timeEnd = $fetchTimeStore?.end?.toISOString();
  $: topListQuery = createQueryServiceMetricsViewToplist(
    $runtime.instanceId,
    metricViewName,
    {
      dimensionName: dimensionName,
      measureNames: [measure?.name],
      timeStart: hasTimeSeries ? timeStart : undefined,
      timeEnd: hasTimeSeries ? timeEnd : undefined,
      filter: filterForDimension,
      limit: "250",
      offset: "0",
      sort: [
        {
          name: measure?.name,
          ascending: false,
        },
      ],
    },
    {
      query: {
        enabled:
          (hasTimeSeries ? !!timeStart && !!timeEnd : true) &&
          !!filterForDimension,
      },
    }
  );

  let values: { value: number; label: string | number }[] = [];
  let comparisonValues = [];

  /** replace data after fetched. */
  $: if (!$topListQuery?.isFetching) {
    values =
      $topListQuery?.data?.data.map((val) => ({
        value: val[measure?.name],
        label: val[dimensionColumn],
      })) ?? [];
  }

  // get all values that are selected but not visible.
  // we'll put these at the bottom w/ a divider.
  $: selectedValuesThatAreBelowTheFold = activeValues
    ?.filter((label) => {
      return (
        // the value is visible within the fold.
        !values.slice(0, slice).some((value) => {
          return value.label === label;
        })
      );
    })
    .map((label) => {
      const existingValue = values.find((value) => value.label === label);
      // return the existing value, or if it does not exist, just return the label.
      // FIX ME return values for label which are not in the query
      return existingValue ? { ...existingValue } : { label };
    })
    .sort((a, b) => {
      return b.value - a.value;
    });

  // Compose the comparison /toplist query
  $: showTimeComparison =
    $dashboardStore?.leaderboardContextColumn ===
      LeaderboardContextColumn.DELTA_CHANGE && $dashboardStore?.showComparison;
  $: showPercentOfTotal =
    $dashboardStore?.leaderboardContextColumn ===
    LeaderboardContextColumn.PERCENT;

  $: showContext = $dashboardStore?.leaderboardContextColumn;

  // add all sliced and active values to the include filter.
  $: currentVisibleValues =
    $topListQuery?.data?.data
      ?.slice(0, slice)
      ?.concat(selectedValuesThatAreBelowTheFold)
      ?.map((v) => v[dimensionColumn]) ?? [];
  $: updatedFilters = getFilterForComparsion(
    filterForDimension,
    dimensionName,
    currentVisibleValues
  );
  $: comparisonTimeStart = $comparisonStore?.start;
  $: comparisonTimeEnd = $comparisonStore?.end;
  $: comparisonTopListQuery = createQueryServiceMetricsViewToplist(
    $runtime.instanceId,
    metricViewName,
    {
      dimensionName: dimensionName,
      measureNames: [measure?.name],
      timeStart: comparisonTimeStart,
      timeEnd: comparisonTimeEnd,
      filter: updatedFilters,
      limit: currentVisibleValues.length.toString(),
      offset: "0",
      sort: [
        {
          name: measure?.name,
          ascending: false,
        },
      ],
    },
    {
      query: {
        enabled: Boolean(
          showTimeComparison &&
            !!comparisonTimeStart &&
            !!comparisonTimeEnd &&
            !!updatedFilters
        ),
      },
    }
  );

  $: if (!$comparisonTopListQuery?.isFetching) {
    comparisonValues =
      $comparisonTopListQuery?.data?.data?.map((val) => ({
        value: val[measure?.name],
        label: val[dimensionColumn],
      })) ?? [];
  }

  let hovered: boolean;

  $: comparisonMap = new Map(comparisonValues?.map((v) => [v.label, v.value]));

  $: aboveTheFoldItems = prepareLeaderboardItemData(
    values.slice(0, slice),
    activeValues,
    comparisonMap
  );

  $: belowTheFoldItems = prepareLeaderboardItemData(
    selectedValuesThatAreBelowTheFold,
    activeValues,
    comparisonMap
  );
</script>

{#if topListQuery}
  <div
    style:width="315px"
    on:mouseenter={() => (hovered = true)}
    on:mouseleave={() => (hovered = false)}
  >
    <LeaderboardHeader
      {showTimeComparison}
      {showPercentOfTotal}
      isFetching={$topListQuery.isFetching}
      {displayName}
      on:toggle-filter-mode={toggleFilterMode}
      {filterExcludeMode}
      {hovered}
      dimensionDescription={dimension?.description}
      on:click={() => selectDimension(dimensionName)}
    />
    {#if values}
      <div class="rounded-b border-gray-200 surface text-gray-800">
        <!-- place the leaderboard entries that are above the fold here -->
        {#each aboveTheFoldItems as itemData (itemData.label)}
          <LeaderboardListItem
            {itemData}
            {showContext}
            {atLeastOneActive}
            {filterExcludeMode}
            {unfilteredTotal}
            {isSummableMeasure}
            {referenceValue}
            {formatPreset}
            on:click
            on:keydown
            on:select-item
          />
        {/each}
        <!-- place the selected values that are not above the fold here -->
        {#if selectedValuesThatAreBelowTheFold?.length}
          <hr />
          {#each belowTheFoldItems as itemData (itemData.label)}
            <LeaderboardListItem
              {itemData}
              {showContext}
              {atLeastOneActive}
              {filterExcludeMode}
              {isSummableMeasure}
              {referenceValue}
              {formatPreset}
              on:click
              on:keydown
              on:select-item
            />
          {/each}

          <hr />
        {/if}
        {#if $topListQuery?.isError}
          <div class="text-red-500">
            {$topListQuery?.error}
          </div>
        {:else if values.length === 0}
          <div style:padding-left="30px" class="p-1 ui-copy-disabled">
            no available values
          </div>
        {/if}
        {#if values.length > slice}
          <Tooltip location="right">
            <button
              on:click={() => selectDimension(dimensionName)}
              class="block flex-row w-full text-left transition-color ui-copy-muted"
              style:padding-left="30px"
            >
              (Expand Table)
            </button>
            <TooltipContent slot="tooltip-content"
              >Expand dimension to see more values</TooltipContent
            >
          </Tooltip>
        {/if}
      </div>
    {/if}
  </div>
{/if}
