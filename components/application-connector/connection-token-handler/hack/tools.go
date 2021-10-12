//go:build tools
// +build tools

// This package imports things required by build scripts, to force `go mod` to see them as dependencies
package hack

import _ "k8s.io/code-generator"
