<script>
	import { t } from "$lib/i18n.js";
</script>

<div class="auth-page" style="justify-content: center; align-items: center; min-height: 50vh;">
	<div class="auth-card" style="text-align: center; padding: 48px 32px; max-width: 400px; margin: 0 auto;">
		<p class="eyebrow">{$t("auth.loginEyebrow")}</p>
		<h1 class="section-title">{$t("auth.loginPendingShort")}</h1>
		<p>{$t("auth.loginPending")}</p>
		<div class="spinner"></div>
	</div>
</div>

<style>
	.spinner {
		width: 40px;
		height: 40px;
		margin: 24px auto 0;
		border: 3px solid var(--border);
		border-top-color: var(--purple-600);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
</style>
