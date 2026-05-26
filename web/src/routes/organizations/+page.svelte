<script>
	import { t } from "$lib/i18n.js";
	import { getTheme, getInitials } from "../../lib/theme.js";

	let { data } = $props();

	function formatLocation(org) {
		if (org.city && org.country) {
			return `${org.city}, ${org.country}`;
		}

		return org.city || org.country || $t("organizationsPage.unknownLocation");
	}

	function organizationPath(org) {
		return `/organizations/${encodeURIComponent(org.slug)}`;
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

<nav class="breadcrumbs">
	<a href="/">{$t("organizationDetail.home")}</a>
	<span>›</span>
	<span>{$t("nav.organizations")}</span>
</nav>

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
			{@const theme = getTheme(org.name)}
			<article class="card" style="padding: 0; overflow: hidden; display: flex; flex-direction: column; height: 100%;">
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
						<a href={organizationPath(org)}>{org.name}</a>
					</h3>
					<p class="small" style="margin: 0; flex-grow: 1; color: var(--muted); display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;">
						{org.description || $t("organizationsPage.noDescription")}
					</p>
					<p class="meta" style="margin: 0; font-size: 12px; font-weight: 600; color: var(--muted);">
						📍 {formatLocation(org)}
					</p>
				</div>
			</article>
		{/each}
	</div>
{/if}
