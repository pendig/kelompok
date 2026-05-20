import { fetchJSONResult } from "../../lib/api.js";

export async function load() {
	const organizationsResult = await fetchJSONResult("/api/v1/organizations?limit=100");

	return {
		organizations: organizationsResult.data,
		loadError: organizationsResult.error,
	};
}
