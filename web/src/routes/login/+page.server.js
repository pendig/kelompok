import { fail, redirect } from "@sveltejs/kit";
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
	return fail(400, {
		ok: false,
		error: error instanceof Error ? error.message : "Unable to log in",
	});
}

export async function load({ cookies, url }) {
	const returnTo = safeReturnTo(url.searchParams.get("return_to"));
	const session = await loadSession(cookies);
	if (session) {
		throw redirect(303, returnTo);
	}

	return { returnTo };
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
