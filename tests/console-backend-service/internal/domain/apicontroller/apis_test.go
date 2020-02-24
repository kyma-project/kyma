// +build acceptance

package apicontroller

import (
	"fmt"
	"testing"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	apiName                    = "test-api"
	apiNamespace               = "console-backend-service-api"
	hostname                   = "test-hostname"
	serviceName                = "test-service-name"
	servicePort                = 8080
	jwksUri                    = "http://test-jwks-uri"
	issuer                     = "test-issuer"
	disableIstioAuthPolicyMTLS = true
	authenticationEnabled      = true
	newHostname                = "different-hostname"
)

type apiQueryResponse struct {
	Api api `json:"api"`
}

type apisQueryResponse struct {
	Apis []api `json:"apis"`
}

type apiCreateResponse struct {
	CreateAPI api `json:"createAPI"`
}

type apiUpdateResponse struct {
	UpdateAPI api `json:"updateAPI"`
}

type apiDeleteResponse struct {
	DeleteAPI api `json:"deleteAPI"`
}

type apiEvent struct {
	Type string
	API  api
}

type service struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type api struct {
	Name                   string                 `json:"name"`
	Hostname               string                 `json:"hostname"`
	Service                service                `json:"service"`
	AuthenticationPolicies []authenticationPolicy `json:"authenticationPolicies"`
}

type authenticationPolicy struct {
	Type    string `json:"type"`
	Issuer  string `json:"issuer"`
	JwksURI string `json:"jwksURI"`
}

func TestApisQuery(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	t.Log("Creating namespace")
	namespace, err := k8sClient.Namespaces().Create(fixNamespace())
	require.NoError(t, err)

	defer func() {
		t.Log("Deleting namespace")
		err = k8sClient.Namespaces().Delete(namespace.Name, &metav1.DeleteOptions{})
		require.NoError(t, err)
	}()

	t.Log("Subscribing On APIs")
	subscription := subscribeApiEvent(c, namespace.Name)
	defer subscription.Close()

	t.Log("Creating API")
	createRes, err := createApi(c, namespace.Name)
	require.NoError(t, err)
	checkOutput(t, createRes.CreateAPI)
	assert.Equal(t, hostname, createRes.CreateAPI.Hostname)

	t.Log("Checking Subscription Event")
	event, err := readApiEvent(subscription)
	assert.NoError(t, err)
	checkApiEvent(t, "ADD", apiName, event)

	t.Log("Updating API")
	updateRes, err := updateApi(c, namespace.Name)
	require.NoError(t, err)
	checkOutput(t, updateRes.UpdateAPI)
	assert.Equal(t, newHostname, updateRes.UpdateAPI.Hostname)

	t.Log("Querying for API")
	var apiRes apiQueryResponse
	err = c.Do(fixAPIQuery(namespace.Name), &apiRes)
	require.NoError(t, err)
	checkOutput(t, apiRes.Api)
	assert.Equal(t, newHostname, apiRes.Api.Hostname)

	t.Log("Querying for APIs")
	var apisRes apisQueryResponse
	err = c.Do(fixAPIsQuery(namespace.Name), &apisRes)
	require.NoError(t, err)
	checkOutput(t, apisRes.Apis[0])

	t.Log("Deleting API")
	deleteRes, err := deleteApi(c, apiName, namespace.Name)
	require.NoError(t, err)
	checkOutput(t, deleteRes.DeleteAPI)

	t.Log("Checking authorization directives...")
	as := auth.New()
	ops := &auth.OperationsInput{
		auth.Get:    {fixAPIQuery(namespace.Name)},
		auth.List:   {fixAPIsQuery(namespace.Name)},
		auth.Create: {fixMutation("createAPI", hostname, namespace.Name)},
		auth.Update: {fixMutation("updateAPI", hostname, namespace.Name)},
		auth.Delete: {fixDeleteMutation(apiName, namespace.Name)},
	}
	as.Run(t, ops)
}

func fixNamespace() *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: apiNamespace,
		},
	}
}

const apiQuery = `
	name
	hostname
	service {
		name
		port
	}
	authenticationPolicies {
		type
		issuer
		jwksURI
	}
	creationTimestamp
`

