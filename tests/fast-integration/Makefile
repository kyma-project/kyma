.PHONY: ci
ci:
	npm install
	npm test
	npm run test-monitoring

.PHONY: ci-no-install
ci-no-install:
	npm install
	npm run test-no-install
	npm run test-monitoring

.PHONY: ci-compass
ci-compass:
	npm install
	npm run test-compass

.PHONY: ci-skr
ci-skr:
	npm install
	npm run test-skr

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
