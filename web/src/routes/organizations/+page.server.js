import { fetchJSONResult } from "../../lib/api.js";

export async function load() {
	const response = await fetchJSONResult("/api/v1/organizations?limit=50");
	return {
		organizations: response.data,
		loadError: response.error,
	};
}
