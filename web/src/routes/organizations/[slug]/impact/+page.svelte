<script>
	import { fallbackDate } from "../../../../lib/api.js";
	import { locale, t } from "$lib/i18n.js";
	import { normalizeSdgGoals } from "$lib/sdgs.js";

	let { data } = $props();
	let org = $derived(data.organization);
</script>

<nav class="breadcrumbs">
	<a href="/">{$t("organizationDetail.home")}</a>
	<span>›</span>
	<a href="/organizations">{$t("nav.organizations")}</a>
	<span>›</span>
	<a href={`/organizations/${encodeURIComponent(org.slug)}`}>{org.name}</a>
	<span>›</span>
	<span>{$t("organizationDetail.impactReports")}</span>
</nav>

<h1 class="section-title">{$t("impactPage.title")}</h1>
<p class="muted">{org.name}</p>

{#if data.impactReports.length === 0}
	<p class="empty">{$t("impactPage.empty")}</p>
{:else}
	<div>
		{#each data.impactReports as report}
			{@const reportGoals = normalizeSdgGoals(report.sdgs || [], $locale)}
			<div class="card list-item">
				<h3 class="title">{report.title}</h3>
				<p class="small muted">
					{$t("impactPage.period")}:
					{report.report_period_start ? fallbackDate(report.report_period_start, $locale) : "—"} →
					{report.report_period_end ? fallbackDate(report.report_period_end, $locale) : "—"}
				</p>
				<p>{report.summary || $t("impactPage.noSummary")}</p>
				{#if reportGoals.length}
					<div class="sdg-grid compact">
						{#each reportGoals as goal}
							<span class="sdg-chip" style={`--sdg-color: ${goal.color}; --sdg-text: ${goal.textColor};`}>
								{#if goal.icon}
									<img src={goal.icon} alt="" loading="lazy" />
								{:else}
									<span>{goal.code}</span>
								{/if}
								{goal.title}
							</span>
						{/each}
					</div>
				{/if}
				<p class="meta">
					{$t("impactPage.status")}: {report.status} · {$t("impactPage.publicAt")}:
					{fallbackDate(report.published_at, $locale)}
				</p>
			</div>
		{/each}
	</div>
{/if}
