.PHONY: ci
ci:
	npm install
	npm test

.PHONY: ci-no-install
ci-no-install:
	npm install
	npm run test-no-install

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
	npm run test-commerce-mock-prep
	npm run test-commerce-mock-tests

.PHONY: ci-post-upgrade
ci-post-upgrade:
	npm install
	npm run test-commerce-mock-tests
	npm run test-commerce-mock-cleanup