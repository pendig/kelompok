import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";

// Hosts allowed to reach the Vite dev/preview server.
// Includes localhost (for in-LXC smoke), the public beta domain, and a
// leading-dot wildcard so future kelompok.id subdomains (e.g. staging.kelompok.id)
// keep working without another config change.
//
// Note: serving via `vite dev` behind a tunnel is acceptable for the current
// staging baseline, but the long-term plan is to switch to a production build
// served by @sveltejs/adapter-node. Until then, this allowlist is the minimum
// needed so Cloudflare tunnel traffic (Host: beta.kelompok.id) is not rejected
// by Vite's host check.
const allowedHosts = [
	"localhost",
	"127.0.0.1",
	"beta.kelompok.id",
	".kelompok.id",
];

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		allowedHosts,
	},
	preview: {
		allowedHosts,
	},
});
