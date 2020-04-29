// +build acceptance

package serverless

import (
	"fmt"
	"testing"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/configurer"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func TestFunctionEventQueries(t *testing.T) {
	c, err := graphql.New()
	assert.NoError(t, err)

	coreCli, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	namespace, err := fixNamespace(coreCli)
	require.NoError(t, err)
	namespaceName := namespace.Name()
	defer namespace.Delete()

	subscription := subscribeFunctionEvent(c, createFunctionEventArguments("1", namespaceName), functionEventDetailsFields())
	defer subscription.Close()

	labels := []string{FunctionLabel}
	err = mutationFunction(c, "createFunction", mutationFunctionArguments("1", namespaceName, labels), functionDetailsFields())
	require.NoError(t, err)

	event, err := readFunctionEvent(subscription)
	require.NoError(t, err)

	expectedFunction := fixFunction("1", namespaceName, labels)
	expectedEvent := fixFunctionEvent("ADD", expectedFunction)
	checkFunctionEvent(t, expectedEvent, event)

	function, err := queryFunction(c, queryFunctionArguments("1", namespaceName), functionDetailsFields())
	require.NoError(t, err)
	checkFunctionQuery(t, expectedFunction, function)

	labels = []string{FunctionLabel, FunctionLabel}
	err = mutationFunction(c, "updateFunction", mutationFunctionArguments("1", namespaceName, labels), functionDetailsFields())
	require.NoError(t, err)

	function, err = queryFunction(c, queryFunctionArguments("1", namespaceName), functionDetailsFields())
	require.NoError(t, err)

	expectedFunction = fixFunction("1", namespaceName, labels)
	checkFunctionQuery(t, expectedFunction, function)

	err = mutationFunction(c, "deleteFunction", deleteFunctionArguments("1", namespaceName), functionMetadataDetailsFields())
	require.NoError(t, err)

	err = mutationFunction(c, "createFunction", mutationFunctionArguments("2", namespaceName, labels), functionDetailsFields())
	require.NoError(t, err)
	err = mutationFunction(c, "createFunction", mutationFunctionArguments("3", namespaceName, labels), functionDetailsFields())
	require.NoError(t, err)

	functions, err := queryFunctions(c, namespaceName, functionDetailsFields())
	names := []string{fmt.Sprintf("%s-%s", FunctionNamePrefix, "2"), fmt.Sprintf("%s-%s", FunctionNamePrefix, "3")}
	checkFunctionList(t, namespaceName, names, functions)

	functionsMeta := []FunctionMetadataInput{
		{Name: fmt.Sprintf("%s-%s", FunctionNamePrefix, "2"), Namespace: namespaceName},
		{Name: fmt.Sprintf("%s-%s", FunctionNamePrefix, "3"), Namespace: namespaceName},
	}
	err = mutationFunction(c, "deleteManyFunctions", deleteManyFunctionsArguments(namespaceName, functionsMeta), functionMetadataDetailsFields())
	require.NoError(t, err)

	functions, err = queryFunctions(c, namespaceName, functionDetailsFields())
	assert.Equal(t, 0, len(functions))
}

func checkFunctionList(t *testing.T, expectedNamespace string, expectedNames []string, actual []Function) {
	assert.Equal(t, len(expectedNames), len(actual))
	for _, function := range actual {
		assert.Equal(t, expectedNamespace, function.Namespace)
		assert.Equal(t, true, isInArray(function.Name, expectedNames))
	}
}

func isInArray(data string, array []string) bool {
	for _, elem := range array {
		if elem == data {
			return true
		}
	}
	return false
}

func queryFunctions(client *graphql.Client, namespace, details string) ([]Function, error) {
	query := fmt.Sprintf(`
		query{
			functions (
				namespace: "%s""
			){
				%s
			}
		}
	`, namespace, details)

	req := graphql.NewRequest(query)
	var res FunctionListQueryResponse
	err := client.Do(req, &res)

	return res.Functions, err
}

func checkFunctionQuery(t *testing.T, expected, actual Function) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Namespace, actual.Namespace)
	assert.Equal(t, expected.Labels, actual.Labels)
}

func checkFunctionEvent(t *testing.T, expected, actual FunctionEvent) {
	assert.Equal(t, expected.Type, actual.Type)
	checkFunctionQuery(t, expected.Function, actual.Function)
}

func fixFunctionEvent(eventType string, function Function) FunctionEvent {
	return FunctionEvent{
		Type:     eventType,
		Function: function,
	}
}

