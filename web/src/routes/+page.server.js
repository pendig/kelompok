import { fetchJSONResult } from "../lib/api.js";

export async function load() {
	const [organizationsResult, postsResult] = await Promise.all([
		fetchJSONResult("/api/v1/organizations?limit=12"),
		fetchJSONResult("/api/v1/posts?limit=8"),
	]);

	return {
		organizations: organizationsResult.data,
		posts: postsResult.data,
		loadErrors: [organizationsResult.error, postsResult.error].filter(Boolean),
	};
}
