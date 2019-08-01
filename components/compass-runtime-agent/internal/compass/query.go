package compass

import "fmt"

func ApplicationsForRuntimeQuery(runtimeID string) string {
	return fmt.Sprintf(`query {
	result: applicationsForRuntime(runtimeID: %s) {
		%s
	}
}`, runtimeID, applicationsQueryData())
}

func applicationsQueryData() string {
	return pageData(applicationData())
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
		spec {%s}
		targetURL
		group
		auth {%s}
		defaultAuth {%s}
		version {%s}`, apiSpecData(), runtimeAuthData(), authData(), versionData())
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
