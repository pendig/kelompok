<script>
	import { locale, t } from "$lib/i18n.js";

	let { data, form } = $props();

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

	function joinList(value) {
		if (Array.isArray(value)) {
			return value.join(", ");
		}
		if (typeof value === "string") {
			return value;
		}
		return "";
	}

	function contactValue(org, key) {
		const contact = org?.profile_data?.public_contact || {};
		return typeof contact === "object" && !Array.isArray(contact) ? contact[key] || "" : "";
	}

	function selectedPath(org) {
		return `/admin?org=${encodeURIComponent(org.slug)}`;
	}

	const releaseCommands = [
		"go run ./cmd/kelompok db migrate",
		"go run ./cmd/kelompok seed demo",
		"go run ./cmd/kelompok-api",
		"cd web && npm run build",
		"git tag -a v1.0-alpha.2 -m \"release: prepare 1.0-alpha.2\"",
	];
	const socialPlaceholder = '{"instagram":"https://instagram.com/example"}';
	const evidencePlaceholder = '{"note":"manual claim test"}';
	const metricsPlaceholder = '{"beneficiaries":120}';
	const metadataPlaceholder = '{"source":"manual"}';

	const checksTotal = $derived(data.checks?.length ?? 0);
	const orgSummary = $derived(data.organizations || []);
	const postSummary = $derived(data.posts || []);
	const impactSummary = $derived(data.impactReports || []);
	const relationshipSummary = $derived(data.relationships || []);
	const selectedOrg = $derived(data.selectedOrganization);
	const currentUser = $derived(data.session?.user);
	const selectedProfile = $derived(selectedOrg?.profile_data || {});
	const selectedSdgs = $derived(selectedOrg?.sdgs_data || {});
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
		<p class="eyebrow">{$t("admin.goal")} - {$t("admin.releaseCandidate")}</p>
		<h1 class="section-title">{$t("admin.releaseTitle")}</h1>
		<p class="section-note">
			{$t("admin.releaseSubtitle")}
			- {$t("admin.updatedAt", { date: checkedAt })}
		</p>
		<p class="admin-release-note">{$t("admin.releaseNote")}</p>
	</div>
</section>

