.PHONY: ci
ci:
	npm install
	npm test
	npm run test-monitoring

.PHONY: ci-compass
ci-compass:
	npm install
	npm run test-compass

.PHONY: ci-skr
ci-skr:
	npm install
	npm run test-skr

.PHONY: ci-skr-svcat-migration
ci-skr-svcat-migration:
	npm install
	npm run test-skr-svcat-migration

.PHONY: ci-kyma-preview
ci-kyma-preview:
	npm install
	npm run test-kyma-preview

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
