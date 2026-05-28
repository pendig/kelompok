import { redirect } from "@sveltejs/kit";
import {
	loadSession,
	logoutSession,
	SESSION_COOKIE,
	SESSION_UNVERIFIED_COOKIE,
} from "$lib/server/session.js";

export async function load({ cookies, url }) {
	const session = await loadSession(cookies);

	if (!session) {
		// SESSION_COOKIE is preserved when the API call fails for transient
		// reasons (e.g. backend restart). loadSession sets the unverified
		// cookie in that case so we can keep the user logged in visually
		// and render an error state instead of bouncing them to /login.
		const unverified = cookies.get(SESSION_UNVERIFIED_COOKIE);
		const stillHasSession = cookies.get(SESSION_COOKIE);
		if (unverified && stillHasSession) {
			return {
				session: null,
				unverified: true,
				claimId: url.searchParams.get("claim") || "",
			};
		}

		throw redirect(303, "/login?return_to=/account");
	}

	return {
		session,
		unverified: false,
		claimId: url.searchParams.get("claim") || "",
	};
}

export const actions = {
	logout: async ({ cookies }) => {
		await logoutSession(cookies);
		throw redirect(303, "/");
	},
};
