.PHONY: ci-certificate-gherkin
ci-certificate-gherkin:
	cd .. && npm install
	../node_modules/.bin/cucumber-js ./certificate/features/certificate.feature -f json:test-report.json

.PHONY: ci-skr-gherkin
ci-skr-gherkin:
	cd .. && npm install
	../node_modules/.bin/cucumber-js ./skr/features/skr.feature -f json:test-report.json