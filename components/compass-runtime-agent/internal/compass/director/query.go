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
	return fmt.Sprintf(`id`)
}

func bundlesData() string {
	return fmt.Sprintf(`id
		name
		description
		instanceAuthRequestInputSchema
		apiDefinitions {%s}
		eventDefinitions {%s}
		documents {%s}
		defaultInstanceAuth {%s}
		`, pageData(bundleApiDefinitions()), pageData(eventAPIData()), pageData(documentData()), authData())
}

func bundleApiDefinitions() string {
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
		title
		displayName
		description
		format
		kind
		data`)
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
	return fmt.Sprintf(`
		username
		password
	`)
}

func oauthCredentialData() string {
	return fmt.Sprintf(`
		clientId
		clientSecret
		url
	`)
}

func requestAuthData() string {
	return fmt.Sprintf(`
		csrf {%s}
		`, csrfData())
}

func csrfData() string {
	return fmt.Sprintf(`
		tokenEndpointURL
		`)
}
