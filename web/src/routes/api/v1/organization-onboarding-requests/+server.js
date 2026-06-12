const DEFAULT_API_BASE_URL = "http://localhost:4621";
const TARGET_PATH = "/api/v1/organization-onboarding-requests";

function apiBaseURL() {
	const configured = (import.meta.env.VITE_API_BASE_URL || "").trim();
	return (configured || DEFAULT_API_BASE_URL).replace(/\/$/, "");
}

function forwardedHeaders(request) {
	const headers = new Headers();

	for (const name of ["accept", "authorization", "content-type"]) {
		const value = request.headers.get(name);
		if (value) {
			headers.set(name, value);
		}
	}

	if (!headers.has("accept")) {
		headers.set("accept", "application/json");
	}

	return headers;
}

function responseHeaders(upstream) {
	const headers = new Headers();

	for (const name of ["content-type", "cache-control"]) {
		const value = upstream.headers.get(name);
		if (value) {
			headers.set(name, value);
		}
	}

	return headers;
}

export async function POST({ request, fetch }) {
	const upstream = await fetch(`${apiBaseURL()}${TARGET_PATH}`, {
		method: "POST",
		headers: forwardedHeaders(request),
		body: await request.text(),
	});

	return new Response(upstream.body, {
		status: upstream.status,
		statusText: upstream.statusText,
		headers: responseHeaders(upstream),
	});
}
