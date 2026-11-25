import { defineConfig } from 'vitepress';
import tabsPlugin from './plugins/tabs-plugin';
import { readFileSync } from 'fs';
import { resolve, dirname, sep } from 'path';
import { fileURLToPath } from 'url';
import { parse } from '@babel/parser';

const __dirname = dirname(fileURLToPath(import.meta.url));

// All statically-resolved files must stay within this root. A relative import
// in a legacy sidebar can only pull in other files from the copied external
// content — never escape to an arbitrary path on the build machine.
const EXTERNAL_CONTENT_ROOT = resolve(__dirname, '../docs/external-content');
const MAX_IMPORT_DEPTH = 5;
// Extensions tried when resolving a relative import like './gen-docs/_sidebar'.
const IMPORT_EXTS = ['', '.json', '.ts', '.js'];

// Names of modules still shipping a legacy JS/TS sidebar; reported once as a
// single migration reminder after all sidebars are loaded.
const legacySidebars = [];

// Turns a Babel AST node into a plain JS value WITHOUT executing any code.
// Only pure literals are accepted. A reference to a locally-imported module
// (e.g. `items: commandsSidebar`) is resolved by statically parsing that
// module too; anything else (function call, spread, dynamic expression, ...)
// throws.
function literalFromNode(node, ctx) {
  switch (node.type) {
    case 'StringLiteral':
    case 'NumericLiteral':
    case 'BooleanLiteral':
      return node.value;
    case 'NullLiteral':
      return null;
    case 'TemplateLiteral':
      if (node.expressions.length === 0) {
        return node.quasis[0].value.cooked;
      }
      throw new Error('template literal with embedded expression');
    case 'UnaryExpression':
      if (node.operator === '-') return -literalFromNode(node.argument, ctx);
      if (node.operator === '+') return +literalFromNode(node.argument, ctx);
      throw new Error(`unsupported unary operator "${node.operator}"`);
    case 'ArrayExpression':
      return node.elements.map((el) => {
        if (el === null) throw new Error('sparse array element');
        if (el.type === 'SpreadElement') throw new Error('spread element');
        return literalFromNode(el, ctx);
      });
    case 'ObjectExpression': {
      const obj = {};
      for (const prop of node.properties) {
        if (prop.type !== 'ObjectProperty') {
          throw new Error(`unsupported property "${prop.type}"`);
        }
        if (prop.computed) throw new Error('computed object key');
        let key;
        if (prop.key.type === 'Identifier') {
          key = prop.key.name;
        } else if (prop.key.type === 'StringLiteral') {
          key = prop.key.value;
        } else if (prop.key.type === 'NumericLiteral') {
          key = String(prop.key.value);
        } else {
          throw new Error(`unsupported key type "${prop.key.type}"`);
        }
        obj[key] = literalFromNode(prop.value, ctx);
      }
      return obj;
    }
    case 'Identifier': {
      const binding = ctx.imports.get(node.name);
      if (!binding) throw new Error(`unresolved identifier "${node.name}"`);
      return resolveImportedExport(binding, ctx);
    }
    default:
      throw new Error(`non-literal expression "${node.type}"`);
  }
}

// Collects import bindings of a parsed module: localName -> { source, imported }.
function collectImports(ast) {
  const imports = new Map();
  for (const node of ast.program.body) {
    if (node.type !== 'ImportDeclaration') continue;
    for (const spec of node.specifiers) {
      if (spec.type === 'ImportDefaultSpecifier') {
        imports.set(spec.local.name, {
          source: node.source.value,
          imported: 'default',
        });
      } else if (spec.type === 'ImportSpecifier') {
        imports.set(spec.local.name, {
          source: node.source.value,
          imported:
            spec.imported.type === 'Identifier'
              ? spec.imported.name
              : spec.imported.value,
        });
      }
      // Namespace imports (`import * as x`) are unsupported and will surface
      // as an "unresolved identifier" if referenced.
    }
  }
  return imports;
}

