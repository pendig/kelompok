import { fetchJSONResult } from "../lib/api.js";
import { loadSession } from "$lib/server/session.js";

export async function load({ cookies }) {
	const [organizationsResult, postsResult, session] = await Promise.all([
		fetchJSONResult("/api/v1/organizations?limit=12"),
		fetchJSONResult("/api/v1/posts?limit=8"),
		loadSession(cookies),
	]);

	return {
		organizations: organizationsResult.data,
		posts: postsResult.data,
		session,
		loadErrors: [organizationsResult.error, postsResult.error].filter(Boolean),
	};
}
