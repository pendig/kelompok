<script>
	import { enhance } from "$app/forms";
	import { t } from "$lib/i18n.js";

	let { data, form } = $props();
	let pending = $state(false);
	let email = $state("");
	let password = $state("");
	let emailTouched = $state(false);
	let passwordTouched = $state(false);

	let emailValid = $derived(/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()));
	let passwordValid = $derived(password.length > 0);

	const FRIENDLY_CODES = new Set([
		"invalid_credentials",
		"session_invalid",
		"session_required",
		"login_failed",
	]);

	function loginErrorMessage() {
		if (!form?.error && !form?.code) return null;
		const code = form?.code;
		if (code && FRIENDLY_CODES.has(code)) {
			return $t(`auth.errors.${code}`);
		}
		return $t("auth.error", { message: form?.error || code || "" });
	}

	function emailHelp() {
		if (!emailTouched || emailValid) return null;
		return email.trim().length === 0 ? $t("auth.emailRequiredHelp") : $t("auth.emailInvalidHelp");
	}

	function submitLogin({ cancel }) {
		emailTouched = true;
		passwordTouched = true;
		if (!emailValid || !passwordValid) {
			cancel();
			return;
		}
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
			<p class="eyebrow">{$t("auth.loginEyebrow")}</p>
			<h1 class="section-title">{$t("auth.loginTitle")}</h1>
			<p class="section-note">{$t("auth.loginSubtitle")}</p>
		</div>

		{#if form?.ok === false}
			<p class="form-banner error compact" role="alert">{loginErrorMessage()}</p>
		{:else if pending}
			<p class="form-banner notice compact" role="status" aria-live="polite">{$t("auth.loginPending")}</p>
		{/if}

		<form class="auth-form" method="POST" use:enhance={submitLogin} novalidate>
			<input type="hidden" name="return_to" value={data.returnTo} />
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
					aria-describedby={emailHelp() ? "login-email-help" : undefined}
				/>
				{#if emailHelp()}
					<span id="login-email-help" class="form-help error-text">{emailHelp()}</span>
				{/if}
			</label>
			<label>
				{$t("auth.password")}
				<input
					name="password"
					type="password"
					autocomplete="current-password"
					bind:value={password}
					onblur={() => passwordTouched = true}
					aria-invalid={passwordTouched && !passwordValid ? "true" : undefined}
					aria-describedby={passwordTouched && !passwordValid ? "login-password-help" : undefined}
				/>
				{#if passwordTouched && !passwordValid}
					<span id="login-password-help" class="form-help error-text">{$t("auth.passwordRequiredHelp")}</span>
				{/if}
			</label>
			<button class="btn primary" type="submit" disabled={pending} aria-busy={pending}>
				{pending ? $t("auth.loginPendingShort") : $t("auth.login")}
			</button>
		</form>

		<p class="auth-switch">
			{$t("auth.noAccount")}
			<a href="/register">{$t("auth.createAccount")}</a>
		</p>
	</div>

	<aside class="auth-side">
		<p class="eyebrow">{$t("auth.claimPathEyebrow")}</p>
		<h2>{$t("auth.claimPathTitle")}</h2>
		<p>{$t("auth.claimPathBody")}</p>
	</aside>
</section>