// Resolves a relative import binding to the exported literal of the target
// file, which is parsed statically (never executed) and must live inside the
// external-content root.
function resolveImportedExport(binding, ctx) {
  if (ctx.depth >= MAX_IMPORT_DEPTH) {
    throw new Error('import nesting too deep');
  }
  const { source, imported } = binding;
  if (!source.startsWith('./') && !source.startsWith('../')) {
    throw new Error(`non-relative import "${source}"`);
  }

  const resolvedBase = resolve(ctx.dir, source);
  let filePath = null;
  for (const ext of IMPORT_EXTS) {
    const candidate = resolvedBase + ext;
    try {
      readFileSync(candidate);
      filePath = candidate;
      break;
    } catch (err) {
      if (err.code !== 'ENOENT' && err.code !== 'EISDIR') throw err;
    }
  }
  if (!filePath) throw new Error(`cannot resolve import "${source}"`);

  if (
    filePath !== EXTERNAL_CONTENT_ROOT &&
    !filePath.startsWith(EXTERNAL_CONTENT_ROOT + sep)
  ) {
    throw new Error(`import "${source}" escapes external-content root`);
  }

  return loadModuleExport(filePath, imported, ctx.depth + 1);
}

// Reads a JS/TS/JSON module file as DATA and statically evaluates one of its
// exports. The file is never imported or executed, so malicious top-level
// code in an external repo cannot run here.
function loadModuleExport(filePath, exportName, depth) {
  const src = readFileSync(filePath, 'utf8');

  if (filePath.endsWith('.json')) {
    if (exportName !== 'default') {
      throw new Error(`JSON module has no named export "${exportName}"`);
    }
    return JSON.parse(src);
  }

  const ast = parse(src, { sourceType: 'module', plugins: ['typescript'] });
  const ctx = { imports: collectImports(ast), dir: dirname(filePath), depth };

  for (const node of ast.program.body) {
    if (exportName === 'default' && node.type === 'ExportDefaultDeclaration') {
      return literalFromNode(node.declaration, ctx);
    }
    if (node.type === 'ExportNamedDeclaration') {
      // export const NAME = [...]
      if (node.declaration && node.declaration.type === 'VariableDeclaration') {
        for (const d of node.declaration.declarations) {
          if (d.id.type === 'Identifier' && d.id.name === exportName) {
            return literalFromNode(d.init, ctx);
          }
        }
      }
      // export { NAME }
      for (const spec of node.specifiers ?? []) {
        if (spec.type !== 'ExportSpecifier') continue;
        const exportedName =
          spec.exported.type === 'Identifier'
            ? spec.exported.name
            : spec.exported.value;
        if (exportedName === exportName) {
          return literalFromNode(
            { type: 'Identifier', name: spec.local.name },
            ctx,
          );
        }
      }
    }
  }
  throw new Error(`export "${exportName}" not found`);
}

// Loads a module's sidebar. Prefers the inert _sidebar.json; falls back to a
// legacy _sidebar.{ts,js} parsed as data (never executed) with a migration
// notice. Any failure is logged loudly and degrades to an empty sidebar
// instead of silently breaking the whole build.
function readSidebar(repoName) {
  const base = resolve(
    __dirname,
    `../docs/external-content/${repoName}/docs/user`,
  );

  // Recommended format: inert JSON, parsed directly.
  try {
    const data = JSON.parse(readFileSync(`${base}/_sidebar.json`, 'utf8'));
    if (!Array.isArray(data)) throw new Error('_sidebar.json is not an array');
    return data;
  } catch (err) {
    if (err.code !== 'ENOENT') {
      console.error(
        `[sidebar] ${repoName}: _sidebar.json is present but invalid (${err.message}) — sidebar will be empty.`,
      );
      return [];
    }
  }

  // Legacy format: JS/TS module read as text and parsed via AST, never run.
  for (const ext of ['_sidebar.ts', '_sidebar.js', '_sidebar']) {
    const filePath = `${base}/${ext}`;
    try {
      readFileSync(filePath);
    } catch (err) {
      if (err.code === 'ENOENT') continue;
      throw err;
    }
    legacySidebars.push(repoName);
    try {
      const data = loadModuleExport(filePath, 'default', 0);
      if (!Array.isArray(data)) {
        throw new Error('default export is not an array');
      }
      return data;
    } catch (err) {
      console.error(
        `[sidebar] ${repoName}: "${ext}" could not be statically parsed (${err.message}). ` +
          `Sidebar will be EMPTY until it is migrated to _sidebar.json.`,
      );
      return [];
    }
  }

  console.error(
    `[sidebar] ${repoName}: no _sidebar.json or _sidebar.{ts,js} found — sidebar will be empty.`,
  );
  return [];
}

