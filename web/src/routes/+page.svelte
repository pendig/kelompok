<script>
	import { fallbackDate } from "../lib/api.js";
	import { locale, t } from "$lib/i18n.js";
	import { getTheme, getInitials } from "../lib/theme.js";

	let { data } = $props();

	const managedRoles = $derived(
		(data.session?.organization_roles || []).filter((role) => ["owner", "admin", "editor"].includes(role.role)),
	);
	const managerPath = $derived(
		data.session ? (managedRoles.length ? "/console" : "/organizations") : "/register",
	);
	const managerActionKey = $derived(
		data.session ? (managedRoles.length ? "home.managerConsoleAction" : "home.managerClaimAction") : "home.managerRegisterAction",
	);

	function formatLocation(org, fallback) {
		if (org.city && org.country) {
			return `${org.city}, ${org.country}`;
		}

		return org.city || org.country || fallback;
	}

	function organizationPath(slug) {
		return `/organizations/${encodeURIComponent(slug)}`;
	}

	function postPath(post) {
		return `${organizationPath(post.organization?.slug || post.organization_slug)}/posts/${encodeURIComponent(post.slug)}`;
	}

	function claimStatusLabel(status) {
		const labels = {
			claimed: $t("organizationDetail.claimStatusClaimed"),
			pending: $t("organizationDetail.claimStatusPending"),
			rejected: $t("organizationDetail.claimStatusRejected"),
			unclaimed: $t("organizationDetail.claimStatusUnclaimed"),
		};

		return labels[status] || status || $t("organizationDetail.claimStatusUnclaimed");
	}
</script>

