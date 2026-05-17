import { loadSession } from "$lib/server/session.js";

export async function load({ cookies }) {
	return {
		session: await loadSession(cookies),
	};
}
