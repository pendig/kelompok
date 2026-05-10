<script>
	import { fallbackDate } from "../../lib/api.js";
	import { locale, t } from "$lib/i18n.js";

	let { data } = $props();
</script>

<section class="page-heading">
	<p class="eyebrow">{$t("nav.posts")}</p>
	<h1 class="section-title">{$t("postsPage.title")}</h1>
</section>

{#if data.loadError}
	<p class="error">{$t("postsPage.loadError")}</p>
{/if}

{#if data.posts.length === 0}
	<p class="empty">{$t("postsPage.empty")}</p>
{:else}
	<div>
		{#each data.posts as post}
			<div class="list-item">
				<a class="title" href={`/posts/${post.slug}`}>{post.title}</a>
				<p class="meta">
					{$t("postsPage.by", { name: post.organization?.name || $t("postsPage.unknownAuthor") })}
					· {fallbackDate(post.published_at, $locale)}
				</p>
				<p class="small">{post.summary || "—"}</p>
			</div>
		{/each}
	</div>
{/if}
