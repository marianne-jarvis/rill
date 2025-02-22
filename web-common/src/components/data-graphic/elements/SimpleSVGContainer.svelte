<!-- @component
A convenience container that creates an svg element and slots in scales, mouseover values, and the configuration object
to the props.
-->
<script lang="ts">
  import { getContext } from "svelte";

  import { mousePositionToDomainActionFactory } from "../actions/mouse-position-to-domain-action-factory";
  import { createScrubAction } from "../actions/scrub-action-factory";
  import { contexts } from "../constants";
  import type { ScaleStore, SimpleConfigurationStore } from "../state/types";

  const config = getContext(contexts.config) as SimpleConfigurationStore;
  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;
  const { coordinates, mousePositionToDomain } =
    mousePositionToDomainActionFactory();

  const scrubActionObject = createScrubAction({
    plotLeft: $config?.plotLeft,
    plotRight: $config?.plotRight,
    plotTop: $config?.plotTop,
    plotBottom: $config?.plotBottom,
    startEventName: "scrub-start",
    moveEventName: "scrub-move",
    endEventName: "scrub-end",
  });

  // pull out the scrub action to be attached to the svg element
  const scrub = scrubActionObject.scrubAction;
  // const scrubCoordinates = scrubActionObject.coordinates;

  // make sure to reactively update the action store
  $: scrubActionObject.updatePlotBounds({
    plotLeft: $config?.plotLeft,
    plotRight: $config?.plotRight,
    plotTop: $config?.plotTop,
    plotBottom: $config?.plotBottom,
  });

  export let mouseoverValue = undefined;
  export let hovered = undefined;
  export let overflowHidden = true;

  $: mouseoverValue = $coordinates;

  $: hovered = $coordinates.x !== undefined;
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<svg
  style="overflow: {overflowHidden ? 'hidden' : 'visible'}"
  use:scrub
  on:scrub-start
  on:scrub-end
  on:scrub-move
  use:mousePositionToDomain
  on:click
  width={$config.width}
  height={$config.height}
>
  <slot
    config={$config}
    xScale={$xScale}
    yScale={$yScale}
    {mouseoverValue}
    {hovered}
  />
</svg>
