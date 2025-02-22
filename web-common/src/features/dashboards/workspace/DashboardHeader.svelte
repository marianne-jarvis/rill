<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { projectShareStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Filters from "../filters/Filters.svelte";
  import { useMetaQuery } from "../selectors";
  import TimeControls from "../time-controls/TimeControls.svelte";

  export let metricViewName: string;
  export let hasTitle: boolean;

  const viewMetrics = (metricViewName: string) => {
    goto(`/dashboard/${metricViewName}/edit`);

    behaviourEvent.fireNavigationEvent(
      metricViewName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.Dashboard,
      MetricsEventScreenName.MetricsDefinition
    );
  };

  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);
  $: displayName = $metaQuery.data?.label;
  $: isEditableDashboard = $featureFlags.readOnly === false;

  function deployModal() {
    projectShareStore.set(true);
  }
</script>

<section class="w-full flex flex-col" id="header">
  <!-- top row: title and call to action -->
  <!-- Rill Local includes the title, Rill Cloud does not -->
  {#if hasTitle}
    <!-- FIXME: adding an -mb-3 fixes the spacing issue incurred by changes to the header
    to accommodate the cloud dashboard. We should go back and reconcile these headers so we don't need
  to do this. -->
    <div
      class="flex items-center justify-between -mb-3 w-full pl-1 pr-4"
      style:height="var(--header-height)"
    >
      <!-- title element -->
      <h1 style:line-height="1.1" style:margin-top="-1px">
        <div class="ui-copy-dashboard-header">
          {displayName || metricViewName}
        </div>
      </h1>
      <!-- top right CTAs -->

      <PanelCTA side="right">
        {#if isEditableDashboard}
          <Tooltip distance={8}>
            <Button
              on:click={() => viewMetrics(metricViewName)}
              type="secondary"
            >
              Edit Metrics <MetricsIcon size="16px" />
            </Button>
            <TooltipContent slot="tooltip-content">
              Edit this dashboard's metrics & settings
            </TooltipContent>
          </Tooltip>
          <Tooltip distance={8}>
            <Button on:click={deployModal} type="primary">Deploy</Button>
            <TooltipContent slot="tooltip-content">
              Deploy this dashboard to Rill Cloud
            </TooltipContent>
          </Tooltip>
        {/if}
      </PanelCTA>
    </div>
  {/if}
  <!-- bottom row -->
  <div class="-ml-3 p-1 py-2 space-y-2">
    <TimeControls {metricViewName} />
    {#key metricViewName}
      <Filters />
    {/key}
  </div>
</section>
