const DEFAULT_API_BASE_URL = "http://localhost:4621";
const DEFAULT_FETCH_TIMEOUT_MS = 3500;

export class APIError extends Error {
	constructor(message, { status, code, details } = {}) {
		super(message);
		this.name = "APIError";
		this.status = status;
		this.code = code;
		this.details = details;
	}
}

function normalize(path) {
	if (!path.startsWith("/")) {
		return `/${path}`;
	}
	return path;
}

function readConfig() {
	const configured = (import.meta.env.VITE_API_BASE_URL || "").trim();
	if (configured.length === 0) {
		return DEFAULT_API_BASE_URL;
	}
	return configured.replace(/\/$/, "");
}

async function readResponse(response) {
	const contentType = response.headers.get("content-type") || "";
	if (!contentType.includes("application/json")) {
		const text = await response.text();
		throw new Error(`Unexpected response from API (${response.status}): ${text.slice(0, 200)}`);
	}

	const payload = await response.json();
	if (!response.ok) {
		const details = payload?.error;
		throw new APIError(details ? `${details.code}: ${details.message}` : `HTTP ${response.status}`, {
			status: response.status,
			code: details?.code,
			details: details?.details,
		});
	}

	return payload;
}

export async function fetchJSON(path, init = {}) {
	const baseUrl = readConfig();
	const url = `${baseUrl}${normalize(path)}`;
	const controller = init.signal ? null : new AbortController();
	const timeout = controller
		? setTimeout(() => controller.abort(), DEFAULT_FETCH_TIMEOUT_MS)
		: null;

	try {
		const response = await fetch(url, {
			...init,
			headers: {
				"accept": "application/json",
				...(init.headers || {}),
			},
			signal: init.signal ?? controller.signal,
		});

		return readResponse(response);
	} finally {
		if (timeout) {
			clearTimeout(timeout);
		}
	}
}

export async function fetchJSONResult(path, fallbackData = []) {
	try {
		const payload = await fetchJSON(path);
		return {
			data: payload.data ?? fallbackData,
			error: null,
		};
	} catch (error) {
		return {
			data: fallbackData,
			error: error instanceof Error ? error.message : "Unable to load data from API",
		};
	}
}

export function fallbackDate(value, locale = "en-US") {
	if (!value) {
		return "—";
	}
	const date = new Date(value);
	if (Number.isNaN(date.getTime())) {
		return value;
	}

	const dateLocale = locale === "id" ? "id-ID" : "en-US";
	// Pin the timezone so SSR (UTC container) and the client (visitor's local zone)
	// format the same calendar day. Without this, an evening-UTC timestamp renders
	// e.g. "30 Mei 2026" on the server but "31 Mei 2026" in WIB, which Svelte reports
	// as a hydration_mismatch. Asia/Jakarta is the platform's primary audience zone.
	return new Intl.DateTimeFormat(dateLocale, { dateStyle: "medium", timeZone: "Asia/Jakarta" }).format(date);
}
