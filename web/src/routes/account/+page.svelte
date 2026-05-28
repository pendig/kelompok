<script>
	import { untrack } from "svelte";
	import { t } from "$lib/i18n.js";

	let { data, form } = $props();

	let user = $derived(data.session.user);
	let roles = $derived(data.session.organization_roles ?? []);

	let nameInput = $state(untrack(() => data.session.user.name));
	$effect(() => {
		nameInput = user.name;
	});

	let isUpdateProfile = $derived(form?.action === "updateProfile");
	let updateSuccess = $derived(isUpdateProfile && form?.ok === true);
	let updateError = $derived(isUpdateProfile && form?.ok === false ? form.error : null);
	let updateErrorCode = $derived(isUpdateProfile && form?.ok === false ? form.code : null);

	function profileErrorMessage(code, message) {
		if (!code && !message) return null;
		const known = [
			"name_required",
			"profile_name_required",
			"profile_name_too_long",
			"session_expired",
		];
		if (known.includes(code)) {
			return $t(`account.errors.${code}`);
		}
		return $t("account.errors.generic", { message: message || code });
	}
</script>

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

<section class="account-profile">
	<div class="section-head single">
		<div>
			<p class="eyebrow">{$t("account.profileEyebrow")}</p>
			<h2 class="section-title">{$t("account.profileTitle")}</h2>
			<p class="section-note">{$t("account.profileSubtitle")}</p>
		</div>
	</div>

	<form
		class="auth-form profile-form"
		method="POST"
		action="?/updateProfile"
		aria-describedby="profile-form-status"
	>
		<label>
			{$t("account.profileName")}
			<input
				name="name"
				type="text"
				autocomplete="name"
				bind:value={nameInput}
				maxlength="120"
				required
				aria-invalid={updateError ? "true" : undefined}
			/>
			<span class="form-help muted">{$t("account.profileNameHelp")}</span>
		</label>

		<label>
			{$t("account.profileEmail")}
			<input type="email" value={user.email} disabled readonly aria-readonly="true" />
			<span class="form-help muted">{$t("account.profileEmailHelp")}</span>
		</label>

		<div id="profile-form-status" class="form-status" aria-live="polite">
			{#if updateSuccess}
				<p class="success compact">{$t("account.profileUpdated")}</p>
			{:else if updateError}
				<p class="error compact">{profileErrorMessage(updateErrorCode, updateError)}</p>
			{/if}
		</div>

		<button class="btn primary" type="submit">{$t("account.profileSave")}</button>
	</form>
</section>

<section>
	<h2 class="section-title">{$t("account.organizations")}</h2>
	{#if roles.length === 0}
		<div class="empty account-empty">
			<strong>{$t("account.noOrganizationsTitle")}</strong>
			<p>{$t("account.noOrganizationsBody")}</p>
			<a href="/organizations" class="ghost-button">{$t("account.findOrganization")}</a>
		</div>
	{:else}
		<div class="account-role-list">
			{#each roles as role}
				<article class="account-role">
					<div>
						<p class="label">{role.role}</p>
						<h3>{role.organization_name}</h3>
						<p class="muted">/{role.organization_slug}</p>
					</div>
					<div class="inline-actions">
						<a class="ghost-button" href={`/organizations/${encodeURIComponent(role.organization_slug)}`}>
							{$t("account.viewPublic")}
						</a>
						<a class="btn primary" href={`/admin?org=${encodeURIComponent(role.organization_slug)}`}>
							{$t("account.manage")}
						</a>
					</div>
				</article>
			{/each}
		</div>
	{/if}
</section>

<style>
	.account-profile {
		margin-top: 32px;
	}
	.profile-form {
		max-width: 520px;
	}
	.profile-form input[disabled] {
		background: var(--surface-soft);
		color: var(--muted);
		cursor: not-allowed;
	}
	.form-help {
		display: block;
		margin-top: 6px;
		font-size: 13px;
	}
	.form-status {
		min-height: 24px;
	}
	.form-status .success {
		color: hsl(150, 60%, 30%);
		background: hsl(150, 60%, 96%);
		border: 1px solid hsl(150, 50%, 80%);
		padding: 8px 12px;
		border-radius: 8px;
		margin: 0;
	}
</style>
