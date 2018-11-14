package apicontroller

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-test/deep"
	istioAuthApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/authentication.istio.io/v1alpha1"
	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	istioNetApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/networking.istio.io/v1alpha3"
	istioAuth "github.com/kyma-project/kyma/components/api-controller/pkg/clients/authentication.istio.io/clientset/versioned"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	istioNet "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type componentTestContext struct{}

func TestComponentSpec(t *testing.T) {

	domainName := os.Getenv(domainNameEnv)
	if domainName == "" {
		t.Fatal("Domain name not set.")
	}

	ctx := componentTestContext{}

	kubeConfig := ctx.defaultConfigOrExit()

	kymaClient := kyma.NewForConfigOrDie(kubeConfig)
	istioNetClient := istioNet.NewForConfigOrDie(kubeConfig)
	istioAuthClient := istioAuth.NewForConfigOrDie(kubeConfig)

	Convey("API Controller should", t, func() {

		Convey("create API with authentication disabled", func() {
			t.Log("create API with authentication disabled")

			testId := ctx.generateTestId(testIdLength)
			t.Logf("Running test: %s", testId)
			api := ctx.apiFor(testId, domainName, apiSecurityDisabled, true)

			lastApi, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpApi(kymaClient, lastApi, t, false)
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			lastApi, err = ctx.awaitApiChanged(kymaClient, lastApi, true, false)
			So(err, ShouldBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
			So(lastApi.Spec, ctx.ShouldDeepEqual, api.Spec)

			lastVs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastApi.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := ctx.virtualServiceFor(testId, domainName)
			So(err, ShouldBeNil)
			So(lastVs.Spec, ctx.ShouldDeepEqual, expectedVs)
		})

		Convey("create API with hostname without domain", func() {
			t.Log("create API with hostname without domain")

			testId := ctx.generateTestId(testIdLength)
			t.Logf("Running test: %s", testId)
			api := ctx.apiFor(testId, domainName, apiSecurityDisabled, false)

			lastApi, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpApi(kymaClient, lastApi, t, false)
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			lastApi, err = ctx.awaitApiChanged(kymaClient, lastApi, true, false)
			So(err, ShouldBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
			So(lastApi.Spec, ctx.ShouldDeepEqual, api.Spec)

			lastVs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastApi.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := ctx.virtualServiceFor(testId, domainName)
			So(err, ShouldBeNil)
			So(lastVs.Spec, ctx.ShouldDeepEqual, expectedVs)
		})

		Convey("not create API with wrong domain", func() {
			t.Log("not create API with wrong domain")

			testId := ctx.generateTestId(testIdLength)
			t.Logf("Running test: %s", testId)
			api := ctx.apiFor(testId, domainName+"x", apiSecurityDisabled, true)

			_, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			So(err, ShouldNotBeNil)
		})

		Convey("create API with default jwt configuration to enable authentication", func() {
			t.Log("create API with default jwt configuration to enable authentication")

			testId := ctx.generateTestId(testIdLength)
			t.Logf("Running test: %s", testId)
			api := ctx.apiFor(testId, domainName, apiSecurityDisabled, true)
			authEnabled := true
			api.Spec.AuthenticationEnabled = &authEnabled

			lastApi, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpApi(kymaClient, lastApi, t, false)
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			lastApi, err = ctx.awaitApiChanged(kymaClient, lastApi, true, true)
			So(err, ShouldBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
			So(lastApi.Spec, ctx.ShouldDeepEqual, api.Spec)

			vs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastApi.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := ctx.virtualServiceFor(testId, domainName)
			So(err, ShouldBeNil)
			So(vs.Spec, ctx.ShouldDeepEqual, expectedVs)

			lastPolicy, err := istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastApi.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			expectedPolicy := ctx.policyFor(testId, fmt.Sprintf("https://dex.%s", domainName))
			So(err, ShouldBeNil)
			So(lastPolicy.Spec, ctx.ShouldDeepEqual, expectedPolicy)
		})

		Convey("update API to disable authentication", func() {
			t.Log("update API to disable authentication")

			testId := ctx.generateTestId(testIdLength)
			t.Logf("Running test: %s", testId)
			api := ctx.apiFor(testId, domainName, apiSecurityEnabled, true)

			createdApi, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpApi(kymaClient, createdApi, t, false)
			So(err, ShouldBeNil)
			So(createdApi, ShouldNotBeNil)
			So(createdApi.ResourceVersion, ShouldNotBeEmpty)

			createdApi, err = ctx.awaitApiChanged(kymaClient, createdApi, true, true)
			So(err, ShouldBeNil)
			So(createdApi.ResourceVersion, ShouldNotBeEmpty)
			So(createdApi.Spec, ctx.ShouldDeepEqual, api.Spec)

			authEnabled := false
			createdApi.Spec.AuthenticationEnabled = &authEnabled

			updatedApi, err := kymaClient.GatewayV1alpha2().Apis(namespace).Update(createdApi)
			So(err, ShouldBeNil)
			So(updatedApi, ShouldNotBeNil)
			So(updatedApi.ResourceVersion, ShouldNotBeEmpty)

			updatedApi, err = ctx.awaitApiChanged(kymaClient, updatedApi, false, true)
			So(err, ShouldBeNil)
			So(updatedApi.ResourceVersion, ShouldNotBeEmpty)
			So(updatedApi.Spec, ctx.ShouldDeepEqual, createdApi.Spec)
			So(updatedApi.Status.AuthenticationStatus.Resource.Uid, ShouldBeEmpty)

			_, err = istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(createdApi.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)
		})

		Convey("create API with custom jwt configuration", func() {
			t.Log("create API with custom jwt configuration")

			testId := ctx.generateTestId(testIdLength)
			t.Logf("Running test: %s", testId)
			api := ctx.apiFor(testId, domainName, apiSecurityDisabled, true)
			ctx.setCustomJwtAuthenticationConfig(api)

			lastApi, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpApi(kymaClient, lastApi, t, false)
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			lastApi, err = ctx.awaitApiChanged(kymaClient, lastApi, true, true)
			So(err, ShouldBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
			So(lastApi.Spec, ctx.ShouldDeepEqual, api.Spec)

			policy, err := istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastApi.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			expectedPolicy := ctx.policyFor(testId, api.Spec.Authentication[0].Jwt.Issuer)
			So(err, ShouldBeNil)
			So(policy.Spec, ctx.ShouldDeepEqual, expectedPolicy)
		})

		Convey("delete API and all its related resources", func() {
			t.Log("delete API and all its related resources")

			testId := ctx.generateTestId(testIdLength)
			t.Logf("Running test: %s", testId)
			api := ctx.apiFor(testId, domainName, apiSecurityEnabled, true)

			lastApi, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpApi(kymaClient, lastApi, t, true)
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			lastApi, err = ctx.awaitApiChanged(kymaClient, lastApi, true, true)
			So(err, ShouldBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
			So(lastApi.Spec, ctx.ShouldDeepEqual, api.Spec)
			policy, err := istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastApi.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			So(err, ShouldBeNil)
			vs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastApi.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			So(err, ShouldBeNil)

			err = kymaClient.GatewayV1alpha2().Apis(namespace).Delete(lastApi.Name, &metav1.DeleteOptions{})
			So(err, ShouldBeNil)

			time.Sleep(5 * time.Second)

			_, err = kymaClient.GatewayV1alpha2().Apis(namespace).Get(lastApi.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)

			_, err = istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(policy.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)

			_, err = istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(vs.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)
		})
	})
}

func (ctx componentTestContext) apiFor(testId, domainName string, secured ApiSecurity, hostWithDomain bool) *kymaApi.Api {

	return &kymaApi.Api{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("sample-app-api-%s", testId),
		},
		Spec: kymaApi.ApiSpec{
			Hostname: ctx.hostnameFor(testId, domainName, hostWithDomain),
			Service: kymaApi.Service{
				Name: fmt.Sprintf("sample-app-svc-%s", testId),
				Port: 80,
			},
			AuthenticationEnabled: (*bool)(&secured),
			Authentication:        []kymaApi.AuthenticationRule{},
		},
	}
}

func (componentTestContext) virtualServiceFor(testId, domainName string) *istioNetApi.VirtualServiceSpec {
	return &istioNetApi.VirtualServiceSpec{
		Hosts:    []string{testId + "." + domainName},
		Gateways: []string{"kyma-gateway.kyma-system.svc.cluster.local"},
		Http: []*istioNetApi.HTTPRoute{
			{
				Match: []*istioNetApi.HTTPMatchRequest{
					{
						Uri: &istioNetApi.StringMatch{Regex: "/.*"},
					},
				},
				Route: []*istioNetApi.DestinationWeight{
					{
						Destination: &istioNetApi.Destination{
							Host: fmt.Sprintf("sample-app-svc-%s.kyma-system.svc.cluster.local", testId),
						},
					},
				},
			},
		},
	}
}

func (componentTestContext) policyFor(testId, issuer string) *istioAuthApi.PolicySpec {
	return &istioAuthApi.PolicySpec{
		Targets: istioAuthApi.Targets{
			{Name: fmt.Sprintf("sample-app-svc-%s", testId)},
		},
		PrincipalBinding: istioAuthApi.UseOrigin,
		Origins: istioAuthApi.Origins{
			{
				Jwt: &istioAuthApi.Jwt{
					Issuer:  issuer,
					JwksUri: "http://dex-service.kyma-system.svc.cluster.local:5556/keys",
				},
			},
		},
	}
}

func (componentTestContext) setCustomJwtAuthenticationConfig(api *kymaApi.Api) {
	// OTHER EXAMPLE OF POSSSIBLE VALUES:
	//issuer := "https://accounts.google.com"
	//jwksUri := "https://www.googleapis.com/oauth2/v3/certs"

	issuer := "https://accounts.google.com"
	jwksUri := "http://dex-service.kyma-system.svc.cluster.local:5556/keys"

	rules := []kymaApi.AuthenticationRule{
		{
			Type: kymaApi.JwtType,
			Jwt: kymaApi.JwtAuthentication{
				Issuer:  issuer,
				JwksUri: jwksUri,
			},
		},
	}

	secured := true
	if api.Spec.AuthenticationEnabled != nil && !(*api.Spec.AuthenticationEnabled) { // optional property, but if set earlier to false it will force auth disabled
		api.Spec.AuthenticationEnabled = &secured
	}
	api.Spec.Authentication = rules
}

func (componentTestContext) hostnameFor(testId, domainName string, hostWithDomain bool) string {
	if hostWithDomain {
		return fmt.Sprintf("%s.%s", testId, domainName)
	}
	return testId
}

func (componentTestContext) awaitApiChanged(iface *kyma.Clientset, api *kymaApi.Api, vsChanged, policyChanged bool) (*kymaApi.Api, error) {
	var result *kymaApi.Api
	err := retry.Do(func() error {
		lastApi, err := iface.GatewayV1alpha2().Apis(namespace).Get(api.Name, metav1.GetOptions{})

		if err != nil {
			return err
		}
		if vsChanged && lastApi.Status.VirtualServiceStatus.Resource.Version == api.Status.VirtualServiceStatus.Resource.Version {
			return fmt.Errorf("VirtualService not created, old: %s, new: %s",
				api.Status.VirtualServiceStatus.Resource.Version,
				lastApi.Status.VirtualServiceStatus.Resource.Version)
		}
		if policyChanged && lastApi.Status.AuthenticationStatus.Resource.Version == api.Status.AuthenticationStatus.Resource.Version {
			return fmt.Errorf("policy not created, old: %s, new: %s",
				api.Status.AuthenticationStatus.Resource.Version,
				lastApi.Status.AuthenticationStatus.Resource.Version)
		}
		result = lastApi
		return nil
	}, retry.Attempts(10))
	return result, err
}

func (componentTestContext) defaultConfigOrExit() *rest.Config {

	kubeConfigLocation := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		log.Debugf("unable to load local kube config. Root cause: %v", err)
		if config, err2 := rest.InClusterConfig(); err2 != nil {
			log.Fatalf("unable to load kube config. Root cause: %v", err2)
		} else {
			kubeConfig = config
		}
	}
	return kubeConfig
}

func (componentTestContext) generateTestId(n int) string {

	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (componentTestContext) ShouldDeepEqual(actual interface{}, expected ...interface{}) string {
	return strings.Join(deep.Equal(actual, expected[0]), "\n")
}

func (componentTestContext) cleanUpApi(kymaClient *kyma.Clientset, api *kymaApi.Api, t *testing.T, allowMissing bool) {
	if api == nil {
		return
	}
	err := kymaClient.GatewayV1alpha2().Apis(namespace).Delete(api.Name, &metav1.DeleteOptions{})
	if !allowMissing && err != nil {
		t.Fatalf("Cannot clean up API %s: %s", api.Name, err)
	}
}
