<script>
	import { locale, t } from "$lib/i18n.js";
	import { getAllSdgGoals, normalizeSdgGoals } from "$lib/sdgs.js";

	let { data } = $props();

	let goals = $derived(getAllSdgGoals($locale));
	let organizationsByGoal = $derived(
		Object.fromEntries(
			goals.map((goal) => [
				goal.number,
				data.organizations.filter((organization) =>
					normalizeSdgGoals(organization.sdgs_data?.primary || [], $locale).some((item) => item.number === goal.number),
				),
			]),
		),
	);
	let tracedCount = $derived(Object.values(organizationsByGoal).filter((items) => items.length > 0).length);

	function organizationPath(organization) {
		return `/organizations/${encodeURIComponent(organization.slug)}`;
	}

	function formatLocation(organization) {
		const parts = [organization.city, organization.region, organization.country].filter(Boolean);
		return parts.length ? parts.join(", ") : $t("organizationsPage.unknownLocation");
	}
</script>

<nav class="breadcrumbs">
	<a href="/">{$t("organizationDetail.home")}</a>
	<span>›</span>
	<span>{$t("sdgsPage.title")}</span>
</nav>

<section class="section sdgs-hero">
	<div>
		<p class="eyebrow">{$t("sdgsPage.eyebrow")}</p>
		<h1 class="section-title">{$t("sdgsPage.title")}</h1>
		<p class="section-note">{$t("sdgsPage.subtitle")}</p>
	</div>
	<div class="sdgs-hero-card">
		<p class="label">{$t("sdgsPage.tracedIssues")}</p>
		<strong>{tracedCount}/17</strong>
		<span>{$t("sdgsPage.tracedIssuesHelp")}</span>
	</div>
</section>

{#if data.loadError}
	<p class="error">{$t("sdgsPage.loadError")}</p>
{/if}

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("sdgsPage.directoryEyebrow")}</p>
			<h2 class="section-title">{$t("sdgsPage.directoryTitle")}</h2>
		</div>
		<p class="section-note">{$t("sdgsPage.directoryNote")}</p>
	</div>

	<div class="sdgs-directory">
		{#each goals as goal}
			{@const organizations = organizationsByGoal[goal.number] || []}
			<article class="sdgs-directory-card">
				<div class="sdgs-directory-head">
					<img src={goal.icon} alt={`SDG ${goal.number}: ${goal.title}`} loading="lazy" />
					<div>
						<p class="label">SDG {goal.code}</p>
						<h3>{goal.title}</h3>
					</div>
				</div>

				{#if organizations.length}
					<div class="sdgs-org-list">
						{#each organizations.slice(0, 4) as organization}
							<a href={organizationPath(organization)}>
								<span>{organization.name}</span>
								<small>{formatLocation(organization)}</small>
							</a>
						{/each}
					</div>
				{:else}
					<p class="empty compact">{$t("sdgsPage.noOrganizations")}</p>
				{/if}
			</article>
		{/each}
	</div>
</section>
