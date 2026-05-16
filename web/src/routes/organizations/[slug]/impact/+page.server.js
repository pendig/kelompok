import { error } from "@sveltejs/kit";
import { APIError, fetchJSON } from "../../../../lib/api.js";

export async function load({ params }) {
	const { slug } = params;
	const encodedSlug = encodeURIComponent(slug);

	try {
		const [orgPayload, impactPayload] = await Promise.all([
			fetchJSON(`/api/v1/organizations/${encodedSlug}`),
			fetchJSON(`/api/v1/organizations/${encodedSlug}/impact-reports?limit=50`),
		]);

		return {
			organization: orgPayload.data,
			impactReports: impactPayload.data ?? [],
		};
	} catch (err) {
		if (err instanceof APIError && err.status === 404) {
			error(404, "Organization impact reports not found");
		}
		throw err;
	}
}
