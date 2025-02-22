<script lang="ts">
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import {
    ColumnSummary,
    getSummaries,
  } from "@rilldata/web-common/components/column-profile/queries";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "@rilldata/web-common/lib/formatters";
  import {
    createQueryServiceTableCardinality,
    createQueryServiceTableColumns,
    createRuntimeServiceGetCatalogEntry,
    V1CatalogEntry,
    V1Source,
  } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { GridCell, LeftRightGrid } from "../../../components/grid";
  import { LIST_SLIDE_DURATION } from "../../../layout/config";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useIsSourceUnsaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  $: runtimeInstanceId = $runtime.instanceId;

  $: getSource = createRuntimeServiceGetCatalogEntry(
    runtimeInstanceId,
    sourceName
  );
  let sourceCatalog: V1CatalogEntry;
  $: sourceCatalog = $getSource?.data?.entry;

  let showColumns = true;

  // get source table references.

  // toggle state for inspector sections

  /** source summary information */
  let rowCount;
  let columnCount;
  let nullPercentage;

  function formatConnectorType(connectorType: string) {
    switch (connectorType) {
      case "s3":
        return "S3";
      case "gcs":
        return "GCS";
      case "https":
        return "http(s)";
      case "local_file":
        return "Local file";
      default:
        return "";
    }
  }

  function getFileExtension(source: V1Source): string {
    const path = source?.properties?.path?.toLowerCase();
    if (path?.includes(".csv")) return "CSV";
    if (path?.includes(".parquet")) return "Parquet";
    if (path?.includes(".json")) return "JSON";
    if (path?.includes(".ndjson")) return "JSON";
    return "";
  }

  $: connectorType = formatConnectorType(sourceCatalog?.source?.connector);
  $: fileExtension = getFileExtension(sourceCatalog);

  $: cardinalityQuery = createQueryServiceTableCardinality(
    $runtime.instanceId,
    sourceName
  );
  $: cardinality = $cardinalityQuery?.data?.cardinality
    ? Number($cardinalityQuery?.data?.cardinality)
    : 0;

  /** get the current row count */
  $: {
    rowCount = `${formatInteger(cardinality)} row${
      cardinality !== 1 ? "s" : ""
    }`;
  }

  /** get the current column count */
  $: {
    columnCount = `${formatInteger(
      sourceCatalog?.source?.schema?.fields?.length
    )} columns`;
  }

  /** total % null cells */

  $: profileColumns = createQueryServiceTableColumns(
    $runtime?.instanceId,
    sourceName,
    {},
    { query: { keepPreviousData: true } }
  );

  let summaries: Readable<Array<ColumnSummary>>;
  $: if ($profileColumns?.data?.profileColumns) {
    summaries = getSummaries(sourceName, $runtime?.instanceId, $profileColumns);
  }

  let totalNulls = undefined;

  $: if (summaries) {
    totalNulls = $summaries.reduce(
      (total, column) => total + (+column.nullCount || 0),
      0
    );
  }
  $: {
    const totalCells =
      sourceCatalog?.source?.schema?.fields?.length * cardinality;
    nullPercentage = formatBigNumberPercentage(totalNulls / totalCells);
  }

  const sourceStore = useSourceStore(sourceName);

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    sourceName,
    $sourceStore.clientYAML
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;
</script>

<div
  class="table-profile {isSourceUnsaved && 'grayscale'} transition duration-200"
>
  {#if sourceCatalog && !$getSource.isError}
    <!-- summary info -->
    <div class="p-4 pt-2">
      <LeftRightGrid>
        <GridCell side="left"
          >{connectorType}
          {fileExtension !== "" ? `(${fileExtension})` : ""}</GridCell
        >
        <GridCell side="right" classes="text-gray-800 font-bold">
          {rowCount}
        </GridCell>

        <Tooltip location="left" alignment="start" distance={24}>
          <GridCell side="left" classes="text-gray-600">
            {#if totalNulls !== undefined}
              {nullPercentage} null
            {/if}
          </GridCell>
          <TooltipContent slot="tooltip-content">
            {#if totalNulls !== undefined}
              {nullPercentage} of table values are null
            {:else}
              awaiting calculation of total null table values
            {/if}
          </TooltipContent>
        </Tooltip>
        <GridCell side="right" classes="text-gray-800 font-bold">
          {columnCount}
        </GridCell>
      </LeftRightGrid>
    </div>

    <hr />

    <div class="pb-4 pt-4">
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="available columns"
          bind:active={showColumns}
        >
          columns
        </CollapsibleSectionTitle>
      </div>

      {#if showColumns}
        <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
          <ColumnProfile objectName={sourceName} indentLevel={0} />
        </div>
      {/if}
    </div>
  {/if}
</div>
