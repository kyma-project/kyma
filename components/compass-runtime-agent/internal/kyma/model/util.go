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
