import { redirect } from "@sveltejs/kit";
import { loadSession, logoutSession } from "$lib/server/session.js";

export async function load({ cookies }) {
	const session = await loadSession(cookies);
	if (!session) {
		throw redirect(303, "/login?return_to=/account");
	}

	return { session };
}

export const actions = {
	logout: async ({ cookies }) => {
		await logoutSession(cookies);
		throw redirect(303, "/");
	},
};
