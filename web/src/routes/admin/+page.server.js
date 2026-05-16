import { fail } from "@sveltejs/kit";
import { env } from "$env/dynamic/private";
import { fetchJSON, fetchJSONResult } from "../../lib/api.js";

const ADMIN_API_KEY = `${env.KELOMPOK_ADMIN_API_KEY || ""}`.trim();

function adminHeaders(headers = {}) {
	return {
		...headers,
		...(ADMIN_API_KEY ? { "x-kelompok-admin-key": ADMIN_API_KEY } : {}),
	};
}

function adminFetchJSON(path, init = {}) {
	return fetchJSON(path, {
		...init,
		headers: adminHeaders(init.headers || {}),
	});
}

async function adminFetchJSONResult(path, fallbackData = []) {
	try {
		const payload = await adminFetchJSON(path);
		return {
			data: payload.data ?? fallbackData,
			error: null,
		};
	} catch (error) {
		return {
			data: fallbackData,
			error: error instanceof Error ? error.message : "Unable to load data from API",
		};
	}
}

function checkEndpoint(path) {
	return fetchJSON(path)
		.then((payload) => ({ ok: true, payload }))
		.catch((error) => ({ ok: false, error: error instanceof Error ? error.message : "Health check failed" }));
}

function value(form, key) {
	return `${form.get(key) || ""}`.trim();
}

function splitList(source) {
	return source
		.split(/[\n,]/)
		.map((item) => item.trim())
		.filter(Boolean);
}

function optionalDate(source) {
	if (!source) {
		return null;
	}

	return `${source}T00:00:00.000Z`;
}

function jsonObject(source, fallback = {}) {
	if (!source) {
		return fallback;
	}

	const parsed = JSON.parse(source);
	if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
		throw new Error("Expected a JSON object");
	}
	return parsed;
}

function jsonArray(source, fallback = []) {
	if (!source) {
		return fallback;
	}

	const parsed = JSON.parse(source);
	if (!Array.isArray(parsed)) {
		throw new Error("Expected a JSON array");
	}
	return parsed;
}

async function mutate(path, body, method = "POST") {
	return adminFetchJSON(path, {
		method,
		headers: {
			"content-type": "application/json",
		},
		body: JSON.stringify(body),
	});
}

function organizationInput(form) {
	const publicContact = {
		email: value(form, "public_contact_email") || value(form, "official_email"),
		instagram: value(form, "public_contact_instagram"),
		phone: value(form, "public_contact_phone"),
	};

	return {
		slug: value(form, "slug"),
		name: value(form, "name"),
		legal_name: value(form, "legal_name"),
		description: value(form, "description"),
		history: value(form, "history"),
		country: value(form, "country"),
		region: value(form, "region"),
		city: value(form, "city"),
		website_url: value(form, "website_url"),
		official_email: value(form, "official_email"),
		claim_status: value(form, "claim_status") || "unclaimed",
		profile_data: {
			focus: splitList(value(form, "focus")),
			programs: splitList(value(form, "programs")),
			languages: splitList(value(form, "languages")),
			public_contact: Object.fromEntries(Object.entries(publicContact).filter(([, item]) => item)),
		},
		sdgs_data: {
			primary: splitList(value(form, "sdgs")),
		},
		impact_data: jsonObject(value(form, "impact_data"), {}),
	};
}

function memberInput(form) {
	return {
		name: value(form, "name"),
		position: value(form, "position"),
		bio: value(form, "bio"),
		email: value(form, "email"),
		phone: value(form, "phone"),
		social_links: jsonObject(value(form, "social_links"), {}),
		start_date: optionalDate(value(form, "start_date")),
		end_date: optionalDate(value(form, "end_date")),
	};
}

