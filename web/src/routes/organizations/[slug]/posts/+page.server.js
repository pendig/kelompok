import { fetchJSON } from "../../../../lib/api.js";

export async function load({ params }) {
	const { slug } = params;
	const encodedSlug = encodeURIComponent(slug);

	const [orgPayload, postsPayload] = await Promise.all([
		fetchJSON(`/api/v1/organizations/${encodedSlug}`),
		fetchJSON(`/api/v1/organizations/${encodedSlug}/posts?limit=50`),
	]);

	return {
		organization: orgPayload.data,
		posts: postsPayload.data ?? [],
	};
}
