package model

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
