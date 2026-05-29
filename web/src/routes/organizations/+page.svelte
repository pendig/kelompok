<script>
	import { t } from "$lib/i18n.js";
	import { getTheme, getInitials } from "../../lib/theme.js";

	let { data } = $props();
	let searchTerm = $state("");
	let locationFilter = $state("");
	let focusFilter = $state("");
	let sdgFilter = $state("");

	const organizations = $derived(data.organizations || []);
	const locationOptions = $derived(uniqueOptions(organizations.flatMap((org) => [org.region, org.city])));
	const focusOptions = $derived(uniqueOptions(organizations.flatMap((org) => organizationFocus(org))));
	const sdgOptions = $derived(uniqueOptions(organizations.flatMap((org) => organizationSdgs(org))));
	const hasActiveFilters = $derived(
		searchTerm.trim() || locationFilter || focusFilter || sdgFilter,
	);
	const filteredOrganizations = $derived(
		organizations.filter((org) => {
			const query = normalize(searchTerm);
			const matchesSearch =
				!query ||
				[
					org.name,
					org.slug,
					org.description,
					org.legal_name,
					org.city,
					org.region,
					org.country,
				]
					.map(normalize)
					.some((value) => value.includes(query));
			const locations = [org.region, org.city].filter(Boolean);
			const matchesLocation = !locationFilter || locations.includes(locationFilter);
			const matchesFocus = !focusFilter || organizationFocus(org).includes(focusFilter);
			const matchesSdg = !sdgFilter || organizationSdgs(org).includes(sdgFilter);

			return matchesSearch && matchesLocation && matchesFocus && matchesSdg;
		}),
	);

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

	function normalize(value) {
		return String(value || "").trim().toLowerCase();
	}

	function uniqueOptions(values) {
		return Array.from(new Set(values.map((value) => String(value || "").trim()).filter(Boolean))).sort(
			(first, second) => first.localeCompare(second),
		);
	}

	function toArray(value) {
		if (Array.isArray(value)) {
			return value;
		}
		if (typeof value === "string" && value.trim()) {
			return [value.trim()];
		}
		return [];
	}

	function organizationFocus(org) {
		const profile = org.profile_data || {};
		return uniqueOptions([
			...toArray(profile.focus),
			...toArray(profile.tags),
			...toArray(profile.programs),
		]);
	}

	function organizationSdgs(org) {
		const sdgs = org.sdgs_data || {};
		return uniqueOptions([
			...toArray(sdgs.primary),
			...toArray(sdgs.secondary),
			...toArray(sdgs.goals),
		]);
	}

	function resetFilters() {
		searchTerm = "";
		locationFilter = "";
		focusFilter = "";
		sdgFilter = "";
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

<section class="directory-tools" aria-label={$t("organizationsPage.filterLabel")}>
	<label>
		<span>{$t("organizationsPage.searchLabel")}</span>
		<input
			type="search"
			bind:value={searchTerm}
			placeholder={$t("organizationsPage.searchPlaceholder")}
		/>
	</label>

	<label>
		<span>{$t("organizationsPage.locationFilter")}</span>
		<select bind:value={locationFilter}>
			<option value="">{$t("organizationsPage.allLocations")}</option>
			{#each locationOptions as location}
				<option value={location}>{location}</option>
			{/each}
		</select>
	</label>

	<label>
		<span>{$t("organizationsPage.focusFilter")}</span>
		<select bind:value={focusFilter}>
			<option value="">{$t("organizationsPage.allFocus")}</option>
			{#each focusOptions as focus}
				<option value={focus}>{focus}</option>
			{/each}
		</select>
	</label>

	<label>
		<span>{$t("organizationsPage.sdgFilter")}</span>
		<select bind:value={sdgFilter}>
			<option value="">{$t("organizationsPage.allSdgs")}</option>
			{#each sdgOptions as sdg}
				<option value={sdg}>{sdg}</option>
			{/each}
		</select>
	</label>

	<button class="ghost-button" type="button" onclick={resetFilters} disabled={!hasActiveFilters}>
		{$t("organizationsPage.resetFilters")}
	</button>
</section>

<p class="directory-count">
	{$t("organizationsPage.resultCount", { count: filteredOrganizations.length })}
</p>

{#if organizations.length === 0}
	<div class="empty directory-empty">
		<h2>{$t("organizationsPage.emptyTitle")}</h2>
		<p>{$t("organizationsPage.empty")}</p>
		<div class="actions">
			<a href="/register">{$t("organizationsPage.createOrgCta")}</a>
		</div>
	</div>
{:else if filteredOrganizations.length === 0}
	<div class="empty directory-empty">
		<h2>{$t("organizationsPage.noResultsTitle")}</h2>
		<p>{$t("organizationsPage.noResultsBody")}</p>
		<div class="actions">
			<button class="ghost-button" type="button" onclick={resetFilters}>
				{$t("organizationsPage.resetFilters")}
			</button>
			<a href="/login">{$t("organizationsPage.claimCta")}</a>
		</div>
	</div>
{:else}
	<div class="grid">
		{#each filteredOrganizations as org}
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
					{#if organizationSdgs(org).length > 0 || organizationFocus(org).length > 0}
						<div class="directory-tags" aria-label={$t("organizationsPage.cardSignals")}>
							{#each organizationSdgs(org).slice(0, 3) as sdg}
								<span>{sdg}</span>
							{/each}
							{#each organizationFocus(org).slice(0, 2) as focus}
								<span>{focus}</span>
							{/each}
						</div>
					{/if}
				</div>
			</article>
		{/each}
	</div>
{/if}

<style>
	.directory-tools {
		display: grid;
		grid-template-columns: minmax(220px, 1.4fr) repeat(3, minmax(160px, 1fr)) auto;
		gap: 12px;
		align-items: end;
		margin: 24px 0 12px;
		padding: 16px;
		border: 1px solid var(--border);
		border-radius: 12px;
		background: var(--surface);
		box-shadow: var(--shadow-sm);
	}

	.directory-tools label {
		display: grid;
		gap: 6px;
		min-width: 0;
		color: var(--muted);
		font-size: 12px;
		font-weight: 700;
	}

	.directory-tools .ghost-button {
		min-height: 40px;
	}

	.directory-count {
		margin: 0 0 16px;
		color: var(--muted);
		font-size: 13px;
		font-weight: 700;
	}

	.directory-empty {
		display: grid;
		justify-items: center;
		gap: 8px;
	}

	.directory-empty h2,
	.directory-empty p {
		margin: 0;
	}

	.directory-empty h2 {
		color: var(--text);
		font-size: 18px;
		letter-spacing: 0;
	}

	.directory-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin-top: 4px;
	}

	.directory-tags span {
		display: inline-flex;
		align-items: center;
		min-height: 22px;
		padding: 0 8px;
		border: 1px solid var(--border);
		border-radius: 999px;
		background: var(--surface-soft);
		color: var(--muted);
		font-size: 11px;
		font-weight: 700;
		line-height: 1.2;
	}

	@media (max-width: 1080px) {
		.directory-tools {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 720px) {
		.directory-tools {
			grid-template-columns: 1fr;
		}
	}
</style>
