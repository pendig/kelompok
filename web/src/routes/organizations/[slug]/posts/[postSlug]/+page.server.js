import { error } from "@sveltejs/kit";
import { APIError, fetchJSON } from "../../../../../lib/api.js";

export async function load({ params }) {
	const organizationSlug = encodeURIComponent(params.slug);
	const postSlug = encodeURIComponent(params.postSlug);
	try {
		const response = await fetchJSON(`/api/v1/organizations/${organizationSlug}/posts/${postSlug}`);

		return {
			post: response.data,
		};
	} catch (err) {
		if (err instanceof APIError && err.status === 404) {
			error(404, "Post not found");
		}
		throw err;
	}
}
