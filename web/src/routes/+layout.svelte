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

	let isOnline = $state(true);
	let showOnlineToast = $state(false);
	let toastTimeout;

	function updateConnectionStatus() {
		const wasOffline = !isOnline;
		isOnline = navigator.onLine;
		if (isOnline && wasOffline) {
			showOnlineToast = true;
			if (toastTimeout) clearTimeout(toastTimeout);
			toastTimeout = setTimeout(() => {
				showOnlineToast = false;
			}, 4000);
		}
	}

	onMount(() => {
		initLocale();
		initTheme();
		isOnline = navigator.onLine;
		window.addEventListener("online", updateConnectionStatus);
		window.addEventListener("offline", updateConnectionStatus);

		return () => {
			window.removeEventListener("online", updateConnectionStatus);
			window.removeEventListener("offline", updateConnectionStatus);
			if (toastTimeout) clearTimeout(toastTimeout);
		};
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
	<meta name="theme-color" content="#ffffff" />
	
	<!-- Favicon set wired from assets -->
	<link rel="icon" href="/favicon.ico" />
	<link rel="icon" type="image/svg+xml" href="/favicon.svg" />
	<link rel="icon" type="image/png" sizes="96x96" href="/favicon-96x96.png" />
	<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
	<link rel="manifest" href="/site.webmanifest" />

	<!-- OpenGraph and Twitter metadata using our brand Banner -->
	<meta property="og:title" content="Kelompok - The Solutions of Movement" />
	<meta property="og:description" content="Open-source free platform for nonprofit organization profiles, posts, and impact reports." />
	<meta property="og:image" content="/brand/kelompokId_Banner.png" />
	<meta property="og:type" content="website" />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content="Kelompok - The Solutions of Movement" />
	<meta name="twitter:description" content="Open-source free platform for nonprofit organization profiles, posts, and impact reports." />
	<meta name="twitter:image" content="/brand/kelompokId_Banner.png" />
</svelte:head>

<a class="skip-link" href="#main-content">{$t("nav.skipToContent")}</a>

<!-- Connection status banner -->
{#if !isOnline}
	<div class="connection-indicator offline visible" role="alert">
		<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m19 19-4-4"/><path d="m15 19-4-4"/><path d="M5 12a7 7 0 0 1 12-4.9"/><path d="M12 18a1.5 1.5 0 0 1-3 0c0-1 1-1.5 1.5-2"/></svg>
		<span>{$t("connection.offlineTitle")} - {$t("connection.offlineBody")}</span>
	</div>
{:else}
	{#if showOnlineToast}
		<div class="connection-indicator online visible" role="status">
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
			<span>{$t("connection.onlineTitle")} - {$t("connection.onlineBody")}</span>
		</div>
	{/if}
{/if}

<div class="topbar">
	<div class="container topbar-inner">
		<a href="/" class="brand" aria-label="Kelompok home">
			<img src="/brand/kelompokLogoHorizontalTransparant.png" alt="Kelompok Logo" class="brand-logo-horizontal" />
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

<!-- Floating Liquid Glass Bottom Navigation for Mobile -->
<div class="floating-bottom-nav">
	<a href="/" class="bottom-nav-link" class:active={isCurrent("/")} aria-label="Home">
		<svg class="nav-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
		<span class="nav-label">{$t("nav.home") || "Home"}</span>
	</a>
	<a href="/organizations" class="bottom-nav-link" class:active={isActive("/organizations")} aria-label="Organizations">
		<svg class="nav-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="16" height="20" x="4" y="2" rx="2" ry="2"/><path d="M9 22v-4h6v4"/><path d="M8 6h.01"/><path d="M16 6h.01"/><path d="M8 10h.01"/><path d="M16 10h.01"/><path d="M12 6h.01"/><path d="M12 10h.01"/><path d="M8 14h.01"/><path d="M16 14h.01"/><path d="M12 14h.01"/></svg>
		<span class="nav-label">{$t("nav.organizations")}</span>
	</a>
	<a href="/posts" class="bottom-nav-link" class:active={isActive("/posts")} aria-label="Posts">
		<svg class="nav-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1-2.5-2.5Z"/><path d="M6 6h10"/><path d="M6 10h10"/></svg>
		<span class="nav-label">{$t("nav.posts")}</span>
	</a>
	<a href="/sdgs" class="bottom-nav-link" class:active={isActive("/sdgs")} aria-label="SDGs">
		<svg class="nav-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 2v20"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
		<span class="nav-label">{$t("nav.sdgs")}</span>
	</a>
	{#if currentUser}
		<a href="/account" class="bottom-nav-link" class:active={isActive("/account")} aria-label="Account">
			<svg class="nav-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
			<span class="nav-label">{$t("nav.account")}</span>
		</a>
	{:else}
		<a href="/login" class="bottom-nav-link" class:active={isActive("/login")} aria-label="Login">
			<svg class="nav-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><polyline points="10 17 15 12 10 7"/><line x1="15" x2="3" y1="12" y2="12"/></svg>
			<span class="nav-label">{$t("nav.login")}</span>
		</a>
	{/if}
</div>
