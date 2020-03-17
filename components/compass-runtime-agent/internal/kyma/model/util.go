package model

func APIExists(id string, application Application) bool {

	for _, apiDefinition := range application.APIs {
		if apiDefinition.ID == id {
			return true
		}
	}

	for _, eventAPIDefinition := range application.EventAPIs {
		if eventAPIDefinition.ID == id {
			return true
		}
	}

	return false
}

func APIPackageExists(id string, application Application) (APIPackage, bool) {

	for _, apiPackage := range application.APIPackages {
		if apiPackage.ID == id {
			return apiPackage, true
		}
	}

	return APIPackage{}, false
}

func PackageContainsAnySpecs(p APIPackage) bool {

	for _, apiDefinition := range p.APIDefinitions {
		if apiDefinition.APISpec != nil {
			return true
		}
	}

	for _, eventApiDefinition := range p.EventDefinitions {
		if eventApiDefinition.EventAPISpec != nil {
			return true
		}
	}

	return false
}
