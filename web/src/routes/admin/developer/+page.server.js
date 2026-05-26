import { env } from "$env/dynamic/private";
import { fetchJSON } from "../../../lib/api.js";

const ADMIN_API_KEY = `${env.KELOMPOK_ADMIN_API_KEY || ""}`.trim();

function checkEndpoint(path) {
	return fetchJSON(path)
		.then((payload) => ({ ok: true, payload }))
		.catch((error) => ({ ok: false, error: error instanceof Error ? error.message : "Health check failed" }));
}

export async function load() {
	const [health, ready, root] = await Promise.all([
		checkEndpoint("/healthz"),
		checkEndpoint("/readyz"),
		checkEndpoint("/"),
	]);

	const checks = [
		{
			label: "adminDeveloper.checkApiHealth",
			status: health.ok ? "pass" : "fail",
			detail: health.ok ? `status=${health.payload.data?.status || "ok"}` : health.error,
		},
		{
			label: "adminDeveloper.checkReadyHealth",
			status: ready.ok ? "pass" : "warn",
			detail: ready.ok ? `status=${ready.payload.data?.status || "ready"}` : ready.error,
		},
		{
			label: "adminDeveloper.checkRootEndpoint",
			status: root.ok ? "pass" : "warn",
			detail: root.ok ? "root endpoint responds" : root.error,
		},
		{
			label: "adminDeveloper.checkAdminAuth",
			status: ADMIN_API_KEY ? "pass" : "fail",
			detail: ADMIN_API_KEY ? "admin key configured for server-side admin requests" : "set KELOMPOK_ADMIN_API_KEY",
		},
	];

	return {
		checks,
		checkedAt: new Date().toISOString(),
	};
}
