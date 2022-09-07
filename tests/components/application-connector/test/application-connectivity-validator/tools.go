package application_connectivity_validator

func validatorURL(app, path string) string {
	return "http://central-application-connectivity-validator.kyma-system:8080/" + app + "/" + path
}
