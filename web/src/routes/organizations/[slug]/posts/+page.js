import { fetchJSON } from "../../../../lib/api.js";

export async function load({ params }) {
	const { slug } = params;
	const [orgPayload, postsPayload] = await Promise.all([
		fetchJSON(`/api/v1/organizations/${slug}`),
		fetchJSON(`/api/v1/organizations/${slug}/posts?limit=50`),
	]);

	return {
		organization: orgPayload.data,
		posts: postsPayload.data ?? [],
	};
}

