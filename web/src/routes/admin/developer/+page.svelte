<script>
	import { locale, t } from "$lib/i18n.js";

	let { data } = $props();

	const releaseCommands = [
		"go run ./cmd/kelompok db migrate",
		"go run ./cmd/kelompok seed demo",
		"go run ./cmd/kelompok-api",
		"cd web && npm run build",
		"git tag -a v1.0-alpha.2 -m \"release: prepare 1.0-alpha.2\"",
	];

	const readinessCount = $derived(data.checks?.filter((check) => check.status === "pass").length ?? 0);
	const checksTotal = $derived(data.checks?.length ?? 0);
	const checkedAt = $derived(
		new Intl.DateTimeFormat($locale === "id" ? "id-ID" : "en-US", {
			dateStyle: "medium",
			timeStyle: "short",
		}).format(new Date(data.checkedAt)),
	);

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
			return $t("adminDeveloper.pass");
		}
		if (status === "warn") {
			return $t("adminDeveloper.warn");
		}
		return $t("adminDeveloper.fail");
	}
</script>

<nav class="breadcrumbs">
	<a href="/admin">{$t("nav.admin")}</a>
	<span>›</span>
	<span>{$t("adminDeveloper.title")}</span>
</nav>

<section class="section">
	<div class="page-heading">
		<p class="eyebrow">{$t("adminDeveloper.eyebrow")}</p>
		<h1 class="section-title">{$t("adminDeveloper.title")}</h1>
		<p class="section-note">{$t("adminDeveloper.subtitle")}</p>
		<p class="admin-release-note">{$t("adminDeveloper.updatedAt", { date: checkedAt })}</p>
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("adminDeveloper.checksEyebrow")}</p>
			<h2 class="section-title">{$t("adminDeveloper.checksTitle")}</h2>
		</div>
		<span class={["admin-release-badge", readinessCount >= 3 ? "admin-status-pass" : "admin-status-warn"].join(" ")}>
			{readinessCount}/{checksTotal} {$t("adminDeveloper.ready")}
		</span>
	</div>

	<div class="admin-grid">
		{#each data.checks as check}
			<article class="admin-card">
				<div class="admin-card-head">
					<p class="label">{$t(check.label)}</p>
					<span class={["admin-status", statusStyle(check.status)].join(" ")}>{statusText(check.status)}</span>
				</div>
				<p>{check.detail}</p>
			</article>
		{/each}
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("adminDeveloper.commandsEyebrow")}</p>
			<h2 class="section-title">{$t("adminDeveloper.commandsTitle")}</h2>
		</div>
		<p class="section-note">{$t("adminDeveloper.commandsNote")}</p>
	</div>

	<div class="admin-command-list">
		{#each releaseCommands as command}
			<pre class="admin-code">{command}</pre>
		{/each}
	</div>
</section>
