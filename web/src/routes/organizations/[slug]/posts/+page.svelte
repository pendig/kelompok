<script>
	import { fallbackDate } from "../../../../lib/api.js";
	import { locale, t } from "$lib/i18n.js";

	let { data } = $props();
	let org = $derived(data.organization);

	function postPath(post) {
		return `/organizations/${encodeURIComponent(org.slug)}/posts/${encodeURIComponent(post.slug)}`;
	}
</script>

<nav class="breadcrumbs">
	<a href="/">{$t("organizationDetail.home")}</a>
	<span>›</span>
	<a href="/organizations">{$t("nav.organizations")}</a>
	<span>›</span>
	<a href={`/organizations/${encodeURIComponent(org.slug)}`}>{org.name}</a>
	<span>›</span>
	<span>{$t("nav.posts")}</span>
</nav>

<h1 class="section-title">{$t("organizationPostsPage.title")}</h1>
<p class="muted">{org.name}</p>

{#if data.posts.length === 0}
	<p class="empty">{$t("organizationPostsPage.empty")}</p>
{:else}
	<div>
		{#each data.posts as post}
			<div class="list-item">
				<a class="title" href={postPath(post)}>{post.title}</a>
				<div class="meta">{fallbackDate(post.published_at, $locale)} · {post.summary || "—"}</div>
			</div>
		{/each}
	</div>
{/if}
