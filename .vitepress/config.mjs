import { defineConfig } from 'vitepress'
import { nodePolyfills } from 'vite-plugin-node-polyfills'
import path from 'path-browserify'

const remoteAliases = {
  '/btp-manager/(.*)': 'https://raw.githubusercontent.com/kyma-project/btp-manager/main/docs/$1',
  '/application-connector-manager/(.*)': 'https://raw.githubusercontent.com/kyma-project/application-connector-manager/main/docs/$1',
  '/keda-manager/(.*)': 'https://raw.githubusercontent.com/kyma-project/keda-manager/main/docs/$1',
  '/serverless-manager/(.*)': 'https://raw.githubusercontent.com/kyma-project/serverless/main/docs/$1',
  '/telemetry-manager/(.*)': 'https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/$1',
  '/istio/(.*)': 'https://raw.githubusercontent.com/kyma-project/istio/main/docs/$1',
  '/nats-manager/(.*)': 'https://raw.githubusercontent.com/kyma-project/nats-manager/main/docs/$1',
  '/eventing-manager/(.*)': 'https://raw.githubusercontent.com/kyma-project/eventing-manager/main/docs/$1',
  '/api-gateway/(.*)': 'https://raw.githubusercontent.com/kyma-project/api-gateway/release-3.1/docs/$1',
  '/cloud-manager/(.*)': 'https://raw.githubusercontent.com/kyma-project/cloud-manager/main/docs/$1',
  '/docker-registry/(.*)': 'https://raw.githubusercontent.com/kyma-project/docker-registry/main/docs/$1',
  '/cli/(.*)': 'https://raw.githubusercontent.com/kyma-project/cli/main/docs/$1',
  '/busola/(.*)': 'https://raw.githubusercontent.com/kyma-project/busola/main/docs/$1',
};

// https://vitepress.dev/reference/site-config
export default defineConfig({
  srcDir: "docs",
  title: "Kyma Project",
  description: "**Kyma** `/'ki.ma/` Kyma is an opinionated set of Kubernetes-based modular building blocks, including all necessary capabilities to develop and run enterprise-grade cloud-native applications. It is the open path to the SAP ecosystem supporting business scenarios end-to-end.",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
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
          { text: 'Istio', link: '/istio/user/README.md' },
          { text: 'API Gateway', link: '/api-gateway/user/README.md' },
          { text: 'SAP BTP Operator', link: '/btp-manager/user/README.md' },
          { text: 'Application Connector', link: '/application-connector-manager/user/README.md' },
          { text: 'Cloud Manager', link: '/cloud-manager/user/README.md' },
          { text: 'Docker Registry', link: '/docker-registry/user/README.md' },
          { text: 'Eventing', link: '/eventing-manager/user/README.md' },
          { text: 'Keda', link: '/keda-manager/user/README.md' },
          { text: 'NATS', link: '/nats-manager/user/README.md' },
          { text: 'Serverless', link: '/serverless-manager/user/README.md' },
          { text: 'Telemetry', link: '/telemetry-manager/user/README.md' }
        ]
      },
      {
        text: 'User Interfaces',
        link: '/01-overview/ui/README',
        items: [
          { text: 'Kyma Dashboard', link: '/busola/user/README' },
          { text: 'Kyma CLI', link: '/cli/user/README' }
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

    socialLinks: [
      { icon: 'github', link: 'https://github.com/vuejs/vitepress' }
    ],
    alias: remoteAliases,
  },
  vite: {
    plugins: [
      nodePolyfills()
    ]
  }
})
