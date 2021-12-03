.PHONY: ci
ci:
	npm ci --loglevel verbose
	npm test
	npm run test-monitoring

.PHONY: ci-compass
ci-compass:
	npm ci --loglevel verbose
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
	npm ci --loglevel verbose
	npm run test-application-connectivity-2

.PHONY: ci-application-connectivity-2-compass
ci-application-connectivity-2-compass:
	npm ci --loglevel verbose
	npm run test-application-connectivity-2-compass

.PHONY: ci-test-eventing
ci-test-eventing:
	npm ci
	npm run test-eventing
