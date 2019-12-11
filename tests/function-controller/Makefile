# Copyright 2019 The Kyma Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

APP_NAME = function-controller-tests
APP_PATH = tests/function-controller
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.12
SCRIPTS_DIR = ../../common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

# generate goal that runs 'all-local' inside a container
$(eval $(call buildpack-mount,all))

verify:: all


# This patch adds support for few extra types to the 'CreateFromManifests' util function.
# https://stackoverflow.com/a/13517811/4716370
vendor/k8s.io/kubernetes/test/e2e/framework/create.go.patched: hack/k8s_framework_create.patch vendor
	@echo '+ Patching vendored kubernetes/test/e2e/framework'
	if [ -e $@ ]; then patch -R -p0 < $@; fi
	patch -p0 < $<
	cp $< $@

vendor: Gopkg.lock
	@echo '+ Pulling vendored dependencies'
	@dep ensure -v --vendor-only


.PHONY: test-local

# The 0 timeout prevents the test from panicking after 10m, leaving stale resources behind.
test-local: vendor/k8s.io/kubernetes/test/e2e/framework/create.go.patched
	@echo '+ Running e2e tests'
ifdef TEST_ARGS
	@go test -timeout=0 -v ./ -args $(TEST_ARGS)
else
	@go test -timeout=0 -v ./
endif
