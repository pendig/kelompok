<script>
	import { t } from "$lib/i18n.js";

	let { form } = $props();
</script>

<section class="auth-page">
	<div class="auth-card">
		<div>
			<p class="eyebrow">{$t("auth.registerEyebrow")}</p>
			<h1 class="section-title">{$t("auth.registerTitle")}</h1>
			<p class="section-note">{$t("auth.registerSubtitle")}</p>
		</div>

		{#if form?.error}
			<p class="error compact">{$t("auth.error", { message: form.error })}</p>
		{/if}

		<form class="auth-form" method="POST">
			<label>
				{$t("auth.name")}
				<input name="name" autocomplete="name" placeholder="Kelompok Admin" />
			</label>
			<label>
				{$t("auth.email")}
				<input name="email" type="email" autocomplete="email" placeholder="you@example.org" required />
			</label>
			<label>
				{$t("auth.password")}
				<input name="password" type="password" autocomplete="new-password" minlength="8" required />
			</label>
			<button class="btn primary" type="submit">{$t("auth.register")}</button>
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
