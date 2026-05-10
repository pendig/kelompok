<script>
	import { fallbackDate } from "../../../lib/api.js";

	let { data } = $props();
	const post = () => data.post;
	const org = () => post().organization;
</script>

<nav class="breadcrumbs">
	<a href="/">Home</a>
	<span>›</span>
	<a href="/posts">Articles</a>
	<span>›</span>
	<span>{post().title}</span>
</nav>

<article>
	<h1>{post().title}</h1>
	<p class="meta">By {org().name} · Dipublikasi {fallbackDate(post().published_at)}</p>
	<p class="small">{post().summary || "Tidak ada summary."}</p>
	{#if post().content}
		<p>{post().content}</p>
	{:else}
		<p class="empty">Konten belum ditambahkan.</p>
	{/if}
</article>
