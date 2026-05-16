<script>
	import { fallbackDate } from "../../../lib/api.js";
	import { locale, t } from "$lib/i18n.js";

	let { data, form } = $props();

	let org = $derived(data.organization);
	let profile = $derived(org?.profile_data || {});
	let sdgs = $derived(org?.sdgs_data || {});
	let claimSubmitted = $derived(form?.ok && form?.action === "submitClaim");
	let claimError = $derived(!form?.ok && form?.action === "submitClaim" ? form.error : "");

	function organizationPath(path = "") {
		return `/organizations/${encodeURIComponent(org.slug)}${path}`;
	}

	function postPath(post) {
		return `${organizationPath("/posts")}/${encodeURIComponent(post.slug)}`;
	}

	function formatLocation() {
		const parts = [org.city, org.region, org.country].filter(Boolean);
		return parts.length ? parts.join(", ") : $t("organizationDetail.unknownLocation");
	}

	function contactItems() {
		const value = profile.public_contact;
		if (!value) {
			return [];
		}

		if (typeof value === "string") {
			return [{ label: $t("organizationDetail.publicContact"), value }];
		}

		if (Array.isArray(value)) {
			return value
				.map((item) => {
					if (typeof item === "string") {
						return { label: $t("organizationDetail.publicContact"), value: item };
					}
					return {
						label: item.label || item.type || $t("organizationDetail.publicContact"),
						value: item.value || item.url || item.email || item.phone || "",
					};
				})
				.filter((item) => item.value);
		}

		return Object.entries(value)
			.map(([label, item]) => {
				if (typeof item === "string") {
					return { label, value: item };
				}
				return { label, value: item?.value || item?.url || item?.email || item?.phone || "" };
			})
			.filter((item) => item.value);
	}

	function firstContactValue(kind) {
		const contact = profile.public_contact;

		if (!contact) {
			return "";
		}

		if (!kind && typeof contact === "string") {
			return contact;
		}

		if (!Array.isArray(contact) && typeof contact === "object") {
			const direct = contact[kind];
			if (typeof direct === "string") {
				return direct;
			}
			if (direct && typeof direct === "object") {
				return direct.value || direct.url || direct.email || direct.phone || "";
			}
		}

		if (Array.isArray(contact)) {
			const match = contact.find((item) => {
				if (typeof item === "string") {
					return !kind || item.includes("@");
				}
				return item.type === kind || item.label === kind || item[kind];
			});

			if (typeof match === "string") {
				return match;
			}
			return match?.value || match?.url || match?.email || match?.phone || "";
		}

		return "";
	}

	function claimTargetDefault() {
		return firstContactValue("email") || firstContactValue("");
	}
</script>

<nav class="breadcrumbs">
	<a href="/">{$t("organizationDetail.home")}</a>
	<span>›</span>
	<a href="/organizations">{$t("nav.organizations")}</a>
	<span>›</span>
	<span>{org.name}</span>
</nav>

