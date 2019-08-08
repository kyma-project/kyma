package compass

import (
	"fmt"
)

type queryProvider struct{}

// TODO: createApplication does not support paging
// it will return only API, EventAPI and Document ids from page one
func (qp queryProvider) createApplication(input string) string {
	return fmt.Sprintf(`mutation {
	result: createApplication(in: %s) {
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
	result: deleteApplication(id: "%s") {
		id
	}
}`, id)
}

func (qp queryProvider) createAPI(applicationId string, input string) string {
	return fmt.Sprintf(`mutation {
	result: addAPI(applicationID: "%s", in: %s) {
		%s
	}
}`, applicationId, input, apiDefinitionData())
}

func (qp queryProvider) updateAPI(apiId string, input string) string {
	return fmt.Sprintf(`mutation {
	result: updateAPI(id: "%s", in: %s) {
		%s
	}
}`, apiId, input, apiDefinitionData())
}

func (qp queryProvider) deleteAPI(apiId string) string {
	return fmt.Sprintf(`mutation {
	result: deleteAPI(id: "%s") {
		id
	}
}`, apiId)
}

func (qp queryProvider) createEventAPI(applicationId string, input string) string {
	return fmt.Sprintf(`mutation {
	result: addEventAPI(applicationID: "%s", in: %s) {
		%s
	}
}`, applicationId, input, apiDefinitionData())
}

func (qp queryProvider) updateEventAPI(apiId string, input string) string {
	return fmt.Sprintf(`mutation {
	result: updateEventAPI(id: "%s", in: %s) {
		%s
	}
}`, apiId, input, apiDefinitionData())
}

func (qp queryProvider) deleteEventAPI(apiId string) string {
	return fmt.Sprintf(`mutation {
	result: deleteEventAPI(id: "%s") {
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
		apis {%s}
		eventAPIs {%s}
		documents {%s}
	`, pageData(apiDefinitionData()), pageData(eventAPIData()), pageData(documentData()))
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
