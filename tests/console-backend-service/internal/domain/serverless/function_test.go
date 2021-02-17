// +build acceptance

package serverless

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/configurer"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const sleepTime = 5 * time.Second

func TestFunctionEventQueries(t *testing.T) {
	c, err := graphql.New()
	assert.NoError(t, err)

	coreCli, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	t.Logf("Create namespace with prefix: %s", namespacePrefix)
	namespace, err := fixNamespace(coreCli)
	require.NoError(t, err)
	namespaceName := namespace.Name()
	defer namespace.Delete()

	t.Logf("Subscribe on function: %s", functionName1)
	subscription := subscribeFunctionEvent(c, createFunctionEventArguments(functionName1, namespaceName), functionEventDetailsFields())
	defer subscription.Close()

	t.Logf("Create function: %s", functionName1)
	err = mutationFunction(c, "createFunction",
		mutationFunctionArguments(functionName1, namespaceName, nil), functionDetailsFields())
	require.NoError(t, err)

	t.Logf("Check subscription event")
	event, err := readFunctionEvent(subscription)
	require.NoError(t, err)

	expectedFunction := fixFunction(functionName1, namespaceName, nil)
	expectedEvent := fixFunctionEvent("ADD", expectedFunction)
	checkFunctionEvent(t, expectedEvent, event)

	//wait for reactions from function controller to function CR
	time.Sleep(sleepTime)

	t.Logf("Query function: %s", functionName1)
	function, err := queryFunction(c, queryFunctionArguments(functionName1, namespaceName), functionDetailsFields())
	require.NoError(t, err)
	checkFunctionQuery(t, expectedFunction, function)

	t.Logf("Update function: %s", functionName1)
	labels := []string{functionLabel}
	err = mutationFunction(c, "updateFunction",
		mutationFunctionArguments(functionName1, namespaceName, labels), functionDetailsFields())
	require.NoError(t, err)

	t.Logf("Query function: %s", functionName1)
	function, err = queryFunction(c, queryFunctionArguments(functionName1, namespaceName), functionDetailsFields())
	require.NoError(t, err)

	expectedFunction = fixFunction(functionName1, namespaceName, labels)
	checkFunctionQuery(t, expectedFunction, function)

	t.Logf("Delete function: %s", functionName1)
	err = mutationFunction(c, "deleteFunction",
		deleteFunctionArguments(functionName1, namespaceName), functionMetadataDetailsFields())
	require.NoError(t, err)

	t.Logf("Create functions: %s, %s", functionName2, functionName3)
	err = mutationFunction(c, "createFunction",
		mutationFunctionArguments(functionName2, namespaceName, labels), functionDetailsFields())
	require.NoError(t, err)
	err = mutationFunction(c, "createFunction",
		mutationFunctionArguments(functionName3, namespaceName, labels), functionDetailsFields())
	require.NoError(t, err)

	t.Logf("Query functions: %s, %s", functionName2, functionName3)
	functions, err := queryFunctions(c, queryFunctionsArguments(namespaceName), functionDetailsFields())
	checkFunctionList(t, namespaceName, []string{functionName2, functionName3}, functions)

	t.Logf("Delete functions: %s, %s", functionName2, functionName3)
	functionsMeta := []FunctionMetadataInput{
		{Name: functionName2, Namespace: namespaceName},
		{Name: functionName3, Namespace: namespaceName},
	}
	err = mutationFunction(c, "deleteManyFunctions",
		deleteManyFunctionsArguments(namespaceName, functionsMeta), functionMetadataDetailsFields())
	require.NoError(t, err)

	t.Logf("Query functions in namespace: %s", namespaceName)
	functions, err = queryFunctions(c, queryFunctionsArguments(namespaceName), functionDetailsFields())
	assert.Equal(t, 0, len(functions))

	t.Log("Check auth connection")
	opts := &auth.OperationsInput{
		auth.Create: {fixFunctionRequest("mutation", "createFunction",
			mutationFunctionArguments(functionName1, namespaceName, nil), functionDetailsFields())},
		auth.Update: {fixFunctionRequest("mutation", "updateFunction",
			mutationFunctionArguments(functionName1, namespaceName, labels), functionDetailsFields())},
		auth.Get: {fixFunctionRequest("query", "function",
			queryFunctionArguments(functionName1, namespaceName), functionDetailsFields())},
		auth.List: {fixFunctionRequest("query", "functions",
			queryFunctionsArguments(namespaceName), functionDetailsFields())},
		auth.Delete: {fixFunctionRequest("mutation", "deleteFunction",
			deleteFunctionArguments(functionName1, namespaceName), functionMetadataDetailsFields())},
		auth.Watch: {fixFunctionRequest("subscription", "functionEvent",
			createFunctionEventArguments(functionName1, namespaceName), functionEventDetailsFields())},
	}
	auth.New().Run(t, opts)
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

func queryFunctions(client *graphql.Client, arguments, details string) ([]Function, error) {
	req := fixFunctionRequest("query", "functions", arguments, details)
	var res FunctionListQueryResponse
	err := client.Do(req, &res)

	return res.Functions, err
}

func checkFunctionQuery(t *testing.T, expected, actual Function) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Namespace, actual.Namespace)
	assert.Equal(t, len(expected.Labels), len(actual.Labels))
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

