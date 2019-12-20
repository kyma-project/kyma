package compass

import (
	"fmt"
)

type queryProvider struct{}

func (qp queryProvider) updateLabelDefinition(input string) string {
	return fmt.Sprintf(`mutation {
  result: updateLabelDefinition( in: %s ) {
    key
    schema
  }
}
`, input)
}

func (qp queryProvider) labelDefinition(key string) string {
	return fmt.Sprintf(`query {
	result: labelDefinition(key: "%s") {
		key
    	schema
	}
}`, key)
}

func (qp queryProvider) setRuntimeLabel(runtimeId, key string, value []string) string {
	return fmt.Sprintf(`mutation {
	result: setRuntimeLabel(runtimeID: "%s", key: "%s", value: %s) {
		key
		value
	}
}`, runtimeId, key, value)
}

func (qp queryProvider) requestOneTimeTokenForApplication(appID string) string {
	return fmt.Sprintf(`mutation {
	result: requestOneTimeTokenForApplication(id: "%s") {
		token
		connectorURL
	}
}`, appID)
}

// TODO: createApplication does not support paging
// it will return only API, EventAPI and Document ids from page one
func (qp queryProvider) createApplication(input string) string {
	return fmt.Sprintf(`mutation {
	result: registerApplication(in: %s) {
		%s
	}
}`, input, applicationData())
}

func (qp queryProvider) updateApplication(applicationId, input string) string {
	return fmt.Sprintf(`mutation {
	result: updateApplication(id: "%s" in: %s) {
		%s
	}
}`, applicationId, input, applicationData())
}

func (qp queryProvider) deleteApplication(id string) string {
	return fmt.Sprintf(`mutation {
	result: unregisterApplication(id: "%s") {
		id
	}
}`, id)
}

func (qp queryProvider) getRuntime(runtimeId string) string {
	return fmt.Sprintf(`query {
	result: runtime(id: "%s") {
		%s
	}
}`, runtimeId, runtimeData())
}

func (qp queryProvider) createAPI(applicationId string, input string) string {
	return fmt.Sprintf(`mutation {
	result: addAPIDefinition(applicationID: "%s", in: %s) {
		%s
	}
}`, applicationId, input, apiDefinitionData())
}

func (qp queryProvider) updateAPI(apiId string, input string) string {
	return fmt.Sprintf(`mutation {
	result: updateAPIDefinition(id: "%s", in: %s) {
		%s
	}
}`, apiId, input, apiDefinitionData())
}

func (qp queryProvider) deleteAPI(apiId string) string {
	return fmt.Sprintf(`mutation {
	result: deleteAPIDefinition(id: "%s") {
		id
	}
}`, apiId)
}

func (qp queryProvider) createEventAPI(applicationId string, input string) string {
	return fmt.Sprintf(`mutation {
	result: addEventDefinition(applicationID: "%s", in: %s) {
		%s
	}
}`, applicationId, input, eventAPIData())
}

func (qp queryProvider) updateEventAPI(apiId string, input string) string {
	return fmt.Sprintf(`mutation {
	result: updateEventDefinition(id: "%s", in: %s) {
		%s
	}
}`, apiId, input, eventAPIData())
}

func (qp queryProvider) deleteEventAPI(apiId string) string {
	return fmt.Sprintf(`mutation {
	result: deleteEventDefinition(id: "%s") {
		id
	}
}`, apiId)
}

func pageData(item string) string {
	return fmt.Sprintf(`data {
		%s
	}
	pageInfo {%s}
	totalCount
	`, item, pageInfoData())
}

func pageInfoData() string {
	return `startCursor
		endCursor
		hasNextPage`
}

func applicationData() string {
	return fmt.Sprintf(`id
		name
		description
		labels
		apiDefinitions {%s}
		eventDefinitions {%s}
		documents {%s}
	`, pageData(apiDefinitionData()), pageData(eventAPIData()), pageData(documentData()))
}

func runtimeData() string {
	return fmt.Sprintf(`
		id
		name
		description
		labels 
		status {condition timestamp}
		auths {%s}`, systemAuthData())
}

func systemAuthData() string {
	return fmt.Sprintf(`
		id
		auth {%s}`, authData())
}

func authData() string {
	return fmt.Sprintf(`credential {
				... on BasicCredentialData {
					username
					password
				}
				...  on OAuthCredentialData {
					clientId
					clientSecret
					url
					
				}
			}
			additionalHeaders
			additionalQueryParams
			requestAuth { 
			  csrf {
				tokenEndpointURL
				credential {
				  ... on BasicCredentialData {
				  	username
					password
				  }
				  ...  on OAuthCredentialData {
					clientId
					clientSecret
					url
					
				  }
			    }
				additionalHeaders
				additionalQueryParams
			}
			}
		`)
}

func apiDefinitionData() string {
	return fmt.Sprintf(`		id
		name
		description
		applicationID
		spec {%s}
		targetURL
		group
		auths {%s}
		defaultAuth {%s}
		version {%s}`, apiSpecData(), runtimeAuthData(), authData(), versionData())
}

func apiSpecData() string {
	return `data
		format
		type`
}

func runtimeAuthData() string {
	return fmt.Sprintf(`runtimeID
		auth {%s}`, authData())
}

func versionData() string {
	return `value
		deprecated
		deprecatedSince
		forRemoval`
}

func eventAPIData() string {
	return fmt.Sprintf(`
			id
			applicationID
			name
			description
			group 
			spec {%s}
			version {%s}
		`, eventSpecData(), versionData())
}

func eventSpecData() string {
	return `data
		type
		format`
}

func documentData() string {
	return `
		id
		applicationID
		title
		displayName
		description
		format
		kind
		data`
}
