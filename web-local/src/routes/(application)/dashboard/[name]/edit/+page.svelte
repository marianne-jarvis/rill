<script lang="ts">
  import { page } from "$app/stores";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { MetricsWorkspace } from "@rilldata/web-common/features/metrics-views";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";
  import { CATALOG_ENTRY_NOT_FOUND } from "../../../../../lib/errors/messages";

  $: metricViewName = $page.params.name;

  onMount(() => {
    if ($featureFlags.readOnly) {
      throw error(404, "Page not found");
    }
  });

  $: fileQuery = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(metricViewName, EntityType.MetricsDefinition),
    {
      query: {
        onError: (err) => {
          if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
            throw error(404, "Dashboard not found");
          }

          throw error(err.response?.status || 500, err.message);
        },
        // this will ensure that any changes done outside our app is pulled in.
        refetchOnWindowFocus: true,
      },
    }
  );

  $: yaml = $fileQuery.data?.blob || "";

  let nonStandardError: string | undefined;
  $: catalogQuery = createRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    metricViewName,
    {
      query: {
        onError: async (err) => {
          // If the catalog entry doesn't exist, the dashboard config is invalid
          // The component should render the specific error
          if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
            nonStandardError = err.message;
          }

          // Throw all other errors
          throw error(err.response?.status || 500, err.message);
        },
      },
    }
  );
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

{#if $fileQuery.data && yaml !== undefined}
  <MetricsWorkspace metricsDefName={metricViewName} />
{/if}