function postInput(form) {
	const summary = value(form, "summary");
	return {
		organization_slug: value(form, "organization_slug"),
		slug: value(form, "slug"),
		title: value(form, "title"),
		summary,
		content: value(form, "content"),
		category_slug: value(form, "category_slug"),
		status: value(form, "status") || "draft",
		post_data: jsonObject(value(form, "post_data"), { kind: "article" }),
		seo_data: jsonObject(value(form, "seo_data"), {
			title: value(form, "title"),
			description: summary,
		}),
	};
}

function impactInput(form) {
	return {
		organization_slug: value(form, "organization_slug"),
		title: value(form, "title"),
		summary: value(form, "summary"),
		report_period_start: optionalDate(value(form, "report_period_start")),
		report_period_end: optionalDate(value(form, "report_period_end")),
		sdgs: jsonArray(value(form, "sdgs_json"), splitList(value(form, "sdgs"))),
		metrics: jsonObject(value(form, "metrics"), {}),
		status: value(form, "status") || "draft",
	};
}

function claimInput(form) {
	return {
		method: value(form, "method") || "official_email",
		target: value(form, "target"),
		requester_email: value(form, "requester_email"),
		evidence: jsonObject(value(form, "evidence"), {
			source: "admin-ui",
		}),
	};
}

function actionError(error) {
	return fail(400, {
		ok: false,
		error: error instanceof Error ? error.message : "Action failed",
	});
}

