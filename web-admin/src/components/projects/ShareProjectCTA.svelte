<script lang="ts">
  import ProjectAccessControls from "@rilldata/web-admin/components/projects/ProjectAccessControls.svelte";
  import CLICommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";

  export let organization: string;
  export let project: string;

  let addUserCommand: string;
  $: addUserCommand = `rill user add --org ${organization} --project ${project} --role viewer`;
</script>

<div>
  <span class="uppercase text-gray-500 font-semibold text-[10px] leading-4"
    >Share</span
  >
  <ProjectAccessControls {organization} {project}>
    <svelte:fragment slot="manage-project">
      <div>
        Run this command in the Rill CLI to invite a teammate to view this
        project.
      </div>
      <CLICommandDisplay command={addUserCommand} />
    </svelte:fragment>
    <svelte:fragment slot="read-project">
      <div>
        Ask your organization’s admin to invite viewers using the Rill CLI.
      </div>
    </svelte:fragment>
  </ProjectAccessControls>
</div>
