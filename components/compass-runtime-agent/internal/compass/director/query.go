package director

import "fmt"

type queryProvider struct{}

func (qp queryProvider) applicationsAndLabelsForRuntimeQuery(runtimeID string) string {
	return fmt.Sprintf(`query {
		runtime(id: "%s") {
			%s
		}
		applicationsForRuntime(runtimeID: "%s") {
			%s
		}
	}`, runtimeID, labels(), runtimeID, applicationsQueryData())
}

func (qp queryProvider) setRuntimeLabelMutation(runtimeId, key, value string) string {
	return fmt.Sprintf(`mutation {
		setRuntimeLabel(runtimeID: "%s", key: "%s", value: "%s") {
			%s
		}
	}`, runtimeId, key, value, labelData())
}

func labels() string {
	return `labels`
}

func applicationsQueryData() string {
	return pageData(applicationData())
}

func labelData() string {
	return `key
			value`
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
		providerName
		description
		labels
		auths {%s}
		bundles {%s}
	`, systemAuthData(), pageData(bundlesData()))
}

func systemAuthData() string {
	return "id"
}

func bundlesData() string {
	return fmt.Sprintf(`id
		name
		description
		instanceAuthRequestInputSchema
		apiDefinitions {%s}
		eventDefinitions {%s}
		defaultInstanceAuth {%s}
		`, pageData(bundleApiDefinitions()), pageData(eventAPIData()), authData())
}

func bundleApiDefinitions() string {
	return fmt.Sprintf(`		id
		name
		description
		targetURL
		group
		version {%s}`, versionData())
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
			name
			description
			group
			version {%s}
		`, versionData())
}

func authData() string {
	return fmt.Sprintf(`
		credential {%s}
		additionalHeaders
		additionalQueryParams
		requestAuth {%s}
		`, credentialData(), requestAuthData())
}

func credentialData() string {
	return fmt.Sprintf(`
		... on BasicCredentialData {%s}
		... on OAuthCredentialData {%s}
	`, basicCredentialData(), oauthCredentialData())
}

func basicCredentialData() string {
	return `
		username
		password
	`
}

func oauthCredentialData() string {
	return `
		clientId
		clientSecret
		url
	`
}

func requestAuthData() string {
	return fmt.Sprintf(`
		csrf {%s}
		`, csrfData())
}

func csrfData() string {
	return `
		tokenEndpointURL
		`
}

func (qp queryProvider) getRuntimeQuery(runtimeID string) string {
	return fmt.Sprintf(`query {
    result: runtime(id: "%s") {
         id name description labels
}}`, runtimeID)
}

func (qp queryProvider) updateRuntimeMutation(runtimeID, runtimeInput string) string {
	return fmt.Sprintf(`mutation {
    result: updateRuntime(id: "%s" in: %s) {
		id
}}`, runtimeID, runtimeInput)
}
