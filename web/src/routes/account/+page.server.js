import { fail, redirect } from "@sveltejs/kit";
import { APIError } from "$lib/api.js";
import {
	SESSION_COOKIE,
	SESSION_UNVERIFIED_COOKIE,
	loadSession,
	logoutSession,
	updateProfile,
} from "$lib/server/session.js";

export async function load({ cookies }) {
	const session = await loadSession(cookies);
	if (!session) {
		throw redirect(303, "/login?return_to=/account");
	}

	return { session };
}

function value(form, key) {
	return `${form.get(key) || ""}`.trim();
}

function actionError(error) {
	if (error instanceof APIError) {
		if (error.status === 401 || error.status === 403) {
			return fail(error.status, {
				ok: false,
				action: "updateProfile",
				code: "session_expired",
				error: "session_expired",
			});
		}
		return fail(error.status || 400, {
			ok: false,
			action: "updateProfile",
			code: error.code || "profile_update_failed",
			error: error.apiMessage || error.message || "Unable to update profile",
		});
	}
	return fail(400, {
		ok: false,
		action: "updateProfile",
		code: "profile_update_failed",
		error: error instanceof Error ? error.message : "profile_update_failed",
	});
}

export const actions = {
	logout: async ({ cookies }) => {
		await logoutSession(cookies);
		throw redirect(303, "/");
	},
	updateProfile: async ({ request, cookies }) => {
		const form = await request.formData();
		const name = value(form, "name");
		if (!name) {
			return fail(400, {
				ok: false,
				action: "updateProfile",
				code: "name_required",
				error: "name_required",
			});
		}

		try {
			await updateProfile(cookies, { name });
		} catch (error) {
			if (error instanceof APIError && (error.status === 401 || error.status === 403)) {
				cookies.delete(SESSION_COOKIE, { path: "/" });
				cookies.delete(SESSION_UNVERIFIED_COOKIE, { path: "/" });
				throw redirect(303, "/login?return_to=/account");
			}
			return actionError(error);
		}

		return {
			ok: true,
			action: "updateProfile",
		};
	},
};
