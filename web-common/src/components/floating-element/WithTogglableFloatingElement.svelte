<script lang="ts">
  import { createEventDispatcher, setContext } from "svelte";
  import { writable } from "svelte/store";
  import Portal from "../Portal.svelte";
  import { FloatingElement } from "./index";

  export let location = "bottom";
  export let alignment = "middle";
  export let relationship = "parent";
  export let distance = 0;
  export let pad = 8;
  export let suppress = false;
  export let active = false;
  export let inline = false;
  export let overflowFlipY = true;

  /** this passes down the dom element used for the "outside click" action.
   * Since this element is not strictly within the parent of the menu (which is in a Portal),
   * we will need to check to see if this element was also clicked before firing the outside click callback.
   */
  const triggerElementStore = writable(undefined);
  $: triggerElementStore.set(parent?.children?.[0]);
  setContext("rill:menu:menuTrigger", triggerElementStore);

  const dispatch = createEventDispatcher();
  $: {
    if (active) dispatch("open");
    if (!active) dispatch("close");
  }

  let parent;
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div class:inline bind:this={parent}>
  <slot
    {active}
    handleClose={() => {
      active = false;
    }}
    toggleFloatingElement={() => {
      active = !active;
    }}
  />
  {#if active && !suppress}
    <Portal>
      <div style="z-index: 50;">
        <FloatingElement
          target={parent}
          {relationship}
          {location}
          {alignment}
          {distance}
          {pad}
          {overflowFlipY}
        >
          <slot name="floating-element" />
        </FloatingElement>
      </div>
    </Portal>
  {/if}
</div>
