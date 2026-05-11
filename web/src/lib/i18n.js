import { browser } from "$app/environment";
import { derived, writable } from "svelte/store";

export const languages = [
	{ code: "id", label: "ID" },
	{ code: "en", label: "EN" },
];

const dictionaries = {
	id: {
		nav: {
			organizations: "Organisasi",
			posts: "Artikel",
			source: "GitHub",
		},
		home: {
			eyebrow: "CRM terbuka untuk organisasi",
			title: "Kelompok",
			subtitle:
				"The Solutions of Movement. Profil publik, artikel, dan laporan dampak untuk organisasi yang ingin bergerak lebih rapi.",
			primaryAction: "Lihat organisasi",
			secondaryAction: "Baca artikel",
			organizationsLoaded: "organisasi dimuat",
			postsLoaded: "artikel dimuat",
			mvpLabel: "MVP",
			mvpDesc: "profil, artikel, dampak",
			previewEyebrow: "Profil publik",
			previewTitle: "Halaman organisasi yang bisa diklaim",
			previewChip: "Tema ungu",
			claimLabel: "Verifikasi klaim",
			claimValue: "Email resmi atau Instagram organisasi",
			claimSignals: ["Klaim email resmi", "Klaim Instagram", "API + CLI siap"],
			noticeTitle: "Data publik sedang dimuat",
			noticeBody:
				"Beberapa data belum bisa dimuat dari API. Halaman tetap aktif dan akan menampilkan data saat koneksi pulih.",
			scopeEyebrow: "Ruang lingkup MVP",
			scopeTitle: "Sederhana dulu, mudah dikembangkan",
			scopeNote:
				"Fokus awalnya adalah CRM organisasi yang enak dipakai, lalu integrasi dan plugin bisa tumbuh di atas fondasi yang sama.",
			features: [
				{
					title: "Profil publik",
					copy:
						"Landing page organisasi dengan sejarah, ketua, kontak, email resmi, lokasi, dan metadata SDGS.",
				},
				{
					title: "Artikel dan berita",
					copy:
						"Update, pengumuman, dan cerita dampak yang langsung terhubung ke profil organisasi.",
				},
				{
					title: "Laporan dampak",
					copy:
						"Pelaporan publik yang tetap terstruktur, tapi fleksibel untuk data dinamis organisasi.",
				},
			],
			liveEyebrow: "Data publik",
			liveTitle: "Organisasi dan artikel terbaru",
			liveNote: "Bagian ini mengambil data dari API lokal agar tampilan tetap dekat dengan data nyata.",
			organizations: "Organisasi",
			posts: "Artikel",
			total: "{count} total",
			noOrganizations: "Belum ada organisasi publik yang terdaftar.",
			noPosts: "Belum ada artikel.",
			noDescription: "Belum ada deskripsi publik.",
			unknownLocation: "Lokasi belum diatur",
			unknownAuthor: "Tidak diketahui",
			by: "Oleh {name}",
			calloutEyebrow: "Dibangun untuk kolaborasi",
			calloutTitle: "Satu core, banyak kebutuhan.",
			calloutBody:
				"Core tetap fokus pada CRM organisasi. API, CLI, dan plugin bisa bertambah tanpa membuat pengalaman publik menjadi rumit.",
			sourceAction: "Lihat kode",
		},
		organizationsPage: {
			title: "Organisasi",
			description: "Daftar profil publik organisasi yang sudah tersedia.",
			loadError: "Data organisasi belum bisa dimuat dari API. Coba refresh beberapa saat lagi.",
			empty: "Belum ada organisasi publik.",
			noDescription: "Belum ada deskripsi publik.",
			unknownLocation: "Lokasi belum diatur",
		},
		postsPage: {
			title: "Artikel Publik",
			loadError: "Data artikel belum bisa dimuat dari API. Coba refresh beberapa saat lagi.",
			empty: "Belum ada artikel publik saat ini.",
			unknownAuthor: "Tidak diketahui",
			by: "Oleh {name}",
		},
		organizationDetail: {
			info: "Informasi",
			location: "Lokasi",
			website: "Website",
			claim: "Claim",
			languages: "Bahasa",
			allPosts: "Semua artikel",
			impactReports: "Laporan dampak",
			vision: "Profil organisasi",
			history: "Sejarah",
			publicContact: "Kontak publik",
			noDescription: "Belum ada deskripsi publik.",
			noHistory: "Belum diisi.",
			noContact: "Informasi kontak belum diisi.",
			noSdgs: "Belum ada data SDGS.",
			noPrograms: "Program belum diisi.",
			noPosts: "Belum ada artikel publik.",
			noReports: "Belum ada laporan publik.",
			focus: "Fokus utama",
			programs: "Program",
			recentPosts: "Artikel terbaru",
			updatedAt: "Diperbarui {date}",
			tagline: "Nama legal",
			unknownLocation: "Lokasi belum diatur",
			home: "Beranda",
		},
		postDetail: {
			eyebrow: "Artikel",
			meta: "Oleh {name} · Dipublikasi {date}",
			noSummary: "Belum ada ringkasan.",
			noContent: "Konten belum ditambahkan.",
		},
		organizationPostsPage: {
			title: "Artikel organisasi",
			empty: "Belum ada artikel publik.",
		},
		impactPage: {
			title: "Laporan dampak",
			empty: "Belum ada laporan publik.",
			period: "Periode",
			noSummary: "Belum ada ringkasan.",
			status: "Status",
			publicAt: "Publik",
		},
	},
	en: {
		nav: {
			organizations: "Organizations",
			posts: "Posts",
			source: "GitHub",
		},
		home: {
			eyebrow: "Open-source CRM for organizations",
			title: "Kelompok",
			subtitle:
				"The Solutions of Movement. Public profiles, posts, and impact reports for organizations that want cleaner movement infrastructure.",
			primaryAction: "Explore organizations",
			secondaryAction: "Read posts",
			organizationsLoaded: "organizations loaded",
			postsLoaded: "posts loaded",
			mvpLabel: "MVP",
			mvpDesc: "profiles, posts, impact",
			previewEyebrow: "Public profile",
			previewTitle: "Claimable organization pages",
			previewChip: "Purple theme",
			claimLabel: "Claim verification",
			claimValue: "Official email or organization Instagram",
			claimSignals: ["Official email claim", "Instagram claim", "API + CLI ready"],
			noticeTitle: "Public data is loading",
			noticeBody:
				"Some data could not be loaded from the API. The page remains available and will show data when the connection recovers.",
			scopeEyebrow: "MVP scope",
			scopeTitle: "Simple first, extensible later",
			scopeNote:
				"The first focus is a usable organization CRM, with integrations and plugins growing on top of the same foundation.",
			features: [
				{
					title: "Public profiles",
					copy:
						"Organization landing pages with history, leadership, contacts, official email, location, and SDGS metadata.",
				},
				{
					title: "Posts and articles",
					copy: "Updates, announcements, and impact stories connected directly to organization profiles.",
				},
				{
					title: "Impact reports",
					copy: "Public reporting that stays structured while leaving room for dynamic organization data.",
				},
			],
			liveEyebrow: "Public data",
			liveTitle: "Organizations and latest posts",
			liveNote: "This section reads from the local API so the interface stays close to real data.",
			organizations: "Organizations",
			posts: "Posts",
			total: "{count} total",
			noOrganizations: "No public organizations are registered yet.",
			noPosts: "No posts yet.",
			noDescription: "No public description yet.",
			unknownLocation: "Location is not set",
			unknownAuthor: "Unknown",
			by: "By {name}",
			calloutEyebrow: "Built for collaboration",
			calloutTitle: "One core, many needs.",
			calloutBody:
				"The core stays focused on organization CRM. API, CLI, and plugins can grow without making the public experience complicated.",
			sourceAction: "View source",
		},
		organizationsPage: {
			title: "Organizations",
			description: "Public organization profiles currently available.",
			loadError: "Organization data could not be loaded from the API. Try refreshing in a moment.",
			empty: "No public organizations yet.",
			noDescription: "No public description yet.",
			unknownLocation: "Location is not set",
		},
		postsPage: {
			title: "Public Posts",
			loadError: "Post data could not be loaded from the API. Try refreshing in a moment.",
			empty: "No public posts yet.",
			unknownAuthor: "Unknown",
			by: "By {name}",
		},
		organizationDetail: {
			info: "Information",
			location: "Location",
			website: "Website",
			claim: "Claim",
			languages: "Languages",
			allPosts: "All posts",
			impactReports: "Impact reports",
			vision: "Organization profile",
			history: "History",
			publicContact: "Public contact",
			noDescription: "No public description yet.",
			noHistory: "Not filled yet.",
			noContact: "Public contact information is not set.",
			noSdgs: "No SDGS data yet.",
			noPrograms: "Programs are not filled yet.",
			noPosts: "No public posts yet.",
			noReports: "No public reports yet.",
			focus: "Primary focus",
			programs: "Programs",
			recentPosts: "Latest posts",
			updatedAt: "Updated {date}",
			tagline: "Legal name",
			unknownLocation: "Location is not set",
			home: "Home",
		},
		postDetail: {
			eyebrow: "Post",
			meta: "By {name} · Published {date}",
			noSummary: "No summary yet.",
			noContent: "Content has not been added yet.",
		},
		organizationPostsPage: {
			title: "Organization posts",
			empty: "No public posts yet.",
		},
		impactPage: {
			title: "Impact reports",
			empty: "No public reports yet.",
			period: "Period",
			noSummary: "No summary yet.",
			status: "Status",
			publicAt: "Public",
		},
	},
};

export const locale = writable("id");

function readPath(source, path) {
	return path.split(".").reduce((value, key) => value?.[key], source);
}

function interpolate(value, params) {
	if (typeof value !== "string") {
		return value;
	}

	return value.replace(/\{(\w+)\}/g, (_, key) => `${params[key] ?? ""}`);
}

export const t = derived(locale, ($locale) => {
	return (path, params = {}) => {
		const current = readPath(dictionaries[$locale], path);
		const fallback = readPath(dictionaries.id, path);
		return interpolate(current ?? fallback ?? path, params);
	};
});

export function setLocale(nextLocale) {
	const selected = dictionaries[nextLocale] ? nextLocale : "id";
	locale.set(selected);

	if (browser) {
		localStorage.setItem("kelompok:locale", selected);
		document.documentElement.lang = selected;
	}
}

export function initLocale() {
	if (!browser) {
		return;
	}

	const saved = localStorage.getItem("kelompok:locale");
	setLocale(saved || "id");
}
