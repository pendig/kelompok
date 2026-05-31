import { loadSession } from "$lib/server/session.js";
import { fetchJSON } from "$lib/api.js";

export async function load({ cookies }) {
	let maintenance = false;
	try {
		const res = await fetchJSON("/api/v1/maintenance");
		maintenance = res?.data?.maintenance || false;
	} catch (e) {
		// If it's a 503 or failed to fetch, then we are under maintenance!
		if (e?.status === 503) {
			maintenance = true;
		}
	}

	return {
		session: await loadSession(cookies),
		maintenance,
	};
}
