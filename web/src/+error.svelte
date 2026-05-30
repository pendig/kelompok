<script>
	import { page } from "$app/state";
	import { t } from "$lib/i18n.js";

	let isNotFound = $derived(page.status === 404);
</script>

<section class="error-page">
	<p class="eyebrow">{$t("errorPage.eyebrow")}</p>
	<h1 class="section-title">{isNotFound ? $t("errorPage.notFoundTitle") : $t("errorPage.title")}</h1>
	<p class="section-note">{isNotFound ? $t("errorPage.notFoundBody") : $t("errorPage.genericBody")}</p>

	{#if page.error?.message}
		<p class="error-page-detail">{page.error.message}</p>
	{/if}
	<p class="meta">{$t("errorPage.status", { status: page.status })}</p>

	<div class="inline-actions">
		<a class="btn primary" href="/">{$t("errorPage.backHome")}</a>
		<a class="ghost-button" href="/organizations">{$t("errorPage.browseOrganizations")}</a>
	</div>
</section>

<style>
	.error-page {
		display: grid;
		gap: 12px;
		justify-items: start;
		max-width: 60ch;
		margin: 48px 0;
	}

	.error-page .inline-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 12px;
		margin-top: 8px;
	}

	.error-page-detail {
		margin: 0;
		padding: 12px 14px;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: var(--surface-soft);
		color: var(--muted);
		font-size: 13.5px;
		overflow-wrap: break-word;
	}
</style>
