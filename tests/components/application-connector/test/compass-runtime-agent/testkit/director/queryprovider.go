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

func (qp queryProvider) registerApplicationNewMutation(appName, displayName string) string {
	return fmt.Sprintf(`mutation {
	result: registerApplicationFromTemplate(in: {
		templateName: "SAP Commerce Cloud"
		values: [
			{ placeholder: "name", value: "%s" }
			{ placeholder: "display-name", value: "%s" }
        ]
	}) { id } }`, appName, displayName)
}

/*
mutation {
  result: registerApplicationFromTemplate(
    in: {
      templateName: "SAP template"
      values: [
        { placeholder: "name", value: "new-name" }
        { placeholder: "display-name", value: "new-display-name" }
      ]
    }
  )



*/

func (qp queryProvider) unregisterApplicationMutation(applicationID string) string {
	return fmt.Sprintf(`mutation {
	result: unregisterApplication(id: "%s") {
		id
	} }`, applicationID)
}
