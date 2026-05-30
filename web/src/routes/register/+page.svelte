<script>
	import { enhance } from "$app/forms";
	import { t } from "$lib/i18n.js";

	let { form } = $props();
	let pending = $state(false);
	let name = $state("");
	let email = $state("");
	let password = $state("");

	let passwordTouched = $state(false);
	let nameTouched = $state(false);
	let emailTouched = $state(false);

	let passwordValid = $derived(password.length >= 8);
	let nameValid = $derived(name.trim().length > 0);
	let emailValid = $derived(/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()));

	const FRIENDLY_CODES = new Set([
		"user_exists",
		"name_required",
	]);

	function registerErrorMessage() {
		if (!form?.error && !form?.code) return null;
		const code = form?.code;
		if (code && FRIENDLY_CODES.has(code)) {
			return $t(`auth.errors.${code}`);
		}
		return $t("auth.error", { message: form?.error || code || "" });
	}

	function submitRegister() {
		pending = true;
		return async ({ update }) => {
			await update();
			pending = false;
		};
	}
</script>

<section class="auth-page">
	<div class="auth-card">
		<div>
			<p class="eyebrow">{$t("auth.registerEyebrow")}</p>
			<h1 class="section-title">{$t("auth.registerTitle")}</h1>
			<p class="section-note">{$t("auth.registerSubtitle")}</p>
		</div>

		{#if form?.ok === false}
			<p class="form-banner error compact" role="alert">{registerErrorMessage()}</p>
		{:else if pending}
			<p class="form-banner notice compact" role="status" aria-live="polite">{$t("auth.registerPending")}</p>
		{/if}

		<form class="auth-form" method="POST" use:enhance={submitRegister}>
			<label>
				{$t("auth.name")}
				<input
					name="name"
					autocomplete="name"
					placeholder="Kelompok Admin"
					bind:value={name}
					onblur={() => nameTouched = true}
					aria-invalid={nameTouched && !nameValid ? "true" : undefined}
					required
				/>
				{#if nameTouched && !nameValid}
					<span class="form-help error-text">{$t("auth.nameRequiredHelp")}</span>
				{/if}
			</label>
			<label>
				{$t("auth.email")}
				<input
					name="email"
					type="email"
					autocomplete="email"
					placeholder="you@example.org"
					bind:value={email}
					onblur={() => emailTouched = true}
					aria-invalid={emailTouched && !emailValid ? "true" : undefined}
					required
				/>
				{#if emailTouched && !emailValid}
					<span class="form-help error-text">{$t("auth.emailInvalidHelp")}</span>
				{/if}
			</label>
			<label>
				{$t("auth.password")}
				<input
					name="password"
					type="password"
					autocomplete="new-password"
					minlength="8"
					bind:value={password}
					onblur={() => passwordTouched = true}
					aria-describedby="password-help"
					aria-invalid={passwordTouched && !passwordValid ? "true" : undefined}
					required
				/>
				<span id="password-help" class:success-text={passwordValid} class:error-text={passwordTouched && !passwordValid} class="form-help">
					{$t("auth.passwordHelp")}
				</span>
			</label>
			<button class="btn primary" type="submit" disabled={pending} aria-busy={pending}>
				{pending ? $t("auth.registerPendingShort") : $t("auth.register")}
			</button>
		</form>

		<p class="auth-switch">
			{$t("auth.hasAccount")}
			<a href="/login">{$t("auth.signIn")}</a>
		</p>
	</div>

	<aside class="auth-side">
		<p class="eyebrow">{$t("auth.onboardingEyebrow")}</p>
		<h2>{$t("auth.onboardingTitle")}</h2>
		<p>{$t("auth.onboardingBody")}</p>
	</aside>
</section>
