import { error, fail } from "@sveltejs/kit";
import { APIError, fetchJSON } from "../../../lib/api.js";
import { SESSION_COOKIE, loadSession } from "$lib/server/session.js";

function value(form, key) {
	return `${form.get(key) || ""}`.trim();
}

function claimInput(form, session) {
	const note = value(form, "evidence_note");

	return {
		method: value(form, "method") || "official_email",
		target: value(form, "target"),
		requester_email: session?.user?.email || value(form, "requester_email"),
		evidence: {
			source: "public-profile",
			...(note ? { note } : {}),
		},
	};
}

function actionError(err) {
	if (err instanceof APIError && (err.status === 401 || err.status === 403)) {
		return fail(err.status, {
			ok: false,
			action: "submitClaim",
			errorCode: "unauthorized",
			error: err.apiMessage || err.message,
		});
	}
	if (err instanceof APIError && err.code === "claim_create_failed") {
		const message = err.apiMessage || err.message;
		const normalized = message.toLowerCase();
		const errorCode = normalized.includes("duplicate") || normalized.includes("unique") || normalized.includes("conflict")
			? "duplicate"
			: "generic";
		return fail(400, {
			ok: false,
			action: "submitClaim",
			errorCode,
			error: message,
		});
	}
	return fail(400, {
		ok: false,
		action: "submitClaim",
		errorCode: "generic",
		error: err instanceof Error ? err.message : "Unable to submit claim",
	});
}

export async function load({ params, cookies }) {
	const { slug } = params;
	const encodedSlug = encodeURIComponent(slug);
	const session = await loadSession(cookies);

	try {
		const [orgPayload, postPayload, impactPayload] = await Promise.all([
			fetchJSON(`/api/v1/organizations/${encodedSlug}`),
			fetchJSON(`/api/v1/organizations/${encodedSlug}/posts?limit=12`),
			fetchJSON(`/api/v1/organizations/${encodedSlug}/impact-reports?limit=12`),
		]);

		return {
			slug,
			organization: orgPayload.data,
			posts: postPayload.data ?? [],
			impactReports: impactPayload.data ?? [],
			session,
		};
	} catch (err) {
		if (err instanceof APIError && err.status === 404) {
			error(404, "Organization not found");
		}
		throw err;
	}
}

export const actions = {
	submitClaim: async ({ request, params, cookies }) => {
		try {
			const session = await loadSession(cookies);
			if (!session?.user?.email) {
				return fail(401, {
					ok: false,
					action: "submitClaim",
					errorCode: "login_required",
					error: "Login required",
				});
			}
			const form = await request.formData();
			const payload = await fetchJSON(`/api/v1/organizations/${encodeURIComponent(params.slug)}/claims`, {
				method: "POST",
				headers: {
					authorization: `Bearer ${cookies.get(SESSION_COOKIE)}`,
					"content-type": "application/json",
				},
				body: JSON.stringify(claimInput(form, session)),
			});

			return { ok: true, action: "submitClaim", item: payload.data };
		} catch (err) {
			return actionError(err);
		}
	},
};
