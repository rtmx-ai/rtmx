// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightClientMermaid from '@pasqal-io/starlight-client-mermaid';

// https://astro.build/config
export default defineConfig({
	site: 'https://rtmx.ai',
	base: '/',
	integrations: [
		starlight({
			plugins: [starlightClientMermaid()],
			title: 'RTMX',
			description: 'Requirements Traceability Matrix for Python - AI-native test traceability',
			expressiveCode: {
				frames: false,
			},
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/iotactical/rtmx' },
			],
			logo: {
				light: './src/assets/rtmx-logo-light.svg',
				dark: './src/assets/rtmx-logo-dark.svg',
				replacesTitle: false,
			},
			customCss: [
				'./src/styles/custom.css',
			],
			defaultLocale: 'en',
			head: [
				{
					tag: 'meta',
					attrs: {
						name: 'theme-color',
						content: '#0a0a0a',
					},
				},
				{
					// Force dark mode immediately before page renders
					tag: 'script',
					content: `document.documentElement.dataset.theme = 'dark';`,
				},
				{
					tag: 'script',
					content: `
						document.addEventListener('DOMContentLoaded', () => {
							// Ensure dark mode stays forced
							document.documentElement.dataset.theme = 'dark';
							const header = document.querySelector('header.header');
							if (!header) return;

							// Check if we already added nav (avoid duplicates)
							if (header.querySelector('.header-nav')) return;

							// Find the title wrapper to insert nav links inside it
							const titleWrapper = header.querySelector('.title-wrapper');
							if (!titleWrapper) return;

							// Create simple nav links
							const nav = document.createElement('nav');
							nav.className = 'header-nav';
							nav.innerHTML = \`
								<a href="/quickstart" class="header-nav-link">Docs</a>
								<a href="/roadmap" class="header-nav-link">Roadmap</a>
								<a href="/security" class="header-nav-link">Security</a>
								<a href="/enterprise" class="header-nav-link">Enterprise</a>
								<a href="/pricing" class="header-nav-link">Pricing</a>
							\`;

							// Append nav inside the title wrapper
							titleWrapper.appendChild(nav);
						});
					`,
				},
			],
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'index' },
						{ label: 'Installation', slug: 'installation' },
						{ label: 'Quickstart', slug: 'quickstart' },
					],
				},
				{
					label: 'Guides',
					items: [
						{ label: 'CLI Reference', slug: 'guides/cli-reference' },
						{ label: 'Test Markers', slug: 'guides/markers' },
						{ label: 'Dependencies', slug: 'guides/dependencies' },
						{ label: 'Schema', slug: 'guides/schema' },
						{ label: 'Configuration', slug: 'guides/configuration' },
					],
				},
				{
					label: 'Adapters',
					items: [
						{ label: 'Overview', slug: 'adapters' },
						{ label: 'GitHub', slug: 'adapters/github' },
						{ label: 'Jira', slug: 'adapters/jira' },
						{ label: 'MCP Server', slug: 'adapters/mcp' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'Lifecycle', slug: 'reference/lifecycle' },
						{ label: 'Architecture', slug: 'reference/architecture' },
					],
				},
			],
		}),
	],
});
