import { fail, redirect } from "@sveltejs/kit";
import { env } from "$env/dynamic/private";
import { APIError, fetchJSON } from "../../lib/api.js";
import {
	loadSession,
	loginWithPassword,
	logoutSession,
	SESSION_COOKIE,
	SESSION_UNVERIFIED_COOKIE,
} from "$lib/server/session.js";

const ADMIN_API_KEY = `${env.KELOMPOK_ADMIN_API_KEY || ""}`.trim();

function adminHeaders(cookies, headers = {}) {
	const sessionToken = cookies?.get(SESSION_COOKIE) || "";
	const sessionUnverified = cookies?.get(SESSION_UNVERIFIED_COOKIE) === "1";
	const useSessionToken = sessionToken && !sessionUnverified;
	return {
		...headers,
		...(useSessionToken ? { authorization: `Bearer ${sessionToken}` } : {}),
		...(!useSessionToken && ADMIN_API_KEY ? { "x-kelompok-admin-key": ADMIN_API_KEY } : {}),
	};
}

function adminFetchJSON(path, init = {}, cookies = null) {
	return fetchJSON(path, {
		...init,
		headers: adminHeaders(cookies, init.headers || {}),
	});
}

async function adminFetchJSONResult(path, fallbackData = [], cookies = null) {
	try {
		const payload = await adminFetchJSON(path, {}, cookies);
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

async function mutate(path, body, method = "POST", cookies = null) {
	return adminFetchJSON(path, {
		method,
		headers: {
			"content-type": "application/json",
		},
		body: JSON.stringify(body),
	}, cookies);
}

function organizationInput(form) {
	const publicContact = {
		email: value(form, "public_contact_email") || value(form, "official_email"),
		instagram: value(form, "public_contact_instagram"),
		phone: value(form, "public_contact_phone"),
	};
	const impactData = jsonObject(value(form, "impact_data"), {});
	const beneficiaries = value(form, "impact_beneficiaries");
	const volunteerHours = value(form, "impact_volunteer_hours");
	const impactNote = value(form, "impact_note");
	if (beneficiaries) {
		impactData.beneficiaries = Number(beneficiaries);
	}
	if (volunteerHours) {
		impactData.volunteer_hours = Number(volunteerHours);
	}
	if (impactNote) {
		impactData.note = impactNote;
	}

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
		source_data: jsonObject(value(form, "source_data"), {}),
		sdgs_data: {
			primary: splitList(value(form, "sdgs")),
		},
		impact_data: impactData,
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

function relationshipInput(form) {
	return {
		parent_organization_slug: value(form, "parent_organization_slug"),
		child_organization_slug: value(form, "child_organization_slug"),
		relationship_type: value(form, "relationship_type") || "related",
		label: value(form, "label"),
		status: value(form, "status") || "active",
		started_at: optionalDate(value(form, "started_at")),
		ended_at: optionalDate(value(form, "ended_at")),
		metadata: jsonObject(value(form, "metadata"), {}),
	};
}

function actionError(error) {
	if (error instanceof APIError && error.code === "organization_slug_taken") {
		return fail(409, {
			ok: false,
			error: "organization_slug_taken",
			error_code: "organization_slug_taken",
		});
	}

	return fail(400, {
		ok: false,
		error: error instanceof Error ? error.message : "Action failed",
	});
}

function canManageOrganizationRole(role) {
	return role === "owner" || role === "admin";
}

export async function load({ url, cookies }) {
	const session = await loadSession(cookies);
	const isScopedSession = Boolean(session && session.user?.role !== "superadmin");
	if (isScopedSession && url.pathname === "/admin") {
		const query = url.searchParams.toString();
		throw redirect(303, `/console${query ? `?${query}` : ""}`);
	}
	const requestedSlug = url.searchParams.get("org");
	const requestedView = url.searchParams.get("view");
	const justCreated = url.searchParams.get("created") === "1";
	const allowedViews = new Set(["dashboard", "organizations", "organization-edit", "members", "relationships", "posts", "impact", "claims", "audit"]);
	let orgPayload = { data: [], error: null };
	let postPayload = { data: [], error: null };
	let impactPayload = { data: [], error: null };
	let selectedPayload = { data: null, error: null };
	let memberPayload = { data: [], error: null };
	let claimPayload = { data: [], error: null };
	let auditPayload = { data: [], error: null };
	let relationshipPayload = { data: [], error: null };
	let selectedSlug = "";

	if (isScopedSession) {
		const roleOrganizations = (session.organization_roles ?? []).filter((role) =>
			canManageOrganizationRole(role.role),
		);
		const allowedSlugs = roleOrganizations.map((role) => role.organization_slug).filter(Boolean);
		selectedSlug = allowedSlugs.includes(requestedSlug) ? requestedSlug : allowedSlugs[0] || "";
		orgPayload = {
			data: roleOrganizations.map((role) => ({
				slug: role.organization_slug,
				name: role.organization_name,
				claim_status: role.role,
			})),
			error: null,
		};

		if (selectedSlug) {
			[selectedPayload, postPayload, impactPayload, memberPayload, claimPayload, auditPayload, relationshipPayload] =
				await Promise.all([
					adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}`, null, cookies),
					adminFetchJSONResult(
						`/api/v1/org-admin/posts?organization_slug=${encodeURIComponent(selectedSlug)}&limit=50`,
						[],
						cookies,
					),
					adminFetchJSONResult(
						`/api/v1/org-admin/impact-reports?organization_slug=${encodeURIComponent(selectedSlug)}&limit=50`,
						[],
						cookies,
					),
					adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}/members?limit=30`, [], cookies),
					adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}/claims?limit=20`, [], cookies),
					adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}/audit-logs?limit=20`, [], cookies),
					adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}/relationships?limit=50`, [], cookies),
				]);
		}
	} else {
		[orgPayload, postPayload, impactPayload] = await Promise.all([
			adminFetchJSONResult("/api/v1/org-admin/organizations?limit=50", [], cookies),
			adminFetchJSONResult("/api/v1/org-admin/posts?limit=50", [], cookies),
			adminFetchJSONResult("/api/v1/org-admin/impact-reports?limit=50", [], cookies),
		]);
		selectedSlug = requestedSlug || orgPayload.data?.[0]?.slug || "";

		if (selectedSlug) {
			[selectedPayload, memberPayload, claimPayload, auditPayload, relationshipPayload] = await Promise.all([
				adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}`, null, cookies),
				adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}/members?limit=30`, [], cookies),
				adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}/claims?limit=20`, [], cookies),
				adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}/audit-logs?limit=20`, [], cookies),
				adminFetchJSONResult(`/api/v1/org-admin/organizations/${encodeURIComponent(selectedSlug)}/relationships?limit=50`, [], cookies),
			]);
		}
	}

	let organizations = orgPayload.data ?? [];
	if (selectedPayload.data && !organizations.some((organization) => organization.slug === selectedPayload.data.slug)) {
		organizations = [selectedPayload.data, ...organizations];
	} else if (selectedPayload.data) {
		organizations = organizations.map((organization) =>
			organization.slug === selectedPayload.data.slug ? selectedPayload.data : organization,
		);
	}
	const posts = postPayload.data ?? [];
	const impactReports = impactPayload.data ?? [];

	return {
		organizations,
		session,
		isAuthenticated: Boolean(session || ADMIN_API_KEY),
		posts,
		impactReports,
		members: memberPayload.data ?? [],
		claims: claimPayload.data ?? [],
		auditLogs: auditPayload.data ?? [],
		relationships: relationshipPayload.data ?? [],
		selectedOrganization: selectedPayload.data,
		selectedSlug,
		justCreated,
		consoleMode: url.pathname === "/console",
		initialTab: allowedViews.has(requestedView) ? requestedView : requestedSlug ? "organization-edit" : "dashboard",
		loadErrors: [
			orgPayload.error,
			postPayload.error,
			impactPayload.error,
			selectedPayload.error,
			memberPayload.error,
			claimPayload.error,
			auditPayload.error,
			relationshipPayload.error,
		].filter(Boolean),
	};
}

export const actions = {
	login: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			await loginWithPassword(cookies, value(form, "email"), value(form, "password"));
			return { ok: true, action: "login" };
		} catch (error) {
			return actionError(error);
		}
	},
	logout: async ({ cookies, url }) => {
		await logoutSession(cookies);
		throw redirect(303, url.pathname === "/console" ? "/login" : "/admin");
	},
	createOrganization: async ({ request, cookies, url }) => {
		let createdSlug = "";
		try {
			const form = await request.formData();
			const payload = await mutate("/api/v1/org-admin/organizations", organizationInput(form), "POST", cookies);
			createdSlug = payload.data?.slug || value(form, "slug");
		} catch (error) {
			return actionError(error);
		}
		throw redirect(303, `${url.pathname}?org=${encodeURIComponent(createdSlug)}&view=organization-edit&created=1`);
	},
	updateOrganization: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const slug = value(form, "current_slug");
			const payload = await mutate(
				`/api/v1/org-admin/organizations/${encodeURIComponent(slug)}`,
				organizationInput(form),
				"PATCH",
				cookies,
			);
			return { ok: true, action: "updateOrganization", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	createMember: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const slug = value(form, "organization_slug");
			const payload = await mutate(
				`/api/v1/org-admin/organizations/${encodeURIComponent(slug)}/members`,
				memberInput(form),
				"POST",
				cookies,
			);
			return { ok: true, action: "createMember", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	createPost: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const organizationSlug = value(form, "organization_slug");
			const payload = await mutate(
				`/api/v1/org-admin/posts?organization_slug=${encodeURIComponent(organizationSlug)}`,
				postInput(form),
				"POST",
				cookies,
			);
			return { ok: true, action: "createPost", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	createImpactReport: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const organizationSlug = value(form, "organization_slug");
			const payload = await mutate(
				`/api/v1/org-admin/impact-reports?organization_slug=${encodeURIComponent(organizationSlug)}`,
				impactInput(form),
				"POST",
				cookies,
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
	createRelationship: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const payload = await mutate("/api/v1/org-admin/organization-relationships", relationshipInput(form), "POST", cookies);
			return { ok: true, action: "createRelationship", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	deleteRelationship: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const payload = await mutate(
				`/api/v1/org-admin/organization-relationships/${encodeURIComponent(value(form, "id"))}`,
				{},
				"DELETE",
				cookies,
			);
			return { ok: true, action: "deleteRelationship", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	approveClaim: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const payload = await mutate(`/api/v1/org-admin/claims/${encodeURIComponent(value(form, "id"))}/approve`, {}, "POST", cookies);
			return { ok: true, action: "approveClaim", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	rejectClaim: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const payload = await mutate(`/api/v1/org-admin/claims/${encodeURIComponent(value(form, "id"))}/reject`, {}, "POST", cookies);
			return { ok: true, action: "rejectClaim", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	publishPost: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const payload = await mutate(`/api/v1/org-admin/posts/${encodeURIComponent(value(form, "id"))}/publish`, {}, "POST", cookies);
			return { ok: true, action: "publishPost", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
	publishImpactReport: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const payload = await mutate(
				`/api/v1/org-admin/impact-reports/${encodeURIComponent(value(form, "id"))}/publish`,
				{},
				"POST",
				cookies,
			);
			return { ok: true, action: "publishImpactReport", item: payload.data };
		} catch (error) {
			return actionError(error);
		}
	},
};