func fixAPIQuery(namespace string) *graphql.Request {
	query := fmt.Sprintf(`
		query ($name: String!, $namespace: String!) {
			api(name: $name, namespace: $namespace) {
				%s
			}
		}
	`, apiQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", apiName)
	req.SetVar("namespace", namespace)

	return req
}

func fixAPIsQuery(namespace string) *graphql.Request {
	query := fmt.Sprintf(`
		query ($namespace: String!) {
			apis(namespace: $namespace) {
				%s
			}
		}
	`, apiQuery)
	req := graphql.NewRequest(query)
	req.SetVar("namespace", namespace)

	return req
}

func fixMutation(mutation string, hostname string, namespace string) *graphql.Request {
	query := fmt.Sprintf(`
		mutation %s($name: String!, $namespace: String!, $servicePort: Int!, $hostname: String!, $serviceName: String!, $jwksUri: String!, $issuer: String!) {
			%s(name: $name, namespace: $namespace, params: {
				servicePort: $servicePort, hostname: $hostname, serviceName: $serviceName, jwksUri: $jwksUri, issuer: $issuer
			}) {
			%s
		  }
		}
	`, mutation, mutation, apiQuery)

	req := graphql.NewRequest(query)
	req.SetVar("name", apiName)
	req.SetVar("namespace", namespace)
	req.SetVar("hostname", hostname)
	req.SetVar("serviceName", serviceName)
	req.SetVar("jwksUri", jwksUri)
	req.SetVar("issuer", issuer)
	req.SetVar("servicePort", servicePort)
	req.SetVar("disableIstioAuthPolicyMTLS", disableIstioAuthPolicyMTLS)
	req.SetVar("authenticationEnabled", authenticationEnabled)

	return req
}

func fixDeleteMutation(name, namespace string) *graphql.Request {
	query := fmt.Sprintf(`
		mutation deleteAPI($name: String!, $namespace: String!) {
			deleteAPI(name: $name, namespace: $namespace) {
				%s
			}
		}
	`, apiQuery)

	req := graphql.NewRequest(query)
	req.SetVar("name", name)
	req.SetVar("namespace", namespace)

	return req
}

func createApi(c *graphql.Client, namespace string) (apiCreateResponse, error) {
	req := fixMutation("createAPI", hostname, namespace)

	var res apiCreateResponse
	err := c.Do(req, &res)

	return res, err
}

func updateApi(c *graphql.Client, namespace string) (apiUpdateResponse, error) {
	req := fixMutation("updateAPI", newHostname, namespace)

	var res apiUpdateResponse
	err := c.Do(req, &res)

	return res, err
}

func deleteApi(c *graphql.Client, name, namespace string) (apiDeleteResponse, error) {
	req := fixDeleteMutation(name, namespace)

	var res apiDeleteResponse
	err := c.Do(req, &res)

	return res, err
}

func subscribeApiEvent(c *graphql.Client, namespace string) *graphql.Subscription {
	query := fmt.Sprintf(`
			subscription ($namespace: String!){
				apiEvent (namespace: $namespace){
					type 
					api {
						%s
					}
				}
			}
		`, apiQuery)
	req := graphql.NewRequest(query)
	req.SetVar("namespace", namespace)

	return c.Subscribe(req)
}

func readApiEvent(sub *graphql.Subscription) (apiEvent, error) {
	type Response struct {
		ApiEvent apiEvent
	}
	var event Response
	err := sub.Next(&event, tester.DefaultSubscriptionTimeout)

	return event.ApiEvent, err
}

func checkOutput(t *testing.T, apiMutation api) {
	assert.Equal(t, apiName, apiMutation.Name)
	assert.Equal(t, serviceName, apiMutation.Service.Name)
	assert.Equal(t, servicePort, apiMutation.Service.Port)
	assert.Equal(t, "JWT", apiMutation.AuthenticationPolicies[0].Type)
	assert.Equal(t, issuer, apiMutation.AuthenticationPolicies[0].Issuer)
	assert.Equal(t, jwksUri, apiMutation.AuthenticationPolicies[0].JwksURI)
}

func checkApiEvent(t *testing.T, expectedType, expectedName string, actual apiEvent) {
	assert.Equal(t, expectedType, actual.Type)
	checkOutput(t, actual.API)
}