const istioSidebar = readSidebar('istio');
const apiGatewaySidebar = readSidebar('api-gateway');
const sapBtpOperatorSidebar = readSidebar('btp-manager');
const applicationConnectorSidebar = readSidebar(
  'application-connector-manager',
);
const cloudManagerSidebar = readSidebar('cloud-manager');
const dockerRegistrySidebar = readSidebar('docker-registry');
const eventingSidebar = readSidebar('eventing-manager');
const kedaSidebar = readSidebar('keda-manager');
const natsSidebar = readSidebar('nats-manager');
const registryCacheSidebar = readSidebar('registry-cache');
const serverlessSidebar = readSidebar('serverless');
const telemetrySidebar = readSidebar('telemetry-manager');
const cliSidebar = readSidebar('cli');
const busolaSidebar = readSidebar('busola');
const registryProxySidebar = readSidebar('registry-proxy');
const communityModulesSidebar = readSidebar('community-modules');

if (legacySidebars.length > 0) {
  console.warn(
    `[sidebar] ${legacySidebars.length} module(s) still use a legacy JS/TS sidebar and ` +
      `should migrate to an inert _sidebar.json: ${legacySidebars.join(', ')}.`,
  );
}

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
            const frontmatterValue = getFrontmatterFieldValue(
              document,
              fieldName,
            );
            if (frontmatterValue !== undefined) {
              return frontmatterValue;
            }

            // Fallback to default field
            return document[fieldName];
          },
          // Custom tokenizer to handle special characters in technical docs
          tokenize: (text) => text.toLowerCase().split(/[\s\-_/]+/),
          // Process terms to improve search (e.g., stemming)
          processTerm: (term) => term.toLowerCase(),
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
            title: 5, // Most important
            text: 3, // Body content
            headings: 4, // Section headings
            tags: 2, // Tags metadata
            categories: 2, // Categories metadata
            description: 4, // Description field
            page_synonyms: 3, // Synonyms/alternate terms
          },
          // Fields to search in
          fields: [
            'title',
            'text',
            'headings',
            'tags',
            'categories',
            'description',
            'page_synonyms',
          ],
        },
      },
    },
  };
}

function getFrontmatterFieldValue(document, fieldName) {
  const frontmatter = document?.frontmatter;
  if (!frontmatter || !frontmatter[fieldName]) return undefined;

  const fieldsToJoin = ['categories', 'tags', 'page_synonyms'];

  if (fieldsToJoin.includes(fieldName)) {
    return Array.isArray(frontmatter[fieldName])
      ? frontmatter[fieldName].join(' ')
      : frontmatter[fieldName];
  }

  if (fieldName === 'description') {
    return frontmatter.description;
  }

  return undefined;
}

function makeSidebarAbsolutePath(sidebar, objectName) {
  return sidebar.map((item) => {
    const newItem = { ...item };

    if (item.link) {
      newItem.link = `/external-content/${objectName}/docs/user/${item.link.replace('./', '')}`;
    }

    if (item.items) {
      newItem.items = makeSidebarAbsolutePath(item.items, objectName);
    }

    return newItem;
  });
}

