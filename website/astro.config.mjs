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
