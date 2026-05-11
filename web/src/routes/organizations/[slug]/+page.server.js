import { fetchJSON } from "../../../lib/api.js";

export async function load({ params }) {
	const { slug } = params;
	const encodedSlug = encodeURIComponent(slug);

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
}