export async function load({ url }) {
	const [orgPayload, postPayload, impactPayload, health, ready, root] = await Promise.all([
		adminFetchJSONResult("/api/v1/org-admin/organizations?limit=50"),
		adminFetchJSONResult("/api/v1/org-admin/posts?limit=50"),
		adminFetchJSONResult("/api/v1/org-admin/impact-reports?limit=50"),
		checkEndpoint("/healthz"),
		checkEndpoint("/readyz"),
		checkEndpoint("/"),
	]);

	const organizations = orgPayload.data ?? [];
	const posts = postPayload.data ?? [];
	const impactReports = impactPayload.data ?? [];
	const requestedSlug = url.searchParams.get("org");
	const selectedSlug = requestedSlug || organizations[0]?.slug || "";

	const [selectedPayload, memberPayload, claimPayload] =
		selectedSlug ?
			await Promise.all([
				adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}`, null),
				adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}/members?limit=30`),
				adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}/claims?limit=20`),
			])
		:	[
				{ data: null, error: null },
				{ data: [], error: null },
				{ data: [], error: null },
			];

	const impactByOrganization = organizations.slice(0, 6).map((org) => {
		const count = impactReports.filter((item) => item.organization_slug === org.slug).length;
		const latest = impactReports.find((item) => item.organization_slug === org.slug)?.title || null;
		return {
			orgSlug: org.slug,
			orgName: org.name || "Unnamed org",
			count,
			latest,
			error: null,
		};
	});

	const missingClaimStatusOrgs = organizations.filter((item) => item.claim_status === undefined).length;
	const checks = [
		{
			label: "admin.checkApiHealth",
			status: health.ok ? "pass" : "fail",
			detail: health.ok ? `status=${health.payload.data?.status || "ok"}` : health.error,
			needsReview: false,
		},
		{
			label: "admin.checkAdminAuth",
			status: ADMIN_API_KEY ? "pass" : "fail",
			detail: ADMIN_API_KEY ? "admin key is configured for server-side admin requests" : "set KELOMPOK_ADMIN_API_KEY",
			needsReview: !ADMIN_API_KEY,
		},
		{
			label: "admin.checkReadyHealth",
			status: ready.ok ? "pass" : "warn",
			detail: ready.ok ? `status=${ready.payload.data?.status || "ready"}` : ready.error,
			needsReview: !ready.ok,
		},
		{
			label: "admin.checkHasOrganizations",
			status: organizations.length > 0 ? "pass" : "warn",
			detail:
				organizations.length > 0 ? `${organizations.length} orgs visible in admin API` : "No organization data found",
			needsReview: true,
		},
		{
			label: "admin.checkHasPosts",
			status: posts.length > 0 ? "pass" : "warn",
			detail: posts.length > 0 ? `${posts.length} posts loaded` : "No posts found",
			needsReview: true,
		},
		{
			label: "admin.checkHasImpact",
			status: impactReports.length > 0 ? "pass" : "warn",
			detail:
				impactReports.length > 0 ?
					`${impactReports.length} impact reports loaded`
				:	"Impact admin endpoint works, but no data found",
			needsReview: true,
		},
		{
			label: "admin.checkHasClaimStatus",
			status: organizations.length > 0 && missingClaimStatusOrgs === 0 ? "pass" : "warn",
			detail:
				organizations.length === 0 ?
					"No organizations loaded yet"
				: missingClaimStatusOrgs > 0
				? `${missingClaimStatusOrgs} organization(s) missing claim status`
				: `${organizations.length} records include claim status`,
			needsReview: missingClaimStatusOrgs > 0 || organizations.length === 0,
		},
	];

	return {
		organizations,
		posts,
		impactReports,
		members: memberPayload.data ?? [],
		claims: claimPayload.data ?? [],
		selectedOrganization: selectedPayload.data,
		selectedSlug,
		health,
		ready,
		root,
		impactByOrganization,
		missingClaimStatusOrgs,
		loadErrors: [
			orgPayload.error,
			postPayload.error,
			impactPayload.error,
			selectedPayload.error,
			memberPayload.error,
			claimPayload.error,
			health.ok ? null : health.error,
			ready.ok ? null : ready.error,
			root.ok ? null : root.error,
		].filter(Boolean),
		checks,
	};
}

export const actions = {
	createOrganization: async ({ request }) => {
		try {
			const form = await request.formData();
			const payload = await mutate("/api/v1/org-admin/organizations", organizationInput(form));
			return { ok: true, action: "createOrganization", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	updateOrganization: async ({ request }) => {
		try {
			const form = await request.formData();
			const slug = value(form, "current_slug");
			const payload = await mutate(
				`/api/v1/org-admin/organizations/${encodeURIComponent(slug)}`,
				organizationInput(form),
				"PATCH",
			);
			return { ok: true, action: "updateOrganization", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	createMember: async ({ request }) => {
		try {
			const form = await request.formData();
			const slug = value(form, "organization_slug");
			const payload = await mutate(
				`/api/v1/org-admin/organizations/${encodeURIComponent(slug)}/members`,
				memberInput(form),
			);
			return { ok: true, action: "createMember", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	createPost: async ({ request }) => {
		try {
			const form = await request.formData();
			const organizationSlug = value(form, "organization_slug");
			const payload = await mutate(
				`/api/v1/org-admin/posts?organization_slug=${encodeURIComponent(organizationSlug)}`,
				postInput(form),
			);
			return { ok: true, action: "createPost", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	createImpactReport: async ({ request }) => {
		try {
			const form = await request.formData();
			const organizationSlug = value(form, "organization_slug");
			const payload = await mutate(
				`/api/v1/org-admin/impact-reports?organization_slug=${encodeURIComponent(organizationSlug)}`,
				impactInput(form),
			);
			return { ok: true, action: "createImpactReport", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	createClaim: async ({ request }) => {
		try {
			const form = await request.formData();
			const slug = value(form, "organization_slug");
			const payload = await mutate(`/api/v1/organizations/${encodeURIComponent(slug)}/claims`, claimInput(form));
			return { ok: true, action: "createClaim", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	publishPost: async ({ request }) => {
		try {
			const form = await request.formData();
			const payload = await mutate(`/api/v1/org-admin/posts/${encodeURIComponent(value(form, "id"))}/publish`, {});
			return { ok: true, action: "publishPost", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	publishImpactReport: async ({ request }) => {
		try {
			const form = await request.formData();
			const payload = await mutate(
				`/api/v1/org-admin/impact-reports/${encodeURIComponent(value(form, "id"))}/publish`,
				{},
			);
			return { ok: true, action: "publishImpactReport", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
};
