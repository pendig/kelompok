import { fetchJSON } from "../../../../lib/api.js";

export async function load({ params }) {
	const { slug } = params;
	const [orgPayload, impactPayload] = await Promise.all([
		fetchJSON(`/api/v1/organizations/${slug}`),
		fetchJSON(`/api/v1/organizations/${slug}/impact-reports?limit=50`),
	]);

	return {
		organization: orgPayload.data,
		impactReports: impactPayload.data ?? [],
	};
}

