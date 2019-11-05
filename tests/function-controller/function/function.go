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

package function

import (
	. "github.com/onsi/ginkgo"

	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/log"

	"github.com/kyma-project/kyma/tests/function-controller/framework/registry"
)

var _ = Describe("Functions", func() {
	var registryURL string

	f := framework.NewDefaultFramework("function")

	BeforeEach(func() {
		registryURL = registry.DeployLocal(f)
	})

	It("should print this", func() {
		log.Logf(registryURL)
	})
})
