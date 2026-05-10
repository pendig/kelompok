import { fetchJSON } from "../../../lib/api.js";

export async function load({ params }) {
	const response = await fetchJSON(`/api/v1/posts/${encodeURIComponent(params.slug)}`);
	return {
		post: response.data,
	};
}