<header class="two-col" style="margin-top: 1rem">
	<section class="card">
		<h1>{org.name}</h1>
		<p class="muted">Slug: <span class="code">{org.slug}</span></p>
		<p>{org.description || $t("organizationDetail.noDescription")}</p>
		<p class="small muted">{$t("organizationDetail.tagline")}: {org.legal_name || "—"}</p>
		<p class="small">
			{$t("organizationDetail.updatedAt", { date: fallbackDate(org.updated_at, $locale) })}
		</p>
	</section>

	<section class="card">
		<div class="label">{$t("organizationDetail.info")}</div>
		<p class="value"><strong>{$t("organizationDetail.location")}:</strong> {formatLocation()}</p>
		<p class="value">
			<strong>{$t("organizationDetail.website")}:</strong>
			{#if org.website_url}
				<a href={org.website_url} target="_blank" rel="noreferrer">{org.website_url}</a>
			{:else}
				—
			{/if}
		</p>
		<p class="value"><strong>{$t("organizationDetail.claim")}:</strong> {org.claim_status}</p>
		{#if profile.languages?.length}
			<div class="label">{$t("organizationDetail.languages")}</div>
			<div class="pill-row">
				{#each profile.languages as language}
					<span class="pill">{language}</span>
				{/each}
			</div>
		{/if}
	</section>
</header>

{#if org.claim_status !== "claimed"}
	<section class="claim-card">
		<div class="claim-copy">
			<p class="eyebrow">{$t("organizationDetail.claimEyebrow")}</p>
			<h2>{$t("organizationDetail.claimTitle")}</h2>
			<p>{$t("organizationDetail.claimDescription")}</p>
			{#if claimSubmitted}
				<p class="success">{$t("organizationDetail.claimSubmitted")}</p>
			{/if}
			{#if claimError}
				<p class="error compact">{$t("organizationDetail.claimError", { message: claimError })}</p>
			{/if}
		</div>
		<form class="claim-form" method="POST" action="?/submitClaim">
			<label>
				{$t("organizationDetail.claimMethod")}
				<select name="method">
					<option value="official_email">{$t("organizationDetail.claimMethodEmail")}</option>
					<option value="instagram">{$t("organizationDetail.claimMethodInstagram")}</option>
				</select>
			</label>
			<label>
				{$t("organizationDetail.claimTarget")}
				<input name="target" value={claimTargetDefault()} placeholder="admin@example.org" required />
			</label>
			<label>
				{$t("organizationDetail.claimRequesterEmail")}
				<input name="requester_email" type="email" placeholder="you@example.org" required />
			</label>
			<label>
				{$t("organizationDetail.claimEvidence")}
				<textarea name="evidence_note" rows="3" placeholder={$t("organizationDetail.claimEvidencePlaceholder")}></textarea>
			</label>
			<button class="btn primary" type="submit">{$t("organizationDetail.claimSubmit")}</button>
		</form>
	</section>
{/if}

<section>
	<div class="actions">
		<a href={organizationPath("/posts")}>{$t("organizationDetail.allPosts")}</a>
		<a href={organizationPath("/impact")}>{$t("organizationDetail.impactReports")}</a>
	</div>

	<h2 class="section-title">{$t("organizationDetail.vision")}</h2>
	<div class="grid">
		<div class="card">
			<div class="label">{$t("organizationDetail.history")}</div>
			<p class="small">{org.history || $t("organizationDetail.noHistory")}</p>
		</div>
		<div class="card">
			<div class="label">{$t("organizationDetail.publicContact")}</div>
			{#if contactItems().length}
				<ul class="detail-list">
					{#each contactItems() as item}
						<li><strong>{item.label}:</strong> {item.value}</li>
					{/each}
				</ul>
			{:else}
				<p class="small">{$t("organizationDetail.noContact")}</p>
			{/if}
		</div>
	</div>
</section>

<section>
	<h2 class="section-title">SDGS</h2>
	<div class="grid">
		<div class="card">
			<div class="label">{$t("organizationDetail.focus")}</div>
			{#if sdgs.primary?.length}
				<div class="pill-row">
					{#each sdgs.primary as goal}
						<span class="pill">{goal}</span>
					{/each}
				</div>
			{:else}
				<p class="small">{$t("organizationDetail.noSdgs")}</p>
			{/if}
		</div>
		<div class="card">
			<div class="label">{$t("organizationDetail.programs")}</div>
			{#if profile.programs?.length}
				<ul>
					{#each profile.programs as item}
						<li>{item}</li>
					{/each}
				</ul>
			{:else}
				<p class="small">{$t("organizationDetail.noPrograms")}</p>
			{/if}
		</div>
	</div>
</section>

<section>
	<h2 class="section-title">{$t("organizationDetail.recentPosts")}</h2>
	{#if data.posts.length === 0}
		<p class="empty">{$t("organizationDetail.noPosts")}</p>
	{:else}
		{#each data.posts.slice(0, 6) as post}
			<div class="list-item">
				<a class="title" href={postPath(post)}>{post.title}</a>
				<div class="meta">
					{fallbackDate(post.published_at, $locale)} · {post.summary || "—"}
				</div>
			</div>
		{/each}
	{/if}
</section>

<section>
	<h2 class="section-title">{$t("organizationDetail.impactReports")}</h2>
	{#if data.impactReports.length === 0}
		<p class="empty">{$t("organizationDetail.noReports")}</p>
	{:else}
		{#each data.impactReports.slice(0, 6) as report}
			<div class="list-item">
				<div class="title">{report.title}</div>
				<div class="meta">{fallbackDate(report.published_at, $locale)} · {report.summary || "—"}</div>
			</div>
		{/each}
	{/if}
</section>
