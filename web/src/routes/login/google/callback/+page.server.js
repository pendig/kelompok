import { redirect } from "@sveltejs/kit";
import { env } from "$env/dynamic/private";
import { loginWithGoogle } from "$lib/server/session.js";

export async function load({ url, cookies }) {
	const code = url.searchParams.get("code");
	const returnTo = url.searchParams.get("state") || "/account";

	if (!code) {
		throw redirect(303, `/login?error=missing_code`);
	}

	try {
		let origin = env.ORIGIN || url.origin;
		if (!origin.includes("localhost") && !origin.includes("127.0.0.1") && origin.startsWith("http://")) {
			origin = origin.replace("http://", "https://");
		}
		const redirectUri = `${origin}${url.pathname}`;
		await loginWithGoogle(cookies, code, redirectUri);
		throw redirect(303, returnTo);
	} catch (error) {
		console.error("Google OAuth callback error:", error);
		const message = error?.apiMessage || error?.message || "google_login_failed";
		throw redirect(303, `/login?error=${encodeURIComponent(message)}`);
	}
}
