<script>
	import { fallbackDate } from "../../../lib/api.js";

	let { data } = $props();

	const org = () => data.organization;
	const parsedProfile = () => org()?.profile_data || {};
	const parsedSDGs = () => org()?.sdgs_data || {};
	const publicContact = () => {
		const value = parsedProfile()?.public_contact;
		if (typeof value === "string") {
			return value;
		}
		if (value) {
			return JSON.stringify(value);
		}
		return "";
	};
</script>

<nav class="breadcrumbs">
	<a href="/">Home</a>
	<span>›</span>
	<a href="/organizations">Organizations</a>
	<span>›</span>
	<span>{org().name}</span>
</nav>

<header class="two-col" style="margin-top: 1rem">
	<section class="card">
		<h1>{org().name}</h1>
		<p class="muted">Slug: <span class="code">{org().slug}</span></p>
		<p>{org().description || "Tidak ada deskripsi publik."}</p>
		<p class="small muted">Tagline: {org().legal_name || "—"}</p>
		<p class="small">Perbarui: {fallbackDate(org().updated_at)}</p>
	</section>

	<section class="card">
		<div class="label">Informasi</div>
		<p class="value"><strong>Lokasi:</strong> {org().city || "—"}{#if org().region}, {org().region}{/if}{#if org().country} · {org().country}{/if}</p>
		<p class="value"><strong>Website:</strong> {#if org().website_url}<a href={org().website_url} target="_blank" rel="noreferrer">{org().website_url}</a>{:else}—{/if}</p>
		<p class="value"><strong>Claim:</strong> {org().claim_status}</p>
		{#if parsedProfile()?.languages?.length}
			<div class="label">Bahasa</div>
			<div class="pill-row">
				{#each parsedProfile().languages as language}
					<span class="pill">{language}</span>
				{/each}
			</div>
		{/if}
	</section>
</header>

<section>
	<div class="actions">
		<a href={`/organizations/${org().slug}/posts`}>Semua Post</a>
		<a href={`/organizations/${org().slug}/impact`}>Laporan Dampak</a>
	</div>

	<h2 class="section-title">Visi dan misi</h2>
	<div class="grid">
		<div class="card">
			<div class="label">History</div>
			<p class="small">{org().history || "Belum diisi."}</p>
		</div>
		<div class="card">
			<div class="label">Kontak publik</div>
			<p class="small">{publicContact() || "Informasi kontak belum diisi."}</p>
		</div>
	</div>
</section>

<section>
	<h2 class="section-title">SDGs</h2>
	<div class="grid">
		<div class="card">
			<div class="label">Fokus utama</div>
			{#if parsedSDGs()?.primary?.length}
				<div class="pill-row">
					{#each parsedSDGs().primary as goal}
						<span class="pill">{goal}</span>
					{/each}
				</div>
			{:else}
				<p class="small">Belum ada data SDGs.</p>
			{/if}
		</div>
		<div class="card">
			<div class="label">Program</div>
			{#if parsedProfile()?.programs?.length}
				<ul>
					{#each parsedProfile().programs as item}
						<li>{item}</li>
					{/each}
				</ul>
			{:else}
				<p class="small">Program belum diisi.</p>
			{/if}
		</div>
	</div>
</section>

<section>
	<h2 class="section-title">Postingan Terbaru</h2>
	{#if data.posts.length === 0}
		<p class="empty">Belum ada post publik.</p>
	{:else}
		{#each data.posts.slice(0, 6) as post}
			<div class="list-item">
				<a class="title" href={`/posts/${post.slug}`}>{post.title}</a>
				<div class="meta">
					{fallbackDate(post.published_at)} · {post.summary || "—"}
				</div>
			</div>
		{/each}
	{/if}
</section>

<section>
	<h2 class="section-title">Laporan Dampak</h2>
	{#if data.impactReports.length === 0}
		<p class="empty">Belum ada laporan publik.</p>
	{:else}
		{#each data.impactReports.slice(0, 6) as report}
			<div class="list-item">
				<div class="title">{report.title}</div>
				<div class="meta">{fallbackDate(report.published_at)} · {report.summary || "—"}</div>
			</div>
		{/each}
	{/if}
</section>
