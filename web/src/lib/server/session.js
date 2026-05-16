import { env } from "$env/dynamic/private";
import { APIError, fetchJSON } from "$lib/api.js";

export const SESSION_COOKIE = "kelompok_session";

export function sessionCookieOptions(expiresAt) {
	return {
		path: "/",
		httpOnly: true,
		sameSite: "lax",
		secure: env.NODE_ENV === "production",
		expires: new Date(expiresAt),
	};
}

export async function loadSession(cookies) {
	const token = cookies.get(SESSION_COOKIE);
	if (!token) {
		return null;
	}

	try {
		const payload = await fetchJSON("/api/v1/auth/me", {
			headers: {
				authorization: `Bearer ${token}`,
			},
		});
		return payload.data ?? null;
	} catch (error) {
		if (error instanceof APIError && (error.status === 401 || error.status === 403)) {
			cookies.delete(SESSION_COOKIE, { path: "/" });
		}
		return null;
	}
}

export async function loginWithPassword(cookies, email, password) {
	const payload = await fetchJSON("/api/v1/auth/login", {
		method: "POST",
		headers: {
			"content-type": "application/json",
		},
		body: JSON.stringify({ email, password }),
	});
	cookies.set(SESSION_COOKIE, payload.data.token, sessionCookieOptions(payload.data.expires_at));
	return payload.data;
}

export async function logoutSession(cookies) {
	const token = cookies.get(SESSION_COOKIE);
	if (token) {
		try {
			await fetchJSON("/api/v1/auth/logout", {
				method: "POST",
				headers: {
					authorization: `Bearer ${token}`,
				},
			});
		} catch {
			// Local cookie cleanup should still happen if the API session is already gone.
		}
	}
	cookies.delete(SESSION_COOKIE, { path: "/" });
}
