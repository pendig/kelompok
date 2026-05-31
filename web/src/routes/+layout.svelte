<script>
	import { page } from "$app/state";
	import { onMount } from "svelte";
	import { initLocale, languages, locale, setLocale, t } from "$lib/i18n.js";
	import "../app.css";

	let { children, data } = $props();
	let currentUser = $derived(data.session?.user);
	let navOpen = $state(false);
	const pathname = $derived(page.url.pathname || "/");

	onMount(() => {
		initLocale();
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

	let isSuperadmin = $derived(currentUser && currentUser.role === "superadmin");
	let showMaintenance = $derived(data.maintenance && !isSuperadmin);

	const maintenanceText = {
		id: {
			title: "Situs Sedang Pemeliharaan",
			body: "Kelompok sedang melakukan pembaruan berkala untuk memberikan pengalaman terbaik. Kami akan kembali online secepatnya.",
			button: "Coba Lagi"
		},
		en: {
			title: "System Maintenance",
			body: "Kelompok is undergoing scheduled updates to provide the best possible experience. We will be back online shortly.",
			button: "Try Again"
		}
	};
	let currentText = $derived(maintenanceText[$locale] || maintenanceText.en);
</script>

{#if showMaintenance}
	<div class="maintenance-wrapper">
		<div class="glowing-blob blob-1"></div>
		<div class="glowing-blob blob-2"></div>
		
		<div class="maintenance-card">
			<div class="maintenance-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="cog-icon">
					<circle cx="12" cy="12" r="3"></circle>
					<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
				</svg>
			</div>
			
			<h1 class="maintenance-title">{currentText.title}</h1>
			<p class="maintenance-body">{currentText.body}</p>
			
			<div class="maintenance-divider"></div>
			
			<div class="maintenance-footer">
				<button type="button" class="btn-refresh" onclick={() => window.location.reload()}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="refresh-icon">
						<path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67"></path>
					</svg>
					{currentText.button}
				</button>
				
				<div class="lang-selector">
					{#each languages as language}
						<button
							type="button"
							class="lang-btn"
							class:active={$locale === language.code}
							onclick={() => setLocale(language.code)}
						>
							{language.label}
						</button>
					{/each}
				</div>
			</div>
		</div>
	</div>
{:else}
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
			</nav>
		</div>
	</div>

	<main id="main-content" class="page" tabindex="-1">
		<div class="container">
			{@render children()}
		</div>
	</main>
{/if}

<style>
	.maintenance-wrapper {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		background: radial-gradient(circle at center, #1e1b4b 0%, #09090b 100%);
		display: flex;
		justify-content: center;
		align-items: center;
		overflow: hidden;
		font-family: 'Outfit', 'Inter', system-ui, sans-serif;
		color: #f4f4f5;
		z-index: 99999;
	}

	.glowing-blob {
		position: absolute;
		border-radius: 50%;
		filter: blur(120px);
		opacity: 0.15;
		z-index: 1;
		animation: pulse-blob 8s ease-in-out infinite alternate;
	}

	.blob-1 {
		width: 400px;
		height: 400px;
		background: #6366f1;
		top: -100px;
		left: -100px;
	}

	.blob-2 {
		width: 450px;
		height: 450px;
		background: #a855f7;
		bottom: -150px;
		right: -100px;
		animation-delay: -4s;
	}

	@keyframes pulse-blob {
		0% { transform: scale(1) translate(0, 0); }
		100% { transform: scale(1.15) translate(30px, 30px); }
	}

	.maintenance-card {
		position: relative;
		z-index: 2;
		background: rgba(24, 24, 27, 0.45);
		backdrop-filter: blur(24px);
		-webkit-backdrop-filter: blur(24px);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 24px;
		padding: 48px;
		max-width: 520px;
		width: 90%;
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
		text-align: center;
		display: flex;
		flex-direction: column;
		align-items: center;
	}

	.maintenance-icon {
		display: flex;
		justify-content: center;
		align-items: center;
		width: 80px;
		height: 80px;
		background: linear-gradient(135deg, rgba(99, 102, 241, 0.15) 0%, rgba(168, 85, 247, 0.15) 100%);
		border: 1px solid rgba(168, 85, 247, 0.3);
		border-radius: 20px;
		margin-bottom: 28px;
		color: #a855f7;
		box-shadow: 0 8px 32px 0 rgba(168, 85, 247, 0.1);
	}

	.cog-icon {
		width: 42px;
		height: 42px;
		animation: spin-cog 12s linear infinite;
	}

	@keyframes spin-cog {
		100% { transform: rotate(360deg); }
	}

	.maintenance-title {
		font-size: 2rem;
		font-weight: 700;
		margin: 0 0 16px 0;
		background: linear-gradient(135deg, #ffffff 30%, #a855f7 100%);
		-webkit-background-clip: text;
		background-clip: text;
		-webkit-text-fill-color: transparent;
		letter-spacing: -0.025em;
	}

	.maintenance-body {
		font-size: 1.05rem;
		line-height: 1.6;
		color: #a1a1aa;
		margin: 0;
		padding: 0 12px;
	}

	.maintenance-divider {
		width: 100%;
		height: 1px;
		background: linear-gradient(to right, transparent, rgba(255, 255, 255, 0.08), transparent);
		margin: 32px 0;
	}

	.maintenance-footer {
		display: flex;
		flex-direction: column;
		gap: 20px;
		width: 100%;
		align-items: center;
	}

	.btn-refresh {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		background: linear-gradient(135deg, #6366f1 0%, #a855f7 100%);
		color: white;
		border: none;
		padding: 12px 32px;
		border-radius: 14px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
		box-shadow: 0 4px 16px rgba(99, 102, 241, 0.35);
	}

	.btn-refresh:hover {
		transform: translateY(-2px);
		box-shadow: 0 6px 20px rgba(99, 102, 241, 0.5);
	}

	.btn-refresh:active {
		transform: translateY(0);
	}

	.refresh-icon {
		width: 18px;
		height: 18px;
		transition: transform 0.5s ease;
	}

	.btn-refresh:hover .refresh-icon {
		transform: rotate(180deg);
	}

	.lang-selector {
		display: flex;
		gap: 8px;
		background: rgba(255, 255, 255, 0.04);
		padding: 4px;
		border-radius: 10px;
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.lang-btn {
		background: transparent;
		border: none;
		color: #71717a;
		padding: 6px 12px;
		border-radius: 8px;
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		outline: none;
	}

	.lang-btn.active {
		background: rgba(255, 255, 255, 0.08);
		color: #ffffff;
	}

	.lang-btn:hover:not(.active) {
		color: #a1a1aa;
	}
</style>
