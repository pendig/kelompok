import { fail, redirect } from "@sveltejs/kit";
import { env } from "$env/dynamic/private";
import { APIError } from "$lib/api.js";
import { loginWithPassword, loadSession } from "$lib/server/session.js";

function value(form, key) {
	return `${form.get(key) || ""}`.trim();
}

function safeReturnTo(source) {
	const value = `${source || ""}`.trim();
	if (value.startsWith("/") && !value.startsWith("//") && !value.startsWith("/login") && !value.startsWith("/register")) {
		return value;
	}
	return "/account";
}

function actionError(error) {
	if (error instanceof APIError) {
		return fail(error.status || 400, {
			ok: false,
			code: error.code || "login_failed",
			error: error.apiMessage || error.message || "Unable to log in",
		});
	}
	return fail(400, {
		ok: false,
		code: "login_failed",
		error: error instanceof Error ? error.message : "Unable to log in",
	});
}

export async function load({ cookies, url }) {
	const returnTo = safeReturnTo(url.searchParams.get("return_to"));
	const session = await loadSession(cookies);
	if (session) {
		throw redirect(303, returnTo);
	}

	return {
		returnTo,
		googleClientId: env.GOOGLE_OAUTH_CLIENT_ID || "",
		error: url.searchParams.get("error") || "",
	};
}

export const actions = {
	default: async ({ request, cookies }) => {
		try {
			const form = await request.formData();
			const returnTo = safeReturnTo(value(form, "return_to"));
			await loginWithPassword(cookies, value(form, "email"), value(form, "password"));
			throw redirect(303, returnTo);
		} catch (error) {
			if (error?.status && error?.location) {
				throw error;
			}
			return actionError(error);
		}
	},
};
