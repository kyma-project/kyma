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

func (qp queryProvider) unregisterApplicationMutation(applicationID string) string {
	return fmt.Sprintf(`mutation {
	result: unregisterApplication(id: "%s") {
		id
	} }`, applicationID)
}