{#if form?.error}
	<p class="error">
		<strong>{$t("admin.actionError")}: </strong>
		{form.error}
	</p>
{:else if form?.ok}
	<p class="notice">
		<strong>{$t("admin.actionSucceeded")}</strong>
		<span>{form.action}</span>
	</p>
{/if}

<section class="section">
	<div class="admin-panel session-panel">
		<div>
			<p class="eyebrow">{$t("admin.session")}</p>
			<h2>{data.isAuthenticated ? $t("admin.sessionActive") : $t("admin.loginTitle")}</h2>
			<p class="section-note">
				{#if currentUser}
					{currentUser.name} · {currentUser.email}
				{:else if data.isAuthenticated}
					{$t("admin.adminKeyMode")}
				{:else}
					{$t("admin.loginSubtitle")}
				{/if}
			</p>
		</div>
		{#if currentUser}
			<form method="POST" action="?/logout">
				<button class="ghost-button" type="submit">{$t("admin.logout")}</button>
			</form>
		{:else if !data.isAuthenticated}
			<form class="login-form" method="POST" action="?/login">
				<label>
					<span>{$t("admin.email")}</span>
					<input name="email" type="email" required autocomplete="email" />
				</label>
				<label>
					<span>{$t("admin.password")}</span>
					<input name="password" type="password" required autocomplete="current-password" />
				</label>
				<button class="btn primary" type="submit">{$t("admin.login")}</button>
			</form>
		{/if}
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("admin.checks")}</p>
			<h2 class="section-title">{$t("admin.checkSection")}</h2>
		</div>
		<span class={["admin-release-badge", readinessCount >= 3 ? "admin-status-pass" : "admin-status-warn"].join(" ")}>
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
			<p class="admin-number">{impactSummary.length}</p>
			<p class="small">{$t("admin.checkHasImpact")}</p>
		</article>
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("admin.operationsSection")}</p>
			<h2 class="section-title">{$t("admin.crmTitle")}</h2>
		</div>
		<p class="section-note">{$t("admin.crmSubtitle")}</p>
	</div>

	<div class="admin-workspace">
		<aside class="admin-sidebar">
			<div class="admin-panel compact">
				<p class="label">{$t("admin.selectedOrg")}</p>
				{#if orgSummary.length === 0}
					<p class="empty">{$t("admin.noOrganizations")}</p>
				{:else}
					<div class="admin-org-nav">
						{#each orgSummary as org}
							<a class:active={org.slug === data.selectedSlug} href={selectedPath(org)}>
								<span>{org.name}</span>
								<small>{org.claim_status}</small>
							</a>
						{/each}
					</div>
				{/if}
			</div>

			<form class="admin-panel compact" method="POST" action="?/createOrganization">
				<h3>{$t("admin.createOrg")}</h3>
				<label>
					<span>{$t("admin.name")}</span>
					<input name="name" required placeholder="Kelompok Nusantara" />
				</label>
				<label>
					<span>{$t("admin.slug")}</span>
					<input name="slug" placeholder="kelompok-nusantara" />
				</label>
				<label>
					<span>{$t("admin.officialEmail")}</span>
					<input name="official_email" type="email" placeholder="hello@example.org" />
				</label>
				<label>
					<span>{$t("admin.city")}</span>
					<input name="city" placeholder="Jakarta" />
				</label>
				<input name="claim_status" type="hidden" value="unclaimed" />
				<button class="btn primary" type="submit">{$t("admin.create")}</button>
			</form>
		</aside>

		<div class="admin-main">
			{#if selectedOrg}
				<form class="admin-panel" method="POST" action="?/updateOrganization">
					<div class="admin-form-head">
						<div>
							<p class="eyebrow">{$t("admin.orgForm")}</p>
							<h3>{selectedOrg.name}</h3>
						</div>
						<a class="ghost-button" href={`/organizations/${encodeURIComponent(selectedOrg.slug)}`} target="_blank" rel="noreferrer">
							{$t("admin.viewPublic")}
						</a>
					</div>
					<input name="current_slug" type="hidden" value={selectedOrg.slug} />
					<div class="admin-field-grid">
						<label>
							<span>{$t("admin.name")}</span>
							<input name="name" required value={selectedOrg.name || ""} />
						</label>
						<label>
							<span>{$t("admin.slug")}</span>
							<input name="slug" value={selectedOrg.slug || ""} />
						</label>
						<label>
							<span>{$t("admin.legalName")}</span>
							<input name="legal_name" value={selectedOrg.legal_name || ""} />
						</label>
						<label>
							<span>{$t("admin.officialEmail")}</span>
							<input name="official_email" type="email" value={selectedOrg.official_email || ""} />
						</label>
						<label>
							<span>{$t("admin.websiteUrl")}</span>
							<input name="website_url" type="url" value={selectedOrg.website_url || ""} />
						</label>
						<label>
							<span>{$t("admin.claimStatus")}</span>
							<select name="claim_status" value={selectedOrg.claim_status || "unclaimed"}>
								<option value="unclaimed">unclaimed</option>
								<option value="pending">pending</option>
								<option value="claimed">claimed</option>
								<option value="rejected">rejected</option>
							</select>
						</label>
						<label>
							<span>{$t("admin.country")}</span>
							<input name="country" value={selectedOrg.country || ""} />
						</label>
						<label>
							<span>{$t("admin.region")}</span>
							<input name="region" value={selectedOrg.region || ""} />
						</label>
						<label>
							<span>{$t("admin.city")}</span>
							<input name="city" value={selectedOrg.city || ""} />
						</label>
					</div>
					<div class="admin-field-grid two">
						<label>
							<span>{$t("admin.description")}</span>
							<textarea name="description" rows="4">{selectedOrg.description || ""}</textarea>
						</label>
						<label>
							<span>{$t("admin.history")}</span>
							<textarea name="history" rows="4">{selectedOrg.history || ""}</textarea>
						</label>
						<label>
							<span>{$t("admin.focus")}</span>
							<textarea name="focus" rows="3">{joinList(selectedProfile.focus)}</textarea>
						</label>
						<label>
							<span>{$t("admin.programs")}</span>
							<textarea name="programs" rows="3">{joinList(selectedProfile.programs)}</textarea>
						</label>
						<label>
							<span>{$t("admin.sdgs")}</span>
							<textarea name="sdgs" rows="3">{joinList(selectedSdgs.primary)}</textarea>
						</label>
						<label>
							<span>{$t("admin.languages")}</span>
							<textarea name="languages" rows="3">{joinList(selectedProfile.languages)}</textarea>
						</label>
					</div>
					<div class="admin-field-grid">
						<label>
							<span>{$t("admin.publicEmail")}</span>
							<input name="public_contact_email" type="email" value={contactValue(selectedOrg, "email")} />
						</label>
						<label>
							<span>Instagram</span>
							<input name="public_contact_instagram" value={contactValue(selectedOrg, "instagram")} />
						</label>
						<label>
							<span>{$t("admin.phone")}</span>
							<input name="public_contact_phone" value={contactValue(selectedOrg, "phone")} />
						</label>
					</div>
					<label>
						<span>{$t("admin.impactData")}</span>
						<textarea name="impact_data" rows="3">{JSON.stringify(selectedOrg.impact_data || {}, null, 2)}</textarea>
					</label>
					<div class="admin-actions">
						<button class="btn primary" type="submit">{$t("admin.update")}</button>
					</div>
					</form>

					<div class="admin-panel-grid">
						<form class="admin-panel" method="POST" action="?/createRelationship">
							<h3>{$t("admin.relationshipEditor")}</h3>
							<div class="admin-field-grid two">
								<label>
									<span>{$t("admin.parentOrganization")}</span>
									<input name="parent_organization_slug" required placeholder="muhammadiyah" />
								</label>
								<label>
									<span>{$t("admin.childOrganization")}</span>
									<input name="child_organization_slug" required value={selectedOrg.slug} />
								</label>
								<label>
									<span>{$t("admin.relationshipType")}</span>
									<select name="relationship_type">
										<option value="structural_parent">structural_parent</option>
										<option value="autonomous_body">autonomous_body</option>
										<option value="affiliated_with">affiliated_with</option>
										<option value="network_member">network_member</option>
										<option value="related">related</option>
									</select>
								</label>
								<label>
									<span>{$t("admin.relationshipStatus")}</span>
									<select name="status">
										<option value="active">active</option>
										<option value="pending">pending</option>
										<option value="inactive">inactive</option>
										<option value="archived">archived</option>
									</select>
								</label>
								<label>
									<span>{$t("admin.startedAt")}</span>
									<input name="started_at" type="date" />
								</label>
								<label>
									<span>{$t("admin.endedAt")}</span>
									<input name="ended_at" type="date" />
								</label>
							</div>
							<label>
								<span>{$t("admin.relationshipLabel")}</span>
								<input name="label" placeholder="Induk struktural" />
							</label>
							<label>
								<span>{$t("admin.metadata")}</span>
								<textarea name="metadata" rows="2" placeholder={metadataPlaceholder}></textarea>
							</label>
							<button class="btn primary" type="submit">{$t("admin.createRelationship")}</button>
						</form>

						<div class="admin-panel">
							<h3>{$t("admin.relationships")}</h3>
							{#if relationshipSummary.length === 0}
								<p class="empty">{$t("admin.noRelationships")}</p>
							{:else}
								<div class="admin-list compact-list">
									{#each relationshipSummary as relationship}
										<div class="admin-list-item">
											<div class="admin-list-item__meta">
												<p class="label">{relationship.parent.name} → {relationship.child.name}</p>
												<span class="mini-badge">{relationship.relationship_type}</span>
											</div>
											<p class="small">
												{relationship.parent.slug} → {relationship.child.slug}
												{#if relationship.label}
													· {relationship.label}
												{/if}
											</p>
											<div class="inline-actions">
												<span class="mini-badge">{relationship.status}</span>
												<form method="POST" action="?/deleteRelationship" class="inline-form">
													<input type="hidden" name="id" value={relationship.id} />
													<button class="ghost-button danger" type="submit">{$t("admin.remove")}</button>
												</form>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</div>
					</div>

					<div class="admin-panel-grid">
						<form class="admin-panel" method="POST" action="?/createMember">
							<h3>{$t("admin.addMember")}</h3>
						<input name="organization_slug" type="hidden" value={selectedOrg.slug} />
						<div class="admin-field-grid two">
							<label>
								<span>{$t("admin.name")}</span>
								<input name="name" required />
							</label>
							<label>
								<span>{$t("admin.position")}</span>
								<input name="position" />
							</label>
							<label>
								<span>{$t("admin.email")}</span>
								<input name="email" type="email" />
							</label>
							<label>
								<span>{$t("admin.phone")}</span>
								<input name="phone" />
							</label>
						</div>
						<label>
							<span>{$t("admin.bio")}</span>
							<textarea name="bio" rows="3"></textarea>
						</label>
						<label>
							<span>{$t("admin.socialLinks")}</span>
							<textarea name="social_links" rows="2" placeholder={socialPlaceholder}></textarea>
						</label>
						<button class="btn primary" type="submit">{$t("admin.create")}</button>
					</form>

					<form class="admin-panel" method="POST" action="?/createClaim">
						<h3>{$t("admin.claimRequest")}</h3>
						<input name="organization_slug" type="hidden" value={selectedOrg.slug} />
						<div class="admin-field-grid two">
							<label>
								<span>{$t("admin.method")}</span>
								<select name="method">
									<option value="official_email">official_email</option>
									<option value="instagram">instagram</option>
								</select>
							</label>
							<label>
								<span>{$t("admin.target")}</span>
								<input name="target" required value={selectedOrg.official_email || ""} />
							</label>
						</div>
						<label>
							<span>{$t("admin.requesterEmail")}</span>
							<input name="requester_email" type="email" value={selectedOrg.official_email || ""} />
						</label>
						<label>
							<span>{$t("admin.evidence")}</span>
							<textarea name="evidence" rows="2" placeholder={evidencePlaceholder}></textarea>
						</label>
						<button class="btn primary" type="submit">{$t("admin.submitClaim")}</button>
					</form>
				</div>

				<div class="admin-panel-grid">
					<form class="admin-panel" method="POST" action="?/createPost">
						<h3>{$t("admin.createPost")}</h3>
						<input name="organization_slug" type="hidden" value={selectedOrg.slug} />
						<div class="admin-field-grid two">
							<label>
								<span>{$t("admin.titleField")}</span>
								<input name="title" required />
							</label>
							<label>
								<span>{$t("admin.slug")}</span>
								<input name="slug" required />
							</label>
							<label>
								<span>{$t("admin.category")}</span>
								<input name="category_slug" placeholder="news" />
							</label>
							<label>
								<span>{$t("admin.status")}</span>
								<select name="status">
									<option value="draft">draft</option>
									<option value="published">published</option>
								</select>
							</label>
						</div>
						<label>
							<span>{$t("admin.summary")}</span>
							<textarea name="summary" rows="3"></textarea>
						</label>
						<label>
							<span>{$t("admin.content")}</span>
							<textarea name="content" rows="6"></textarea>
						</label>
						<button class="btn primary" type="submit">{$t("admin.create")}</button>
					</form>

					<form class="admin-panel" method="POST" action="?/createImpactReport">
						<h3>{$t("admin.createImpact")}</h3>
						<input name="organization_slug" type="hidden" value={selectedOrg.slug} />
						<div class="admin-field-grid two">
							<label>
								<span>{$t("admin.titleField")}</span>
								<input name="title" required />
							</label>
							<label>
								<span>{$t("admin.status")}</span>
								<select name="status">
									<option value="draft">draft</option>
									<option value="published">published</option>
								</select>
							</label>
							<label>
								<span>{$t("admin.periodStart")}</span>
								<input name="report_period_start" type="date" />
							</label>
							<label>
								<span>{$t("admin.periodEnd")}</span>
								<input name="report_period_end" type="date" />
							</label>
						</div>
						<label>
							<span>{$t("admin.summary")}</span>
							<textarea name="summary" rows="3"></textarea>
						</label>
						<label>
							<span>{$t("admin.sdgs")}</span>
							<textarea name="sdgs" rows="2" placeholder="SDG 4, SDG 13"></textarea>
						</label>
						<label>
							<span>{$t("admin.metrics")}</span>
							<textarea name="metrics" rows="3" placeholder={metricsPlaceholder}></textarea>
						</label>
						<button class="btn primary" type="submit">{$t("admin.create")}</button>
					</form>
				</div>
			{:else}
				<p class="empty">{$t("admin.noOrganizations")}</p>
			{/if}
		</div>
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("admin.members")}</p>
			<h2 class="section-title">{$t("admin.members")}</h2>
		</div>
		<p class="section-note">{selectedOrg?.name || $t("admin.noOrganizations")}</p>
	</div>
	<div class="surface-card admin-list">
		{#if data.members.length === 0}
			<p class="empty">{$t("admin.noMembers")}</p>
		{:else}
			{#each data.members as member}
				<div class="admin-list-item">
					<div class="admin-list-item__meta">
						<p class="label">{member.name}</p>
						<p class="muted small">{member.position || "-"}</p>
					</div>
					<p class="small">{member.email || member.phone || member.bio || "-"}</p>
				</div>
			{/each}
		{/if}
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("admin.claims")}</p>
			<h2 class="section-title">{$t("admin.claimRequests")}</h2>
		</div>
		<p class="section-note">{selectedOrg?.name || $t("admin.noOrganizations")}</p>
	</div>
	<div class="surface-card admin-list">
		{#if data.claims.length === 0}
			<p class="empty">{$t("admin.noClaims")}</p>
		{:else}
			{#each data.claims as claim}
				<div class="admin-list-item">
					<div class="admin-list-item__meta">
						<p class="label">{claim.method} · {claim.target}</p>
						<span class="mini-badge">{claim.status}</span>
					</div>
					<p class="small">{claim.id}</p>
					{#if claim.status === "pending"}
						<div class="inline-actions">
							<form method="POST" action="?/approveClaim" class="inline-form">
								<input type="hidden" name="id" value={claim.id} />
								<button class="ghost-button" type="submit">{$t("admin.approve")}</button>
							</form>
							<form method="POST" action="?/rejectClaim" class="inline-form">
								<input type="hidden" name="id" value={claim.id} />
								<button class="ghost-button danger" type="submit">{$t("admin.reject")}</button>
							</form>
						</div>
					{/if}
				</div>
			{/each}
		{/if}
	</div>
</section>

<section class="section">
	<div class="section-head">
		<div>
			<p class="eyebrow">{$t("admin.audit")}</p>
			<h2 class="section-title">{$t("admin.auditLogs")}</h2>
		</div>
		<p class="section-note">{selectedOrg?.name || $t("admin.noOrganizations")}</p>
	</div>
	<div class="surface-card admin-list">
		{#if data.auditLogs.length === 0}
			<p class="empty">{$t("admin.noAuditLogs")}</p>
		{:else}
			{#each data.auditLogs as log}
				<div class="admin-list-item">
					<div class="admin-list-item__meta">
						<p class="label">{log.entity_type} · {log.action}</p>
						<span class="mini-badge">{new Date(log.created_at).toLocaleDateString($locale === "id" ? "id-ID" : "en-US")}</span>
					</div>
					<p class="small">{log.id}</p>
				</div>
			{/each}
		{/if}
	</div>
</section>

<section class="section content-grid">
	<div>
		<div class="section-head single">
			<div>
				<p class="eyebrow">{$t("admin.posts")}</p>
				<h2 class="section-title">{$t("admin.posts")}</h2>
			</div>
		</div>
		<div class="surface-card admin-list">
			{#if postSummary.length === 0}
				<p class="empty">{$t("postsPage.empty")}</p>
			{:else}
				{#each postSummary.slice(0, 8) as post}
					<div class="admin-list-item">
						<div class="admin-list-item__meta">
							<p class="label">{post.title}</p>
							<span class="mini-badge">{post.status}</span>
						</div>
						<p class="small">{post.organization_name} - {post.summary || "-"}</p>
						{#if post.status !== "published"}
							<form method="POST" action="?/publishPost" class="inline-form">
								<input type="hidden" name="id" value={post.id} />
								<button class="ghost-button" type="submit">{$t("admin.publish")}</button>
							</form>
						{/if}
					</div>
				{/each}
			{/if}
		</div>
	</div>

	<div>
		<div class="section-head single">
			<div>
				<p class="eyebrow">{$t("admin.impactReports")}</p>
				<h2 class="section-title">{$t("admin.impactReports")}</h2>
			</div>
		</div>
		<div class="surface-card admin-list">
			{#if impactSummary.length === 0}
				<p class="empty">{$t("admin.noImpact")}</p>
			{:else}
				{#each impactSummary.slice(0, 8) as report}
					<div class="admin-list-item">
						<div class="admin-list-item__meta">
							<p class="label">{report.title}</p>
							<span class="mini-badge">{report.status}</span>
						</div>
						<p class="small">{report.organization_name} - {report.summary || "-"}</p>
						{#if report.status !== "published"}
							<form method="POST" action="?/publishImpactReport" class="inline-form">
								<input type="hidden" name="id" value={report.id} />
								<button class="ghost-button" type="submit">{$t("admin.publish")}</button>
							</form>
						{/if}
					</div>
				{/each}
			{/if}
		</div>
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
