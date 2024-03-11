package director

import "fmt"

type queryProvider struct{}

func (qp queryProvider) createFormation(formationName string) string {
	return fmt.Sprintf(`mutation {
	result: createFormation(formation: {
		name: "%s"
	}) { id } }`, formationName)
}

func (qp queryProvider) deleteFormation(formationName string) string {
	return fmt.Sprintf(`mutation {
	result: deleteFormation(formation: {
		name: "%s"
	}) { id } }`, formationName)
}

func (qp queryProvider) registerApplicationFromTemplateMutation(appName, displayName string) string {
	return fmt.Sprintf(`mutation {
	result: registerApplicationFromTemplate(in: {
		templateName: "SAP Commerce Cloud"
		values: [
			{ placeholder: "name", value: "%s" }
			{ placeholder: "display-name", value: "%s" }
        ]
	}) { id } }`, appName, displayName)
}

func (qp queryProvider) addBundleMutation(appID string) string {
	return fmt.Sprintf(`mutation {
	result: addBundle(
		applicationID: "%s"
		in: {
			name: "bndl-app-1"
			description: "Foo bar"
			apiDefinitions: [
			  {
				name: "comments-v1"
				description: "api for adding comments"
				targetURL: "http://mywordpress.com/comments"
				group: "comments"
				spec: {
				  data: "{\"openapi\":\"3.0.2\"}"
				  type: OPEN_API
				  format: YAML
				}
				version: {
				  value: "1.0.0"
				  deprecated: true
				  deprecatedSince: "v5"
				  forRemoval: false
				}
			  }
			]
		}
	) { id } }`, appID)
}

func (qp queryProvider) updateApplicationMutation(id, description string) string {
	return fmt.Sprintf(`mutation {
	result: updateApplication(
		id: "%s"
		in: {description: "%s"
			}) { id } }`,
		id, description)
}

func (qp queryProvider) assignFormationForAppMutation(applicationId, formationName string) string {
	return fmt.Sprintf(`mutation {
	result: assignFormation(
		objectID: "%s"
		objectType: APPLICATION
		formation: { name: "%s" }
	) { id } }`, applicationId, formationName)
}

func (qp queryProvider) unassignFormation(applicationId, formationName string) string {
	return fmt.Sprintf(`mutation {
		result: unassignFormation(
		objectID: "%s"
		objectType: APPLICATION
		formation: { name: "%s" }
		) {
			name
		}
}`, applicationId, formationName)
}

func (qp queryProvider) assignFormationForRuntimeMutation(runtimeId, formationName string) string {
	return fmt.Sprintf(`mutation {
	result: assignFormation(
		objectID: "%s"
		objectType: RUNTIME
		formation: { name: "%s" }
	) { id } }`, runtimeId, formationName)
}

func (qp queryProvider) unregisterApplicationMutation(applicationID string) string {
	return fmt.Sprintf(`mutation {
	result: unregisterApplication(id: "%s") {
		id
	} }`, applicationID)
}

func (qp queryProvider) deleteRuntimeMutation(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: unregisterRuntime(id: "%s") {
		id
	}}`, runtimeID)
}

func (qp queryProvider) registerRuntimeMutation(runtimeName string) string {
	return fmt.Sprintf(`mutation {
	result: registerRuntime(in: {
		name: "%s"
	}) { id } }`, runtimeName)
}

func (qp queryProvider) requestOneTimeTokenMutation(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: requestOneTimeTokenForRuntime(id: "%s") {
		token connectorURL
	}}`, runtimeID)
}
