<script>
	import { locale } from "$lib/i18n.js";

	let { status = "unclaimed", tone, label, size = "default" } = $props();

	const labels = {
		id: {
			active: "Aktif",
			admin: "Admin",
			approved: "Disetujui",
			claimed: "Terklaim",
			draft: "Draft",
			editor: "Editor",
			fail: "Gagal",
			owner: "Owner",
			pass: "Lolos",
			pending: "Menunggu review",
			published: "Terbit",
			rejected: "Ditolak",
			unclaimed: "Belum diklaim",
			verified: "Terverifikasi",
			viewer: "Viewer",
			warn: "Perlu cek",
		},
		en: {
			active: "Active",
			admin: "Admin",
			approved: "Approved",
			claimed: "Claimed",
			draft: "Draft",
			editor: "Editor",
			fail: "Fail",
			owner: "Owner",
			pass: "Pass",
			pending: "Pending review",
			published: "Published",
			rejected: "Rejected",
			unclaimed: "Unclaimed",
			verified: "Verified",
			viewer: "Viewer",
			warn: "Check",
		},
	};

	const successStatuses = ["active", "admin", "approved", "claimed", "owner", "pass", "published", "verified"];
	const warningStatuses = ["draft", "pending", "unclaimed", "warn"];

	let normalizedStatus = $derived(String(status || "unclaimed").toLowerCase());
	let resolvedTone = $derived(
		tone || (successStatuses.includes(normalizedStatus) ? "success" : warningStatuses.includes(normalizedStatus) ? "warning" : "danger"),
	);
	let resolvedLabel = $derived(
		label || labels[$locale]?.[normalizedStatus] || normalizedStatus.replace(/[_-]+/g, " "),
	);
	let classes = $derived(
		["status-badge", `status-badge--${resolvedTone}`, size !== "default" ? `status-badge--${size}` : ""]
			.filter(Boolean)
			.join(" "),
	);
</script>

<span class={classes} data-status={normalizedStatus}>{resolvedLabel}</span>
