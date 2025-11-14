# VitePress Project Structure and Automation for Kyma

This documentation describes the structure, configuration, and automation of the VitePress-based documentation for the Kyma project. All examples and folder references relate to the Kyma documentation setup.

## Prerequisites

`Node.js` is required to run and build the site.

## Project Structure

### `.vitepress/` Directory

Contains all VitePress configuration files for the Kyma project:

- **config.mjs**: Main configuration file, which includes the following elements:
  - `getSearchConfig` function that configures the search component
  - `defineConfig` with site metadata (title: "Kyma", description, base path "/Kyma/")
  - Theme configuration including logo, navigation bar, sidebar structure, and social links
  - Sidebar imports from external modules
  - `makeSidebarAbsolutePath` function to transform relative sidebar paths to absolute paths
  - Markdown configuration with custom tabs plugin
- **theme/Features-add.vue**: Custom component that renders additional features section on the homepage.
- **theme/style.css**: Defines custom CSS styling that overrides VitePress default theme variables. Includes:
  - SAP 72 font family integration with multiple weights (Regular, Bold, Light)
  - Kyma brand colors (primary: #0b74de, primary-dark: #0a549d)
  - Custom color variables for brand, tip, warning, and danger elements
  - Button styling with Kyma brand colors
  - Home page hero styling with gradients and background effects
  - Features grid layout with hover effects and animations
  - Custom footer styling
  - Search component (Algolia) color customization
- **theme/index.js**: Used to register additional components, other than the ones provided by the VitePress framework. It includes:
  - Import and extension of VitePress DefaultTheme
  - Custom `Layout.vue` component integration
  - `Style.css` import for custom styling
  - Global component registration for Tabs
- **theme/MyFooter.vue**: Custom footer component that appears at the bottom of every page. Contains:
  - Copyright notice for Kyma project authors (2018-2025)
  - Legal links to SAP Terms of Use, Privacy Statement, and Legal Disclosure
  - Scoped CSS styling with centered alignment, padding, and border styling
  - Integration with VitePress CSS variables for consistent theming
- **theme/Layout.vue**: Custom layout component that extends the default VitePress layout. It includes:
  - Integration of custom footer (`MyFooter`) at the bottom of every page
  - Additional features section (`FeaturesAdd`) after home page features
  - Client-side routing compatibility for handling legacy URL redirects
- **components/Tabs.vue**: Advanced tabs component with dual functionality modes.
- **components/Tab.vue**: Simple tab content wrapper component that works with `Tabs.vue`.
- **plugins/tabs-plugin.js**: Markdown plugin that processes tab syntax in markdown files. Features:
  - Regex-based parsing of `<!-- tabs:start -->` and `<!-- tabs:end -->` blocks
  - Automatic detection of tab headers using `#### **Title**` format
  - Content extraction and markdown rendering for each tab
  - Base64 encoding of tab data for component consumption
  - Integration with custom Tabs Vue component via `tabs-data` prop
  - Extends markdown renderer while preserving original functionality

### Documentation Content

- **index.md**: Landing page configuration using VitePress home layout. Contains:
  - Hero section with Kyma branding, tagline with phonetic pronunciation, and call-to-action buttons
  - Main features grid showcasing six key value propositions
  - Custom `features_additonal` section displaying company adopters/users with logos
  - Custom CSS classes for styling (`home-kyma-heavy-bold`, `home-kyma-phonetic`)
  - Asset references for icons and logos with specified dimensions
- **external-content/**: Aggregates documentation from different repositories.
- **_sidebar.ts**: Sidebar configuration files imported from the individual modules repositories. Each module has its own sidebar configuration that gets transformed by the `makeSidebarAbsolutePath` function in `config.mjs` to work with the absolute path structure

### Automation and Workflows

- **workflows/deploy.yml**: GitHub Actions workflow that runs every day at midnight to update the documentation. It can also be triggered manually. The CronJob executes the following actions:
  - Copies content from the documentation repository into the `external-content` folder
  - Builds and deploys documentation artifacts
  - Sets up Node.js and installs dependencies
- **copy_external_content.sh**: Bash script for local development that automates external content aggregation. Process includes:
  - Clones specific Kyma project repositories
  - Copies `docs/user` and `docs/assets` folders from each repository to `../docs/external-content/[repo]/docs/`
  - Provides error handling for missing directories
  - Automatically cleans up cloned repositories after copying
  - Target directory structure: `../docs/external-content/[repo-name]/docs/user` and `../docs/external-content/[repo-name]/docs/assets`
- **move-to-public.sh**: Asset management script that processes files for public distribution. Features:
  - Scans `./docs/` directory for non-source files using extensive exclusion filters
  - Excludes common development files (`*.md`, `*.ts`, `*.js`, `*.vue`, `*.css`, config files, etc.)
  - Skips VitePress internal directories (`.vitepress`, `Node_modules`, `.git`, `.github`)
  - Copies qualifying files to `./docs/public/` maintaining directory structure
  - Specifically handles the entire `/assets/` folder copying to the public directory
  - Provides console output showing file movement operations
  - Creates necessary target directories automatically

### Configuration and Assets

- **package.json**: Contains scripts for development, build, and preview
- **public/**: Should contain references to assets that will be propagated on the website (managed by `move-to-public.sh`)

## Building and Previewing Documentation

To verify the build process:

```bash
npm run docs:build
npm run docs:preview
```

## Summary

The VitePress project for Kyma is organized for easy expansion, automated updates, and customization of both appearance and functionality. Automated workflows and dedicated scripts allow for fast deployment of changes both locally and in production. This setup is tailored specifically for the Kyma documentation ecosystem.
