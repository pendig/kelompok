<script>
	import { locale, t } from "$lib/i18n.js";

	let { data } = $props();

	function statusStyle(status) {
		switch (status) {
			case "pass":
				return "admin-status-pass";
			case "warn":
				return "admin-status-warn";
			default:
				return "admin-status-fail";
		}
	}

	function statusText(status) {
		if (status === "pass") {
			return $t("admin.pass");
		}
		if (status === "warn") {
			return $t("admin.checksWarn");
		}
		return $t("admin.checksFail");
	}

	const releaseCommands = [
		"go run ./cmd/kelompok db migrate",
		"go run ./cmd/kelompok seed demo",
		"go run ./cmd/kelompok-api",
		"cd web && npm run build",
		"git tag -a v1.0-alpha.1 -m \"release: prepare 1.0-alpha.1\"",
	];

	const checksTotal = $derived(data.checks?.length ?? 0);
	const orgSummary = $derived(data.organizations || []);
	const postSummary = $derived(data.posts || []);
	const readinessCount = $derived(data.checks?.filter((check) => check.status === "pass").length ?? 0);
	const checkedAt = $derived(
		new Intl.DateTimeFormat($locale === "id" ? "id-ID" : "en-US", {
		dateStyle: "medium",
		timeStyle: "short",
		}).format(new Date()),
	);
</script>

<section class="section">
	<div class="page-heading">
		<p class="eyebrow">{$t("admin.goal")} · {$t("admin.releaseCandidate")}</p>
		<h1 class="section-title">{$t("admin.releaseTitle")}</h1>
		<p class="section-note">
			{$t("admin.releaseSubtitle")}
			· {$t("admin.updatedAt", { date: checkedAt })}
		</p>
		<p class="admin-release-note">{$t("admin.releaseNote")}</p>
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("admin.checks")}</p>
			<h2 class="section-title">{$t("admin.checkSection")}</h2>
		</div>
		<span class={ ["admin-release-badge", readinessCount >= 3 ? "admin-status-pass" : "admin-status-warn"].join(" ") }>
			{readinessCount}/{checksTotal} {$t("admin.releaseTitle")}
		</span>
	</div>

	{#if data.loadErrors.length}
		<p class="error">
			<strong>{$t("admin.checksFail")}: </strong>
			{data.loadErrors[0]}
		</p>
	{/if}

	<div class="admin-grid">
		{#each data.checks as check}
			<article class="admin-card">
				<div class="admin-card-head">
					<p class="label">{$t(check.label)}</p>
					<span class={ ["admin-status", statusStyle(check.status)].join(" ") }>{statusText(check.status)}</span>
				</div>
				<p>{check.detail}</p>
			</article>
		{/each}
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("admin.dataSection")}</p>
			<h2 class="section-title">{$t("admin.publicDataQuality")}</h2>
		</div>
	</div>
	<div class="admin-grid">
		<article class="card">
			<p class="label">{$t("admin.totalOrganizations")}</p>
			<p class="admin-number">{orgSummary.length}</p>
			<p class="small">{$t("admin.checkHasOrganizations")}</p>
		</article>
		<article class="card">
			<p class="label">{$t("admin.totalPosts")}</p>
			<p class="admin-number">{postSummary.length}</p>
			<p class="small">{$t("admin.checkHasPosts")}</p>
		</article>
		<article class="card">
			<p class="label">{$t("admin.totalImpactPreview")}</p>
			<p class="admin-number">{data.impactByOrganization.reduce((total, item) => total + (item.count || 0), 0)}</p>
			<p class="small">{$t("admin.checkHasImpact")}</p>
		</article>
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("admin.impactByOrgTitle")}</p>
			<h2 class="section-title">{$t("admin.impactByOrgTitle")}</h2>
		</div>
	</div>
	<div class="surface-card admin-list">
		{#if data.impactByOrganization.length === 0}
			<p class="empty">{$t("admin.noImpact")}</p>
		{:else}
			{#each data.impactByOrganization as item}
				<div class="admin-list-item">
					<div class="admin-list-item__meta">
						<p class="label">{item.orgName}</p>
						<p class="muted small">{item.count || 0} item{item.count === 1 ? "" : "s"}</p>
					</div>
					<p class="small">{item.error ? item.error : item.latest || $t("admin.noImpact")}</p>
				</div>
			{/each}
		{/if}
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("admin.commandSection")}</p>
			<h2 class="section-title">{$t("admin.cliCommandTitle")}</h2>
		</div>
		<p class="section-note">{$t("admin.cliCommandNotes")}</p>
	</div>
	<div class="admin-command-list">
		{#each releaseCommands as command}
			<pre class="admin-code">{command}</pre>
		{/each}
	</div>
	<p class="small">{$t("admin.cliHelp")}</p>
</section>