func fixFunction(nameSuffix, namespace string, labels []string) Function {
	labelTemplate := map[string]string{}
	for _, label := range labels {
		labelTemplate[label] = label
	}

	return Function{
		Name:      fmt.Sprintf("%s-%s", FunctionNamePrefix, nameSuffix),
		Namespace: namespace,
		Labels:    labelTemplate,
	}
}

func readFunctionEvent(sub *graphql.Subscription) (FunctionEvent, error) {
	type Response struct {
		FunctionEvent FunctionEvent
	}

	var response Response
	err := sub.Next(&response, tester.DefaultDeletionTimeout)

	return response.FunctionEvent, err
}

func queryFunction(client *graphql.Client, arguments, details string) (Function, error) {
	query := fmt.Sprintf(`
		query{
			function (
				%s
			){
				%s
			}
		}
	`, arguments, details)
	req := graphql.NewRequest(query)
	var res Function
	err := client.Do(req, &res)

	return res, err
}

func mutationFunction(client *graphql.Client, requestType, arguments, resourceDetailsQuery string) error {
	query := fmt.Sprintf(`
		mutation {
			%s (
				%s
			){
				%s
			}
		}
	`, requestType, arguments, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	err := client.Do(req, nil)

	return err
}

func queryFunctionArguments(nameSuffix, namespace string) string {
	return fmt.Sprintf(`
		name: "%s-%s",
		namespace: "%s"
	`, FunctionNamePrefix, nameSuffix, namespace)
}

func deleteManyFunctionsArguments(namespace string, functionsMetadata []FunctionMetadataInput) string {
	functions := ""
	for i, meta := range functionsMetadata {
		functions += fmt.Sprintf(`{name: "%s", namespace: "%s"}`, meta.Name, meta.Namespace)
		if i+1 < len(functionsMetadata) {
			functions += ", "
		}
	}

	return fmt.Sprintf(`
		namespace: "%s",
		functions: [ %s ]
	`, namespace, functions)
}

func deleteFunctionArguments(functionSuffix, namespace string) string {
	return fmt.Sprintf(`
		namespace: "%s",
		function: {
			name: "%s-%s",
    		namespace: "%s"
		}
	`, namespace, FunctionNamePrefix, functionSuffix, namespace)
}

func mutationFunctionArguments(functionNameSuffix, namespaceName string, labels []string) string {
	labelTemplate := ""
	if len(labels) != 0 {
		for i, label := range labels {
			labelTemplate += fmt.Sprintf(`{"%s": "%s"}`, label, label)
			if i+1 < len(labels) {
				labelTemplate += ", "
			}
		}
	}

	return fmt.Sprintf(`
		name: "%s-%s",
		namespace: "%s",
		params: {
			labels: [ %s ],
			source: "module.exports = { main: function(event, context) { return 'Hello World' } }",
			dependencies: "",
			env: [  ],
			replicas: {  },
			resources: { limits: { memory: "100m", cpu: "128Mi" }, requests: { memory: "50m", cpu: "64Mi" } },
		},
	`, FunctionNamePrefix, functionNameSuffix, namespaceName, labelTemplate)
}

func createFunctionEventArguments(functionSuffix, namespace string) string {
	return fmt.Sprintf(`
		namespace: "%s"
		functionName: "%s-%s"
	`, namespace, FunctionNamePrefix, functionSuffix)
}

func functionEventDetailsFields() string {
	return fmt.Sprintf(`
        type
        function {
			%s
        }
    `, functionDetailsFields())
}

func functionMetadataDetailsFields() string {
	return `
		name
		namespace
	`
}

func functionDetailsFields() string {
	return `
		name
		namespace
		UID
		labels
		source
		dependencies
		env {
			name
			value
		}
		replicas {
			min
			max
		}
		resources {
			limits {
				memory
    			cpu
			}
			requests {
				memory
    			cpu
			}
		}
		status {
			phase
			reason
			message
		}
	`
}

func subscribeFunctionEvent(client *graphql.Client, arguments, resourceDetailsQuery string) *graphql.Subscription {
	query := fmt.Sprintf(`
		subscription {
			functionEvent (
				%s
			){
				%s
			}
		}
	`, arguments, resourceDetailsQuery)
	req := graphql.NewRequest(query)

	return client.Subscribe(req)
}

func fixNamespace(dynamicCli *corev1.CoreV1Client) (*configurer.NamespaceConfigurer, error) {
	namespace := configurer.NewNamespace(NamespacePrefix, dynamicCli)
	err := namespace.Create(nil)

	return namespace, err
}
