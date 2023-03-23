<script lang="ts">
  import { createForm } from "svelte-forms-lib";
  import { notifications } from "../../../components/notifications";
  import {
    CONFIG_SELECTOR,
    CONFIG_TOP_LEVEL_LABEL_CLASSES,
    SELECTOR_BUTTON_TEXT_CLASSES,
    SELECTOR_CONTAINER,
  } from "./styles";

  export let metricsInternalRep;

  $: policiesNode = $metricsInternalRep.getMetricKey("policies");
  $: firstPolicy = policiesNode[0].expression;

  const { form, handleSubmit } = createForm({
    initialValues: {
      newPolicy: firstPolicy || "",
    },
    onSubmit: async (values) => {
      try {
        $metricsInternalRep.updatePolicies(values.newPolicy);
      } catch (err) {
        console.error(err);
        notifications.send({ message: err.response.data.message });
      }
    },
  });

  function handleKeydown(event: KeyboardEvent) {
    if (event.code == "Enter") {
      event.preventDefault();
      handleSubmit(event);
      (event.target as HTMLInputElement).blur();
    }
  }

  function updateFormWithNewPolicy(policy: string) {
    $form.newPolicy = policy;
  }

  // This kicks in when the user changes the policies via code artifact
  $: updateFormWithNewPolicy(firstPolicy);
</script>

<div class="flex flex-col gap-y-2">
  <div class={CONFIG_TOP_LEVEL_LABEL_CLASSES}>Policies</div>
  <div class="w-96" style={SELECTOR_CONTAINER.style}>
    <form id="policy-form" autocomplete="off">
      <input
        type="text"
        bind:value={$form["newPolicy"]}
        on:keydown={handleKeydown}
        on:blur={handleSubmit}
        placeholder={"Enter a SQL expression..."}
        class="{SELECTOR_BUTTON_TEXT_CLASSES} placeholder:font-normal placeholder:text-gray-600 bg-white w-full hover:bg-gray-200 rounded border border-6 border-gray-200 hover:border-gray-300  hover:text-gray-900 px-2 py-1 h-[34px] {CONFIG_SELECTOR.focus}"
      />
    </form>
  </div>
</div>
