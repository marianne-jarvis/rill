<script lang="ts">
  import { createQuery, CreateQueryResult } from "@tanstack/svelte-query";
  import { getBlock } from "./util";
  import { fetchData, TCellData } from "./mock-data";
  import FormattedNumberCell from "./FormattedNumberCell.svelte";
  import type { SvelteComponent } from "svelte";
  import DimensionCell from "./DimensionCell.svelte";
  import { useTDTContext } from "./context";

  export let rowIdx: number;
  export let colIdx: number;
  export let fixed = false;
  export let lastFixed = false;

  const { store } = useTDTContext();
  // If the current data block has this cell, get the data. Otherwise for now assume "loading" state (can handle errors later)
  let cellData: TCellData & { isLoading?: boolean } = {
    text: "...",
    isLoading: true,
  };
  let block = getBlock(100, rowIdx, rowIdx);
  $: block = getBlock(100, rowIdx, rowIdx);

  const cellQuery = createQuery({
    queryKey: ["time-dimension-details", block[0], block[1]],
    queryFn: fetchData(block, 1000),
  }) as CreateQueryResult<{
    block: number[];
    data: { text?: string; value?: number; sparkline?: number[] }[][];
  }>;

  $: {
    if (
      $cellQuery.data &&
      rowIdx >= $cellQuery.data.block[0] &&
      rowIdx < $cellQuery.data.block[1]
    ) {
      cellData =
        $cellQuery.data.data[rowIdx - $cellQuery.data.block[0]][colIdx];
    } else cellData = { text: "...", isLoading: true };
  }

  let _class = "";
  $: {
    _class = "h-full flex items-center px-2";
    if (fixed) _class += ` z-2`;
    if (lastFixed) _class += ` right-shadow`;

    // Determine background color based on store
    const isRowHighlighted = $store.highlightedRow === rowIdx;
    const isColHighlighted = $store.highlightedCol === colIdx;
    const isHighlighted = isRowHighlighted || isColHighlighted;
    const isDoubleHighlighted = isRowHighlighted && isColHighlighted;
    const isScrubbed =
      $store.scrubbedCols &&
      colIdx >= $store.scrubbedCols.at(0) &&
      colIdx <= $store.scrubbedCols.at(1);

    let bgColors = {
      fixed: {
        base: "bg-slate-50",
        highlighted: "bg-slate-100",
        doubleHighlighted: "bg-slate-200",
      },
      scrubbed: {
        base: "bg-blue-50",
        highlighted: "bg-blue-100",
        doubleHighlighted: "bg-blue-200",
      },
      default: {
        base: "bg-white",
        highlighted: "bg-gray-100",
        doubleHighlighted: "bg-gray-200",
      },
    };

    // Choose palette based on type of cell state
    let palette = bgColors.default;
    if (colIdx === 0) palette = bgColors.default;
    else if (fixed) palette = bgColors.fixed;
    else if (isScrubbed) palette = bgColors.scrubbed;

    // Choose color within palette based on highlighted state
    let colorName = palette.base;
    if (isDoubleHighlighted) colorName = palette.doubleHighlighted;
    else if (isHighlighted) colorName = palette.highlighted;
    _class += ` ${colorName}`;
  }

  const handleMouseEnter = () => {
    $store.highlightedRow = rowIdx;
    $store.highlightedCol = colIdx;
  };
  const handleMouseLeave = () => {
    $store.highlightedCol = null;
    $store.highlightedRow = null;
  };

  // TODO: with real data, this should be dependent on selected metric
  const format = "0.2f";
  let cellComponent: typeof SvelteComponent;
  let cellComponentDefaultProps = {};
  $: {
    // Dimension cell
    if (colIdx === 0) {
      cellComponent = DimensionCell;
      cellComponentDefaultProps = {};
    }
    // Measure cell and delta cell
    else if ([1, 3].includes(colIdx)) {
      cellComponent = FormattedNumberCell;
      cellComponentDefaultProps = { negClass: "text-red-500", format };
    }
    // Percent of total and delta percent cell
    else if ([2, 4].includes(colIdx)) {
      cellComponent = FormattedNumberCell;
      cellComponentDefaultProps = { negClass: "text-red-500", format: "0.1%" };
    }
    // Pivotted measure cells
    else if (!fixed) {
      cellComponent = FormattedNumberCell;
      cellComponentDefaultProps = { format };
    }
    // No match
    else {
      cellComponent = null;
      cellComponentDefaultProps = {};
    }
  }
</script>

<div
  class={_class}
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
>
  {#if cellComponent && !cellData.isLoading}
    <svelte:component
      this={cellComponent}
      {...cellComponentDefaultProps}
      cell={cellData}
    />
  {:else}
    {cellData?.text ?? cellData?.value}
  {/if}
</div>

<style>
  .right-shadow:after {
    content: "";
    width: 1px;
    height: 100%;
    position: absolute;
    top: 0px;
    right: 0px;
    background: #e5e7eb;
    filter: drop-shadow(3px 0px 3px rgb(0 0 0 / 0.27));
  }
</style>
