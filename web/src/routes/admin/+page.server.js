import { fetchJSON, fetchJSONResult } from "../../lib/api.js";

function checkEndpoint(path) {
	return fetchJSON(path)
		.then((payload) => ({ ok: true, payload }))
		.catch((error) => ({ ok: false, error: error instanceof Error ? error.message : "Health check failed" }));
}

export async function load() {
	const [orgPayload, postPayload, health, ready, root] = await Promise.all([
		fetchJSONResult("/api/v1/organizations?limit=30"),
		fetchJSONResult("/api/v1/posts?limit=30"),
		checkEndpoint("/healthz"),
		checkEndpoint("/readyz"),
		checkEndpoint("/"),
	]);

	const organizations = orgPayload.data ?? [];
	const posts = postPayload.data ?? [];

	const impactByOrganization = await Promise.all(
		(organizations || [])
			.slice(0, 6)
			.map((org) => {
				if (!org.slug) {
					return Promise.resolve({
						orgSlug: null,
						orgName: org.name || "Unnamed org",
						count: 0,
						latest: null,
						error: "No valid slug available for impact lookup",
					});
				}

				return fetchJSONResult(`/api/v1/organizations/${encodeURIComponent(org.slug)}/impact-reports?limit=6`).then(
					(impactPayload) => ({
						orgSlug: org.slug,
						orgName: org.name || "Unnamed org",
						count: impactPayload.data?.length ?? 0,
						latest: impactPayload.data?.[0]?.title ?? null,
						error: impactPayload.error,
					}),
				);
			}),
	);

	const missingClaimStatusOrgs = organizations.filter((item) => item.claim_status === undefined).length;
	const checks = [
		{
			label: "admin.checkApiHealth",
			status: health.ok ? "pass" : "fail",
			detail: health.ok ? `status=${health.payload.data?.status || "ok"}` : health.error,
			needsReview: false,
		},
		{
			label: "admin.checkReadyHealth",
			status: ready.ok ? "pass" : "warn",
			detail: ready.ok ? `status=${ready.payload.data?.status || "ready"}` : ready.error,
			needsReview: !ready.ok,
		},
		{
			label: "admin.checkHasOrganizations",
			status: organizations.length > 0 ? "pass" : "warn",
			detail:
				organizations.length > 0 ? `${organizations.length} orgs visible in public API` : "No organization data found",
			needsReview: true,
		},
		{
			label: "admin.checkHasPosts",
			status: posts.length > 0 ? "pass" : "warn",
			detail: posts.length > 0 ? `${posts.length} public posts loaded` : "No public posts found",
			needsReview: true,
		},
		{
			label: "admin.checkHasImpact",
			status: impactByOrganization.some((entry) => entry.count > 0) ? "pass" : "warn",
			detail:
				impactByOrganization.some((entry) => entry.count > 0) ?
					"Impact reports available for at least one organization"
				:	"Impact endpoint works, but no data found in sampled organizations",
			needsReview: true,
		},
		{
			label: "admin.checkHasClaimStatus",
			status: organizations.length > 0 && missingClaimStatusOrgs === 0 ? "pass" : "warn",
			detail:
				organizations.length === 0 ?
					"No public organizations loaded yet"
				: missingClaimStatusOrgs > 0
				? `${missingClaimStatusOrgs} organization(s) missing claim status`
				: `${organizations.length} records include claim status`,
			needsReview: missingClaimStatusOrgs > 0 || organizations.length === 0,
		},
	];

	return {
		organizations,
		posts,
		health,
		ready,
		root,
		impactByOrganization,
		missingClaimStatusOrgs,
		loadErrors: [orgPayload.error, postPayload.error, health.ok ? null : health.error, ready.ok ? null : ready.error, root.ok ? null : root.error].filter(Boolean),
		checks,
	};
}
