import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'do',
  tagline: 'Typesafe dependency injection for Go',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  url: 'https://do.samber.dev',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'samber', // Usually your GitHub org/user name.
  projectName: 'do', // Usually your repo name.

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/samber/do/tree/master/docs/',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
        gtag: {
          trackingID: 'G-ZQ0MR5WG9T',
          anonymizeIP: false,
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    // image: 'img/docusaurus-social-card.jpg',
    navbar: {
      title: '⚙️ do',
      // logo: {
      //   alt: 'My Site Logo',
      //   src: 'img/logo.svg',
      // },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docSidebar',
          position: 'left',
          label: 'Guides',
          // label: 'Docs',
        },
        {
          to: 'examples',
          label: 'Examples',
          position: 'left',
        },
        {
          to: 'https://pkg.go.dev/github.com/samber/do/v2',
          label: 'GoDoc',
          position: 'left',
        },
        {
          to: 'community',
          label: 'Community',
          position: 'left',
        },
        {
          to: 'https://github.com/samber/do/releases',
          label: 'Changelog',
          position: 'right',
        },
        {
          to: 'https://github.com/samber/do',
          // label: 'GitHub',
          position: 'right',
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Project',
          items: [
            {
              label: 'Documentation',
              to: '/docs/getting-started',
            },
            {
              label: 'Changelog',
              to: 'https://github.com/samber/do/releases',
            },
            {
              label: 'Godoc',
              to: 'https://pkg.go.dev/github.com/samber/do/v2',
            },
            {
              label: 'License',
              to: 'https://github.com/samber/do/blob/master/LICENSE',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'New issue',
              to: 'https://github.com/samber/do/issues',
            },
            {
              label: 'GitHub',
              to: 'https://github.com/samber/do',
            },
            {
              label: 'Stack Overflow',
              to: 'https://stackoverflow.com/search?q=samber+do',
            },
            {
              label: 'Twitter',
              to: 'https://twitter.com/samuelberthe',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} do.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
    algolia: {
      appId: 'VDJWQ4V7HW',
      apiKey: '529072f02f10644f56797297725f57db',
      externalUrlRegex: 'do\\.samber\\.dev',
      indexName: 'doc',
      contextualSearch: true,
    },
  } satisfies Preset.ThemeConfig,

  plugins: [
  ],
};

export default config;
