<script>
	import { t } from "$lib/i18n.js";

	let { data } = $props();

	function formatLocation(org) {
		if (org.city && org.country) {
			return `${org.city}, ${org.country}`;
		}

		return org.city || org.country || $t("organizationsPage.unknownLocation");
	}
</script>

<section class="page-heading">
	<p class="eyebrow">{$t("nav.organizations")}</p>
	<h1 class="section-title">{$t("organizationsPage.title")}</h1>
	<p class="muted">{$t("organizationsPage.description")}</p>
</section>

{#if data.loadError}
	<p class="error">{$t("organizationsPage.loadError")}</p>
{/if}

{#if data.organizations.length === 0}
	<p class="empty">{$t("organizationsPage.empty")}</p>
{:else}
	<div class="grid">
		{#each data.organizations as org}
			<article class="card">
				<h3><a href={`/organizations/${org.slug}`}>{org.name}</a></h3>
				<p class="small">{org.description || $t("organizationsPage.noDescription")}</p>
				<p class="small muted">{formatLocation(org)}</p>
			</article>
		{/each}
	</div>
{/if}
