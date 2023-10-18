.PHONY: lint
lint:
	npm install
	npm run lint

.PHONY: lint-fix
lint-fix:
	npm install
	npm run lint-fix

.PHONY: ci
ci:
	npm install
	npm test

.PHONY: ci-compass
ci-compass:
	npm install
	npm run test-compass

.PHONY: ci-pre-upgrade
ci-pre-upgrade:
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default_serverless_cr.yaml -n kyma-system
	npm install
	kubectl wait --for condition=Installed -n kyma-system serverless default --timeout=120s
	npm run upgrade-test-prep
	npm run upgrade-test-tests

.PHONY: ci-post-upgrade
ci-post-upgrade:
	npm install
	npm run upgrade-test-tests
	npm run upgrade-test-cleanup

.PHONY: ci-application-connectivity-2-compass
ci-application-connectivity-2-compass:
	npm install
	npm run test-compass

.PHONY: ci-test-eventing
ci-test-eventing:
	npm install
	npm run eventing-test-prep
	npm run eventing-tests
	npm run eventing-test-cleanup

.PHONY: ci-test-eventing-pre-upgrade
ci-test-eventing-pre-upgrade:
	npm install
	npm run eventing-test-prep
	npm run eventing-tests

.PHONY: ci-test-eventing-post-upgrade
ci-test-eventing-post-upgrade:
	npm install
	npm run eventing-tests
	npm run eventing-test-cleanup

.PHONY: ci-test-eventing-tests
ci-test-eventing-tests:
	npm install
	npm run eventing-tests

.PHONY: ci-test-eventing-provision-skr
ci-test-eventing-provision-skr:
	npm ci
	npm run test-eventing-provision-skr

.PHONY: ci-test-eventing-deprovision-skr
ci-test-eventing-deprovision-skr:
	npm ci
	npm run test-eventing-deprovision-skr
