// +build acceptance

package apicontroller

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/dex"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	tester "github.com/kyma-project/kyma/tests/console-backend-service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	apiName      = "test-api"
	apiNamespace = "console-backend-service-api"
	hostname = "test-hostname"
	serviceName = "test-service-name"
	servicePort = 8080
	jwksUri = "http://test-jwks-uri"
	issuer = "test-issuer"
)

type apiQueryResponse struct {
	Api api `json:"api"`
}

type apisQueryResponse struct {
	Apis []api `json:"apis"`
}

type ApiCreateResponse struct {
	CreateAPI api `json:"createAPI"`
}

type ApiUpdateResponse struct {
	UpdateAPI api `json:"updateAPI"`
}

type ApiDeleteResponse struct {
	DeleteAPI api `json:"deleteAPI"`
}

type apiEvent struct {
	Type        string
	API api
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
	Type string `json:"type"`
	Issuer   string `json:"issuer"`
	JwksURI  string `json:"jwksURI"`
}

func TestApisQuery(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

	c, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	t.Log("Creating namespace")
	_, err = k8sClient.Namespaces().Create(fixNamespace())
	require.NoError(t, err)

	defer func() {
		t.Log("Deleting namespace")
		err = k8sClient.Namespaces().Delete(apiNamespace, &metav1.DeleteOptions{})
		require.NoError(t, err)
	}()

	t.Log("Subscribing On APIs")
	subscription := subscribeApiEvent(c, apiNamespace)
	defer subscription.Close()

	t.Log("Creating API")
	createRes, err := createApi(c, apiName, apiNamespace, hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
	require.NoError(t, err)
	checkOutput(t, createRes.CreateAPI)
	assert.Equal(t, hostname, createRes.CreateAPI.Hostname)

	t.Log("Checking Subscription Event")
	event, err := readApiEvent(subscription)
	assert.NoError(t, err)
	checkApiEvent(t, "ADD", apiName, event)

	newHostname := "different-hostname"
	t.Log("Updating API")
	updateRes, err := updateApi(c, apiName, apiNamespace, newHostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
	require.NoError(t, err)
	checkOutput(t, updateRes.UpdateAPI)
	assert.Equal(t, newHostname, updateRes.UpdateAPI.Hostname)

	t.Log("Querying for API")
	var apiRes apiQueryResponse
	err = c.Do(fixAPIQuery(), &apiRes)
	require.NoError(t, err)
	checkOutput(t, apiRes.Api)
	assert.Equal(t, newHostname, apiRes.Api.Hostname)

	t.Log("Querying for APIs")
	var apisRes apisQueryResponse
	err = c.Do(fixAPIsQuery(), &apisRes)
	require.NoError(t, err)
	checkOutput(t, apisRes.Apis[0])

	t.Log("Deleting API")
	deleteRes, err := deleteApi(c, apiName, apiNamespace)
	require.NoError(t, err)
	checkOutput(t, deleteRes.DeleteAPI)

	t.Log("Checking authorization directives...")
	as := auth.New()
	ops := &auth.OperationsInput{
		auth.List: {fixAPIQuery()},
	}
	as.Run(t, ops)
}

func fixNamespace() *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: apiNamespace,
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
`

func fixAPIQuery() *graphql.Request {
	query := fmt.Sprintf(`
		query ($name: String!, $namespace: String!) {
			api(name: $name, namespace: $namespace) {
				%s
			}
		}
	`, apiQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", apiName)
	req.SetVar("namespace", apiNamespace)

	return req
}

func fixAPIsQuery() *graphql.Request {
	query := fmt.Sprintf(`
		query ($namespace: String!) {
			apis(namespace: $namespace) {
				%s
			}
		}
	`, apiQuery)
	req := graphql.NewRequest(query)
	req.SetVar("namespace", apiNamespace)

	return req
}

func fixMutation(mutation string, name, namespace, hostname, serviceName, jwksUri, issuer string, servicePort int, disableIstioAuthPolicyMTLS, authenticationEnabled *bool) *graphql.Request {
	query := fmt.Sprintf(`
		mutation %s($name: String!, $namespace: String!, $servicePort: Int!, $hostname: String!, $serviceName: String!, $jwksUri: String!, $issuer: String!) {
			%s(params: {
				name: $name, namespace: $namespace, servicePort: $servicePort, hostname: $hostname, serviceName: $serviceName, jwksUri: $jwksUri, issuer: $issuer
			}) {
			%s
		  }
		}
	`, mutation, mutation, apiQuery)

	req := graphql.NewRequest(query)
	req.SetVar("name", name)
	req.SetVar("namespace", namespace)
	req.SetVar("hostname", hostname)
	req.SetVar("serviceName", serviceName)
	req.SetVar("jwksUri", jwksUri)
	req.SetVar("issuer", issuer)
	req.SetVar("servicePort", servicePort)
	req.SetVar("disableIstioAuthPolicyMTLS", nil)
	req.SetVar("authenticationEnabled", nil)

	return req
}

func createApi(c *graphql.Client, name, namespace, hostname, serviceName, jwksUri, issuer string, servicePort int, disableIstioAuthPolicyMTLS, authenticationEnabled *bool) (ApiCreateResponse, error) {
	req := fixMutation("createAPI", name, namespace, hostname, serviceName, jwksUri, issuer, servicePort, disableIstioAuthPolicyMTLS, authenticationEnabled)

	var res ApiCreateResponse
	err := c.Do(req, &res)

	return res, err
}

func updateApi(c *graphql.Client, name, namespace, hostname, serviceName, jwksUri, issuer string, servicePort int, disableIstioAuthPolicyMTLS, authenticationEnabled *bool) (ApiUpdateResponse, error) {
	req := fixMutation("updateAPI", name, namespace, hostname, serviceName, jwksUri, issuer, servicePort, disableIstioAuthPolicyMTLS, authenticationEnabled)

	var res ApiUpdateResponse
	err := c.Do(req, &res)

	return res, err
}

func deleteApi(c *graphql.Client, name, namespace string) (ApiDeleteResponse, error) {
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
	var res ApiDeleteResponse
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

func checkOutput(t *testing.T, apiMutation api)  {
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
