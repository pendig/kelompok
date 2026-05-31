import { redirect } from "@sveltejs/kit";
import { loginWithGoogle } from "$lib/server/session.js";

export async function load({ url, cookies }) {
	const code = url.searchParams.get("code");
	const returnTo = url.searchParams.get("state") || "/account";

	if (!code) {
		throw redirect(303, `/login?error=missing_code`);
	}

	try {
		const redirectUri = `${url.origin}${url.pathname}`;
		await loginWithGoogle(cookies, code, redirectUri);
		throw redirect(303, returnTo);
	} catch (error) {
		console.error("Google OAuth callback error:", error);
		const message = error?.apiMessage || error?.message || "google_login_failed";
		throw redirect(303, `/login?error=${encodeURIComponent(message)}`);
	}
}
