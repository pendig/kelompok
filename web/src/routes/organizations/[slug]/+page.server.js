import { error } from "@sveltejs/kit";
import { APIError, fetchJSON } from "../../../lib/api.js";

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
