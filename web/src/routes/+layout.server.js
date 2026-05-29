import { loadSession } from "$lib/server/session.js";

export async function load({ cookies, url }) {
	return {
		session: await loadSession(cookies),
		pathname: url.pathname,
	};
}
