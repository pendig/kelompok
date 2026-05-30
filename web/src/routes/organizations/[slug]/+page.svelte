<script>
	import { fallbackDate } from "../../../lib/api.js";
	import StatusBadge from "$lib/components/StatusBadge.svelte";
	import { locale, t } from "$lib/i18n.js";
	import { normalizeSdgGoals } from "$lib/sdgs.js";
	import { getTheme, getInitials } from "../../../lib/theme.js";

	let { data, form } = $props();

	let org = $derived(data.organization);
	let profile = $derived(org?.profile_data || {});
	let sdgs = $derived(org?.sdgs_data || {});
	let visibleSdgGoals = $derived(normalizeSdgGoals(sdgs.primary || [], $locale));
	let relationships = $derived(org?.relationships || { parents: [], children: [], related: [] });
	let claimSubmitted = $derived(form?.ok && form?.action === "submitClaim");
	let submittedClaim = $derived(claimSubmitted ? form?.item : null);
	let claimError = $derived(!form?.ok && form?.action === "submitClaim" ? form.error : "");
	let claimErrorCode = $derived(!form?.ok && form?.action === "submitClaim" ? form.errorCode : "");
	let sessionUser = $derived(data?.session?.user ?? null);
	let sessionClaims = $derived(data?.session?.organization_claims ?? []);
	let pendingClaim = $derived(
		sessionClaims.find((claim) => claim.organization_slug === org.slug && claim.status === "pending"),
	);
	let requesterEmailDefault = $derived(sessionUser?.email ?? "");

	// svelte-ignore state_referenced_locally
	let activeTab = $state(form?.action === "submitClaim" ? "claim" : "profile");
	let theme = $derived(getTheme(org.name));

	$effect.pre(() => {
		if (form?.action === "submitClaim") {
			activeTab = "claim";
		}
	});

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

	function claimErrorLabel() {
		if (claimErrorCode) {
			return $t(`organizationDetail.claimErrors.${claimErrorCode}`);
		}
		return claimError;
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
	<!-- Modern Profile Cover Banner (Deterministic Gradient) -->
	<div class="profile-cover" style="background: {theme.cover};"></div>

	<!-- Modern Profile Header Actions Row -->
	<div class="profile-header-row">
		<div class="profile-avatar" style="color: {theme.avatarText}; background: {theme.avatarBg}; flex-shrink: 0; box-shadow: var(--shadow-lg);">
			{getInitials(org.name)}
		</div>
		<div class="inline-actions" style="margin-bottom: 8px;">
			{#if org.website_url}
				<a href={org.website_url} target="_blank" rel="noreferrer" class="btn secondary" style="min-height: 38px; padding-inline: 16px; font-weight: 700;">
					{$t("organizationDetail.website")}
				</a>
			{/if}
			{#if org.claim_status !== "claimed"}
				{#if sessionUser}
					<button onclick={() => activeTab = "claim"} class="btn primary" style="min-height: 38px; padding-inline: 16px; font-weight: 700;">
						{$t("organizationDetail.claimTitle")}
					</button>
				{:else}
					<a href={`/login?next=${encodeURIComponent(organizationPath())}`} class="btn primary" style="min-height: 38px; padding-inline: 16px; font-weight: 700;">
						{$t("organizationDetail.claimLoginAction")}
					</a>
				{/if}
			{/if}
		</div>
	</div>

	<!-- Modern Profile Details Block (Completely on White Background, Slug Removed) -->
	<div class="profile-info-block" style="margin-top: 16px; display: flex; flex-direction: column; gap: 8px;">
		<h1 style="margin: 0; font-size: 32px; font-weight: 800; color: var(--text); line-height: 1.1;">{org.name}</h1>
		{#if org.legal_name}
			<p style="margin: 0; font-size: 14.5px; color: var(--muted); font-weight: 500;">{org.legal_name}</p>
		{/if}
		
		<div class="profile-meta-row" style="margin-top: 4px;">
			<span class="profile-meta-badge">
				<strong>{$t("organizationDetail.location")}:</strong> {formatLocation()}
			</span>
				<span class="profile-meta-badge">
					<strong>{$t("organizationDetail.claim")}:</strong>
					<StatusBadge status={org.claim_status} />
				</span>
				{#if visibleSdgGoals.length}
					<span class="profile-meta-badge sdg-meta-badge">
						<strong>SDGs:</strong>
						<span class="profile-sdg-strip" aria-label="SDG focus">
							{#each visibleSdgGoals.slice(0, 5) as goal}
								{#if goal.icon}
									<img src={goal.icon} alt={`SDG ${goal.number}: ${goal.title}`} loading="lazy" />
								{:else}
									<span>{goal.code}</span>
								{/if}
							{/each}
							{#if visibleSdgGoals.length > 5}
								<small>+{visibleSdgGoals.length - 5}</small>
							{/if}
						</span>
					</span>
				{/if}
			</div>
		</div>

	<!-- Tab Switcher Navigation -->
	<nav class="profile-tabs-nav">
		<button class="profile-tab-trigger" class:active={activeTab === 'profile'} onclick={() => activeTab = 'profile'}>
			{$t("organizationDetail.tabProfile")}
		</button>
		<button class="profile-tab-trigger" class:active={activeTab === 'relationships'} onclick={() => activeTab = 'relationships'}>
			{$t("organizationDetail.tabRelationships")}
		</button>
		<button class="profile-tab-trigger" class:active={activeTab === 'sdg'} onclick={() => activeTab = 'sdg'}>
			{$t("organizationDetail.tabSdgs")}
		</button>
			<button class="profile-tab-trigger" class:active={activeTab === 'posts'} onclick={() => activeTab = 'posts'}>
				{$t("organizationDetail.recentPosts")}
			</button>
			<button class="profile-tab-trigger" class:active={activeTab === 'impact'} onclick={() => activeTab = 'impact'}>
				{$t("organizationDetail.impactReports")}
			</button>
		{#if org.claim_status !== "claimed" && sessionUser}
			<button class="profile-tab-trigger" class:active={activeTab === 'claim'} onclick={() => activeTab = 'claim'}>
				{$t("organizationDetail.tabClaim")}
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
					<h3 class="section-title" style="margin-top: 0; font-size: 20px;">{$t("organizationDetail.sdgsFocus")}</h3>
					<div class="label" style="margin-top: 12px">{$t("organizationDetail.focus")}</div>
					{#if visibleSdgGoals.length}
							<div class="sdg-grid sdg-grid-icons" style="margin-top: 12px">
							{#each visibleSdgGoals as goal}
								<article
									class="sdg-card"
									style={`--sdg-color: ${goal.color}; --sdg-text: ${goal.textColor};`}
									aria-label={`SDG ${goal.number || goal.raw}: ${goal.title}`}
								>
									{#if goal.icon}
										<img class="sdg-card-icon" src={goal.icon} alt={`SDG ${goal.number}: ${goal.title}`} loading="lazy" />
									{:else}
										<span class="sdg-card-number">{goal.code}</span>
										<span class="sdg-card-label">{goal.title}</span>
									{/if}
								</article>
							{/each}
						</div>
					{:else}
						<p class="small" style="margin-top: 10px">{$t("organizationDetail.noSdgs")}</p>
					{/if}
				</div>
				
				<div class="card">
					<h3 class="section-title" style="margin-top: 0; font-size: 20px;">{$t("organizationDetail.programs")}</h3>
					<div class="label" style="margin-top: 12px">{$t("organizationDetail.programWork")}</div>
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

		{#if activeTab === 'posts'}
			<div class="tab-panel">
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
			</div>
		{/if}

		{#if activeTab === 'impact'}
			<div class="tab-panel">
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
		{/if}

	{#if activeTab === 'claim' && org.claim_status !== 'claimed'}
		<div class="tab-panel">
			<div class="claim-card" style="margin-top: 0;">
				<div class="claim-copy">
					<p class="eyebrow">{$t("organizationDetail.claimEyebrow")}</p>
					<h2>{$t("organizationDetail.claimTitle")}</h2>
					<p>{$t("organizationDetail.claimDescription")}</p>
					{#if claimSubmitted}
						<div class="claim-submitted-card" role="status" aria-live="polite">
							<p class="success" style="margin: 0;">
								{$t("organizationDetail.claimSubmitted")}
							</p>
							{#if submittedClaim?.id}
								<p class="muted small claim-submitted-id">
									{$t("organizationDetail.claimSubmittedClaimId", { id: submittedClaim.id })}
								</p>
							{/if}
							<p class="muted small">
								{$t("organizationDetail.claimSubmittedReviewExpectation")}
							</p>
							{#if submittedClaim?.id}
								<a class="ghost-button" href={`/account?claim=${encodeURIComponent(submittedClaim.id)}`}>
									{$t("organizationDetail.claimSubmittedAccountLink")}
								</a>
							{:else}
								<a class="ghost-button" href="/account">
									{$t("organizationDetail.claimSubmittedAccountLink")}
								</a>
							{/if}
						</div>
					{/if}
					{#if claimError}
						<p class="error compact" style="margin-top: 16px;">{$t("organizationDetail.claimError", { message: claimErrorLabel() })}</p>
					{/if}
				</div>
				{#if !sessionUser}
					<div class="claim-form claim-state-card">
						<h3>{$t("organizationDetail.claimLoginTitle")}</h3>
						<p class="muted small">{$t("organizationDetail.claimLoginBody")}</p>
						<div class="inline-actions">
							<a class="btn primary" href={`/login?next=${encodeURIComponent(organizationPath())}`}>
								{$t("organizationDetail.claimLoginAction")}
							</a>
							<a class="ghost-button" href={`/register?next=${encodeURIComponent(organizationPath())}`}>
								{$t("organizationDetail.claimRegisterAction")}
							</a>
						</div>
					</div>
				{:else if pendingClaim}
					<div class="claim-form claim-state-card" role="status" aria-live="polite">
						<p class="eyebrow">{$t("organizationDetail.claimPendingEyebrow")}</p>
						<h3>{$t("organizationDetail.claimPendingTitle")}</h3>
						<p class="muted small">{$t("organizationDetail.claimPendingBody")}</p>
						<dl class="claim-mini-meta">
							<div>
								<dt>{$t("organizationDetail.claimIdLabel")}</dt>
								<dd><code>{pendingClaim.id}</code></dd>
							</div>
							<div>
								<dt>{$t("organizationDetail.claimMethod")}</dt>
								<dd>{pendingClaim.method}</dd>
							</div>
						</dl>
						<a class="ghost-button" href={`/account?claim=${encodeURIComponent(pendingClaim.id)}`}>
							{$t("organizationDetail.claimSubmittedAccountLink")}
						</a>
					</div>
				{:else}
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
							<input
								name="requester_email"
								type="email"
								value={requesterEmailDefault}
								placeholder="you@example.org"
								readonly
								required
							/>
						</label>
						<label>
							{$t("organizationDetail.claimEvidence")}
							<textarea name="evidence_note" rows="3" placeholder={$t("organizationDetail.claimEvidencePlaceholder")}></textarea>
						</label>
						<button class="btn primary" type="submit">{$t("organizationDetail.claimSubmit")}</button>
					</form>
				{/if}
			</div>
		</div>
	{/if}
</div>


<style>
	.claim-submitted-card {
		display: grid;
		gap: 8px;
		margin-top: 16px;
		padding: 14px 16px;
		border: 1px solid hsl(142, 70%, 85%);
		border-radius: 12px;
		background: hsl(142, 70%, 98%);
	}

	.claim-submitted-card .ghost-button {
		justify-self: start;
		margin-top: 4px;
	}

	.claim-submitted-id {
		font-family:
			ui-monospace,
			SFMono-Regular,
			"SF Mono",
			Consolas,
			"Liberation Mono",
			Menlo,
			monospace;
		word-break: break-all;
	}

	.muted.small {
		font-size: 12.5px;
	}

	.claim-state-card {
		align-content: start;
	}

	.claim-mini-meta {
		display: grid;
		gap: 10px;
		margin: 14px 0;
	}

	.claim-mini-meta dt {
		color: var(--muted);
		font-size: 12px;
		font-weight: 700;
		text-transform: uppercase;
	}

	.claim-mini-meta dd {
		margin: 2px 0 0;
	}
</style>
