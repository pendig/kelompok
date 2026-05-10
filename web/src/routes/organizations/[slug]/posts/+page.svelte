<script>
	import { fallbackDate } from "../../../../lib/api.js";

	let { data } = $props();
	const org = () => data.organization;
</script>

<nav class="breadcrumbs">
	<a href="/">Home</a>
	<span>›</span>
	<a href="/organizations">Organizations</a>
	<span>›</span>
	<a href={`/organizations/${org().slug}`}>{org().name}</a>
	<span>›</span>
	<span>Posts</span>
</nav>

<h1 class="section-title">Postingan Organisasi</h1>
<p class="muted">{org().name}</p>

{#if data.posts.length === 0}
	<p class="empty">Belum ada postingan publik.</p>
{:else}
	<div>
		{#each data.posts as post}
			<div class="list-item">
				<a class="title" href={`/posts/${post.slug}`}>{post.title}</a>
				<div class="meta">{fallbackDate(post.published_at)} · {post.summary || "—"}</div>
			</div>
		{/each}
	</div>
{/if}
