<script>
	import { t } from "$lib/i18n.js";

	let { data, form } = $props();

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
</script>

<section class="auth-page">
	<div class="auth-card">
		<div>
			<p class="eyebrow">{$t("auth.loginEyebrow")}</p>
			<h1 class="section-title">{$t("auth.loginTitle")}</h1>
			<p class="section-note">{$t("auth.loginSubtitle")}</p>
		</div>

		{#if form?.ok === false}
			<p class="error compact" role="alert">{loginErrorMessage()}</p>
		{/if}

		<form class="auth-form" method="POST">
			<input type="hidden" name="return_to" value={data.returnTo} />
			<label>
				{$t("auth.email")}
				<input name="email" type="email" autocomplete="email" placeholder="you@example.org" required />
			</label>
			<label>
				{$t("auth.password")}
				<input name="password" type="password" autocomplete="current-password" required />
			</label>
			<button class="btn primary" type="submit">{$t("auth.login")}</button>
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
