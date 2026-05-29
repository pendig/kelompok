import { redirect } from "@sveltejs/kit";
import { loadSession } from "$lib/server/session.js";
import { actions, load as adminLoad } from "../admin/+page.server.js";

// The console route is org-scoped: only authenticated, non-superadmin users
// may see it. We gate before delegating to the shared admin loader so the
// admin-key fallback (used for unauthenticated `/admin`) cannot leak the
// global admin view to anonymous visitors of `/console`.
export async function load(event) {
	const { url, cookies } = event;
	const session = await loadSession(cookies);

	if (!session) {
		const returnTo = `${url.pathname}${url.search}`;
		throw redirect(303, `/login?return_to=${encodeURIComponent(returnTo)}`);
	}

	if (session.user?.role === "superadmin") {
		const query = url.searchParams.toString();
		throw redirect(303, `/admin${query ? `?${query}` : ""}`);
	}

	return adminLoad(event);
}

export { actions };
