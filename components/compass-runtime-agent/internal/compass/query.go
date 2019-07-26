package compass

import "fmt"

// TODO: consider adding some tests and exposing as pkg (if it might be useful in tests)

func ApplicationsQuery() string {
	return fmt.Sprintf(`query {
	result: applications {
		%s
	}
}`, applicationsQueryData())
}

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
		status {condition timestamp}
		webhooks {%s}
		healthCheckURL
		apis {%s}
		eventAPIs {%s}
		documents {%s}
	`, webhookData(), pageData(apiDefinitionData()), pageData(eventAPIData()), pageData(documentData()))
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

func webhookData() string {
	return fmt.Sprintf(
		`id
		applicationID
		type
		url
		auth {
		  %s
		}`, authData())
}

func apiDefinitionData() string {
	return fmt.Sprintf(`		id
		name
		description
		spec {%s}
		targetURL
		group
		auths {%s}
		defaultAuth {%s}
		version {%s}`, apiSpecData(), runtimeAuthData(), authData(), versionData())
}

func apiSpecData() string {
	return fmt.Sprintf(`data
		format
		type
		fetchRequest {%s}`, fetchRequestData())
}

func fetchRequestData() string {
	return fmt.Sprintf(`url
		auth {%s}
		mode
		filter
		status {condition timestamp}`, authData())
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
		format
		fetchRequest {%s}`, fetchRequestData())
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
		data
		fetchRequest {%s}`, fetchRequestData())
}

func runtimeData() string {
	return fmt.Sprintf(`id
		name
		description
		labels 
		status {condition timestamp}
		agentAuth {%s}`, authData())
}
