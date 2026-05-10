<script>
	import { fallbackDate } from "../../../../lib/api.js";

	let { data } = $props();
	const org = () => data.organization;
</script>

<nav class="breadcrumbs">
	<a href="/">Home</a>
	<span>›</span>
	<a href="/organizations">Organizations</a>
	<span>›</span>
	<a href={`/organizations/${org().slug}`}>{org().name}</a>
	<span>›</span>
	<span>Impact Reports</span>
</nav>

<h1 class="section-title">Laporan Dampak</h1>
<p class="muted">{org().name}</p>

{#if data.impactReports.length === 0}
	<p class="empty">Belum ada laporan publik.</p>
{:else}
	<div>
		{#each data.impactReports as report}
			<div class="card list-item">
				<h3 class="title">{report.title}</h3>
				<p class="small muted">Periode:
					{report.report_period_start ? fallbackDate(report.report_period_start) : "—"} →
					{report.report_period_end ? fallbackDate(report.report_period_end) : "—"}
				</p>
				<p>{report.summary || "Tidak ada ringkasan."}</p>
				<p class="meta">Status: {report.status} · Publik: {fallbackDate(report.published_at)}</p>
			</div>
		{/each}
	</div>
{/if}
