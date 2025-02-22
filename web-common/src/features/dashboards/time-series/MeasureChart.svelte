<script lang="ts">
  import Body from "@rilldata/web-common/components/data-graphic/elements/Body.svelte";
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import WithBisector from "@rilldata/web-common/components/data-graphic/functional-components/WithBisector.svelte";
  import WithRoundToTimegrain from "@rilldata/web-common/components/data-graphic/functional-components/WithRoundToTimegrain.svelte";
  import {
    Axis,
    Grid,
  } from "@rilldata/web-common/components/data-graphic/guides";
  import {
    ClippedChunkedLine,
    ChunkedLine,
  } from "@rilldata/web-common/components/data-graphic/marks";
  import { NumberKind } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { previousValueStore } from "@rilldata/web-common/lib/store-utils";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { extent } from "d3-array";
  import { cubicOut } from "svelte/easing";
  import { writable } from "svelte/store";
  import { fly } from "svelte/transition";
  import MeasureValueMouseover from "./MeasureValueMouseover.svelte";
  import {
    getBisectedTimeFromCordinates,
    getOrderedStartEnd,
    localToTimeZoneOffset,
    niceMeasureExtents,
  } from "./utils";
  import {
    TimeRangePreset,
    TimeRoundingStrategy,
  } from "../../../lib/time/types";
  import { getContext } from "svelte";
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import type { ScaleStore } from "@rilldata/web-common/components/data-graphic/state/types";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import MeasureScrub from "@rilldata/web-common/features/dashboards/time-series/MeasureScrub.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";

  export let metricViewName: string;
  export let width: number = undefined;
  export let height: number = undefined;
  export let xMin: Date = undefined;
  export let xMax: Date = undefined;
  export let yMin: number = undefined;
  export let yMax: number = undefined;

  export let timeGrain: V1TimeGrain;
  export let zone: string;

  export let showComparison = false;
  export let data;
  export let xAccessor = "ts";
  export let labelAccessor = "label";
  export let yAccessor = "value";
  export let mouseoverValue;
  export let hovered = false;

  // control point for scrub functionality.
  export let isScrubbing = false;
  export let scrubStart;
  export let scrubEnd;

  export let mouseoverFormat: (d: number | Date | string) => string = (v) =>
    v.toString();
  export let mouseoverTimeFormat: (d: number | Date | string) => string = (v) =>
    v.toString();
  export let numberKind: NumberKind = NumberKind.ANY;
  export let tweenProps = { duration: 400, easing: cubicOut };

  const xScale = getContext(contexts.scale("x")) as ScaleStore;

  // scrub local control points
  let justCreatedScrub = false;
  let isMovingScrub = false;
  let moveStartDelta = 0;
  let moveEndDelta = 0;
  let isResizing = false;
  let isOverStart = false;
  let isOverEnd = false;

  $: hasSubrangeSelected = Boolean(scrubStart && scrubEnd);

  $: mainLineColor =
    isScrubbing || hasSubrangeSelected
      ? "hsla(217, 10%, 60%, 1)"
      : "hsla(217,60%, 55%, 1)";

  $: areaColor =
    isScrubbing || hasSubrangeSelected
      ? "hsla(225, 20%, 80%, .2)"
      : "hsla(217,70%, 80%, .4)";

  $: cursorClass = isMovingScrub
    ? "cursor-grabbing"
    : isScrubbing
    ? "cursor-ew-resize"
    : "";
  $: [xExtentMin, xExtentMax] = extent(data, (d) => d[xAccessor]);
  $: [yExtentMin, yExtentMax] = extent(data, (d) => d[yAccessor]);
  let comparisonExtents;

  /** if we are making a comparison, factor this into the extents calculation.*/
  $: if (showComparison) {
    comparisonExtents = extent(data, (d) => d[`comparison.${yAccessor}`]);

    yExtentMin = Math.min(yExtentMin, comparisonExtents[0] || yExtentMin);
    yExtentMax = Math.max(yExtentMax, comparisonExtents[1] || yExtentMax);
  }

  $: [internalYMin, internalYMax] = niceMeasureExtents(
    [
      yMin !== undefined ? yMin : yExtentMin,
      yMax !== undefined ? yMax : yExtentMax,
    ],
    6 / 5
  );

  $: internalXMin = xMin || xExtentMin;
  $: internalXMax = xMax || xExtentMax;
  // we delay the tween if previousYMax < yMax
  let yMaxStore = writable(yExtentMax);
  let previousYMax = previousValueStore(yMaxStore);

  $: yMaxStore.set(yExtentMax);
  const timeRangeKey = writable(`${xMin}-${xMax}`);

  const previousTimeRangeKey = previousValueStore(timeRangeKey);

  // FIXME: move this function to utils.ts
  /** reset the keys to trigger animations on time range changes */
  let syncTimeRangeKey;
  $: {
    timeRangeKey.set(`${xMin}-${xMax}`);
    if ($previousTimeRangeKey !== $timeRangeKey) {
      if (syncTimeRangeKey) clearTimeout(syncTimeRangeKey);
      syncTimeRangeKey = setTimeout(() => {
        previousTimeRangeKey.set($timeRangeKey);
      }, 400);
    }
  }

  $: delay =
    $previousTimeRangeKey === $timeRangeKey && $previousYMax < yExtentMax
      ? 100
      : 0;

  function inBounds(min, max, value) {
    return value >= min && value <= max;
  }

  function resetScrub() {
    metricsExplorerStore.setSelectedScrubRange(metricViewName, undefined);
  }

  function zoomScrub() {
    resetScrub();

    const { start, end } = getOrderedStartEnd(scrubStart, scrubEnd);
    metricsExplorerStore.setSelectedTimeRange(metricViewName, {
      name: TimeRangePreset.CUSTOM,
      start,
      end,
    });
  }

  function updateScrub(start, end, isScrubbing) {
    const adjustedStart = start ? localToTimeZoneOffset(start, zone) : start;
    const adjustedEnd = end ? localToTimeZoneOffset(end, zone) : end;

    metricsExplorerStore.setSelectedScrubRange(metricViewName, {
      name: TimeRangePreset.CUSTOM,
      start: adjustedStart,
      end: adjustedEnd,
      isScrubbing: isScrubbing,
    });
  }
  function startScrub(event) {
    if (hasSubrangeSelected) {
      const startX = event.detail?.start?.x;
      const scrubStartDate = getBisectedTimeFromCordinates(
        startX,
        $xScale,
        labelAccessor,
        data,
        TIME_GRAIN[timeGrain].label
      );

      // check if we are scrubbing on the edges of scrub rect

      /***
       * Currently being calculated by checking if the bisected dates match
       * the existing scrub start/end dates. Although with this approach it would
       * be hard to identify resizing when we have a small grain on a larger
       * time range window (e.g. 1h on 14D) We can increase the flexibility
       * by adding additional few pixels on both sides for threshold
       */
      isOverStart = scrubStartDate?.getTime() == scrubStart?.getTime();
      isOverEnd = scrubStartDate?.getTime() == scrubEnd?.getTime();

      const { start, end } = getOrderedStartEnd(scrubStart, scrubEnd);

      /**
       * We define a moving action when the start date is in middle of
       * the scrub start and end dates. There is an edge case where we
       * have a scrub of unit grain length, in that case we don't have
       * any dates in the middle of the scrub start and end dates.
       */
      isMovingScrub = scrubStartDate > start && scrubStartDate < end;

      if (isOverStart || isOverEnd) {
        isResizing = true;
        updateScrub(scrubStart, scrubEnd, true);
        return;
      } else if (isMovingScrub) {
        moveStartDelta = startX - $xScale(scrubStart);
        moveEndDelta = startX - $xScale(scrubEnd);

        return;
      }
    }
    // updateScrub(scrubStartDate, undefined, true);
  }

  function moveScrub(event) {
    const startX = event.detail?.start?.x;
    const scrubStartDate = getBisectedTimeFromCordinates(
      startX,
      $xScale,
      labelAccessor,
      data,
      TIME_GRAIN[timeGrain].label
    );

    let stopX = event.detail?.stop?.x;
    let intermediateScrubVal = getBisectedTimeFromCordinates(
      stopX,
      $xScale,
      labelAccessor,
      data,
      TIME_GRAIN[timeGrain].label
    );

    if (hasSubrangeSelected && (isResizing || isMovingScrub)) {
      if (
        isResizing &&
        intermediateScrubVal?.getTime() !== scrubEnd?.getTime() &&
        intermediateScrubVal?.getTime() !== scrubStart?.getTime()
      ) {
        /**
         * Adjust the ends of the subrange by dragging either end.
         * This snaps to the nearest time grain.
         */
        const newStart = isOverStart ? intermediateScrubVal : scrubStart;
        const newEnd = isOverEnd ? intermediateScrubVal : scrubEnd;

        updateScrub(newStart, newEnd, true);
      } else if (!isResizing && isMovingScrub) {
        /**
         * Pick up and shift the entire subrange left/right
         * This snaps to the nearest time grain
         */

        const startX = event.detail?.start?.x;
        const delta = stopX - startX;

        const newStart = getBisectedTimeFromCordinates(
          startX - moveStartDelta + delta,
          $xScale,
          labelAccessor,
          data,
          TIME_GRAIN[timeGrain].label
        );

        const newEnd = getBisectedTimeFromCordinates(
          startX - moveEndDelta + delta,
          $xScale,
          labelAccessor,
          data,
          TIME_GRAIN[timeGrain].label
        );

        const insideBounds = $xScale(newStart) >= 0 && $xScale(newEnd) >= 0;
        if (insideBounds && newStart?.getTime() !== scrubStart?.getTime()) {
          updateScrub(newStart, newEnd, true);
        }
      }
    } else {
      // Only make state changes when the bisected value changes
      if (
        scrubStartDate?.getTime() !== scrubStart?.getTime() ||
        intermediateScrubVal?.getTime() !== scrubEnd?.getTime()
      ) {
        updateScrub(scrubStartDate, intermediateScrubVal, true);
      }
    }
  }

  function endScrub() {
    // Remove scrub if start and end are same
    if (hasSubrangeSelected && scrubStart?.getTime() === scrubEnd?.getTime()) {
      resetScrub();
      return;
    }

    isResizing = false;
    isMovingScrub = false;
    justCreatedScrub = true;

    // reset justCreatedScrub after 100 milliseconds
    setTimeout(() => {
      justCreatedScrub = false;
    }, 100);

    updateScrub(scrubStart, scrubEnd, false);
  }

  function onMouseClick() {
    // skip if still scrubbing
    if (justCreatedScrub || isScrubbing || isResizing) return;

    // skip if no scrub range selected
    if (!hasSubrangeSelected) return;

    const { start, end } = getOrderedStartEnd(scrubStart, scrubEnd);

    if (mouseoverValue?.x < start || mouseoverValue?.x > end) {
      resetScrub();
    }
  }
