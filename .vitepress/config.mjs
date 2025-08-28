import { defineConfig } from 'vitepress'
import istioSidebar from '../docs/externalContent/istio/docs/user/_sidebar';
import apiGatewaySidebar from '../docs/externalContent/api-gateway/docs/user/_sidebar';
import sapBtpOperatorSidebar from '../docs/externalContent/btp-manager/docs/user/_sidebar';
import applicationConnectorSidebar from '../docs/externalContent/application-connector-manager/docs/user/_sidebar';
import cloudManagerSidebar from '../docs/externalContent/cloud-manager/docs/user/_sidebar';
import dockerRegistrySidebar from '../docs/externalContent/docker-registry/docs/user/_sidebar';
import eventingSidebar from '../docs/externalContent/eventing-manager/docs/user/_sidebar';
import kedaSidebar from '../docs/externalContent/keda-manager/docs/user/_sidebar';
import natsSidebar from '../docs/externalContent/nats-manager/docs/user/_sidebar';
import serverlessSidebar from '../docs/externalContent/serverless-manager/docs/user/_sidebar';
import telemetrySidebar from '../docs/externalContent/telemetry-manager/docs/user/_sidebar';
import cliSidebar from '../docs/externalContent/cli/docs/user/_sidebar';
import busolaSidebar from '../docs/externalContent/busola/docs/user/_sidebar';

export function getSearchConfig() {
  return {
      provider: 'local',
      detailedView: true,
      options: {
        detailedView: true,
        miniSearch: {
          /**
           * @type {Pick<import('minisearch').Options, 'extractField' | 'tokenize' | 'processTerm'>}
           */
          options: {
            // Configure how fields are extracted from documents
            extractField: (document, fieldName) => {
              // Extract frontmatter metadata for search
              if (fieldName === 'categories' && document.frontmatter?.categories) {
                return Array.isArray(document.frontmatter.categories)
                    ? document.frontmatter.categories.join(' ')
                    : document.frontmatter.categories;
              }
              if (fieldName === 'tags' && document.frontmatter?.tags) {
                return Array.isArray(document.frontmatter.tags)
                    ? document.frontmatter.tags.join(' ')
                    : document.frontmatter.tags;
              }
              if (fieldName === 'description' && document.frontmatter?.description) {
                return document.frontmatter.description;
              }
              if (fieldName === 'page_synonyms' && document.frontmatter?.page_synonyms) {
                return Array.isArray(document.frontmatter.page_synonyms)
                    ? document.frontmatter.page_synonyms.join(' ')
                    : document.frontmatter.page_synonyms;
              }
              // Extract default fields
              return document[fieldName];
            },
            // Custom tokenizer to handle special characters in technical docs
            tokenize: (text) => text.toLowerCase().split(/[\s\-_/]+/),
            // Process terms to improve search (e.g., stemming)
            processTerm: (term) => term.toLowerCase()
          },
          /**
           * @type {import('minisearch').SearchOptions}
           */
          searchOptions: {
            // Fuzzy search with prefix matching for better results
            fuzzy: 0.2,
            prefix: true,
            // Boosting: Give more weight to title, less to tags/categories
            boost: {
              title: 5,        // Most important
              text: 3,         // Body content
              headings: 4,     // Section headings
              tags: 2,         // Tags metadata
              categories: 2,   // Categories metadata
              description: 4,  // Description field
              page_synonyms: 3 // Synonyms/alternate terms
            },
            // Fields to search in
            fields: ['title', 'text', 'headings', 'tags', 'categories', 'description', 'page_synonyms']
          }
        }
      }
    }
}

function makeSidebarAbsolutePath(sidebar, objectName) {
  return sidebar.map(item => {
    const newItem = { ...item };

    if (item.link) {
      newItem.link = `/externalContent/${objectName}/docs/user/${item.link.replace('./', '')}`;
    }

    if (item.items) {
      newItem.items = makeSidebarAbsolutePath(item.items, objectName);
    }

    return newItem;
  });
}

