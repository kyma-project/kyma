package director

import "fmt"

type queryProvider struct{}

func (qp queryProvider) applicationsForRuntimeQuery(runtimeID string) string {
	return fmt.Sprintf(`query {
	result: applicationsForRuntime(runtimeID: "%s") {
		%s
	}
}`, runtimeID, applicationsQueryData(runtimeID))
}

func (qp queryProvider) setRuntimeLabelMutation(runtimeId, key, value string) string {
	return fmt.Sprintf(`mutation {
		result: setRuntimeLabel(runtimeID: "%s", key: "%s", value: "%s") {
			%s
		}
	}`, runtimeId, key, value, labelData())
}

func labelData() string {
	return `key
			value`
}

func applicationsQueryData(runtimeID string) string {
	return pageData(applicationData(runtimeID))
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

func applicationData(runtimeID string) string {
	return fmt.Sprintf(`id
		name
		providerName
		description
		labels
		apiDefinitions {%s}
		eventDefinitions {%s}
		documents {%s}
		auths {%s}
		packages {%s}
	`, pageData(apiDefinitionData(runtimeID)), pageData(eventAPIData()), pageData(documentData()), systemAuthData(), pageData(packagesData(runtimeID)))
}

func systemAuthData() string {
	return fmt.Sprintf(`id`)
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

func apiDefinitionData(runtimeID string) string {
	return fmt.Sprintf(`		id
		name
		description
		spec {%s}
		targetURL
		group
		auth(runtimeID: "%s") {%s}
		defaultAuth {%s}
		version {%s}`, apiSpecData(), runtimeID, runtimeAuthData(), authData(), versionData())
}

func packagesData(runtimeID string) string {
	return fmt.Sprintf(`id
		name
		description
		instanceAuthRequestInputSchema
		defaultInstanceAuth {%s}
		apiDefinitions {%s}
		eventDefinitions {%s}
		documents {%s}
		`, authData(), pageData(packageApiDefinitions()), pageData(eventAPIData()), pageData(documentData()))
}

func packageApiDefinitions() string {
	return fmt.Sprintf(`		id
		name
		description
		spec {%s}
		targetURL
		group
		version {%s}`, apiSpecData(), versionData())
}

func apiSpecData() string {
	return fmt.Sprintf(`data
		format
		type`)
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
	return fmt.Sprintf(`data
		type
		format`)
}

func documentData() string {
	return fmt.Sprintf(`
		id
		applicationID
		title
		displayName
		description
		format
		kind
		data`)
}
