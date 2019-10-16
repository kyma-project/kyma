# Backup Plugins

## Overview

Backup plugins provide the functionality necessary to properly restore the Kyma cluster and its resources using Velero. They focus mainly on resources related to the Service Catalog, such as instances or bindings. Each plugin is defined in a separate file inside the [`internal/plugins`](internal/plugins) folder.

The structure of folders and files is based on the [Velero plugin example repository](https://github.com/heptio/velero-plugin-example).

## Installation

For Velero to use plugins, run a Docker image built from the [Dockerfile](Dockerfile) as an init container inside the [Velero chart](../../resources/velero) Deployment.

During development, you can push the image to your own image repository and add the plugin by running:

```bash
velero plugin add {yourRepo/imageName:tag}
```

## Development

### Create a new plugin

To create a new plugin for Velero to use, perform the following steps:

1. Go to the [`internal/plugins`](internal/plugins) folder:

This folder includes the backup and restore subfolders, where you can define plugins based on the following object types:

- **Backup Item Action** - performs arbitrary logic on individual items before storing them in the backup file.
- **Restore Item Action** - performs arbitrary logic on individual items before restoring them in the Kyma cluster.

```bash
  ├── internal
    ├── plugins
      ├── backup    # new Backup Item Action plugins
      ├── restore   # new Restore Item Action plugins
  ```

2. Implement plugins as in the following example:

```go
package restore

import (
...
)

// FunctionPluginRestore is a plugin for velero to ...
type FunctionPluginRestore struct {
  Log logrus.FieldLogger
}

// AppliesTo return list of resource kinds which should be handled by this plugin
func (p *FunctionPluginRestore) AppliesTo() (restore.ResourceSelector, error) {
  return restore.ResourceSelector{...}, nil
}

// Execute does ... on the item being restored.
// nolint
func (p *FunctionPluginRestore) Execute(item runtime.Unstructured, restore *v1.Restore) (runtime.Unstructured, error, error) {
    ...
  return item, nil, nil
}

```

3. Register created plugins in the `components/backup-plugins/main.go` file.