// https://vitepress.dev/reference/site-config
export default defineConfig({
  srcDir: 'docs',
  title: 'Kyma',
  description: 'Kyma documentation',
  lastUpdated: true,
  ignoreDeadLinks: true,
  base: '/',
  assetsDir: 'vite-assets',
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    editLink: {
      pattern: ({ filePath }) => {
        if (filePath.startsWith('external-content/')) {
          const [, repoName, ...repoPathParts] = filePath.split('/');
          return `https://github.com/kyma-project/${repoName}/edit/main/${repoPathParts.join('/')}`;
        }

        return `https://github.com/kyma-project/kyma/edit/main/docs/${filePath}`;
      },
      text: 'Edit this page on GitHub',
    },
    logo: { src: '/assets/logo_icon.svg', width: 24, height: 24 },
    nav: [
      { text: 'Documentation', link: '/README' },
      { text: 'Release Notes', link: '/release-notes' },
      { text: 'Support & Contribution', link: '/support-contribution' },
      {
        text: 'Kyma on SAP Community',
        link: 'https://community.sap.com/topics/kyma',
      },
    ],

    sidebar: [
      { text: 'Quick Install', link: '/02-get-started/01-quick-install' },
      {
        text: 'Modules',
        link: '/06-modules/README',
        items: [
          {
            text: 'Istio',
            link: '/external-content/istio/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(istioSidebar, 'istio'),
          },
          {
            text: 'API Gateway',
            link: '/external-content/api-gateway/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(apiGatewaySidebar, 'api-gateway'),
          },
          {
            text: 'SAP BTP Operator',
            link: '/external-content/btp-manager/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(
              sapBtpOperatorSidebar,
              'btp-manager',
            ),
          },
          {
            text: 'Application Connector',
            link: '/external-content/application-connector-manager/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(
              applicationConnectorSidebar,
              'application-connector-manager',
            ),
          },
          {
            text: 'Cloud Manager',
            link: '/external-content/cloud-manager/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(
              cloudManagerSidebar,
              'cloud-manager',
            ),
          },
          {
            text: 'Eventing',
            link: '/external-content/eventing-manager/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(eventingSidebar, 'eventing-manager'),
          },
          {
            text: 'Keda',
            link: '/external-content/keda-manager/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(kedaSidebar, 'keda-manager'),
          },
          {
            text: 'NATS',
            link: '/external-content/nats-manager/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(natsSidebar, 'nats-manager'),
          },
          {
            text: 'Registry Cache',
            link: '/external-content/registry-cache/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(
              registryCacheSidebar,
              'registry-cache',
            ),
          },
          {
            text: 'Serverless',
            link: '/external-content/serverless/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(serverlessSidebar, 'serverless'),
          },
          {
            text: 'Telemetry',
            link: '/external-content/telemetry-manager/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(
              telemetrySidebar,
              'telemetry-manager',
            ),
          },
        ],
      },
      {
        text: 'Community Modules',
        link: '/external-content/community-modules/docs/user/README.md',
        items: [
          ...makeSidebarAbsolutePath(
            communityModulesSidebar,
            'community-modules',
          ),
          {
            text: 'Registry Proxy',
            link: '/external-content/registry-proxy/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(
              registryProxySidebar,
              'registry-proxy',
            ),
          },
          {
            text: 'Docker Registry',
            link: '/external-content/docker-registry/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(
              dockerRegistrySidebar,
              'docker-registry',
            ),
          },
          {
            text: 'CAP Operator',
            link: 'https://sap.github.io/cap-operator/',
          },
        ],
      },
      {
        text: 'User Interfaces',
        link: '/01-overview/ui/README',
        items: [
          {
            text: 'Kyma Dashboard',
            link: '/external-content/busola/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(busolaSidebar, 'busola'),
          },
          {
            text: 'Kyma CLI',
            link: '/external-content/cli/docs/user/README.md',
            collapsed: true,
            items: makeSidebarAbsolutePath(cliSidebar, 'cli'),
          },
        ],
      },
      {
        text: 'Operation Guides',
        link: '/04-operation-guides/README',
        items: [
          {
            text: 'Operations',
            link: '/04-operation-guides/operations/README',
          },
          {
            text: 'Troubleshooting',
            link: '/04-operation-guides/troubleshooting/README',
          },
        ],
      },
      { text: 'Glossary', link: '/glossary' },
    ],
    socialLinks: [
      {
        icon: {
          svg: '<svg aria-hidden="true" class="svg-icon iconLogoGlyphMd native native" width="32" height="37" viewBox="0 0 32 37"><path d="M26 33v-9h4v13H0V24h4v9h22Z" fill="#BCBBBB"></path><path d="m21.5 0-2.7 2 9.9 13.3 2.7-2L21.5 0ZM26 18.4 13.3 7.8l2.1-2.5 12.7 10.6-2.1 2.5ZM9.1 15.2l15 7 1.4-3-15-7-1.4 3Zm14 10.79.68-2.95-16.1-3.35L7 23l16.1 2.99ZM23 30H7v-3h16v3Z" fill="#F48024"></path></svg>',
        },
        link: 'https://stackoverflow.com/questions/tagged/kyma',
      },
      {
        icon: {
          svg: '<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256"><path fill="#e01e5a" d="M53.841 161.32c0 14.832-11.987 26.82-26.819 26.82S.203 176.152.203 161.32c0-14.831 11.987-26.818 26.82-26.818H53.84zm13.41 0c0-14.831 11.987-26.818 26.819-26.818s26.819 11.987 26.819 26.819v67.047c0 14.832-11.987 26.82-26.82 26.82c-14.83 0-26.818-11.988-26.818-26.82z"/><path fill="#36c5f0" d="M94.07 53.638c-14.832 0-26.82-11.987-26.82-26.819S79.239 0 94.07 0s26.819 11.987 26.819 26.819v26.82zm0 13.613c14.832 0 26.819 11.987 26.819 26.819s-11.987 26.819-26.82 26.819H26.82C11.987 120.889 0 108.902 0 94.069c0-14.83 11.987-26.818 26.819-26.818z"/><path fill="#2eb67d" d="M201.55 94.07c0-14.832 11.987-26.82 26.818-26.82s26.82 11.988 26.82 26.82s-11.988 26.819-26.82 26.819H201.55zm-13.41 0c0 14.832-11.988 26.819-26.82 26.819c-14.831 0-26.818-11.987-26.818-26.82V26.82C134.502 11.987 146.489 0 161.32 0s26.819 11.987 26.819 26.819z"/><path fill="#ecb22e" d="M161.32 201.55c14.832 0 26.82 11.987 26.82 26.818s-11.988 26.82-26.82 26.82c-14.831 0-26.818-11.988-26.818-26.82V201.55zm0-13.41c-14.831 0-26.818-11.988-26.818-26.82c0-14.831 11.987-26.818 26.819-26.818h67.25c14.832 0 26.82 11.987 26.82 26.819s-11.988 26.819-26.82 26.819z"/></svg>',
        },
        link: 'https://kyma-community.slack.com/',
      },
      { icon: 'github', link: 'https://github.com/kyma-project' },
    ],
    search: getSearchConfig(),
  },
  markdown: {
    config: (md) => {
      md.use(tabsPlugin);
      // Prevent Vue from interpreting {{ }} (e.g., Go templates) in inline code from external docs.
      const defaultCodeInline = md.renderer.rules.code_inline;
      md.renderer.rules.code_inline = (tokens, idx, options, env, self) => {
        tokens[idx].attrSet('v-pre', '');
        return defaultCodeInline(tokens, idx, options, env, self);
      };
    },
  },
});
