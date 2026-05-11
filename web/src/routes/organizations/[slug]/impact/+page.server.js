import { fetchJSON } from "../../../../lib/api.js";

export async function load({ params }) {
	const { slug } = params;
	const encodedSlug = encodeURIComponent(slug);

	const [orgPayload, impactPayload] = await Promise.all([
		fetchJSON(`/api/v1/organizations/${encodedSlug}`),
		fetchJSON(`/api/v1/organizations/${encodedSlug}/impact-reports?limit=50`),
	]);

	return {
		organization: orgPayload.data,
		impactReports: impactPayload.data ?? [],
	};
}
