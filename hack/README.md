# Hack Directory

## Overview

This package contains various scripts that are used by Kyma developers.

### Purpose

This directory contains tools, such as linters, that help to maintain the source code compliant with Go best coding practices. It also includes utility scripts that generate code and scripts executed on CI pipelines.

### How to Run it Locally

#### Prerequisite

Import the external (Modules') documentation by executing the script `copy_external_content.sh` from the folder `hack`.

This process will copy all the `docs/user` folder and `docs/assets` folder from the repositories specified in the sh file. Everything will be copied to the folder `external-content`; all the existing files will be overritten.

#### Run the Server in Development Mode

Execute the following commands:

```
npm install
npm run docs:dev
```

#### Run the Server in Production-Like Mode

Execute the following commands:

```
npm run docs:build
npm run docs:preview
```

**Note:** The `npm run docs:build` will copy all the unreferenced assets and non-grphical files (like scripts, documents, etc.) to the `docs/public` [directory](https://vitepress.dev/guide/asset-handling#the-public-directory) to grent to get them included in the `dist` folder of the project.
