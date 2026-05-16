<script>
	import { onMount } from "svelte";
	import { initLocale, languages, locale, setLocale, t } from "$lib/i18n.js";
	import "../app.css";

	let { children, data } = $props();
	let currentUser = $derived(data.session?.user);

	onMount(() => {
		initLocale();
	});
</script>

<svelte:head>
	<title>Kelompok - The Solutions of Movement</title>
	<meta
		name="description"
		content="Open-source free platform for nonprofit organization profiles, posts, and impact reports."
	/>
	<meta name="theme-color" content="#5b21b6" />
	<link rel="icon" href="/favicon.ico" />
</svelte:head>

<div class="topbar">
	<div class="container topbar-inner">
		<a href="/" class="brand" aria-label="Kelompok home">
			<img src="/brand/logo-square.png" alt="" class="brand-mark" />
			<span class="brand-copy">
				<span class="brand-name">Kelompok</span>
				<span class="brand-tag">The Solutions of Movement</span>
			</span>
		</a>

		<nav class="nav" aria-label="Primary">
			<a href="/organizations" class="nav-link">{$t("nav.organizations")}</a>
			<a href="/posts" class="nav-link">{$t("nav.posts")}</a>
			{#if currentUser}
				<a href="/account" class="nav-link">{$t("nav.account")}</a>
				<a href="/admin" class="nav-link">{$t("nav.admin")}</a>
			{:else}
				<a href="/login" class="nav-link">{$t("nav.login")}</a>
				<a href="/register" class="nav-link">{$t("nav.register")}</a>
			{/if}
			<a href="https://github.com/pendig/kelompok" class="nav-link">{$t("nav.source")}</a>
			<div class="language-switch" aria-label="Language">
				{#each languages as language}
					<button
						type="button"
						class:active={$locale === language.code}
						aria-pressed={$locale === language.code}
						onclick={() => setLocale(language.code)}
					>
						{language.label}
					</button>
				{/each}
			</div>
		</nav>
	</div>
</div>

<main class="page">
	<div class="container">
		{@render children()}
	</div>
</main>
