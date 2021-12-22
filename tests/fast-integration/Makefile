.PHONY: ci
ci:
	npm install
	npm test

.PHONY: ci-compass
ci-compass:
	npm install
	npm run test-compass

.PHONY: ci-skr
ci-skr:
	npm install
	npm run test-skr

.PHONY: ci-skr-nightly
ci-skr-nightly:
	npm install
	npm run nightly-skr

.PHONY: ci-test-skr-nightly
ci-test-skr-nightly:
	npm install
	npm run test-nightly-skr

.PHONY: ci-skr-kyma-to-kyma2-upgrade
ci-skr-kyma-to-kyma2-upgrade:
	npm install
	npm run test-skr-kyma-to-kyma2-upgrade

.PHONY: ci-skr-svcat-migration
ci-skr-svcat-migration:
	npm install
	npm run test-skr-svcat-migration

.PHONY: ci-pre-upgrade
ci-pre-upgrade:
	npm install
	npm run upgrade-test-prep
	npm run upgrade-test-tests

.PHONY: ci-post-upgrade
ci-post-upgrade:
	npm install
	npm run upgrade-test-tests
	npm run upgrade-test-cleanup

.PHONY: ci-application-connectivity-2
ci-application-connectivity-2:
	npm install
	npm run test-application-connectivity-2

.PHONY: ci-application-connectivity-2-compass
ci-application-connectivity-2-compass:
	npm install
	npm run test-application-connectivity-2-compass

.PHONY: ci-test-eventing
ci-test-eventing:
	npm install
	npm run eventing-upgrade-test-prep
	npm run eventing-upgrade-test-tests
	npm run eventing-upgrade-test-cleanup

.PHONY: ci-test-eventing-pre-upgrade
ci-test-eventing-pre-upgrade:
	npm install
	npm run eventing-upgrade-test-prep
	npm run eventing-upgrade-test-tests

.PHONY: ci-test-eventing-pre-upgrade
ci-test-eventing-post-upgrade:
	npm install
	npm run eventing-upgrade-test-tests
	npm run eventing-upgrade-test-cleanup

.PHONY: ci-test-eventing-tests
ci-test-eventing-post-tests:
	npm install
	npm run eventing-upgrade-test-tests
