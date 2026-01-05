// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightClientMermaid from '@pasqal-io/starlight-client-mermaid';

// https://astro.build/config
export default defineConfig({
	site: 'https://iotactical.github.io',
	base: '/rtmx',
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
						content: '#0f172a',
					},
				},
				{
					tag: 'script',
					content: `
						document.addEventListener('DOMContentLoaded', () => {
							const header = document.querySelector('header.header');
							if (!header) return;
							const rightGroup = header.querySelector('.right-group, .sl-flex');
							if (!rightGroup) return;

							// Find the social icons container (GitHub link)
							const socialIcons = header.querySelector('.social-icons') ||
								header.querySelector('a[href*="github.com"]')?.parentElement;

							// "Don't Use RTMX" link - insert before GitHub icon
							const dontUseLink = document.createElement('a');
							dontUseLink.href = '/rtmx/#dont-use';
							dontUseLink.className = 'dont-use-link';
							dontUseLink.innerHTML = "<span class='strikethrough'>Don't</span>&nbsp;Use RTMX";

							// Sync button - insert before "Don't Use RTMX"
							const buySyncButton = document.createElement('a');
							buySyncButton.href = '/rtmx/pricing';
							buySyncButton.className = 'buy-sync-button';
							buySyncButton.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>Sync';

							if (socialIcons) {
								socialIcons.parentElement.insertBefore(dontUseLink, socialIcons);
								socialIcons.parentElement.insertBefore(buySyncButton, dontUseLink);
							} else {
								rightGroup.appendChild(buySyncButton);
								rightGroup.appendChild(dontUseLink);
							}
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
