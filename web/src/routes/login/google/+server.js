import { redirect } from "@sveltejs/kit";
import { env } from "$env/dynamic/private";

export function GET({ url }) {
	const returnTo = url.searchParams.get("return_to") || "/account";
	const googleClientId = env.GOOGLE_OAUTH_CLIENT_ID;
	if (!googleClientId) {
		throw redirect(303, `/login?error=google_not_configured`);
	}

	const redirectUri = `${url.origin}/login/google/callback`;
	const scope = "openid email profile";
	const googleUrl = `https://accounts.google.com/o/oauth2/v2/auth?client_id=${googleClientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=${encodeURIComponent(scope)}&state=${encodeURIComponent(returnTo)}`;

	throw redirect(302, googleUrl);
}
