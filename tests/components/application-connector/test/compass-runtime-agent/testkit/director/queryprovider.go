package director

import "fmt"

type queryProvider struct{}

func (qp queryProvider) registerApplicationMutation(appName, _ string) string {
	return fmt.Sprintf(`mutation {
	result: registerApplication(in: {
		name: "%s"
	}) { id } }`, appName)
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

func (qp queryProvider) assignFormationForAppMutation(applicationId, formationName string) string {
	return fmt.Sprintf(`mutation {
	result: assignFormation(
		objectID: "%s"
		objectType: APPLICATION
		formation: { name: "%s" }
	) { id } }`, applicationId, formationName)
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
