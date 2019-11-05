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

package registry

import (
	"fmt"

	"github.com/onsi/ginkgo"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/log"
)

const (
	etchostsManifest  = "framework/registry/manifests/sysadmin/nodes-etc-hosts-update.yaml"
	etchostsMountPath = "/host-etc/hosts"

	etchostsDaemonSetName = "registry-etc-hosts-update"
)

// patchHostEtcHosts adds an entry to the /etc/hosts file on all schedulable
// Kubernetes nodes, and returns a function that can undo that action.
// We do this because the DNS record of the Service is not resolvable by the
// container runtime in most environments, yet we *must* use a domain with the
// .local extension in order to circumvent the container runtime's requirement
// to communicate securely (TLS) with its registry.
func patchHostEtcHosts(f *framework.Framework, host string) (undo func()) {
	ginkgo.By("adding custom /etc/hosts entries")

	rmPatchEtcHostsObjects := createWithPatchAndWait(f, setEtcHostsEntry(host))
	log.Logf("Custom /etc/hosts entry added successfully (%s)", host)

	undo = func() {
		rmPatchEtcHostsObjects()

		ginkgo.By("removing custom /etc/hosts entries")

		// recreate the DaemonSet with a remove command
		createWithPatchAndWait(f, rmEtcHostsEntry(host))
		log.Logf("Custom /etc/hosts entry removed successfully (%s)", host)
	}

	return undo
}

type patchFunction = func(interface{}) error

// setEtcHostsEntry defines the hostname to be added/removed to/from the
// /etc/hosts file.
func setEtcHostsEntry(host string) patchFunction {
	var p patchFunction = func(item interface{}) error {
		ds := item.(*appsv1.DaemonSet)

		// location of the REGISTRY_SERVICE_HOSTS env var
		hostsEnvVal := &(ds.Spec.Template.Spec.InitContainers[0].Env[0].Value)
		*hostsEnvVal = host

		return nil
	}

	return p
}

// rmEtcHostsEntry defines a script that removes the given host from the
// /etc/hosts file.
func rmEtcHostsEntry(host string) patchFunction {
	var p patchFunction = func(item interface{}) error {
		setEtcHostsEntry(host)(item)

		ds := item.(*appsv1.DaemonSet)

		// location of the bash script to execute
		bashScript := &(ds.Spec.Template.Spec.InitContainers[0].Command[2])
		*bashScript = makeRmEtcHostsEntryScript(host)

		return nil
	}

	return p
}

func createWithPatchAndWait(f *framework.Framework, p patchFunction) (cleanupFn func()) {
	var retryCreateFn wait.ConditionFunc = func() (done bool, err error) {
		cleanupFn, err = f.CreateFromManifests(p, etchostsManifest)
		switch {
		case apierrors.IsAlreadyExists(err):
			return false, nil
		case err != nil:
			return false, err
		}
		return true, nil
	}

	if err := wait.PollImmediate(framework.Poll, framework.PollShortTimeout, retryCreateFn); err != nil {
		framework.Failf("Failed to patch /etc/hosts: %v", err)
	}

	ns := f.Namespace.Name

	framework.ExpectNoErrorWithRetries(func() error {
		ds, err := f.ClientSet.AppsV1().DaemonSets(ns).Get(etchostsDaemonSetName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		// mitigate race where WaitForDaemonSets succeeds with '0/0 pods ready'
		if ds.Status.CurrentNumberScheduled == 0 {
			return fmt.Errorf("no Pod scheduled, delaying")
		}

		return err

	}, 50, "DaemonSet should exist")

	if err := framework.WaitForDaemonSets(f.ClientSet, ns, 0, framework.PollShortTimeout); err != nil {
		framework.Failf("Error waiting for readiness of /etc/hosts patch Pods: %v", err)
	}

	return cleanupFn
}

func makeRmEtcHostsEntryScript(pattern string) (script string) {
	// we avoid using `sed -i` because it modifies the file's inode
	script = "HOSTS=\"$(sed '/%s/d' %s)\"\n" +
		"echo \"$HOSTS\" > %s\n"

	return fmt.Sprintf(script,
		pattern, etchostsMountPath,
		etchostsMountPath)
}
