<script lang="ts">
  import { page } from "$app/stores";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";
  import {
    createAdminServiceGetOrganization,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import OrganizationHero from "../../components/organizations/OrganizationHero.svelte";
  import ProjectList from "../../components/projects/ProjectList.svelte";

  $: orgName = $page.params.organization;

  $: org = createAdminServiceGetOrganization(orgName);
  $: projs = createAdminServiceListProjectsForOrganization(orgName, undefined, {
    query: { enabled: !!$org.data?.organization },
  });
</script>

<svelte:head>
  <title>{orgName} overview - Rill</title>
</svelte:head>

{#if $org.data && $org.data.organization && $projs.data}
  <VerticalScrollContainer>
    <section
      class="mx-8 my-8 sm:my-16 sm:mx-16 lg:mx-32 lg:my-24 2xl:mx-64 mx-auto flex flex-col gap-y-4"
    >
      <OrganizationHero organization={orgName} />
      {#if $projs.data.projects?.length === 0}
        <span
          >This organization has no projects yet. <a
            href="https://docs.rilldata.com/get-started"
            target="_blank"
            rel="noreferrer">See docs</a
          ></span
        >
      {:else}
        <div class="py-2 px-1.5">
          <ProjectList organization={$page.params.organization} />
        </div>
      {/if}
    </section>
  </VerticalScrollContainer>
{/if}
