# Ark plugins

## Overview

Ark plugins provide an implementation of a few plugins required to properly restore the Kyma cluster and resources created inside it. They focus mainly on resources related to the Service Catalog, such as instances or bindings. Each plugin is defined in a separate file inside the [`internal/plugins`](internal/plugins) folder. The purpose of each plugin is defined in the file's comments.

The structure of folders and files is based on the [Ark plugin example repository](https://github.com/heptio/ark-plugin-example).

## Installation

Run a Docker image build from the [Dockerfile](Dockerfile) as an init container inside the deployment of the [Ark chart](../../resources/ark). During development, you can push a built image to your own image repository and run:

```bash
ark plugin add {yourRepo/imageName:tag}
```

## Create a new plugin

- New plugins must be added under [`internal/plugins`](internal/plugins) :

```             
  ├── internal                                                                  
    ├── plugins
      ├── backup    # new Backup Item Action plugins 
      ├── restore   # new Restore Item Action plugins 
  ```

- All new plugins must be registered in the `tools/ark-plugins/main.go` file.
- **Backup Item Action** - performs arbitrary logic on individual items prior to storing them in the backup file.
- **Restore Item Action** - performs arbitrary logic on individual items prior to restoring them in the Kyma cluster.
- All plugins must be implemented as the following example:

```go
package restore

import (
...
)

// FunctionPluginRestore is a plugin for ark to ...
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
