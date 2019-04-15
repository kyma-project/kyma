// +build acceptance

package apicontroller

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/dex"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"

	gateway "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/waiter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	apiName      = "test-api"
	apiNamespace = "console-backend-service-api"
)

type apisQueryResponse struct {
	Apis []api `json:"apis"`
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
	AuthType string `json:"type"`
	Issuer   string `json:"issuer"`
	JwksURI  string `json:"jwksURI"`
}

func TestApisQuery(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

	c, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	gatewayClient, _, err := client.NewGatewayClientWithConfig()
	require.NoError(t, err)

	t.Log("Creating namespace...")
	_, err = k8sClient.Namespaces().Create(fixNamespace())
	require.NoError(t, err)

	defer func() {
		t.Log("Deleting namespace...")
		err = k8sClient.Namespaces().Delete(apiNamespace, &metav1.DeleteOptions{})
		require.NoError(t, err)
	}()

	t.Log("Creating API...")
	_, err = gatewayClient.GatewayV1alpha2().Apis(apiNamespace).Create(fixAPI(c.Config.EnvConfig.Domain))
	require.NoError(t, err)

	t.Log("Retrieving API...")
	var api *gateway.Api
	err = waiter.WaitAtMost(func() (bool, error) {
		var err error
		api, err = gatewayClient.GatewayV1alpha2().Apis(apiNamespace).Get(apiName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Querying for APIs...")
	var apisRes apisQueryResponse
	err = c.Do(fixAPIQuery(), &apisRes)
	require.NoError(t, err)

	assert.Equal(t, api.Name, apisRes.Apis[0].Name)
	assert.Equal(t, api.Spec.Hostname, apisRes.Apis[0].Hostname)
	assert.Equal(t, api.Spec.Service.Name, apisRes.Apis[0].Service.Name)
	assert.Equal(t, api.Spec.Service.Port, apisRes.Apis[0].Service.Port)
	assert.Equal(t, 1, len(apisRes.Apis[0].AuthenticationPolicies))
	assert.Equal(t, string(api.Spec.Authentication[0].Type), apisRes.Apis[0].AuthenticationPolicies[0].AuthType)
	assert.Equal(t, api.Spec.Authentication[0].Jwt.Issuer, apisRes.Apis[0].AuthenticationPolicies[0].Issuer)
	assert.Equal(t, api.Spec.Authentication[0].Jwt.JwksUri, apisRes.Apis[0].AuthenticationPolicies[0].JwksURI)

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

func fixAPI(domain string) *gateway.Api {
	return &gateway.Api{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiName,
			Namespace: apiNamespace,
		},
		Spec: gateway.ApiSpec{
			Service: gateway.Service{
				Name: "some-service",
				Port: 80,
			},
			Hostname: fmt.Sprintf("some-service.%s", domain),
			Authentication: []gateway.AuthenticationRule{
				{
					Type: "JWT",
					Jwt: gateway.JwtAuthentication{
						JwksUri: "https://dex.kyma.domain/keys",
						Issuer:  "aaa",
					},
				},
			},
		},
	}
}

func fixAPIQuery() *graphql.Request {
	query := `query ($namespace: String!) {
				apis(namespace: $namespace) {
					name
    				hostname
    				service {
						name
						port
					}
    				authenticationPolicies {
						type
						jwksURI
						issuer
					}
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("namespace", apiNamespace)

	return req
}
