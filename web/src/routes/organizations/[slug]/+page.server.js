import { error, fail } from "@sveltejs/kit";
import { APIError, fetchJSON } from "../../../lib/api.js";

function value(form, key) {
	return `${form.get(key) || ""}`.trim();
}

function claimInput(form) {
	const note = value(form, "evidence_note");

	return {
		method: value(form, "method") || "official_email",
		target: value(form, "target"),
		requester_email: value(form, "requester_email"),
		evidence: {
			source: "public-profile",
			...(note ? { note } : {}),
		},
	};
}

function actionError(err) {
	return fail(400, {
		ok: false,
		action: "submitClaim",
		error: err instanceof Error ? err.message : "Unable to submit claim",
	});
}

export async function load({ params }) {
	const { slug } = params;
	const encodedSlug = encodeURIComponent(slug);

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
		};
	} catch (err) {
		if (err instanceof APIError && err.status === 404) {
			error(404, "Organization not found");
		}
		throw err;
	}
}

export const actions = {
	submitClaim: async ({ request, params }) => {
		try {
			const form = await request.formData();
			const payload = await fetchJSON(`/api/v1/organizations/${encodeURIComponent(params.slug)}/claims`, {
				method: "POST",
				headers: {
					"content-type": "application/json",
				},
				body: JSON.stringify(claimInput(form)),
			});

			return { ok: true, action: "submitClaim", item: payload.data };
		} catch (err) {
			return actionError(err);
		}
	},
};
