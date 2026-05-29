<script>
	import AdminWorkspace from "../admin/+page.svelte";

	let { data, form } = $props();
</script>

<AdminWorkspace {data} {form} />