</script>

<SimpleDataGraphic
  overflowHidden={false}
  yMin={internalYMin}
  yMax={internalYMax}
  shareYScale={false}
  yType="number"
  xType="date"
  {width}
  {height}
  top={4}
  left={0}
  right={50}
  bind:mouseoverValue
  bind:hovered
  let:config
  let:yScale
  customClass={cursorClass}
  yMinTweenProps={tweenProps}
  yMaxTweenProps={tweenProps}
  xMaxTweenProps={tweenProps}
  xMinTweenProps={tweenProps}
  on:click={() => onMouseClick()}
  on:scrub-start={(e) => startScrub(e)}
  on:scrub-move={(e) => moveScrub(e)}
  on:scrub-end={() => endScrub()}
>
  <Axis side="right" {numberKind} />
  <Grid />
  <Body>
    <!-- key on the time range itself to prevent weird tweening animations.
    We'll need to migrate this to a more robust solution once we've figured out
    the right way to "tile" together a time series with multiple pages of data.
    -->
    <!-- {#key $timeRangeKey} -->
    {#if showComparison}
      <g
        class="transition-opacity"
        class:opacity-80={mouseoverValue?.x}
        class:opacity-40={!mouseoverValue?.x}
      >
        <ChunkedLine
          area={false}
          lineColor={`hsl(217, 10%, 60%)`}
          delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
          duration={$timeRangeKey !== $previousTimeRangeKey ? 0 : 200}
          {data}
          {xAccessor}
          yAccessor="comparison.{yAccessor}"
        />
      </g>
    {/if}
    <ChunkedLine
      lineColor={mainLineColor}
      {areaColor}
      delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
      duration={$timeRangeKey !== $previousTimeRangeKey ? 0 : 200}
      {data}
      {xAccessor}
      {yAccessor}
    />
    {#if hasSubrangeSelected}
      <ClippedChunkedLine
        start={Math.min(scrubStart, scrubEnd)}
        end={Math.max(scrubStart, scrubEnd)}
        lineColor="hsla(217,60%, 55%, 1)"
        areaColor="hsla(217,70%, 80%, .4)"
        delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
        duration={$timeRangeKey !== $previousTimeRangeKey ? 0 : 200}
        {data}
        {xAccessor}
        {yAccessor}
      />
    {/if}
    <!-- {/key} -->
    <line
      x1={config.plotLeft}
      x2={config.plotLeft + config.plotRight}
      y1={yScale(0)}
      y2={yScale(0)}
      class="stroke-blue-200"
    />
  </Body>
  {#if !isScrubbing && mouseoverValue?.x}
    <WithRoundToTimegrain
      strategy={TimeRoundingStrategy.PREVIOUS}
      value={mouseoverValue.x}
      {timeGrain}
      let:roundedValue
    >
      <WithBisector
        {data}
        callback={(d) => d[xAccessor]}
        value={roundedValue}
        let:point
      >
        {#if point && inBounds(internalXMin, internalXMax, point[xAccessor])}
          <g transition:fly|local={{ duration: 100, x: -4 }}>
            <text
              class="fill-gray-600"
              style:paint-order="stroke"
              stroke="white"
              stroke-width="3px"
              x={config.plotLeft + config.bodyBuffer + 6}
              y={config.plotTop + 10 + config.bodyBuffer}
            >
              {mouseoverTimeFormat(point[labelAccessor])}
            </text>
            {#if showComparison && point[`comparison.${labelAccessor}`]}
              <text
                style:paint-order="stroke"
                stroke="white"
                stroke-width="3px"
                class="fill-gray-400"
                x={config.plotLeft + config.bodyBuffer + 6}
                y={config.plotTop + 24 + config.bodyBuffer}
              >
                {mouseoverTimeFormat(point[`comparison.${labelAccessor}`])} prev.
              </text>
            {/if}
          </g>
          <g transition:fly|local={{ duration: 100, x: -4 }}>
            <MeasureValueMouseover
              {point}
              {xAccessor}
              {yAccessor}
              {showComparison}
              {mouseoverFormat}
              {numberKind}
            />
          </g>
        {/if}
      </WithBisector>
    </WithRoundToTimegrain>
  {/if}
  {#if hasSubrangeSelected}
    <MeasureScrub
      start={scrubStart}
      stop={scrubEnd}
      {isScrubbing}
      {mouseoverTimeFormat}
      on:zoom={() => zoomScrub()}
      on:reset={() => resetScrub()}
    />
  {/if}
</SimpleDataGraphic>
