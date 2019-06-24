package apicontroller

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/common/ingressgateway"

	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type integrationTestContext struct{}

func TestIntegrationSpec(t *testing.T) {

	domainName := os.Getenv(domainNameEnv)
	if domainName == "" {
		t.Fatal("Domain name not set.")
	}

	ctx := integrationTestContext{}
	testID := ctx.generateTestID(testIDLength)

	log.Infof("Running test: %s", testID)

	httpClient, err := ingressgateway.FromEnv().Client()
	if err != nil {
		t.Fatalf("Cannot get ingressgateway client: %s", err)
	}

	t.Logf("Set up...")
	fixture := setUpOrExit(k8sClient, namespace, testID)

	var lastAPI *kymaApi.Api

	suiteFinished := false

	Convey("API Controller should", t, func() {

		Reset(func() {
			if suiteFinished {
				t.Logf("Tear down...")
				fixture.tearDown()
			}
		})

		kymaInterface, kymaErr := kyma.NewForConfig(kubeConfig)
		if kymaErr != nil {
			log.Fatalf("can create kyma clientset. Root cause: %v", kymaErr)
		}

		suiteFinished = false

		Convey("create API with authentication disabled", func() {

			api := ctx.apiFor(testID, domainName, namespace, fixture.SampleAppService, apiSecurityDisabled, true)

			lastAPI, err = kymaInterface.GatewayV1alpha2().Apis(namespace).Create(api)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			ctx.validateAPINotSecured(httpClient, lastAPI.Spec.Hostname, "/")
			lastAPI, err = kymaInterface.GatewayV1alpha2().Apis(namespace).Get(lastAPI.Name, metav1.GetOptions{})
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)
		})

		Convey("update API with custom jwt configuration", func() {

			api := *lastAPI
			ctx.setCustomJwtAuthenticationConfig(&api)

			lastAPI, err = kymaInterface.GatewayV1alpha2().Apis(namespace).Update(&api)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			ctx.validateAPISecured(httpClient, lastAPI.Spec.Hostname, "/")
			ctx.validateAPINotSecured(httpClient, lastAPI.Spec.Hostname, "/do/not/use/in/production")
			ctx.validateAPINotSecured(httpClient, lastAPI.Spec.Hostname, "/web/static/favicon.ico")
			lastAPI, err = kymaInterface.GatewayV1alpha2().Apis(namespace).Get(lastAPI.Name, metav1.GetOptions{})
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)
		})

		Convey("delete API", func() {

			suiteFinished = true
			ctx.checkPreconditions(lastAPI, t)

			err := kymaInterface.GatewayV1alpha2().Apis(namespace).Delete(lastAPI.Name, &metav1.DeleteOptions{})
			So(err, ShouldBeNil)

			_, err = kymaInterface.GatewayV1alpha2().Apis(namespace).Get(lastAPI.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)
		})
	})
}

func (ctx integrationTestContext) apiFor(testID string, domainName string, namespace string, svc *apiv1.Service, secured APISecurity, hostWithDomain bool) *kymaApi.Api {

	return &kymaApi.Api{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("sample-app-api-%s", testID),
		},
		Spec: kymaApi.ApiSpec{
			Hostname: ctx.hostnameFor(testID, domainName, hostWithDomain),
			Service: kymaApi.Service{
				Name: svc.Name,
				Port: int(svc.Spec.Ports[0].Port),
			},
			AuthenticationEnabled: (*bool)(&secured),
			Authentication:        []kymaApi.AuthenticationRule{},
		},
	}
}

func (integrationTestContext) setCustomJwtAuthenticationConfig(api *kymaApi.Api) {
	// OTHER EXAMPLE OF POSSIBLE VALUES:
	//issuer := "https://accounts.google.com"
	//jwksURI := "https://www.googleapis.com/oauth2/v3/certs"

	issuer := "https://accounts.google.com"
	jwksURI := "http://dex-service.kyma-system.svc.cluster.local:5556/keys"

	triggerRule := kymaApi.TriggerRule{
		ExcludedPaths: []kymaApi.MatchExpression{
			kymaApi.MatchExpression{ExprType: kymaApi.ExactMatch, Value: "/do/not/use/in/production1"},
			kymaApi.MatchExpression{ExprType: kymaApi.SuffixMatch, Value: "/favicons.ico"},
		},
	}

	rules := []kymaApi.AuthenticationRule{
		{
			Type: kymaApi.JwtType,
			Jwt: kymaApi.JwtAuthentication{
				Issuer:      issuer,
				JwksUri:     jwksURI,
				TriggerRule: &triggerRule,
			},
		},
	}

	secured := true
	if api.Spec.AuthenticationEnabled != nil && !(*api.Spec.AuthenticationEnabled) { // optional property, but if set earlier to false it will force auth disabled
		api.Spec.AuthenticationEnabled = &secured
	}
	api.Spec.Authentication = rules
}

func (integrationTestContext) checkPreconditions(lastAPI *kymaApi.Api, t *testing.T) {
	if lastAPI == nil {
		t.Fatal("Precondition failed - last API not set")
	}
}

func (integrationTestContext) hostnameFor(testID, domainName string, hostWithDomain bool) string {
	if hostWithDomain {
		return fmt.Sprintf("%s.%s", testID, domainName)
	}
	return testID
}

func (ctx integrationTestContext) validateAPISecured(httpClient *http.Client, hostname, path string) {

	response, err := ctx.withRetries(func() (*http.Response, error) {
		return httpClient.Get(fmt.Sprintf("https://%s%s", hostname, path))
	}, ctx.httpUnauthorizedPredicate)

	So(err, ShouldBeNil)
	So(response.StatusCode, ShouldEqual, http.StatusUnauthorized)
}

func (ctx integrationTestContext) validateAPINotSecured(httpClient *http.Client, hostname, path string) {

	response, err := ctx.withRetries(func() (*http.Response, error) {
		return httpClient.Get(fmt.Sprintf("https://%s%s", hostname, path))
	}, ctx.httpOkPredicate)

	So(err, ShouldBeNil)
	So(response.StatusCode, ShouldEqual, http.StatusOK)
}

func (integrationTestContext) withRetries(httpCall func() (*http.Response, error), shouldRetry func(*http.Response) bool) (*http.Response, error) {
	var retries uint = 120
	delay := 1 * time.Second

	var response *http.Response

	err := retry.Do(func() error {
		var err error
		response, err = httpCall()

		if err != nil {
			return err
		}
		if shouldRetry(response) {
			return errors.Errorf("unexpected response: %s", response.Status)
		}
		return nil
	},
		retry.Attempts(retries),
		retry.Delay(delay),
		retry.DelayType(retry.FixedDelay),
		retry.OnRetry(func(retryNo uint, err error) {
			log.Errorf("[%d / %d] Status: %s", retryNo, retries, err)
		}),
	)

	return response, err
}

func (integrationTestContext) httpOkPredicate(response *http.Response) bool {
	return response.StatusCode < 200 || response.StatusCode > 299
}

func (integrationTestContext) httpUnauthorizedPredicate(response *http.Response) bool {
	return response.StatusCode != 401
}

func (integrationTestContext) generateTestID(n int) string {

	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