<section class="hero">
	<div class="hero-grid">
		<div class="hero-copy">
			<p class="eyebrow inverse">{$t("home.eyebrow")}</p>
			<h1 class="hero-title">{$t("home.title")}</h1>
			<p>{$t("home.subtitle")}</p>

			<div class="hero-actions">
				<a class="btn primary" href={managerPath}>{$t("home.managerAction")}</a>
				<a class="btn secondary" href="/organizations">{$t("home.seekerAction")}</a>
				<a class="btn secondary" href="/sdgs">{$t("home.sdgsAction")}</a>
			</div>

			<div class="hero-stats">
				<div class="stat">
					<span class="stat-value">{data.organizations.length}</span>
					<span class="stat-label">{$t("home.organizationsLoaded")}</span>
				</div>
				<div class="stat">
					<span class="stat-value">{data.posts.length}</span>
					<span class="stat-label">{$t("home.postsLoaded")}</span>
				</div>
				<div class="stat">
					<span class="stat-value">{$t("home.mvpLabel")}</span>
					<span class="stat-label">{$t("home.mvpDesc")}</span>
				</div>
			</div>
		</div>

		<aside class="hero-preview">
			<div class="preview-head">
				<div class="preview-brand">
					<img src="/brand/logo-square.png" alt="Kelompok logo" class="preview-logo" />
					<div>
						<p class="eyebrow inverse">{$t("home.previewEyebrow")}</p>
						<h2 class="preview-title">{$t("home.previewTitle")}</h2>
					</div>
				</div>

				<span class="preview-chip">{$t("home.previewChip")}</span>
			</div>

			<div class="preview-profile-card" aria-label={$t("home.previewTitle")}>
				<div class="preview-profile-cover"></div>
				<div class="preview-profile-body">
					<div class="preview-profile-main">
						<img src="/brand/logo-square.png" alt="" class="preview-profile-avatar" />
						<div>
							<p class="preview-profile-name">{$t("home.previewOrgName")}</p>
							<p class="preview-profile-meta">{$t("home.previewOrgMeta")}</p>
						</div>
					</div>

					<div class="preview-profile-tabs" aria-hidden="true">
						<span>{$t("organizationDetail.tabProfile")}</span>
						<span>{$t("home.posts")}</span>
						<span>{$t("organizationDetail.impactReports")}</span>
					</div>

					<div class="preview-profile-grid">
						<div>
							<p>{$t("organizationDetail.location")}</p>
							<strong>{$t("home.previewLocation")}</strong>
						</div>
						<div>
							<p>{$t("home.previewSdgLabel")}</p>
							<div class="preview-sdg-row">
								<span>4</span>
								<span>11</span>
								<span>13</span>
							</div>
						</div>
					</div>

					<div class="preview-profile-story">
						<p>{$t("home.previewPostTitle")}</p>
						<strong>{$t("home.previewPostBody")}</strong>
					</div>
				</div>
			</div>

			<div class="preview-foot">
				<div>
					<p class="label inverse">{$t("home.claimLabel")}</p>
					<p class="value inverse">{$t("home.claimValue")}</p>
				</div>

				<ul class="chip-list" aria-label="Core capabilities">
					{#each $t("home.claimSignals") as signal}
						<li>{signal}</li>
					{/each}
				</ul>
			</div>
		</aside>
	</div>
</section>

<section class="section">
	<div class="pathway-grid">
		<article class="pathway-panel">
			<p class="eyebrow">{$t("home.managerEyebrow")}</p>
			<h2>{$t("home.managerTitle")}</h2>
			<p>{$t(data.session ? "home.managerBodyAuthenticated" : "home.managerBodyAnonymous")}</p>
			<a class="ghost-button" href={managerPath}>{$t(managerActionKey)}</a>
		</article>

		<article class="pathway-panel">
			<p class="eyebrow">{$t("home.seekerEyebrow")}</p>
			<h2>{$t("home.seekerTitle")}</h2>
			<p>{$t("home.seekerBody")}</p>
			<a class="ghost-button" href="/organizations">{$t("home.seekerBrowseAction")}</a>
		</article>
	</div>
</section>

{#if data.loadErrors?.length}
	<section class="notice" aria-live="polite">
		<p class="eyebrow">{$t("home.noticeTitle")}</p>
		<p>{$t("home.noticeBody")}</p>
	</section>
{/if}

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("home.whyEyebrow")}</p>
			<h2 class="section-title">{$t("home.whyTitle")}</h2>
		</div>
		<p class="section-note">{$t("home.whyNote")}</p>
	</div>

	<div class="feature-grid">
		{#each $t("home.whyReasons") as reason}
			<article class="feature-card quiet">
				<h3>{reason.title}</h3>
				<p>{reason.copy}</p>
			</article>
		{/each}
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("home.scopeEyebrow")}</p>
			<h2 class="section-title">{$t("home.scopeTitle")}</h2>
		</div>
		<p class="section-note">{$t("home.scopeNote")}</p>
	</div>

	<div class="feature-grid">
		{#each $t("home.features") as card, index}
			<article class="feature-card">
				<p class="feature-index">0{index + 1}</p>
				<h3>{card.title}</h3>
				<p>{card.copy}</p>
			</article>
		{/each}
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("home.liveEyebrow")}</p>
			<h2 class="section-title">{$t("home.liveTitle")}</h2>
		</div>
		<p class="section-note">{$t("home.liveNote")}</p>
	</div>

	<div class="content-grid">
		<div class="content-column">
			<div class="mini-head">
				<h3 class="section-title">{$t("home.organizations")}</h3>
				<span class="mini-badge">{$t("home.total", { count: data.organizations.length })}</span>
			</div>

			{#if data.organizations.length === 0}
				<p class="empty">{$t("home.noOrganizations")}</p>
			{:else}
				<div style="display: grid; gap: 16px;">
					{#each data.organizations.slice(0, 4) as org}
						{@const theme = getTheme(org.name)}
						<article class="surface-card" style="padding: 0; overflow: hidden; display: flex; flex-direction: column; height: 100%;">
							<!-- Mini Cover Banner -->
							<div class="mini-card-cover" style="background: {theme.cover};"></div>
							
							<!-- Mini Avatar Overlapping Banner -->
							<div style="padding-inline: 16px; margin-top: -24px; display: flex; align-items: flex-end; justify-content: space-between; position: relative; z-index: 2;">
								<div class="mini-card-avatar" style="width: 48px; height: 48px; font-size: 16px; color: {theme.avatarText}; background: {theme.avatarBg};">
									{getInitials(org.name)}
								</div>
								{#if org.claim_status}
									<span class="admin-status {org.claim_status === 'claimed' ? 'admin-status-pass' : 'admin-status-warn'}" style="font-size: 9.5px; padding: 2px 8px;">
										{claimStatusLabel(org.claim_status)}
									</span>
								{/if}
							</div>

							<!-- Card Content -->
							<div style="padding: 16px; display: flex; flex-direction: column; flex-grow: 1; gap: 8px;">
								<h3 style="margin: 0; font-size: 17px; font-weight: 700; line-height: 1.3;">
									<a href={organizationPath(org.slug)}>{org.name}</a>
								</h3>
								<p class="small" style="margin: 0; flex-grow: 1; color: var(--muted); display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;">
									{org.description || $t("home.noDescription")}
								</p>
								<p class="meta" style="margin: 0; font-size: 12px; font-weight: 600; color: var(--muted);">
									📍 {formatLocation(org, $t("home.unknownLocation"))}{org.region ? ` · ${org.region}` : ""}
								</p>
							</div>
						</article>
					{/each}
				</div>
			{/if}
		</div>

		<div class="content-column">
			<div class="mini-head">
				<h3 class="section-title">{$t("home.posts")}</h3>
				<span class="mini-badge">{$t("home.total", { count: data.posts.length })}</span>
			</div>

			{#if data.posts.length === 0}
				<p class="empty">{$t("home.noPosts")}</p>
			{:else}
				{#each data.posts.slice(0, 4) as post}
					<article class="surface-card">
						<div class="card-top">
							<h3><a href={postPath(post)}>{post.title}</a></h3>
							<span class="pill">{fallbackDate(post.published_at, $locale)}</span>
						</div>

						<p class="small">{post.summary || "—"}</p>
						<p class="meta">
							{$t("home.by", { name: post.organization?.name || $t("home.unknownAuthor") })}
						</p>
					</article>
				{/each}
			{/if}
		</div>
	</div>
</section>

<section class="section callout">
	<div class="callout-inner">
		<div>
			<p class="eyebrow">{$t("home.calloutEyebrow")}</p>
			<h2>{$t("home.calloutTitle")}</h2>
			<p>{$t("home.calloutBody")}</p>
		</div>

		<a class="btn primary" href="https://github.com/pendig/kelompok">{$t("home.sourceAction")}</a>
	</div>
</section>
