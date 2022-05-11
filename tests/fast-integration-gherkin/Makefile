.PHONY: ci-certificate-gherkin
ci-certificate-gherkin:
	cd .. && npm install
	../node_modules/.bin/cucumber-js ./features/Ingress-certificate/certificate.feature

.PHONY: ci-skr-gherkin
ci-skr-gherkin:
	cd .. && npm install
	../node_modules/.bin/cucumber-js ./features/skr-test/skr.feature