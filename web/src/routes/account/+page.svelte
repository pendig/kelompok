<script>
	import { fallbackDate } from "$lib/api.js";
	import { locale, t } from "$lib/i18n.js";

	let { data } = $props();

	let session = $derived(data.session);
	let user = $derived(session?.user ?? null);
	let roles = $derived(session?.organization_roles ?? []);
	let claims = $derived(session?.organization_claims ?? []);
	let unverified = $derived(Boolean(data.unverified));
	let claimId = $derived(data.claimId ?? "");

	let pendingClaims = $derived(claims.filter((claim) => claim.status === "pending"));
	let rejectedClaims = $derived(claims.filter((claim) => claim.status === "rejected"));

	// Hide approved claims from the rejection/pending lists. Approved claims
	// are already represented as full organization roles (they get an
	// owner/admin entry in organization_user_roles when admins approve a
	// claim), so duplicating them in the claims list would be noisy.
	let approvedClaimIds = $derived(
		new Set(claims.filter((claim) => claim.status === "approved").map((claim) => claim.id)),
	);

	let recentSubmittedClaim = $derived(
		claimId ? pendingClaims.find((claim) => claim.id === claimId) : null,
	);

	let hasAnyOrganizationData = $derived(
		roles.length > 0 || pendingClaims.length > 0 || rejectedClaims.length > 0,
	);

	function organizationProfilePath(slug) {
		return `/organizations/${encodeURIComponent(slug)}`;
	}

	function adminPath(slug) {
		return `/admin?org=${encodeURIComponent(slug)}`;
	}

	function methodLabel(method) {
		if (method === "instagram") {
			return $t("account.claimMethodInstagram");
		}
		return $t("account.claimMethodEmail");
	}
</script>

