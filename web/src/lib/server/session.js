import { env } from "$env/dynamic/private";
import { APIError, fetchJSON } from "$lib/api.js";

export const SESSION_COOKIE = "kelompok_session";
export const SESSION_UNVERIFIED_COOKIE = "kelompok_session_unverified";

export function sessionCookieOptions(expiresAt) {
	return {
		path: "/",
		httpOnly: true,
		sameSite: "lax",
		secure: env.NODE_ENV === "production",
		expires: new Date(expiresAt),
	};
}

function transientSessionCookieOptions() {
	return {
		path: "/",
		httpOnly: true,
		sameSite: "lax",
		secure: env.NODE_ENV === "production",
		maxAge: 300,
	};
}

export async function loadSession(cookies) {
	const token = cookies.get(SESSION_COOKIE);
	if (!token) {
		cookies.delete(SESSION_UNVERIFIED_COOKIE, { path: "/" });
		return null;
	}

	try {
		const payload = await fetchJSON("/api/v1/auth/me", {
			headers: {
				authorization: `Bearer ${token}`,
			},
		});
		cookies.delete(SESSION_UNVERIFIED_COOKIE, { path: "/" });
		return payload.data ?? null;
	} catch (error) {
		if (error instanceof APIError && (error.status === 401 || error.status === 403)) {
			cookies.delete(SESSION_COOKIE, { path: "/" });
			cookies.delete(SESSION_UNVERIFIED_COOKIE, { path: "/" });
		} else {
			cookies.set(SESSION_UNVERIFIED_COOKIE, "1", transientSessionCookieOptions());
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
	const session = payload?.data;
	if (!session?.token || !session?.expires_at) {
		throw new Error("Login response is missing session data");
	}

	cookies.set(SESSION_COOKIE, session.token, sessionCookieOptions(session.expires_at));
	cookies.delete(SESSION_UNVERIFIED_COOKIE, { path: "/" });
	return session;
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
	cookies.delete(SESSION_UNVERIFIED_COOKIE, { path: "/" });
}

export async function updateProfile(cookies, { name }) {
	const token = cookies.get(SESSION_COOKIE);
	if (!token) {
		throw new APIError("session_required", { status: 401, code: "session_required" });
	}

	const payload = await fetchJSON("/api/v1/auth/me", {
		method: "PATCH",
		headers: {
			authorization: `Bearer ${token}`,
			"content-type": "application/json",
		},
		body: JSON.stringify({ name }),
	});
	return payload?.data ?? null;
}
