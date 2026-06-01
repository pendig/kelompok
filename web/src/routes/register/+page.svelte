<script>
	import { enhance } from "$app/forms";
	import { t } from "$lib/i18n.js";

	let { data, form } = $props();
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
		const errorVal = form?.error || data?.error;
		const code = form?.code || data?.error;
		if (!errorVal && !code) return null;
		if (code && FRIENDLY_CODES.has(code)) {
			return $t(`auth.errors.${code}`);
		}
		return $t("auth.error", { message: errorVal || code || "" });
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

		{#if form?.ok === false || data?.error}
			<p class="form-banner error compact" role="alert">{registerErrorMessage()}</p>
		{:else if pending}
			<p class="form-banner notice compact" role="status" aria-live="polite">{$t("auth.registerPending")}</p>
		{/if}

		{#if data?.googleClientId}
			<a href="/login/google?return_to=/account" class="google-login-btn">
				<svg class="google-icon" viewBox="0 0 24 24" width="18" height="18">
					<path fill="#EA4335" d="M12 5.04c1.62 0 3.08.56 4.22 1.65l3.15-3.15C17.45 1.8 14.95 1 12 1 7.37 1 3.4 3.66 1.48 7.54l3.75 2.91C6.11 7.5 8.84 5.04 12 5.04z"/>
					<path fill="#4285F4" d="M23.49 12.27c0-.81-.07-1.59-.2-2.35H12v4.47h6.44c-.28 1.47-1.11 2.71-2.36 3.55l3.66 2.84c2.14-1.97 3.75-4.87 3.75-8.51z"/>
					<path fill="#FBBC05" d="M5.23 10.45c-.24-.72-.38-1.49-.38-2.29s.14-1.57.38-2.29L1.48 2.96C.54 4.85 0 6.99 0 9.25s.54 4.4 1.48 6.29l3.75-2.91c-.24-.72-.38-1.49-.38-2.29z"/>
					<path fill="#34A853" d="M12 18.96c-3.16 0-5.89-2.46-6.77-5.41L1.48 16.46C3.4 20.34 7.37 23 12 23c2.95 0 5.45-1 7.28-2.73l-3.66-2.84c-1 .67-2.28 1.53-3.62 1.53z"/>
				</svg>
				<span>{$t("auth.loginWithGoogle")}</span>
			</a>
			
			<div class="auth-divider">
				<span>{$t("auth.or")}</span>
			</div>
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

		<p class="auth-terms">{$t("auth.agreePrefix")} <a href="/terms">{$t("auth.agreeLink")}</a>.</p>
	</div>

	<aside class="auth-side">
		<p class="eyebrow">{$t("auth.onboardingEyebrow")}</p>
		<h2>{$t("auth.onboardingTitle")}</h2>
		<p>{$t("auth.onboardingBody")}</p>
	</aside>
</section>

<style>
	.google-login-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		width: 100%;
		background-color: var(--surface);
		color: var(--text);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 12px 16px;
		font-size: 15px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
		text-decoration: none;
		box-shadow: var(--shadow-sm);
	}

	.google-login-btn:hover {
		background-color: var(--surface-soft);
		border-color: var(--border-strong);
		transform: translateY(-1px);
		box-shadow: var(--shadow-md);
	}

	.google-login-btn:active {
		transform: translateY(0);
	}

	.google-icon {
		flex-shrink: 0;
	}

	.auth-divider {
		display: flex;
		align-items: center;
		text-align: center;
		margin: 8px 0;
		color: var(--muted);
		font-size: 13px;
		font-weight: 500;
	}

	.auth-divider::before,
	.auth-divider::after {
		content: '';
		flex: 1;
		border-bottom: 1px solid var(--border);
	}

	.auth-divider:not(:empty)::before {
		margin-right: 16px;
	}

	.auth-divider:not(:empty)::after {
		margin-left: 16px;
	}
</style>