{#if unverified}
	<section class="section">
		<div class="section-head single">
			<div>
				<p class="eyebrow">{$t("account.eyebrow")}</p>
				<h1 class="section-title">{$t("account.title")}</h1>
				<p class="section-note">{$t("account.subtitle")}</p>
			</div>
		</div>

		<div class="error account-state-card" role="alert">
			<strong>{$t("account.errorTitle")}</strong>
			<p>{$t("account.errorBody")}</p>
			<a class="ghost-button" href="/account">{$t("account.errorRetry")}</a>
		</div>
	</section>
{:else if !user}
	<section class="section">
		<div class="section-head single">
			<div>
				<p class="eyebrow">{$t("account.eyebrow")}</p>
				<h1 class="section-title">{$t("account.title")}</h1>
				<p class="section-note">{$t("account.subtitle")}</p>
			</div>
		</div>

		<div class="notice account-state-card" role="status" aria-live="polite">
			<strong>{$t("account.loadingTitle")}</strong>
			<p>{$t("account.loadingBody")}</p>
		</div>
	</section>
{:else}
	<section class="section">
		<div class="section-head single">
			<div>
				<p class="eyebrow">{$t("account.eyebrow")}</p>
				<h1 class="section-title">{$t("account.title")}</h1>
				<p class="section-note">{$t("account.subtitle")}</p>
			</div>
		</div>

		<div class="account-grid">
			<section class="account-card">
				<div>
					<p class="label">{$t("account.signedInAs")}</p>
					<h2>{user.name}</h2>
					<p class="muted">{user.email}</p>
					<p class="mini-badge">{user.role}</p>
				</div>
				<form method="POST" action="?/logout">
					<button class="ghost-button danger" type="submit">{$t("auth.logout")}</button>
				</form>
			</section>

			<section class="account-card">
				<div>
					<p class="label">{$t("account.nextStep")}</p>
					<h2>{$t("account.crmTitle")}</h2>
					<p class="muted">{$t("account.crmBody")}</p>
				</div>
				<a class="btn primary" href="/admin">{$t("account.openAdmin")}</a>
			</section>
		</div>
	</section>

	{#if recentSubmittedClaim}
		<section class="section">
			<div class="claim-success-card" role="status" aria-live="polite">
				<div class="claim-success-copy">
					<p class="eyebrow">{$t("account.eyebrow")}</p>
					<h2>{$t("account.submittedTitle")}</h2>
					<p class="muted">{$t("account.submittedBody")}</p>
					<dl class="claim-success-meta">
						<div>
							<dt>{$t("account.submittedClaimId")}</dt>
							<dd><code>{recentSubmittedClaim.id}</code></dd>
						</div>
						<div>
							<dt>{$t("account.submittedLabel")}</dt>
							<dd>
								{recentSubmittedClaim.organization_name}
								<span class="muted small">
									· {fallbackDate(recentSubmittedClaim.created_at, $locale)}
								</span>
							</dd>
						</div>
					</dl>
					<p class="muted small claim-success-expectation">
						{$t("account.submittedReviewExpectation")}
					</p>
					<div class="inline-actions">
						<a class="ghost-button" href={organizationProfilePath(recentSubmittedClaim.organization_slug)}>
							{$t("account.submittedViewProfile")}
						</a>
					</div>
				</div>
			</div>
		</section>
	{/if}

	<section class="section">
		<div class="section-head single">
			<div>
				<p class="eyebrow">{$t("account.eyebrow")}</p>
				<h2 class="section-title">{$t("account.claimsTitle")}</h2>
				<p class="section-note">{$t("account.claimsSubtitle")}</p>
			</div>
		</div>

		{#if !hasAnyOrganizationData}
			<div class="empty account-empty">
				<strong>{$t("account.noOrganizationsTitle")}</strong>
				<p>{$t("account.noOrganizationsBody")}</p>
				<a href="/organizations" class="ghost-button">{$t("account.findOrganization")}</a>
			</div>
		{:else}
			<div class="claim-journey">
				<article class="claim-journey-section">
					<header>
						<div>
							<p class="label">{$t("account.approvedSection")}</p>
							<p class="muted small">{$t("account.approvedHelp")}</p>
						</div>
						<span class="admin-status admin-status-pass">{roles.length}</span>
					</header>

					{#if roles.length === 0}
						<p class="empty claim-journey-empty">{$t("account.noOrganizationsBody")}</p>
					{:else}
						<ul class="account-role-list" role="list">
							{#each roles as role}
								<li class="account-role">
									<div>
										<p class="label">{role.role}</p>
										<h3>{role.organization_name}</h3>
										<p class="muted small">/{role.organization_slug}</p>
										<p class="muted small">
											{$t("account.updatedAt", { date: fallbackDate(role.updated_at, $locale) })}
										</p>
									</div>
									<div class="inline-actions">
										<a class="ghost-button" href={organizationProfilePath(role.organization_slug)}>
											{$t("account.viewPublic")}
										</a>
										<a class="btn primary" href={adminPath(role.organization_slug)}>
											{$t("account.manage")}
										</a>
									</div>
								</li>
							{/each}
						</ul>
					{/if}
				</article>

				<article class="claim-journey-section">
					<header>
						<div>
							<p class="label">{$t("account.pendingSection")}</p>
							<p class="muted small">{$t("account.pendingHelp")}</p>
						</div>
						<span class="admin-status admin-status-warn">{pendingClaims.length}</span>
					</header>

					{#if pendingClaims.length === 0}
						<p class="empty claim-journey-empty">{$t("account.noClaimsBody")}</p>
					{:else}
						<ul class="claim-card-list" role="list">
							{#each pendingClaims as claim}
								<li class="claim-status-card claim-status-card--pending">
									<div class="claim-status-card-head">
										<div>
											<h3>{claim.organization_name}</h3>
											<p class="muted small">/{claim.organization_slug}</p>
										</div>
										<span class="admin-status admin-status-warn">
											{$t("account.claimStatusPending")}
										</span>
									</div>
									<dl class="claim-status-card-meta">
										<div>
											<dt>{$t("account.claimIdLabel")}</dt>
											<dd><code>{claim.id}</code></dd>
										</div>
										<div>
											<dt>{$t("account.claimMethodLabel")}</dt>
											<dd>{methodLabel(claim.method)}</dd>
										</div>
										<div>
											<dt>{$t("account.claimTargetLabel")}</dt>
											<dd>{claim.target}</dd>
										</div>
										<div>
											<dt>{$t("account.submittedLabel")}</dt>
											<dd>{fallbackDate(claim.created_at, $locale)}</dd>
										</div>
									</dl>
									<p class="claim-status-card-help muted small">
										{$t("account.pendingReviewExpectation")}
									</p>
									<div class="inline-actions">
										<a class="ghost-button" href={organizationProfilePath(claim.organization_slug)}>
											{$t("account.submittedViewProfile")}
										</a>
									</div>
								</li>
							{/each}
						</ul>
					{/if}
				</article>

				<article class="claim-journey-section">
					<header>
						<div>
							<p class="label">{$t("account.rejectedSection")}</p>
							<p class="muted small">{$t("account.rejectedHelp")}</p>
						</div>
						<span class="admin-status admin-status-fail">{rejectedClaims.length}</span>
					</header>

					{#if rejectedClaims.length === 0}
						<p class="empty claim-journey-empty">{$t("account.noClaimsBody")}</p>
					{:else}
						<ul class="claim-card-list" role="list">
							{#each rejectedClaims as claim}
								<li class="claim-status-card claim-status-card--rejected" hidden={approvedClaimIds.has(claim.id)}>
									<div class="claim-status-card-head">
										<div>
											<h3>{claim.organization_name}</h3>
											<p class="muted small">/{claim.organization_slug}</p>
										</div>
										<span class="admin-status admin-status-fail">
											{$t("account.claimStatusRejected")}
										</span>
									</div>
									<dl class="claim-status-card-meta">
										<div>
											<dt>{$t("account.claimIdLabel")}</dt>
											<dd><code>{claim.id}</code></dd>
										</div>
										<div>
											<dt>{$t("account.claimMethodLabel")}</dt>
											<dd>{methodLabel(claim.method)}</dd>
										</div>
										<div>
											<dt>{$t("account.claimTargetLabel")}</dt>
											<dd>{claim.target}</dd>
										</div>
										{#if claim.reviewed_at}
											<div>
												<dt>{$t("account.reviewedLabel")}</dt>
												<dd>{fallbackDate(claim.reviewed_at, $locale)}</dd>
											</div>
										{/if}
									</dl>
									<p class="claim-status-card-help">
										{$t("account.rejectedRecoveryBody")}
									</p>
									<div class="inline-actions">
										<a class="btn primary" href={organizationProfilePath(claim.organization_slug)}>
											{$t("account.resubmitClaim")}
										</a>
										<a class="ghost-button" href={organizationProfilePath(claim.organization_slug)}>
											{$t("account.viewPublic")}
										</a>
									</div>
								</li>
							{/each}
						</ul>
					{/if}
				</article>
			</div>
		{/if}
	</section>
{/if}

<style>
	.account-state-card {
		display: grid;
		gap: 12px;
		margin-top: 24px;
	}

	.account-state-card .ghost-button {
		justify-self: start;
	}

	.claim-success-card {
		margin-top: 24px;
		padding: 28px;
		border: 1px solid hsl(142, 70%, 85%);
		border-radius: 16px;
		background: linear-gradient(135deg, hsl(142, 70%, 98%) 0%, var(--surface) 100%);
		box-shadow: var(--shadow-md);
	}

	.claim-success-copy h2 {
		margin: 6px 0 10px;
		font-size: 24px;
		font-weight: 800;
		line-height: 1.2;
	}

	.claim-success-copy .muted {
		max-width: 60ch;
	}

	.claim-success-meta {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 12px 24px;
		margin: 16px 0 8px;
		padding: 0;
	}

	.claim-success-meta div {
		display: grid;
		gap: 4px;
		min-width: 0;
	}

	.claim-success-meta dt {
		margin: 0;
		color: var(--muted);
		font-size: 11px;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.claim-success-meta dd {
		margin: 0;
		color: var(--text);
		font-size: 14.5px;
		font-weight: 600;
		word-break: break-all;
	}

	.claim-success-meta code {
		font-family:
			ui-monospace,
			SFMono-Regular,
			"SF Mono",
			Consolas,
			"Liberation Mono",
			Menlo,
			monospace;
		font-size: 13.5px;
	}

	.claim-success-expectation {
		margin: 0;
	}

	.claim-success-card .inline-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 12px;
		margin-top: 16px;
	}

	.claim-journey {
		display: grid;
		gap: 20px;
		margin-top: 16px;
	}

	.claim-journey-section {
		border: 1px solid var(--border);
		border-radius: 16px;
		background: var(--surface);
		box-shadow: var(--shadow-sm);
		padding: 22px;
		display: grid;
		gap: 16px;
	}

	.claim-journey-section > header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 16px;
	}

	.claim-journey-section header .label {
		margin: 0 0 4px;
	}

	.claim-journey-empty {
		margin: 0;
		padding: 16px;
	}

	.claim-card-list {
		display: grid;
		gap: 12px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.claim-status-card {
		display: grid;
		gap: 14px;
		padding: 18px;
		border: 1px solid var(--border);
		border-radius: 12px;
		background: var(--surface-soft);
	}

	.claim-status-card--pending {
		border-color: hsl(38, 90%, 88%);
		background: linear-gradient(135deg, hsl(38, 90%, 99%) 0%, var(--surface) 60%);
	}

	.claim-status-card--rejected {
		border-color: hsl(0, 85%, 90%);
		background: linear-gradient(135deg, hsl(0, 80%, 99%) 0%, var(--surface) 60%);
	}

	.claim-status-card-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
	}

	.claim-status-card-head h3 {
		margin: 0;
		font-size: 17px;
		font-weight: 800;
		line-height: 1.25;
	}

	.claim-status-card-meta {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 8px 18px;
		margin: 0;
	}

	.claim-status-card-meta div {
		display: grid;
		gap: 2px;
		min-width: 0;
	}

	.claim-status-card-meta dt {
		margin: 0;
		color: var(--muted);
		font-size: 11px;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.claim-status-card-meta dd {
		margin: 0;
		color: var(--text);
		font-size: 14px;
		font-weight: 600;
		word-break: break-all;
	}

	.claim-status-card-meta code {
		font-family:
			ui-monospace,
			SFMono-Regular,
			"SF Mono",
			Consolas,
			"Liberation Mono",
			Menlo,
			monospace;
		font-size: 13px;
	}

	.claim-status-card-help {
		margin: 0;
		color: var(--muted);
		font-size: 13.5px;
		line-height: 1.55;
	}

	.claim-status-card .inline-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
	}

	@media (max-width: 720px) {
		.claim-success-meta,
		.claim-status-card-meta {
			grid-template-columns: minmax(0, 1fr);
		}

		.claim-journey-section > header {
			flex-direction: column;
			align-items: flex-start;
		}

		.claim-status-card-head {
			flex-direction: column;
			align-items: flex-start;
		}
	}
</style>
