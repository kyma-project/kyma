.PHONY: ci-certificate-gherkin
ci-certificate-gherkin:
	cd ../fast-integration && npm install
	cd ../ fast-integration-gherkin && npm install
	./node_modules/.bin/cucumber-js features/Ingress-certificate/certificate.feature