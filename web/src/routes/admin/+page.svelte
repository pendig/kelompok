<script>
	import { locale, t } from "$lib/i18n.js";
	import { getTheme, getInitials } from "$lib/theme.js";

	let { data, form } = $props();

	// svelte-ignore state_referenced_locally
	let activeTab = $state(data.initialTab || "dashboard");
	let expandedSections = $state({
		organization: false,
		content: false,
		governance: false,
	});

	const organizationTabs = ["organizations", "organization-edit", "members", "relationships"];
	const contentTabs = ["posts", "impact"];
	const governanceTabs = ["claims", "audit"];

	let organizationNavOpen = $derived(expandedSections.organization || organizationTabs.includes(activeTab));
	let contentNavOpen = $derived(expandedSections.content || contentTabs.includes(activeTab));
	let governanceNavOpen = $derived(expandedSections.governance || governanceTabs.includes(activeTab));

	$effect.pre(() => {
		activeTab = data.initialTab || "dashboard";
	});

	function setActiveTab(tab, section = "") {
		activeTab = tab;
		if (section) {
			expandedSections[section] = true;
		}
	}

	function toggleSection(section) {
		expandedSections[section] = !expandedSections[section];
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

	function selectedPath(org, view = "organization-edit") {
		return `/admin?org=${encodeURIComponent(org.slug)}&view=${encodeURIComponent(view)}`;
	}

	function actionPath(action, view = activeTab) {
		const params = new URLSearchParams();
		if (selectedOrg?.slug) {
			params.set("org", selectedOrg.slug);
		}
		if (view) {
			params.set("view", view);
		}
		const query = params.toString();
		return `?${query ? `${query}&` : ""}/${action}`;
	}

	function formatAdminLocation(org) {
		const parts = [org?.city, org?.region, org?.country].filter(Boolean);
		return parts.length ? parts.join(", ") : "-";
	}

	const socialPlaceholder = '{"instagram":"https://instagram.com/example"}';
	const evidencePlaceholder = '{"note":"manual claim test"}';
	const metricsPlaceholder = '{"beneficiaries":120}';
	const metadataPlaceholder = '{"source":"manual"}';

	const orgSummary = $derived(data.organizations || []);
	const postSummary = $derived(data.posts || []);
	const impactSummary = $derived(data.impactReports || []);
	const relationshipSummary = $derived(data.relationships || []);
	const selectedOrg = $derived(data.selectedOrganization);
	const currentUser = $derived(data.session?.user);
	const selectedProfile = $derived(selectedOrg?.profile_data || {});
	const selectedSdgs = $derived(selectedOrg?.sdgs_data || {});
</script>

<section class="section">
	<div class="page-heading">
		<p class="eyebrow">{$t("admin.workspaceEyebrow")}</p>
		<h1 class="section-title">{$t("admin.workspaceTitle")}</h1>
		<p class="section-note">
			{$t("admin.workspaceSubtitle")}
		</p>
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

<div class="admin-shell">
	<aside class="admin-side-nav" aria-label={$t("admin.adminNavigation")}>
		<p class="label">{$t("admin.adminNavigation")}</p>

		<button
			class="admin-nav-link"
			class:active={activeTab === "dashboard"}
			data-testid="admin-nav-dashboard"
			type="button"
			onclick={() => setActiveTab("dashboard")}
		>
			<span>{$t("admin.dashboard")}</span>
		</button>

		<div class="admin-nav-group">
				<button
					class="admin-nav-group-trigger"
					class:active={organizationTabs.includes(activeTab)}
					data-testid="admin-nav-organization"
				type="button"
				aria-expanded={organizationNavOpen}
				onclick={() => toggleSection("organization")}
			>
				<span>{$t("admin.navOrganization")}</span>
				<span class="admin-nav-chevron" aria-hidden="true">{organizationNavOpen ? "−" : "+"}</span>
			</button>
			{#if organizationNavOpen}
				<div class="admin-nav-submenu">
						<button
							class="admin-nav-subitem"
							class:active={activeTab === "organizations"}
							data-testid="admin-nav-organization-profile"
							type="button"
							onclick={() => setActiveTab("organizations", "organization")}
						>
							{$t("admin.orgForm")}
						</button>
						{#if selectedOrg}
							<button
								class="admin-nav-subitem"
								class:active={activeTab === "organization-edit"}
								data-testid="admin-nav-organization-edit"
								type="button"
								onclick={() => setActiveTab("organization-edit", "organization")}
							>
								{$t("admin.editOrgProfile")}
							</button>
						{/if}
						<button
							class="admin-nav-subitem"
							class:active={activeTab === "members"}
							data-testid="admin-nav-organization-members"
							type="button"
							onclick={() => setActiveTab("members", "organization")}
						>
							{$t("admin.members")}
						</button>
						<button
							class="admin-nav-subitem"
							class:active={activeTab === "relationships"}
							data-testid="admin-nav-organization-relationships"
							type="button"
							onclick={() => setActiveTab("relationships", "organization")}
						>
							{$t("admin.relationships")}
						</button>
					</div>
				{/if}
		</div>

		<div class="admin-nav-group">
				<button
					class="admin-nav-group-trigger"
					class:active={contentTabs.includes(activeTab)}
					data-testid="admin-nav-content"
				type="button"
				aria-expanded={contentNavOpen}
				onclick={() => toggleSection("content")}
			>
				<span>{$t("admin.navContent")}</span>
				<span class="admin-nav-chevron" aria-hidden="true">{contentNavOpen ? "−" : "+"}</span>
			</button>
			{#if contentNavOpen}
				<div class="admin-nav-submenu">
						<button
							class="admin-nav-subitem"
							class:active={activeTab === "posts"}
							data-testid="admin-nav-content-posts"
							type="button"
							onclick={() => setActiveTab("posts", "content")}
						>
							{$t("admin.posts")}
						</button>
						<button
							class="admin-nav-subitem"
							class:active={activeTab === "impact"}
							data-testid="admin-nav-content-impact"
							type="button"
							onclick={() => setActiveTab("impact", "content")}
						>
							{$t("admin.impactReports")}
						</button>
					</div>
				{/if}
		</div>

		<div class="admin-nav-group">
				<button
					class="admin-nav-group-trigger"
					class:active={governanceTabs.includes(activeTab)}
					data-testid="admin-nav-governance"
				type="button"
				aria-expanded={governanceNavOpen}
				onclick={() => toggleSection("governance")}
			>
				<span>{$t("admin.navGovernance")}</span>
				<span class="admin-nav-chevron" aria-hidden="true">{governanceNavOpen ? "−" : "+"}</span>
			</button>
			{#if governanceNavOpen}
				<div class="admin-nav-submenu">
						<button
							class="admin-nav-subitem"
							class:active={activeTab === "claims"}
							data-testid="admin-nav-governance-claims"
							type="button"
							onclick={() => setActiveTab("claims", "governance")}
						>
							{$t("admin.claims")}
						</button>
						<button
							class="admin-nav-subitem"
							class:active={activeTab === "audit"}
							data-testid="admin-nav-governance-audit"
							type="button"
							onclick={() => setActiveTab("audit", "governance")}
						>
							{$t("admin.audit")}
						</button>
					</div>
				{/if}
		</div>

		<a class="admin-nav-link" href="/admin/developer">{$t("admin.developerPanel")}</a>
	</aside>

	<div class="admin-content">
{#if activeTab === "dashboard"}
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
				<p class="eyebrow">{$t("admin.dataSection")}</p>
				<h2 class="section-title">{$t("admin.dashboardTitle")}</h2>
			</div>
			<p class="section-note">{$t("admin.dashboardSubtitle")}</p>
		</div>

		{#if data.loadErrors.length}
			<p class="error">
				<strong>{$t("admin.actionError")}: </strong>
				{data.loadErrors[0]}
			</p>
		{/if}

		<div class="admin-grid">
			<article class="card">
				<p class="label">{$t("admin.totalOrganizations")}</p>
				<p class="admin-number">{orgSummary.length}</p>
				<p class="small">{$t("admin.totalOrganizationsHelp")}</p>
			</article>
			<article class="card">
				<p class="label">{$t("admin.totalPosts")}</p>
				<p class="admin-number">{postSummary.length}</p>
				<p class="small">{$t("admin.totalPostsHelp")}</p>
			</article>
			<article class="card">
				<p class="label">{$t("admin.totalImpactPreview")}</p>
				<p class="admin-number">{impactSummary.length}</p>
				<p class="small">{$t("admin.totalImpactHelp")}</p>
			</article>
		</div>
	</section>

	<section class="section">
		<div class="section-head">
			<div>
				<p class="eyebrow">{$t("admin.nextActionEyebrow")}</p>
				<h2 class="section-title">{$t("admin.nextActionTitle")}</h2>
			</div>
			<p class="section-note">{$t("admin.nextActionSubtitle")}</p>
		</div>
		<div class="admin-grid">
			<article class="card">
				<p class="label">{$t("admin.orgForm")}</p>
				<h3>{$t("admin.manageProfileTitle")}</h3>
				<p class="small">{$t("admin.manageProfileBody")}</p>
				<button class="ghost-button" type="button" onclick={() => setActiveTab("organizations", "organization")}>{$t("admin.openSection")}</button>
			</article>
			<article class="card">
				<p class="label">{$t("admin.posts")}</p>
				<h3>{$t("admin.manageContentTitle")}</h3>
				<p class="small">{$t("admin.manageContentBody")}</p>
				<button class="ghost-button" type="button" onclick={() => setActiveTab("posts", "content")}>{$t("admin.openSection")}</button>
			</article>
			<article class="card">
				<p class="label">{$t("admin.claims")}</p>
				<h3>{$t("admin.manageClaimsTitle")}</h3>
				<p class="small">{$t("admin.manageClaimsBody")}</p>
				<button class="ghost-button" type="button" onclick={() => setActiveTab("claims", "governance")}>{$t("admin.openSection")}</button>
			</article>
		</div>
	</section>
{/if}

	{#if activeTab === "organizations"}
		<section class="section">
			<div class="section-head">
				<div>
					<p class="eyebrow">{$t("admin.operationsSection")}</p>
					<h2 class="section-title">{$t("admin.organizationDirectoryTitle")}</h2>
				</div>
				<p class="section-note">{$t("admin.organizationDirectorySubtitle")}</p>
			</div>

			<div class="admin-directory-layout">
				<div class="admin-org-card-grid">
					{#if orgSummary.length === 0}
						<div class="admin-empty-state">
							<h3>{$t("admin.noOrganizations")}</h3>
							<p>{$t("admin.createOrganizationHint")}</p>
						</div>
					{:else}
						{#each orgSummary as org}
							{@const theme = getTheme(org.name)}
							<a
								class="admin-org-card"
								class:active={org.slug === data.selectedSlug}
								href={selectedPath(org, "organization-edit")}
							>
								<div class="mini-card-cover" style="background: {theme.cover};"></div>
								<div class="admin-org-card-body">
									<div class="admin-org-card-top">
										<div class="mini-card-avatar" style="width: 48px; height: 48px; font-size: 16px; color: {theme.avatarText}; background: {theme.avatarBg};">
											{getInitials(org.name)}
										</div>
										<span class="admin-status {org.claim_status === 'claimed' ? 'admin-status-pass' : 'admin-status-warn'}">
											{org.claim_status || "unclaimed"}
										</span>
									</div>
									<h3>{org.name}</h3>
									<p class="small">{org.description || $t("organizationsPage.noDescription")}</p>
									<p class="meta">{formatAdminLocation(org)}</p>
									<span class="admin-org-card-action">{$t("admin.openEditor")}</span>
								</div>
							</a>
						{/each}
					{/if}
				</div>

				<form
					class="admin-panel admin-create-org"
					method="POST"
					action={actionPath("createOrganization", "organizations")}
					aria-labelledby="admin-create-org-heading"
					aria-describedby="admin-create-org-intro"
				>
					<header class="admin-create-org__head">
						<h3 id="admin-create-org-heading">{$t("admin.createOrgPanelTitle")}</h3>
						<p id="admin-create-org-intro" class="small section-note">{$t("admin.createOrgIntro")}</p>
					</header>

					{#if form?.error && form?.action === "createOrganization"}
						<div class="admin-form-error" role="alert" aria-live="polite">
							<strong>{$t("admin.createOrgErrorTitle")}: </strong>{form.error}
						</div>
					{/if}

					<fieldset class="admin-create-org__group">
						<legend>{$t("admin.editOrgGroupBasics")}</legend>
						<div class="admin-field-grid two">
							<label>
								<span>
									{$t("admin.name")}
									<small class="field-tag required">{$t("admin.createOrgRequired")}</small>
								</span>
								<input
									name="name"
									required
									autocomplete="organization"
									placeholder="Kelompok Nusantara"
								/>
							</label>
							<label>
								<span>
									{$t("admin.slug")}
									<small class="field-tag">{$t("admin.createOrgOptional")}</small>
								</span>
								<input
									name="slug"
									pattern="[a-z0-9](?:[a-z0-9-]*[a-z0-9])?"
									placeholder="kelompok-nusantara"
									aria-describedby="admin-create-slug-help"
								/>
								<small id="admin-create-slug-help" class="field-help">
									{$t("admin.createOrgSlugHelper")}
								</small>
							</label>
							<label>
								<span>
									{$t("admin.legalName")}
									<small class="field-tag">{$t("admin.createOrgOptional")}</small>
								</span>
								<input name="legal_name" placeholder="Yayasan Kelompok" />
							</label>
							<label>
								<span>
									{$t("admin.claimStatus")}
									<small class="field-tag">{$t("admin.createOrgOptional")}</small>
								</span>
								<select name="claim_status">
									<option value="unclaimed">unclaimed</option>
									<option value="pending">pending</option>
									<option value="claimed">claimed</option>
									<option value="rejected">rejected</option>
								</select>
							</label>
						</div>
					</fieldset>

					<fieldset class="admin-create-org__group">
						<legend>{$t("admin.editOrgGroupLocation")}</legend>
						<div class="admin-field-grid two">
							<label>
								<span>{$t("admin.country")}</span>
								<input name="country" placeholder="Indonesia" autocomplete="country-name" />
							</label>
							<label>
								<span>{$t("admin.region")}</span>
								<input name="region" placeholder="DKI Jakarta" />
							</label>
							<label>
								<span>{$t("admin.city")}</span>
								<input name="city" placeholder="Jakarta" />
							</label>
							<label>
								<span>{$t("admin.websiteUrl")}</span>
								<input
									name="website_url"
									type="url"
									inputmode="url"
									placeholder="https://kelompok.id"
									aria-describedby="admin-create-website-help"
								/>
								<small id="admin-create-website-help" class="field-help">
									{$t("admin.createOrgWebsiteHelper")}
								</small>
							</label>
							<label class="admin-field-grid__full">
								<span>{$t("admin.officialEmail")}</span>
								<input
									name="official_email"
									type="email"
									autocomplete="email"
									placeholder="hello@kelompok.id"
									aria-describedby="admin-create-email-help"
								/>
								<small id="admin-create-email-help" class="field-help">
									{$t("admin.createOrgEmailHelper")}
								</small>
							</label>
						</div>
					</fieldset>

					<fieldset class="admin-create-org__group">
						<legend>{$t("admin.editOrgGroupNarrative")}</legend>
						<label>
							<span>{$t("admin.description")}</span>
							<textarea name="description" rows="3" placeholder="Misi dan ringkasan organisasi"></textarea>
						</label>
						<label>
							<span>{$t("admin.history")}</span>
							<textarea name="history" rows="3" placeholder="Latar belakang dan perjalanan organisasi"></textarea>
						</label>
					</fieldset>

					<fieldset class="admin-create-org__group">
						<legend>{$t("admin.createOrgFocusSection")}</legend>
						<div class="admin-field-grid two">
							<label>
								<span>{$t("admin.focus")}</span>
								<textarea
									name="focus"
									rows="2"
									placeholder="Pendidikan, Kesehatan"
									aria-describedby="admin-create-list-help"
								></textarea>
							</label>
							<label>
								<span>{$t("admin.programs")}</span>
								<textarea
									name="programs"
									rows="2"
									placeholder="Beasiswa, Kelas literasi"
									aria-describedby="admin-create-list-help"
								></textarea>
							</label>
							<label>
								<span>{$t("admin.languages")}</span>
								<textarea
									name="languages"
									rows="2"
									placeholder="id, en"
									aria-describedby="admin-create-list-help"
								></textarea>
							</label>
							<label>
								<span>{$t("admin.sdgs")}</span>
								<textarea
									name="sdgs"
									rows="2"
									placeholder="SDG 4, SDG 13"
									aria-describedby="admin-create-list-help"
								></textarea>
							</label>
						</div>
						<small id="admin-create-list-help" class="field-help">
							{$t("admin.createOrgListHelper")}
						</small>
					</fieldset>

					<fieldset class="admin-create-org__group">
						<legend>{$t("admin.createOrgPublicContact")}</legend>
						<p class="small section-note">{$t("admin.createOrgPublicContactHelper")}</p>
						<div class="admin-field-grid two">
							<label>
								<span>{$t("admin.publicEmail")}</span>
								<input name="public_contact_email" type="email" placeholder="halo@kelompok.id" />
							</label>
							<label>
								<span>Instagram</span>
								<input name="public_contact_instagram" placeholder="@kelompok" />
							</label>
							<label>
								<span>{$t("admin.phone")}</span>
								<input name="public_contact_phone" type="tel" placeholder="+62..." />
							</label>
						</div>
					</fieldset>

					<aside class="admin-create-org__deferred" aria-label={$t("admin.createOrgDeferred")}>
						<p class="label">{$t("admin.createOrgDeferred")}</p>
						<p class="small">{$t("admin.createOrgDeferredBody")}</p>
					</aside>

					<div class="admin-actions">
						<button class="btn primary" type="submit">{$t("admin.create")}</button>
					</div>
				</form>
			</div>
		</section>
	{/if}

	{#if activeTab === "organization-edit"}
		{#if !selectedOrg}
			<div class="admin-empty-state">
				<h3>{$t("admin.noOrganizations")}</h3>
				<p>{$t("admin.selectOrganizationFirst")}</p>
				<button class="btn primary" onclick={() => setActiveTab("organizations", "organization")}>
					{$t("admin.chooseOrganization")}
				</button>
			</div>
		{:else}
			<section class="section">
				<div class="section-head">
					<div>
						<p class="eyebrow">{$t("admin.orgForm")}</p>
						<h2 class="section-title">{$t("admin.editOrgProfile")}</h2>
					</div>
					<p class="section-note">{selectedOrg.name}</p>
				</div>

				{#if data.justCreated}
					<p class="notice" role="status" aria-live="polite">
						<strong>{$t("admin.createOrgSuccess")}</strong>
					</p>
				{/if}

				<form class="admin-panel" method="POST" action={actionPath("updateOrganization", "organization-edit")}>
					<div class="admin-form-head">
						<div>
							<p class="eyebrow">{$t("admin.orgForm")}</p>
							<h3>{selectedOrg.name}</h3>
						</div>
						<a class="ghost-button" href={`/organizations/${encodeURIComponent(selectedOrg.slug)}`} target="_blank" rel="noreferrer">
							{$t("admin.viewPublic")}
						</a>
					</div>
					<p class="small section-note">{$t("admin.editOrgPanelIntro")}</p>
					<input name="current_slug" type="hidden" value={selectedOrg.slug} />
					<div class="admin-field-grid">
						<label>
							<span>{$t("admin.name")}</span>
							<input name="name" required value={selectedOrg.name || ""} autocomplete="organization" />
						</label>
						<label>
							<span>{$t("admin.slug")}</span>
							<input
								name="slug"
								value={selectedOrg.slug || ""}
								pattern="[a-z0-9](?:[a-z0-9-]*[a-z0-9])?"
								aria-describedby="admin-edit-slug-help"
							/>
							<small id="admin-edit-slug-help" class="field-help">{$t("admin.createOrgSlugHelper")}</small>
						</label>
						<label>
							<span>{$t("admin.legalName")}</span>
							<input name="legal_name" value={selectedOrg.legal_name || ""} />
						</label>
						<label>
							<span>{$t("admin.officialEmail")}</span>
							<input name="official_email" type="email" value={selectedOrg.official_email || ""} autocomplete="email" />
						</label>
						<label>
							<span>{$t("admin.websiteUrl")}</span>
							<input
								name="website_url"
								type="url"
								inputmode="url"
								value={selectedOrg.website_url || ""}
								aria-describedby="admin-edit-website-help"
							/>
							<small id="admin-edit-website-help" class="field-help">{$t("admin.createOrgWebsiteHelper")}</small>
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
							<input name="country" value={selectedOrg.country || ""} autocomplete="country-name" />
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
						<textarea
							name="impact_data"
							rows="3"
							aria-describedby="admin-edit-impact-help">{JSON.stringify(selectedOrg.impact_data || {}, null, 2)}</textarea>
						<small id="admin-edit-impact-help" class="field-help">{$t("admin.editOrgAdvancedHelper")}</small>
					</label>
					<aside class="admin-create-org__deferred" aria-label={$t("admin.createOrgDeferred")}>
						<p class="label">{$t("admin.createOrgDeferred")}</p>
						<p class="small">{$t("admin.createOrgDeferredBody")}</p>
					</aside>
					<div class="admin-actions">
						<button class="btn primary" type="submit">{$t("admin.update")}</button>
					</div>
				</form>
			</section>
		{/if}
	{/if}

	{#if activeTab === "relationships"}
		{#if !selectedOrg}
			<div class="admin-empty-state">
				<h3>{$t("admin.noOrganizations")}</h3>
				<p>{$t("admin.selectOrganizationFirst")}</p>
				<button class="btn primary" onclick={() => setActiveTab("organizations", "organization")}>
					{$t("admin.chooseOrganization")}
				</button>
			</div>
		{:else}
		<section class="section">
			<div class="section-head">
				<div>
					<p class="eyebrow">{$t("admin.relationships")}</p>
					<h2 class="section-title">{$t("admin.relationships")}</h2>
				</div>
				<p class="section-note">{selectedOrg.name}</p>
			</div>

			<div class="admin-panel-grid">
				<form class="admin-panel" method="POST" action={actionPath("createRelationship", "relationships")}>
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
										<form method="POST" action={actionPath("deleteRelationship", "relationships")} class="inline-form">
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
			</section>
		{/if}
	{/if}

	{#if activeTab === "members"}
		{#if !selectedOrg}
			<div class="admin-empty-state">
				<h3>{$t("admin.noOrganizations")}</h3>
				<p>{$t("admin.selectOrganizationFirst")}</p>
				<button class="btn primary" onclick={() => setActiveTab("organizations", "organization")}>
					{$t("admin.chooseOrganization")}
				</button>
			</div>
		{:else}
			<section class="section">
				<div class="section-head">
					<div>
					<p class="eyebrow">{$t("admin.members")}</p>
					<h2 class="section-title">{$t("admin.members")}</h2>
				</div>
				<p class="section-note">{selectedOrg.name}</p>
			</div>

			<div class="admin-panel-grid">
				<form class="admin-panel" method="POST" action={actionPath("createMember", "members")}>
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

				<div class="admin-panel">
					<h3>{$t("admin.members")}</h3>
					<div class="admin-list">
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
				</div>
				</div>
			</section>
		{/if}
	{/if}

	{#if activeTab === "posts"}
		{#if !selectedOrg}
			<div class="admin-empty-state">
				<h3>{$t("admin.noOrganizations")}</h3>
				<p>{$t("admin.selectOrganizationFirst")}</p>
				<button class="btn primary" onclick={() => setActiveTab("organizations", "organization")}>
					{$t("admin.chooseOrganization")}
				</button>
			</div>
		{:else}
		<section class="section">
			<div class="section-head">
				<div>
					<p class="eyebrow">{$t("admin.posts")}</p>
					<h2 class="section-title">{$t("admin.posts")}</h2>
				</div>
				<p class="section-note">{selectedOrg.name}</p>
			</div>

			<div class="admin-panel-grid">
				<form class="admin-panel" method="POST" action={actionPath("createPost", "posts")}>
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

				<div class="admin-panel">
					<h3>{$t("admin.posts")}</h3>
					<div class="admin-list">
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
										<form method="POST" action={actionPath("publishPost", "posts")} class="inline-form" style="margin-top: 8px;">
											<input type="hidden" name="id" value={post.id} />
											<button class="ghost-button" type="submit">{$t("admin.publish")}</button>
										</form>
									{/if}
								</div>
							{/each}
						{/if}
					</div>
				</div>
				</div>
			</section>
		{/if}
	{/if}

	{#if activeTab === "impact"}
		{#if !selectedOrg}
			<div class="admin-empty-state">
				<h3>{$t("admin.noOrganizations")}</h3>
				<p>{$t("admin.selectOrganizationFirst")}</p>
				<button class="btn primary" onclick={() => setActiveTab("organizations", "organization")}>
					{$t("admin.chooseOrganization")}
				</button>
			</div>
		{:else}
			<section class="section">
				<div class="section-head">
					<div>
					<p class="eyebrow">{$t("admin.impactReports")}</p>
					<h2 class="section-title">{$t("admin.impactReports")}</h2>
				</div>
				<p class="section-note">{selectedOrg.name}</p>
			</div>

			<div class="admin-panel-grid">
				<form class="admin-panel" method="POST" action={actionPath("createImpactReport", "impact")}>
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

				<div class="admin-panel">
					<h3>{$t("admin.impactReports")}</h3>
					<div class="admin-list">
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
										<form method="POST" action={actionPath("publishImpactReport", "impact")} class="inline-form" style="margin-top: 8px;">
											<input type="hidden" name="id" value={report.id} />
											<button class="ghost-button" type="submit">{$t("admin.publish")}</button>
										</form>
									{/if}
								</div>
							{/each}
						{/if}
					</div>
				</div>
			</div>
		</section>
	{/if}
{/if}

	{#if activeTab === "claims"}
		{#if !selectedOrg}
			<div class="admin-empty-state">
				<h3>{$t("admin.noOrganizations")}</h3>
				<p>{$t("admin.selectOrganizationFirst")}</p>
				<button class="btn primary" onclick={() => setActiveTab("organizations", "organization")}>
					{$t("admin.chooseOrganization")}
				</button>
			</div>
		{:else}
		<section class="section">
			<div class="section-head">
				<div>
					<p class="eyebrow">{$t("admin.claims")}</p>
					<h2 class="section-title">{$t("admin.claimRequests")}</h2>
				</div>
				<p class="section-note">{selectedOrg.name}</p>
			</div>

			<div class="admin-panel-grid">
				<form class="admin-panel" method="POST" action={actionPath("createClaim", "claims")}>
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

				<div class="admin-panel">
					<h3>{$t("admin.claims")}</h3>
					<div class="admin-list">
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
										<div class="inline-actions" style="margin-top: 8px;">
											<form method="POST" action={actionPath("approveClaim", "claims")} class="inline-form">
												<input type="hidden" name="id" value={claim.id} />
												<button class="ghost-button" type="submit">{$t("admin.approve")}</button>
											</form>
											<form method="POST" action={actionPath("rejectClaim", "claims")} class="inline-form">
												<input type="hidden" name="id" value={claim.id} />
												<button class="ghost-button danger" type="submit">{$t("admin.reject")}</button>
											</form>
										</div>
									{/if}
								</div>
							{/each}
						{/if}
					</div>
				</div>
				</div>
			</section>
		{/if}
	{/if}

	{#if activeTab === "audit"}
		{#if !selectedOrg}
			<div class="admin-empty-state">
				<h3>{$t("admin.noOrganizations")}</h3>
				<p>{$t("admin.selectOrganizationFirst")}</p>
				<button class="btn primary" onclick={() => setActiveTab("organizations", "organization")}>
					{$t("admin.chooseOrganization")}
				</button>
			</div>
		{:else}
			<section class="section">
				<div class="section-head">
					<div>
					<p class="eyebrow">{$t("admin.audit")}</p>
					<h2 class="section-title">{$t("admin.auditLogs")}</h2>
				</div>
				<p class="section-note">{selectedOrg.name}</p>
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
		{/if}
	{/if}
		</div>
	</div>
