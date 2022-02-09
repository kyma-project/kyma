package model

func APIBundleExists(id string, application Application) (APIBundle, bool) {
	for _, apiBundle := range application.ApiBundles {
		if apiBundle.ID == id {
			return apiBundle, true
		}
	}

	return APIBundle{}, false
}

func BundleContainsAnySpecs(p APIBundle) bool {
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
