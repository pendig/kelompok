<script>
	import { t } from "$lib/i18n.js";

	const currentYear = new Date().getFullYear();

	// Social media URLs fetched from client-side environment variables
	const facebookUrl = import.meta.env.VITE_FOOTER_FACEBOOK_URL;
	const twitterUrl = import.meta.env.VITE_FOOTER_TWITTER_URL;
	const instagramUrl = import.meta.env.VITE_FOOTER_INSTAGRAM_URL;
	const githubUrl = import.meta.env.VITE_FOOTER_GITHUB_URL;
	const linkedinUrl = import.meta.env.VITE_FOOTER_LINKEDIN_URL;

	// Check if at least one social media link exists
	const hasSocials = facebookUrl || twitterUrl || instagramUrl || githubUrl || linkedinUrl;
</script>

<footer class="site-footer">
	<div class="container footer-inner">
		<div class="footer-grid">
			<!-- Brand and About Section -->
			<div class="footer-section brand-section">
				<a href="/" class="footer-logo" aria-label="Kelompok home">
					<img src="/brand/kelompokLogoHorizontalTransparant.png" alt="Kelompok Logo" class="footer-logo-img" />
				</a>
				<p class="footer-about-text">
					{$t("footer.aboutText")}
				</p>
				
				{#if hasSocials}
					<div class="social-links" aria-label="Social media links">
						{#if facebookUrl}
							<a href={facebookUrl} target="_blank" rel="noopener noreferrer" class="social-icon" aria-label="Facebook">
								<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 2h-3a5 5 0 0 0-5 5v3H7v4h3v8h4v-8h3l1-4h-4V7a1 1 0 0 1 1-1h3z"/></svg>
							</a>
						{/if}
						{#if twitterUrl}
							<a href={twitterUrl} target="_blank" rel="noopener noreferrer" class="social-icon" aria-label="Twitter">
								<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 4s-.7 2.1-2 3.4c1.6 10-9.4 17.3-18 11.6 2.2.1 4.4-.6 6-2C3 15.5.5 9.6 3 5c2.2 2.6 5.6 4.1 9 4-.9-4.2 4-6.6 7-3.8 1.1 0 3-1.2 3-1.2z"/></svg>
							</a>
						{/if}
						{#if instagramUrl}
							<a href={instagramUrl} target="_blank" rel="noopener noreferrer" class="social-icon" aria-label="Instagram">
								<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="20" rx="5" ry="5"/><path d="M16 11.37A4 4 0 1 1 12.63 8 4 4 0 0 1 16 11.37z"/><line x1="17.5" y1="6.5" x2="17.51" y2="6.5"/></svg>
							</a>
						{/if}
						{#if githubUrl}
							<a href={githubUrl} target="_blank" rel="noopener noreferrer" class="social-icon" aria-label="GitHub">
								<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22"/></svg>
							</a>
						{/if}
						{#if linkedinUrl}
							<a href={linkedinUrl} target="_blank" rel="noopener noreferrer" class="social-icon" aria-label="LinkedIn">
								<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 8a6 6 0 0 1 6 6v7h-4v-7a2 2 0 0 0-2-2 2 2 0 0 0-2 2v7h-4v-7a6 6 0 0 1 6-6z"/><rect x="2" y="9" width="4" height="12"/><circle cx="4" cy="4" r="2"/></svg>
							</a>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Quick Links Section -->
			<div class="footer-section links-section">
				<h3 class="footer-heading">{$t("footer.linksTitle")}</h3>
				<nav class="footer-nav" aria-label="Footer Quick Links">
					<a href="/organizations" class="footer-link">{$t("nav.organizations")}</a>
					<a href="/posts" class="footer-link">{$t("nav.posts")}</a>
					<a href="/sdgs" class="footer-link">{$t("nav.sdgs")}</a>
				</nav>
			</div>

			<!-- Legal & Terms Section -->
			<div class="footer-section links-section">
				<h3 class="footer-heading">{$t("footer.legalTitle")}</h3>
				<nav class="footer-nav" aria-label="Footer Legal Links">
					<a href="/terms" class="footer-link">{$t("auth.agreeLink")}</a>
				</nav>
			</div>
		</div>

		<!-- Footer Bottom Divider and Info -->
		<div class="footer-bottom">
			<p class="copyright-text">
				{$t("footer.rights", { year: currentYear })}
			</p>
			<p class="built-with-text">
				{$t("footer.builtWith")}
			</p>
		</div>
	</div>
</footer>

<style>
	.site-footer {
		background: var(--surface-soft);
		border-top: 1px solid var(--border);
		padding: 64px 0 48px;
		margin-top: auto;
		width: 100%;
		color: var(--muted);
	}

	.footer-inner {
		display: flex;
		flex-direction: column;
		gap: 48px;
	}

	.footer-grid {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr;
		gap: 48px;
	}

	.brand-section {
		display: flex;
		flex-direction: column;
		gap: 20px;
		max-width: 440px;
	}

	.footer-logo {
		display: inline-block;
		transition: opacity 0.2s ease;
	}

	.footer-logo:hover {
		opacity: 0.85;
	}

	.footer-logo-img {
		height: 36px;
		width: auto;
		object-fit: contain;
	}

	.footer-about-text {
		font-size: 14.5px;
		line-height: 1.6;
		margin: 0;
	}

	.social-links {
		display: flex;
		gap: 12px;
		margin-top: 8px;
	}

	.social-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background: var(--surface);
		border: 1px solid var(--border);
		color: var(--muted);
		transition: all 250ms cubic-bezier(0.4, 0, 0.2, 1);
	}

	.social-icon:hover {
		color: var(--purple-700);
		border-color: var(--purple-600);
		background: var(--surface-tint);
		transform: translateY(-3px);
		box-shadow: var(--shadow-sm);
	}

	:global([data-theme="dark"]) .social-icon:hover {
		color: var(--purple-600);
		border-color: var(--purple-700);
		background: var(--surface-tint);
	}

	.footer-heading {
		color: var(--text);
		font-size: 16px;
		font-weight: 700;
		margin: 0 0 20px 0;
		letter-spacing: 0.5px;
	}

	.footer-nav {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.footer-link {
		font-size: 14.5px;
		text-decoration: none;
		color: var(--muted);
		width: fit-content;
		transition: all 200ms ease;
	}

	.footer-link:hover {
		color: var(--purple-700);
		transform: translateX(3px);
	}

	:global([data-theme="dark"]) .footer-link:hover {
		color: var(--purple-600);
	}

	.footer-bottom {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-top: 24px;
		border-top: 1px solid var(--border);
		font-size: 13.5px;
		gap: 16px;
	}

	.copyright-text, .built-with-text {
		margin: 0;
	}

	/* Responsive Styles */
	@media (max-width: 720px) {
		.site-footer {
			padding: 48px 0 110px; /* Extra bottom padding so it is not covered by floating-bottom-nav */
		}

		.footer-grid {
			grid-template-columns: 1fr;
			gap: 36px;
		}

		.footer-bottom {
			flex-direction: column;
			align-items: flex-start;
			gap: 12px;
		}
	}
</style>
