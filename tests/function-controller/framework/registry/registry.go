/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package registry contains utilities to work with container registries.
package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/replicaset"
)

const (
	manifestsPath = "framework/registry/manifests"

	replicaSetName = "registry"
	serviceName    = "builds"
	statusEndpoint = "/v2/"
	apiHeader      = "Docker-Distribution-Api-Version"
)

// DeployLocal deploys a local container registry inside the Framework
// namespace and waits until it becomes available. See also
// https://github.com/triggermesh/knative-local-registry
func DeployLocal(f *framework.Framework) (url string, cleanup func()) {
	ns := f.Namespace.Name

	if _, err := f.CreateFromManifests(nil, readManifests(f)...); err != nil {
		framework.Failf("Error creating registry objects from manifests: %v", err)
	}

	if err := replicaset.WaitForReadyReplicaSet(f.ClientSet, ns, replicaSetName); err != nil {
		framework.Failf("Error waiting for readiness of registry Pods: %v", err)
	}

	hostname := fmt.Sprintf("builds.%s.svc.cluster.local", ns)

	return hostname, patchHostEtcHosts(f, hostname)
}

func readManifests(f *framework.Framework) (manifestsPaths []string) {
	var walkFn filepath.WalkFunc = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "accessing path %q", path)
		}

		if filepath.Dir(path) == manifestsPath && // do not descend into sub-directories
			filepath.Ext(path) == ".yaml" {

			manifestsPaths = append(manifestsPaths, path)
		}

		return nil
	}

	if err := filepath.Walk(manifestsPath, walkFn); err != nil {
		framework.Failf("Error reading registry manifests from filesystem: %v", err)
	}

	return manifestsPaths
}
