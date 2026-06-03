<script>
	import { t } from "$lib/i18n.js";

	const sections = [
		"intro",
		"data",
		"usage",
		"sharing",
		"security",
		"contact",
	];
</script>

<svelte:head>
	<title>{$t("privacyPage.title")} - Kelompok</title>
</svelte:head>

<nav class="breadcrumbs" aria-label="Breadcrumb">
	<a href="/">{$t("privacyPage.home")}</a>
	<span class="breadcrumb-separator" aria-hidden="true">›</span>
	<span class="breadcrumb-current" aria-current="page">{$t("privacyPage.eyebrow")}</span>
</nav>

<section class="section">
	<div>
		<p class="eyebrow">{$t("privacyPage.eyebrow")}</p>
		<h1 class="section-title">{$t("privacyPage.title")}</h1>
		<p class="section-note">{$t("privacyPage.subtitle")}</p>
		<p class="section-note">{$t("privacyPage.lastUpdatedLabel")}: {$t("privacyPage.lastUpdated")}</p>
	</div>

	<div class="card privacy-body">
		{#each sections as key}
			<h2>{$t(`privacyPage.${key}Title`)}</h2>
			<p>{$t(`privacyPage.${key}Body`)}</p>
		{/each}
	</div>
</section>

<style>
	.privacy-body {
		margin-top: 24px;
		padding: 32px;
		display: grid;
		gap: 8px;
		max-width: 760px;
	}

	.privacy-body h2 {
		margin: 20px 0 0;
		font-size: 18px;
		font-weight: 800;
	}

	.privacy-body h2:first-child {
		margin-top: 0;
	}

	.privacy-body p {
		margin: 0;
		color: var(--muted);
		line-height: 1.7;
	}
</style>
