package director

import "fmt"

type queryProvider struct{}

// at the moment just register application
// scenario will be added later
func (qp queryProvider) registerApplicationMutation(appName, _ string) string {
	return fmt.Sprintf(`mutation {
	result: registerApplication(in: {
		name: "%s"
	}) { id } }`, appName)
}

// from template
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
