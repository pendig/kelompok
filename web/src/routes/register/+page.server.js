import { fail, redirect } from "@sveltejs/kit";
import { APIError, fetchJSON } from "$lib/api.js";
import { loadSession, loginWithPassword } from "$lib/server/session.js";

function value(form, key) {
	return `${form.get(key) || ""}`.trim();
}

function actionError(error) {
	if (error instanceof APIError) {
		return fail(error.status || 400, {
			ok: false,
			code: error.code || "register_failed",
			error: error.apiMessage || error.message || "Unable to register",
		});
	}
	return fail(400, {
		ok: false,
		code: "register_failed",
		error: error instanceof Error ? error.message : "Unable to register",
	});
}

export async function load({ cookies }) {
	const session = await loadSession(cookies);
	if (session) {
		throw redirect(303, "/account");
	}

	return {};
}

export const actions = {
	default: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const name = value(form, "name");
			const email = value(form, "email");
			const password = value(form, "password");
			if (!name) {
				return fail(400, {
					ok: false,
					code: "name_required",
					error: "Name is required",
				});
			}

			await fetchJSON("/api/v1/auth/register", {
				method: "POST",
				headers: {
					"content-type": "application/json",
				},
				body: JSON.stringify({
					name,
					email,
					password,
				}),
			});
			await loginWithPassword(cookies, email, password);
			throw redirect(303, "/account");
		} catch (error) {
			if (error?.status && error?.location) {
				throw error;
			}
			return actionError(error);
		}
	},
};