func fixFunction(name, namespace string, labels []string) Function {
	labelTemplate := map[string]string{}
	for _, label := range labels {
		labelTemplate[label] = label
	}

	return Function{
		Name:      name,
		Namespace: namespace,
		Labels:    labelTemplate,
	}
}

func readFunctionEvent(sub *graphql.Subscription) (FunctionEvent, error) {
	type Response struct {
		FunctionEvent FunctionEvent
	}

	var response Response
	err := sub.Next(&response, tester.DefaultSubscriptionTimeout)

	return response.FunctionEvent, err
}

func queryFunction(client *graphql.Client, arguments, details string) (Function, error) {
	type Response struct {
		Function Function
	}

	req := fixFunctionRequest("query", "function", arguments, details)
	var res Response
	err := client.Do(req, &res)

	return res.Function, err
}

func mutationFunction(client *graphql.Client, requestType, arguments, resourceDetailsQuery string) error {
	req := fixFunctionRequest("mutation", requestType, arguments, resourceDetailsQuery)
	err := client.Do(req, nil)

	return err
}

func subscribeFunctionEvent(client *graphql.Client, arguments, resourceDetailsQuery string) *graphql.Subscription {
	req := fixFunctionRequest("subscription", "functionEvent", arguments, resourceDetailsQuery)

	return client.Subscribe(req)
}

func fixFunctionRequest(requestType, requestName, arguments, details string) *graphql.Request {
	query := fmt.Sprintf(`
		%s {
			%s (
				%s
			){
				%s
			}
		}
	`, requestType, requestName, arguments, details)
	return graphql.NewRequest(query)
}

func queryFunctionArguments(name, namespace string) string {
	return fmt.Sprintf(`
		name: "%s",
		namespace: "%s"
	`, name, namespace)
}

func queryFunctionsArguments(namespace string) string {
	return fmt.Sprintf(`
		namespace: "%s"
	`, namespace)
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

func deleteFunctionArguments(name, namespace string) string {
	return fmt.Sprintf(`
		namespace: "%s",
		function: {
			name:      "%s",
    		namespace: "%s"
		}
	`, namespace, name, namespace)
}

func mutationFunctionArguments(name, namespaceName string, labels []string) string {
	labelTemplate := ""
	if len(labels) != 0 {
		for i, label := range labels {
			labelTemplate += fmt.Sprintf(`%s: "%s"`, label, label)
			if i+1 < len(labels) {
				labelTemplate += ", "
			}
		}
	}

	return fmt.Sprintf(`
		name:      "%s",
		namespace: "%s",
		params: {
			labels: { %s },
			source: "module.exports = { main: function (event, context) { return \"Hello World!\"; } }",
			dependencies: "{ \"name\": \"asd\", \"version\": \"1.0.0\", \"dependencies\": {} }",
			env: [  ],
			replicas: { min: 1, max: 1 },
			resources: { limits: { memory: "128Mi", cpu: "100m" }, requests: { memory: "64Mi", cpu: "50m" } },
			buildResources: { limits: { memory: "1100Mi", cpu: "1100m" }, requests: { memory: "700Mi", cpu: "700m" } },
		},
	`, name, namespaceName, labelTemplate)
}

func createFunctionEventArguments(name, namespace string) string {
	return fmt.Sprintf(`
		namespace: "%s"
		functionName: "%s"
	`, namespace, name)
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
		labels
	`
}

func fixNamespace(dynamicCli *corev1.CoreV1Client) (*configurer.NamespaceConfigurer, error) {
	namespace := configurer.NewNamespace(namespacePrefix, dynamicCli)
	err := namespace.Create(nil)

	return namespace, err
}
