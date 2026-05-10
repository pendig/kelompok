<script>
	import { page } from "$app/state";
</script>

<div class="container">
	<h1 class="section-title">Aduh, sesuatu nggak beres</h1>
	<p class="meta">Status: {page.status}</p>
	<p class="small">{page.error?.message || "Terjadi error saat memuat data."}</p>
</div>

