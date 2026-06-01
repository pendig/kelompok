<script>
	import { page } from "$app/state";
	import { onMount } from "svelte";
	import { initLocale, languages, locale, setLocale, t } from "$lib/i18n.js";
	import { theme, toggleTheme, initTheme } from "$lib/theme-mode.js";
	import "../app.css";

	let { children, data } = $props();
	let currentUser = $derived(data.session?.user);
	let navOpen = $state(false);
	const pathname = $derived(page.url.pathname || "/");

	onMount(() => {
		initLocale();
		initTheme();
	});

	function isActive(path) {
		if (path === "/") {
			return pathname === "/";
		}
		return pathname === path || pathname.startsWith(`${path}/`);
	}

	function isCurrent(path) {
		return pathname === path;
	}

	function closeNav() {
		navOpen = false;
	}
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

<a class="skip-link" href="#main-content">{$t("nav.skipToContent")}</a>

<div class="topbar">
	<div class="container topbar-inner">
		<a href="/" class="brand" aria-label="Kelompok home">
			<img src="/brand/logo-square.png" alt="" class="brand-mark" />
			<span class="brand-copy">
				<span class="brand-name">Kelompok</span>
				<span class="brand-tag">The Solutions of Movement</span>
			</span>
		</a>

		<button
			type="button"
			class="mobile-menu-button"
			aria-label={$t("nav.toggleMenu")}
			aria-controls="primary-navigation"
			aria-expanded={navOpen}
			onclick={() => (navOpen = !navOpen)}
		>
			<span aria-hidden="true"></span>
			<span aria-hidden="true"></span>
			<span aria-hidden="true"></span>
		</button>

		<nav id="primary-navigation" class="nav" class:open={navOpen} aria-label="Primary">
			<a href="/organizations" class="nav-link" class:active={isActive("/organizations")} aria-current={isCurrent("/organizations") ? "page" : undefined} onclick={closeNav}>{$t("nav.organizations")}</a>
			<a href="/posts" class="nav-link" class:active={isActive("/posts")} aria-current={isCurrent("/posts") ? "page" : undefined} onclick={closeNav}>{$t("nav.posts")}</a>
			<a href="/sdgs" class="nav-link" class:active={isActive("/sdgs")} aria-current={isCurrent("/sdgs") ? "page" : undefined} onclick={closeNav}>{$t("nav.sdgs")}</a>
			{#if currentUser}
				<a href="/account" class="nav-link" class:active={isActive("/account")} aria-current={isCurrent("/account") ? "page" : undefined} onclick={closeNav}>{$t("nav.account")}</a>
				{#if currentUser.role === "superadmin"}
					<a href="/admin" class="nav-link" class:active={isActive("/admin")} aria-current={isCurrent("/admin") ? "page" : undefined} onclick={closeNav}>{$t("nav.admin")}</a>
				{:else}
					<a href="/console" class="nav-link" class:active={isActive("/console")} aria-current={isCurrent("/console") ? "page" : undefined} onclick={closeNav}>{$t("nav.console")}</a>
				{/if}
			{:else}
				<a href="/login" class="nav-link" class:active={isActive("/login")} aria-current={isCurrent("/login") ? "page" : undefined} onclick={closeNav}>{$t("nav.login")}</a>
				<a href="/register" class="nav-link" class:active={isActive("/register")} aria-current={isCurrent("/register") ? "page" : undefined} onclick={closeNav}>{$t("nav.register")}</a>
			{/if}
			<a href="https://github.com/pendig/kelompok" class="nav-link" onclick={closeNav}>{$t("nav.source")}</a>
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
			<button
				type="button"
				class="theme-toggle"
				aria-label={$t("nav.themeToggle")}
				aria-pressed={$theme === "dark"}
				title={$theme === "dark" ? $t("nav.themeLight") : $t("nav.themeDark")}
				onclick={toggleTheme}
			>
				{#if $theme === "dark"}
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
						<circle cx="12" cy="12" r="4" />
						<path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
					</svg>
				{:else}
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
						<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
					</svg>
				{/if}
			</button>
		</nav>
	</div>
</div>

<main id="main-content" class="page" tabindex="-1">
	<div class="container">
		{@render children()}
	</div>
</main>
