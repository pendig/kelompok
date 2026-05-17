<script>
	import { t } from "$lib/i18n.js";

	let { data } = $props();

	let user = $derived(data.session.user);
	let roles = $derived(data.session.organization_roles ?? []);
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
