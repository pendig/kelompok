import { fail, redirect } from "@sveltejs/kit";
import { APIError } from "$lib/api.js";
import {
	SESSION_COOKIE,
	SESSION_UNVERIFIED_COOKIE,
	loadSession,
	submitOrganizationOnboardingRequest,
} from "$lib/server/session.js";

function value(form, key) {
	return `${form.get(key) || ""}`.trim();
}

function nullableValue(form, key) {
	const text = value(form, key);
	return text.length > 0 ? text : undefined;
}

function requestInput(form) {
	const method = value(form, "method") || "official_email";
	const target = value(form, "target");
	const note = value(form, "evidence_note");
	const website = nullableValue(form, "website_url");

	return {
		name: value(form, "name"),
		legal_name: nullableValue(form, "legal_name"),
		description: nullableValue(form, "description"),
		country: nullableValue(form, "country"),
		region: nullableValue(form, "region"),
		city: nullableValue(form, "city"),
		website_url: website,
		official_email: method === "official_email" ? target : nullableValue(form, "official_email"),
		method,
		target,
		evidence: {
			source: "organization-onboarding-form",
			...(note ? { note } : {}),
		},
	};
}

function validateInput(input) {
	if (!input.name) {
		return { code: "organization_name_required", field: "name" };
	}
	if (!input.target) {
		return { code: "organization_onboarding_claim_target_required", field: "target" };
	}
	if (input.method === "official_email" && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(input.target)) {
		return { code: "organization_onboarding_claim_target_invalid", field: "target" };
	}
	if (input.website_url) {
		try {
			const parsed = new URL(input.website_url);
			if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
				return { code: "organization_website_url_invalid", field: "website_url" };
			}
		} catch {
			return { code: "organization_website_url_invalid", field: "website_url" };
		}
	}
	return null;
}

function formValues(input) {
	return {
		name: input.name || "",
		legal_name: input.legal_name || "",
		description: input.description || "",
		country: input.country || "",
		region: input.region || "",
		city: input.city || "",
		website_url: input.website_url || "",
		official_email: input.official_email || "",
		method: input.method || "official_email",
		target: input.target || "",
		evidence_note: input.evidence?.note || "",
	};
}

function actionError(error, values) {
	if (error instanceof APIError && (error.status === 401 || error.status === 403)) {
		return fail(error.status, {
			ok: false,
			action: "createOrganization",
			code: "session_expired",
			values,
		});
	}
	if (error instanceof APIError) {
		return fail(error.status || 400, {
			ok: false,
			action: "createOrganization",
			code: error.code || "organization_onboarding_create_failed",
			error: error.apiMessage || error.message,
			values,
		});
	}
	return fail(400, {
		ok: false,
		action: "createOrganization",
		code: "organization_onboarding_create_failed",
		error: error instanceof Error ? error.message : "Unable to submit organization onboarding request",
		values,
	});
}

export async function load({ cookies }) {
	const session = await loadSession(cookies);
	if (!session) {
		const unverified = cookies.get(SESSION_UNVERIFIED_COOKIE);
		const stillHasSession = cookies.get(SESSION_COOKIE);
		if (unverified && stillHasSession) {
			return { session: null, unverified: true };
		}
		throw redirect(303, "/login?return_to=/organizations/new");
	}

	return { session, unverified: false };
}

export const actions = {
	createOrganization: async ({ request, cookies }) => {
		const form = await request.formData();
		const input = requestInput(form);
		const values = formValues(input);
		const invalid = validateInput(input);
		if (invalid) {
			return fail(400, {
				ok: false,
				action: "createOrganization",
				code: invalid.code,
				field: invalid.field,
				values,
			});
		}

		try {
			const item = await submitOrganizationOnboardingRequest(cookies, input);
			return {
				ok: true,
				action: "createOrganization",
				item,
			};
		} catch (error) {
			if (error instanceof APIError && (error.status === 401 || error.status === 403)) {
				cookies.delete(SESSION_COOKIE, { path: "/" });
				cookies.delete(SESSION_UNVERIFIED_COOKIE, { path: "/" });
				throw redirect(303, "/login?return_to=/organizations/new");
			}
			return actionError(error, values);
		}
	},
};
