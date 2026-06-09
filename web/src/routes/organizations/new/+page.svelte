<script>
	import { enhance } from "$app/forms";
	import { t } from "$lib/i18n.js";

	let { data, form } = $props();

	const defaultValues = {
		name: "",
		legal_name: "",
		description: "",
		country: "Indonesia",
		region: "",
		city: "",
		website_url: "",
		official_email: "",
		method: "official_email",
		target: "",
		evidence_note: "",
	};

	let values = $state({ ...defaultValues });
	let touched = $state({});
	let pending = $state(false);
	let created = $derived(form?.action === "createOrganization" && form?.ok === true ? form.item : null);
	let errorCode = $derived(form?.action === "createOrganization" && form?.ok === false ? form.code : null);
	let errorMessage = $derived(form?.action === "createOrganization" && form?.ok === false ? form.error : null);
	let fieldError = $derived(form?.action === "createOrganization" && form?.ok === false ? form.field : null);

	$effect(() => {
		if (form?.values) {
			values = { ...defaultValues, ...form.values };
		}
	});

	let nameValid = $derived(values.name.trim().length > 0 && values.name.trim().length <= 160);
	let targetValid = $derived(
		values.target.trim().length > 0 &&
			(values.method !== "official_email" || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(values.target.trim())),
	);
	let websiteValid = $derived(
		!values.website_url.trim() || /^https?:\/\/[^\s]+\.[^\s]+/i.test(values.website_url.trim()),
	);
	let canSubmit = $derived(nameValid && targetValid && websiteValid && !pending);

	function mark(field) {
		touched[field] = true;
	}

	function methodTargetLabel() {
		if (values.method === "instagram") return $t("organizationOnboarding.instagramTarget");
		if (values.method === "manual_review") return $t("organizationOnboarding.manualTarget");
		return $t("organizationOnboarding.emailTarget");
	}

	function errorText(code, message) {
		const known = [
			"organization_name_required",
			"organization_official_email_invalid",
			"organization_onboarding_slug_taken",
			"organization_onboarding_claim_method_invalid",
			"organization_onboarding_claim_target_required",
			"organization_onboarding_claim_target_invalid",
			"organization_website_url_invalid",
			"session_expired",
		];
		if (known.includes(code)) {
			return $t(`organizationOnboarding.errors.${code}`);
		}
		return $t("organizationOnboarding.errors.generic", { message: message || code });
	}

	function submitForm() {
		pending = true;
		return async ({ update }) => {
			await update();
			pending = false;
		};
	}
</script>

<nav class="breadcrumbs">
	<a href="/">{$t("organizationDetail.home")}</a>
	<span>›</span>
	<a href="/organizations">{$t("nav.organizations")}</a>
	<span>›</span>
	<span>{$t("organizationOnboarding.breadcrumb")}</span>
</nav>