// https://vitepress.dev/reference/site-config
export default defineConfig({
  srcDir: "docs",
  title: "Kyma",
  description: "Kyma documentation",
  lastUpdated: true,
  ignoreDeadLinks: true,
  base: '/kyma/documentation/',
  assetsDir:'vite-assets',
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    logo: {src: '/assets/logo_icon.svg', width: 24, height: 24},
    nav: [
      { text: 'Release Notes', link: '/release-notes' },
      { text: 'Support & Contribution', link: '/support-contribution' },
      { text: 'Kyma on SAP Community', link: 'https://community.sap.com/topics/kyma' }
    ],

    sidebar: [
      { text: 'Kyma Home', link: '/' },
      { text: 'Quick Install', link: '/02-get-started/01-quick-install' },
      {
        text: 'Modules',
        link: '/06-modules/README',
        items: [
          { text: 'Istio', link: '/externalContent/istio/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(istioSidebar, 'istio')},
          { text: 'API Gateway', link: '/externalContent/api-gateway/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(apiGatewaySidebar, 'api-gateway')},
          { text: 'SAP BTP Operator', link: '/externalContent/btp-manager/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(sapBtpOperatorSidebar, 'btp-manager')},
          { text: 'Application Connector', link: '/externalContent/application-connector-manager/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(applicationConnectorSidebar, 'application-connector-manager')},
          { text: 'Cloud Manager', link: '/externalContent/cloud-manager/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(cloudManagerSidebar, 'cloud-manager')},
          { text: 'Docker Registry', link: '/externalContent/docker-registry/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(dockerRegistrySidebar, 'docker-registry')},
          { text: 'Eventing', link: '/externalContent/eventing-manager/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(eventingSidebar, 'eventing-manager')},
          { text: 'Keda', link: '/externalContent/keda-manager/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(kedaSidebar, 'keda-manager')},
          { text: 'NATS', link: '/externalContent/nats-manager/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(natsSidebar, 'nats-manager')},
          { text: 'Serverless', link: '/externalContent/serverless-manager/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(serverlessSidebar, 'serverless-manager')},
          { text: 'Telemetry', link: '/externalContent/telemetry-manager/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(telemetrySidebar, 'telemetry-manager')}
        ]
      },
      {
        text: 'User Interfaces',
        link: '/01-overview/ui/README',
        items: [
          { text: 'Kyma Dashboard', link: '/externalContent/busola/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(busolaSidebar, 'busola') },
          { text: 'Kyma CLI', link: '/externalContent/cli/docs/user/README.md', collapsed: true, items: makeSidebarAbsolutePath(cliSidebar, 'cli') }
        ]
      },
      {
        text: 'Operation Guides',
        link: '/04-operation-guides/README',
        items: [
          { text: 'Operations', link: '/04-operation-guides/operations/README' },
          { text: 'Troubleshooting', link: '/04-operation-guides/troubleshooting/README' }
        ]
      },
      { text: 'Glossary', link: '/glossary' }
    ],
    head: [[
      'link',
      {rel: 'icon', type: 'image/svg+xml', href: '/assets/logo.svg'}
    ],['link', { rel: 'icon', href: '/assets/favicon.ico' }]
    ],
    socialLinks: [{
        icon: {
          svg: '<svg aria-hidden="true" class="svg-icon iconLogoGlyphMd native native" width="32" height="37" viewBox="0 0 32 37"><path d="M26 33v-9h4v13H0V24h4v9h22Z" fill="#BCBBBB"></path><path d="m21.5 0-2.7 2 9.9 13.3 2.7-2L21.5 0ZM26 18.4 13.3 7.8l2.1-2.5 12.7 10.6-2.1 2.5ZM9.1 15.2l15 7 1.4-3-15-7-1.4 3Zm14 10.79.68-2.95-16.1-3.35L7 23l16.1 2.99ZM23 30H7v-3h16v3Z" fill="#F48024"></path></svg>'
        },
        link: 'https://stackoverflow.com/questions/tagged/kyma'
      },
      {
        icon: {
          svg: '<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256"><path fill="#e01e5a" d="M53.841 161.32c0 14.832-11.987 26.82-26.819 26.82S.203 176.152.203 161.32c0-14.831 11.987-26.818 26.82-26.818H53.84zm13.41 0c0-14.831 11.987-26.818 26.819-26.818s26.819 11.987 26.819 26.819v67.047c0 14.832-11.987 26.82-26.82 26.82c-14.83 0-26.818-11.988-26.818-26.82z"/><path fill="#36c5f0" d="M94.07 53.638c-14.832 0-26.82-11.987-26.82-26.819S79.239 0 94.07 0s26.819 11.987 26.819 26.819v26.82zm0 13.613c14.832 0 26.819 11.987 26.819 26.819s-11.987 26.819-26.82 26.819H26.82C11.987 120.889 0 108.902 0 94.069c0-14.83 11.987-26.818 26.819-26.818z"/><path fill="#2eb67d" d="M201.55 94.07c0-14.832 11.987-26.82 26.818-26.82s26.82 11.988 26.82 26.82s-11.988 26.819-26.82 26.819H201.55zm-13.41 0c0 14.832-11.988 26.819-26.82 26.819c-14.831 0-26.818-11.987-26.818-26.82V26.82C134.502 11.987 146.489 0 161.32 0s26.819 11.987 26.819 26.819z"/><path fill="#ecb22e" d="M161.32 201.55c14.832 0 26.82 11.987 26.82 26.818s-11.988 26.82-26.82 26.82c-14.831 0-26.818-11.988-26.818-26.82V201.55zm0-13.41c-14.831 0-26.818-11.988-26.818-26.82c0-14.831 11.987-26.818 26.819-26.818h67.25c14.832 0 26.82 11.987 26.82 26.819s-11.988 26.819-26.82 26.819z"/></svg>'
        },
        link: 'https://kyma-community.slack.com/'
      },
      { icon: 'github', link: 'https://github.com/kyma-project' }
    ],
    search: getSearchConfig()
  }
})
