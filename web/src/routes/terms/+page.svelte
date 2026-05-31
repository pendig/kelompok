<script>
	import { t } from "$lib/i18n.js";

	const sections = [
		"acceptance",
		"accounts",
		"use",
		"content",
		"liability",
		"changes",
		"contact",
	];
</script>

<svelte:head>
	<title>{$t("termsPage.title")} - Kelompok</title>
</svelte:head>

<nav class="breadcrumbs" aria-label="Breadcrumb">
	<a href="/">{$t("termsPage.home")}</a>
	<span aria-hidden="true">›</span>
	<span>{$t("termsPage.eyebrow")}</span>
</nav>

<section class="section">
	<div>
		<p class="eyebrow">{$t("termsPage.eyebrow")}</p>
		<h1 class="section-title">{$t("termsPage.title")}</h1>
		<p class="section-note">{$t("termsPage.subtitle")}</p>
		<p class="section-note">{$t("termsPage.lastUpdatedLabel")}: {$t("termsPage.lastUpdated")}</p>
	</div>

	<div class="card terms-body">
		{#each sections as key}
			<h2>{$t(`termsPage.${key}Title`)}</h2>
			<p>{$t(`termsPage.${key}Body`)}</p>
		{/each}
	</div>
</section>

<style>
	.terms-body {
		margin-top: 24px;
		padding: 32px;
		display: grid;
		gap: 8px;
		max-width: 760px;
	}

	.terms-body h2 {
		margin: 20px 0 0;
		font-size: 18px;
		font-weight: 800;
	}

	.terms-body h2:first-child {
		margin-top: 0;
	}

	.terms-body p {
		margin: 0;
		color: var(--muted);
		line-height: 1.7;
	}
</style>