{#if data.unverified}
	<section class="section">
		<div class="error account-state-card" role="alert">
			<strong>{$t("account.errorTitle")}</strong>
			<p>{$t("account.errorBody")}</p>
			<a class="ghost-button" href="/organizations/new">{$t("account.errorRetry")}</a>
		</div>
	</section>
{:else if created}
	<section class="onboarding-success">
		<div>
			<p class="eyebrow">{$t("organizationOnboarding.successEyebrow")}</p>
			<h1>{$t("organizationOnboarding.successTitle")}</h1>
			<p class="muted">{$t("organizationOnboarding.successBody")}</p>
			<dl class="onboarding-meta">
				<div>
					<dt>{$t("organizationOnboarding.organizationLabel")}</dt>
					<dd>{created.organization?.name}</dd>
				</div>
				<div>
					<dt>{$t("organizationOnboarding.claimIdLabel")}</dt>
					<dd><code>{created.claim?.id}</code></dd>
				</div>
			</dl>
			<div class="inline-actions">
				<a class="btn primary" href="/account">{$t("organizationOnboarding.trackAccount")}</a>
				<a class="ghost-button" href="/organizations">{$t("organizationOnboarding.backToDirectory")}</a>
			</div>
		</div>
	</section>
{:else}
	<section class="page-heading onboarding-heading">
		<p class="eyebrow">{$t("organizationOnboarding.eyebrow")}</p>
		<h1 class="section-title">{$t("organizationOnboarding.title")}</h1>
		<p class="muted">{$t("organizationOnboarding.subtitle")}</p>
	</section>

	<section class="onboarding-layout">
		<div class="onboarding-steps" aria-label={$t("organizationOnboarding.stepsLabel")}>
			<article>
				<span>1</span>
				<div>
					<h2>{$t("organizationOnboarding.stepProfileTitle")}</h2>
					<p>{$t("organizationOnboarding.stepProfileBody")}</p>
				</div>
			</article>
			<article>
				<span>2</span>
				<div>
					<h2>{$t("organizationOnboarding.stepEvidenceTitle")}</h2>
					<p>{$t("organizationOnboarding.stepEvidenceBody")}</p>
				</div>
			</article>
			<article>
				<span>3</span>
				<div>
					<h2>{$t("organizationOnboarding.stepReviewTitle")}</h2>
					<p>{$t("organizationOnboarding.stepReviewBody")}</p>
				</div>
			</article>
		</div>

		<form
			class="auth-form onboarding-form"
			method="POST"
			action="?/createOrganization"
			aria-describedby="onboarding-form-status"
			use:enhance={submitForm}
		>
			<div class="form-section">
				<div>
					<p class="label">{$t("organizationOnboarding.profileSection")}</p>
					<p class="muted small">{$t("organizationOnboarding.profileSectionHelp")}</p>
				</div>
				<label>
					{$t("organizationOnboarding.name")}
					<input
						name="name"
						type="text"
						autocomplete="organization"
						bind:value={values.name}
						onblur={() => mark("name")}
						maxlength="160"
						required
						aria-invalid={fieldError === "name" || (touched.name && !nameValid) ? "true" : undefined}
					/>
					<span class:error-text={touched.name && !nameValid} class:muted={!(touched.name && !nameValid)}>
						{touched.name && !nameValid ? $t("organizationOnboarding.nameRequired") : $t("organizationOnboarding.nameHelp")}
					</span>
				</label>
				<label>
					{$t("organizationOnboarding.legalName")}
					<input name="legal_name" type="text" bind:value={values.legal_name} maxlength="180" />
				</label>
				<label>
					{$t("organizationOnboarding.description")}
					<textarea name="description" bind:value={values.description} rows="4" maxlength="600"></textarea>
					<span class="muted">{$t("organizationOnboarding.descriptionHelp")}</span>
				</label>
			</div>

			<div class="form-section compact-grid">
				<label>
					{$t("organizationOnboarding.country")}
					<input name="country" type="text" bind:value={values.country} />
				</label>
				<label>
					{$t("organizationOnboarding.region")}
					<input name="region" type="text" bind:value={values.region} />
				</label>
				<label>
					{$t("organizationOnboarding.city")}
					<input name="city" type="text" bind:value={values.city} />
				</label>
				<label>
					{$t("organizationOnboarding.website")}
					<input
						name="website_url"
						type="url"
						inputmode="url"
						bind:value={values.website_url}
						onblur={() => mark("website_url")}
						placeholder="https://"
						aria-invalid={fieldError === "website_url" || (touched.website_url && !websiteValid) ? "true" : undefined}
					/>
					<span class:error-text={touched.website_url && !websiteValid} class:muted={!(touched.website_url && !websiteValid)}>
						{touched.website_url && !websiteValid ? $t("organizationOnboarding.websiteInvalid") : $t("organizationOnboarding.websiteHelp")}
					</span>
				</label>
			</div>

			<div class="form-section">
				<div>
					<p class="label">{$t("organizationOnboarding.evidenceSection")}</p>
					<p class="muted small">{$t("organizationOnboarding.evidenceSectionHelp")}</p>
				</div>
				<label>
					{$t("organizationOnboarding.method")}
					<select name="method" bind:value={values.method}>
						<option value="official_email">{$t("organizationOnboarding.methodEmail")}</option>
						<option value="instagram">{$t("organizationOnboarding.methodInstagram")}</option>
						<option value="manual_review">{$t("organizationOnboarding.methodManual")}</option>
					</select>
				</label>
				<label>
					{methodTargetLabel()}
					<input
						name="target"
						type={values.method === "official_email" ? "email" : "text"}
						bind:value={values.target}
						onblur={() => mark("target")}
						required
						aria-invalid={fieldError === "target" || (touched.target && !targetValid) ? "true" : undefined}
					/>
					<span class:error-text={touched.target && !targetValid} class:muted={!(touched.target && !targetValid)}>
						{touched.target && !targetValid ? $t("organizationOnboarding.targetInvalid") : $t("organizationOnboarding.targetHelp")}
					</span>
				</label>
				<label>
					{$t("organizationOnboarding.officialEmail")}
					<input name="official_email" type="email" bind:value={values.official_email} />
					<span class="muted">{$t("organizationOnboarding.officialEmailHelp")}</span>
				</label>
				<label>
					{$t("organizationOnboarding.evidenceNote")}
					<textarea name="evidence_note" bind:value={values.evidence_note} rows="3" maxlength="500"></textarea>
				</label>
			</div>

			<div id="onboarding-form-status" class="form-status" aria-live="polite">
				{#if pending}
					<p class="notice compact">{$t("organizationOnboarding.submitting")}</p>
				{:else if errorCode}
					<p class="error compact">{errorText(errorCode, errorMessage)}</p>
				{/if}
			</div>

			<button class="btn primary" type="submit" disabled={!canSubmit} aria-busy={pending}>
				{pending ? $t("organizationOnboarding.submittingShort") : $t("organizationOnboarding.submit")}
			</button>
		</form>
	</section>
{/if}

<style>
	.onboarding-heading {
		max-width: 780px;
	}

	.onboarding-layout {
		display: grid;
		grid-template-columns: minmax(240px, 0.75fr) minmax(0, 1.45fr);
		gap: 24px;
		align-items: start;
		margin-top: 24px;
	}

	.onboarding-steps {
		display: grid;
		gap: 12px;
		position: sticky;
		top: 92px;
	}

	.onboarding-steps article {
		display: grid;
		grid-template-columns: 36px 1fr;
		gap: 12px;
		padding: 16px;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: var(--surface);
		box-shadow: var(--shadow-sm);
	}

	.onboarding-steps span {
		display: inline-grid;
		place-items: center;
		width: 32px;
		height: 32px;
		border-radius: 8px;
		background: var(--purple-100);
		color: var(--purple-800);
		font-weight: 800;
	}

	.onboarding-steps h2 {
		margin: 0 0 4px;
		font-size: 15px;
		line-height: 1.3;
	}

	.onboarding-steps p {
		margin: 0;
		color: var(--muted);
		font-size: 13px;
		line-height: 1.5;
	}

	.onboarding-form {
		padding: 24px;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: var(--surface);
		box-shadow: var(--shadow-md);
	}

	.form-section {
		display: grid;
		gap: 14px;
		padding-bottom: 18px;
		border-bottom: 1px solid var(--border);
	}

	.form-section:last-of-type {
		border-bottom: 0;
		padding-bottom: 0;
	}

	.compact-grid {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.onboarding-form textarea,
	.onboarding-form select {
		width: 100%;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: var(--surface);
		color: var(--text);
		font: inherit;
	}

	.onboarding-form textarea {
		min-height: 96px;
		padding: 11px 12px;
		resize: vertical;
	}

	.onboarding-form select {
		min-height: 44px;
		padding: 0 12px;
	}

	.onboarding-form span {
		font-size: 13px;
		text-transform: none;
		letter-spacing: 0;
	}

	.error-text {
		color: var(--state-danger-text);
	}

	.onboarding-success {
		margin-top: 32px;
		padding: 32px;
		border: 1px solid var(--state-success-border);
		border-radius: 8px;
		background: linear-gradient(135deg, var(--state-success-bg) 0%, var(--surface) 72%);
		box-shadow: var(--shadow-md);
	}

	.onboarding-success h1 {
		margin: 6px 0 10px;
		font-size: clamp(28px, 5vw, 48px);
		line-height: 1.05;
	}

	.onboarding-success .muted {
		max-width: 62ch;
	}

	.onboarding-meta {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 12px;
		margin: 20px 0;
		padding: 0;
	}

	.onboarding-meta div {
		min-width: 0;
		padding: 14px;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: var(--surface);
	}

	.onboarding-meta dt {
		margin: 0 0 4px;
		color: var(--muted);
		font-size: 11px;
		font-weight: 800;
		text-transform: uppercase;
	}

	.onboarding-meta dd {
		margin: 0;
		font-weight: 800;
		word-break: break-word;
	}

	.inline-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 12px;
	}

	@media (max-width: 900px) {
		.onboarding-layout {
			grid-template-columns: 1fr;
		}

		.onboarding-steps {
			position: static;
		}
	}

	@media (max-width: 640px) {
		.compact-grid,
		.onboarding-meta {
			grid-template-columns: 1fr;
		}

		.onboarding-form,
		.onboarding-success {
			padding: 20px;
		}
	}
</style>
