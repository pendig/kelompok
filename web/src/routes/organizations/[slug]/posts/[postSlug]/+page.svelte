<script>
	import { fallbackDate } from "../../../../../lib/api.js";
	import { locale, t } from "$lib/i18n.js";

	let { data } = $props();

	let post = $derived(data.post);
	let org = $derived(post.organization);
</script>

<nav class="breadcrumbs">
	<a href="/">{$t("organizationDetail.home")}</a>
	<span>›</span>
	<a href="/organizations">{$t("nav.organizations")}</a>
	<span>›</span>
	<a href={`/organizations/${encodeURIComponent(org.slug)}`}>{org.name}</a>
	<span>›</span>
	<span>{post.title}</span>
</nav>

<article class="card detail-article">
	<p class="eyebrow">{$t("postDetail.eyebrow")}</p>
	<h1>{post.title}</h1>
	<p class="meta">
		{$t("postDetail.meta", {
			name: org.name,
			date: fallbackDate(post.published_at, $locale),
		})}
	</p>
	<p class="small">{post.summary || $t("postDetail.noSummary")}</p>
	{#if post.content}
		<p>{post.content}</p>
	{:else}
		<p class="empty">{$t("postDetail.noContent")}</p>
	{/if}
</article>
