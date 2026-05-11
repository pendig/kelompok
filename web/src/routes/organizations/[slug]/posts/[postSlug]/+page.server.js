import { fetchJSON } from "../../../../../lib/api.js";

export async function load({ params }) {
	const organizationSlug = encodeURIComponent(params.slug);
	const postSlug = encodeURIComponent(params.postSlug);
	const response = await fetchJSON(`/api/v1/organizations/${organizationSlug}/posts/${postSlug}`);

	return {
		post: response.data,
	};
}
