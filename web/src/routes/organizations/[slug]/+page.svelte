<script>
	import { fallbackDate } from "../../../lib/api.js";
	import { locale, t } from "$lib/i18n.js";

	let { data, form } = $props();

	let org = $derived(data.organization);
	let profile = $derived(org?.profile_data || {});
	let sdgs = $derived(org?.sdgs_data || {});
	let relationships = $derived(org?.relationships || { parents: [], children: [], related: [] });
	let claimSubmitted = $derived(form?.ok && form?.action === "submitClaim");
	let claimError = $derived(!form?.ok && form?.action === "submitClaim" ? form.error : "");

	let activeTab = $state("profile");

	function organizationPath(path = "") {
		return `/organizations/${encodeURIComponent(org.slug)}${path}`;
	}

	function postPath(post) {
		return `${organizationPath("/posts")}/${encodeURIComponent(post.slug)}`;
	}

	function relationshipPath(item) {
		return `/organizations/${encodeURIComponent(item.organization.slug)}`;
	}

	function formatLocation() {
		const parts = [org.city, org.region, org.country].filter(Boolean);
		return parts.length ? parts.join(", ") : $t("organizationDetail.unknownLocation");
	}

	function getInitials(name) {
		if (!name) return "O";
		return name
			.split(" ")
			.map((word) => word[0])
			.slice(0, 2)
			.join("")
			.toUpperCase();
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

<div class="page" style="margin-top: 1.5rem">
	<!-- Modern Profile Cover Banner -->
	<div class="profile-cover"></div>

	<!-- Modern Profile Details Block -->
	<div class="profile-header-container">
		<div class="profile-avatar">
			{getInitials(org.name)}
		</div>
		<div class="profile-header-details">
			<div class="profile-title-row">
				<div>
					<h1 class="profile-title">{org.name}</h1>
					{#if org.legal_name}
						<p class="profile-tagline">{org.legal_name}</p>
					{/if}
				</div>
				<div class="inline-actions">
					{#if org.website_url}
						<a href={org.website_url} target="_blank" rel="noreferrer" class="btn secondary" style="min-height: 38px; padding-inline: 16px; font-weight: 700;">
							{$t("organizationDetail.website")}
						</a>
					{/if}
					{#if org.claim_status !== "claimed"}
						<button onclick={() => activeTab = "claim"} class="btn primary" style="min-height: 38px; padding-inline: 16px; font-weight: 700;">
							{$t("organizationDetail.claimTitle")}
						</button>
					{/if}
				</div>
			</div>
			
			<div class="profile-meta-row">
				<span class="profile-meta-badge">
					<strong>📍 {$t("organizationDetail.location")}:</strong> {formatLocation()}
				</span>
				<span class="profile-meta-badge">
					<strong>🏷️ Slug:</strong> <span class="code">{org.slug}</span>
				</span>
				<span class="profile-meta-badge">
					<strong>🛡️ {$t("organizationDetail.claim")}:</strong>
					<span class="admin-status {org.claim_status === 'claimed' ? 'admin-status-pass' : 'admin-status-warn'}">
						{org.claim_status === 'claimed' ? 'claimed' : 'unclaimed'}
					</span>
				</span>
			</div>
		</div>
	</div>

	<!-- Tab Switcher Navigation -->
	<nav class="profile-tabs-nav">
		<button class="profile-tab-trigger" class:active={activeTab === 'profile'} onclick={() => activeTab = 'profile'}>
			Profil
		</button>
		<button class="profile-tab-trigger" class:active={activeTab === 'relationships'} onclick={() => activeTab = 'relationships'}>
			Relasi
		</button>
		<button class="profile-tab-trigger" class:active={activeTab === 'sdg'} onclick={() => activeTab = 'sdg'}>
			SDGs & Program
		</button>
		<button class="profile-tab-trigger" class:active={activeTab === 'content'} onclick={() => activeTab = 'content'}>
			Artikel & Laporan
		</button>
		{#if org.claim_status !== "claimed"}
			<button class="profile-tab-trigger" class:active={activeTab === 'claim'} onclick={() => activeTab = 'claim'}>
				Klaim
			</button>
		{/if}
	</nav>

	<!-- Conditional Tab Panels -->
	{#if activeTab === 'profile'}
		<div class="tab-panel">
			<div class="profile-info-grid">
				<div class="card">
					<h3 class="section-title" style="margin-top: 0; font-size: 20px;">{$t("organizationDetail.vision")}</h3>
					<p style="font-size: 15px; line-height: 1.7; margin-top: 10px;">{org.description || $t("organizationDetail.noDescription")}</p>
					
					<h3 class="section-title" style="margin-top: 32px; font-size: 20px;">{$t("organizationDetail.history")}</h3>
					<p class="small" style="line-height: 1.7; margin-top: 10px;">{org.history || $t("organizationDetail.noHistory")}</p>
				</div>
				
				<div style="display: grid; gap: 16px; align-content: start;">
					{#if profile.languages?.length}
						<div class="card" style="padding: 18px;">
							<div class="label">{$t("organizationDetail.languages")}</div>
							<div class="pill-row" style="margin-top: 8px">
								{#each profile.languages as language}
									<span class="pill">{language}</span>
								{/each}
							</div>
						</div>
					{/if}
					
					<div class="card" style="padding: 18px;">
						<div class="label">{$t("organizationDetail.publicContact")}</div>
						{#if contactItems().length}
							<ul class="detail-list" style="margin-top: 8px">
								{#each contactItems() as item}
									<li><strong>{item.label}:</strong> {item.value}</li>
								{/each}
							</ul>
						{:else}
							<p class="small" style="margin-top: 8px;">{$t("organizationDetail.noContact")}</p>
						{/if}
					</div>

					<div class="card" style="padding: 14px 18px;">
						<span class="muted" style="font-size: 12px; font-weight: 500;">
							{$t("organizationDetail.updatedAt", { date: fallbackDate(org.updated_at, $locale) })}
						</span>
					</div>
				</div>
			</div>
		</div>
	{/if}

	{#if activeTab === 'relationships'}
		<div class="tab-panel">
			<div class="mini-head">
				<h2 class="section-title" style="margin: 0; font-size: 22px;">{$t("organizationDetail.relationships")}</h2>
			</div>
			{#if relationships.parents.length === 0 && relationships.children.length === 0 && relationships.related.length === 0}
				<p class="empty">{$t("organizationDetail.noRelationships")}</p>
			{:else}
				<div class="grid">
					{#if relationships.parents.length}
						<div class="card">
							<div class="label">{$t("organizationDetail.parentOrganizations")}</div>
							<ul class="detail-list" style="margin-top: 12px">
								{#each relationships.parents as item}
									<li>
										<a href={relationshipPath(item)}>{item.organization.name}</a>
										<span class="muted"> · {item.label || item.relationship_type}</span>
									</li>
								{/each}
							</ul>
						</div>
					{/if}
					{#if relationships.children.length}
						<div class="card">
							<div class="label">{$t("organizationDetail.childOrganizations")}</div>
							<ul class="detail-list" style="margin-top: 12px">
								{#each relationships.children as item}
									<li>
										<a href={relationshipPath(item)}>{item.organization.name}</a>
										<span class="muted"> · {item.label || item.relationship_type}</span>
									</li>
								{/each}
							</ul>
						</div>
					{/if}
					{#if relationships.related.length}
						<div class="card">
							<div class="label">{$t("organizationDetail.relatedOrganizations")}</div>
							<ul class="detail-list" style="margin-top: 12px">
								{#each relationships.related as item}
									<li>
										<a href={relationshipPath(item)}>{item.organization.name}</a>
										<span class="muted"> · {item.label || item.relationship_type}</span>
									</li>
								{/each}
							</ul>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	{/if}

	{#if activeTab === 'sdg'}
		<div class="tab-panel">
			<div class="admin-panel-grid">
				<div class="card">
					<h3 class="section-title" style="margin-top: 0; font-size: 20px;">SDGs Focus</h3>
					<div class="label" style="margin-top: 12px">{$t("organizationDetail.focus")}</div>
					{#if sdgs.primary?.length}
						<div class="pill-row" style="margin-top: 10px">
							{#each sdgs.primary as goal}
								<span class="pill">{goal}</span>
							{/each}
						</div>
					{:else}
						<p class="small" style="margin-top: 10px">{$t("organizationDetail.noSdgs")}</p>
					{/if}
				</div>
				
				<div class="card">
					<h3 class="section-title" style="margin-top: 0; font-size: 20px;">{$t("organizationDetail.programs")}</h3>
					<div class="label" style="margin-top: 12px">Program Kerja</div>
					{#if profile.programs?.length}
						<ul style="margin-top: 12px; padding-left: 20px; line-height: 1.8;">
							{#each profile.programs as item}
								<li>{item}</li>
							{/each}
						</ul>
					{:else}
						<p class="small" style="margin-top: 12px">{$t("organizationDetail.noPrograms")}</p>
					{/if}
				</div>
			</div>
		</div>
	{/if}

	{#if activeTab === 'content'}
		<div class="tab-panel">
			<div class="admin-panel-grid">
				<div class="card">
					<div class="mini-head" style="border-bottom: 1px solid var(--border); padding-bottom: 12px; margin-bottom: 16px;">
						<h3 class="section-title" style="margin: 0; font-size: 20px;">{$t("organizationDetail.recentPosts")}</h3>
						<a href={organizationPath("/posts")} class="ghost-button">{$t("organizationDetail.allPosts")}</a>
					</div>
					{#if data.posts.length === 0}
						<p class="empty">{$t("organizationDetail.noPosts")}</p>
					{:else}
						<div style="display: grid; gap: 12px;">
							{#each data.posts.slice(0, 6) as post}
								<div class="list-item" style="padding: 12px; margin-bottom: 0;">
									<a class="title" style="font-size: 15px;" href={postPath(post)}>{post.title}</a>
									<div class="meta" style="margin-top: 4px;">
										{fallbackDate(post.published_at, $locale)} · {post.summary || "—"}
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<div class="card">
					<div class="mini-head" style="border-bottom: 1px solid var(--border); padding-bottom: 12px; margin-bottom: 16px;">
						<h3 class="section-title" style="margin: 0; font-size: 20px;">{$t("organizationDetail.impactReports")}</h3>
						<a href={organizationPath("/impact")} class="ghost-button">{$t("organizationDetail.impactReports")}</a>
					</div>
					{#if data.impactReports.length === 0}
						<p class="empty">{$t("organizationDetail.noReports")}</p>
					{:else}
						<div style="display: grid; gap: 12px;">
							{#each data.impactReports.slice(0, 6) as report}
								<div class="list-item" style="padding: 12px; margin-bottom: 0;">
									<div class="title" style="font-size: 15px; font-weight: 700;">{report.title}</div>
									<div class="meta" style="margin-top: 4px;">{fallbackDate(report.published_at, $locale)} · {report.summary || "—"}</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}

	{#if activeTab === 'claim' && org.claim_status !== 'claimed'}
		<div class="tab-panel">
			<div class="claim-card" style="margin-top: 0;">
				<div class="claim-copy">
					<p class="eyebrow">{$t("organizationDetail.claimEyebrow")}</p>
					<h2>{$t("organizationDetail.claimTitle")}</h2>
					<p>{$t("organizationDetail.claimDescription")}</p>
					{#if claimSubmitted}
						<p class="success" style="margin-top: 16px;">{$t("organizationDetail.claimSubmitted")}</p>
					{/if}
					{#if claimError}
						<p class="error compact" style="margin-top: 16px;">{$t("organizationDetail.claimError", { message: claimError })}</p>
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
			</div>
		</div>
	{/if}
</div>
