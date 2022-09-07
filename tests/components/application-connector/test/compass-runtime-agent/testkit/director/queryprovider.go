package director

import "fmt"

type queryProvider struct{}

func (qp queryProvider) registerApplicationMutation(appName, scenarioName string) string {
	return fmt.Sprintf(`mutation {
	result: registerApplication(in: {
		name: "%s" 
	}) { id } }`, appName)
	//return fmt.Sprintf(`mutation {
	//result: registerApplication(in: %s) { id } }`, appInput)
}

func (qp queryProvider) unregisterApplicationMutation(applicationID string) string {
	return fmt.Sprintf(`mutation {
	result: unregisterApplication(id: "%s") {
		id
	} }`, applicationID)
}
