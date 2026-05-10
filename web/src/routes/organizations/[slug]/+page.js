import { fetchJSON } from "../../../lib/api.js";

export async function load({ params }) {
	const { slug } = params;

	const [orgPayload, postPayload, impactPayload] = await Promise.all([
		fetchJSON(`/api/v1/organizations/${slug}`),
		fetchJSON(`/api/v1/organizations/${slug}/posts?limit=12`),
		fetchJSON(`/api/v1/organizations/${slug}/impact-reports?limit=12`),
	]);

	return {
		slug,
		organization: orgPayload.data,
		posts: postPayload.data ?? [],
		impactReports: impactPayload.data ?? [],
	};
}

