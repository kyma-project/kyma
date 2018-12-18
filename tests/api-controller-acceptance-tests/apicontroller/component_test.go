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

			testID := ctx.generatetestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, apiSecurityDisabled, true)

			lastAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, lastAPI, t, false)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			lastAPI, err = ctx.awaitAPIChanged(kymaClient, lastAPI, true, false)
			So(err, ShouldBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)
			So(lastAPI.Spec, ctx.ShouldDeepEqual, api.Spec)

			lastVs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastAPI.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := ctx.virtualServiceFor(testID, domainName)
			So(err, ShouldBeNil)
			So(lastVs.Spec, ctx.ShouldDeepEqual, expectedVs)
		})

		Convey("create API with hostname without domain", func() {
			t.Log("create API with hostname without domain")

			testID := ctx.generatetestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, apiSecurityDisabled, false)

			lastAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, lastAPI, t, false)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			lastAPI, err = ctx.awaitAPIChanged(kymaClient, lastAPI, true, false)
			So(err, ShouldBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)
			So(lastAPI.Spec, ctx.ShouldDeepEqual, api.Spec)

			lastVs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastAPI.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := ctx.virtualServiceFor(testID, domainName)
			So(err, ShouldBeNil)
			So(lastVs.Spec, ctx.ShouldDeepEqual, expectedVs)
		})

		Convey("not create API with wrong domain", func() {
			t.Log("not create API with wrong domain")

			testID := ctx.generatetestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName+"x", apiSecurityDisabled, true)

			_, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			So(err, ShouldNotBeNil)
		})

		Convey("create API with default jwt configuration to enable authentication", func() {
			t.Log("create API with default jwt configuration to enable authentication")

			testID := ctx.generatetestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, apiSecurityDisabled, true)
			authEnabled := true
			api.Spec.AuthenticationEnabled = &authEnabled

			lastAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, lastAPI, t, false)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			lastAPI, err = ctx.awaitAPIChanged(kymaClient, lastAPI, true, true)
			So(err, ShouldBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)
			So(lastAPI.Spec, ctx.ShouldDeepEqual, api.Spec)

			vs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastAPI.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := ctx.virtualServiceFor(testID, domainName)
			So(err, ShouldBeNil)
			So(vs.Spec, ctx.ShouldDeepEqual, expectedVs)

			lastPolicy, err := istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastAPI.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			expectedPolicy := ctx.policyFor(testID, fmt.Sprintf("https://dex.%s", domainName))
			So(err, ShouldBeNil)
			So(lastPolicy.Spec, ctx.ShouldDeepEqual, expectedPolicy)
		})

		Convey("update API to disable authentication", func() {
			t.Log("update API to disable authentication")

			testID := ctx.generatetestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, apiSecurityEnabled, true)

			createdAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, createdAPI, t, false)
			So(err, ShouldBeNil)
			So(createdAPI, ShouldNotBeNil)
			So(createdAPI.ResourceVersion, ShouldNotBeEmpty)

			createdAPI, err = ctx.awaitAPIChanged(kymaClient, createdAPI, true, true)
			So(err, ShouldBeNil)
			So(createdAPI.ResourceVersion, ShouldNotBeEmpty)
			So(createdAPI.Spec, ctx.ShouldDeepEqual, api.Spec)

			authEnabled := false
			createdAPI.Spec.AuthenticationEnabled = &authEnabled

			updatedAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Update(createdAPI)
			So(err, ShouldBeNil)
			So(updatedAPI, ShouldNotBeNil)
			So(updatedAPI.ResourceVersion, ShouldNotBeEmpty)

			updatedAPI, err = ctx.awaitAPIChanged(kymaClient, updatedAPI, false, true)
			So(err, ShouldBeNil)
			So(updatedAPI.ResourceVersion, ShouldNotBeEmpty)
			So(updatedAPI.Spec, ctx.ShouldDeepEqual, createdAPI.Spec)
			So(updatedAPI.Status.AuthenticationStatus.Resource.Uid, ShouldBeEmpty)

			_, err = istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(createdAPI.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)
		})

		Convey("create API with custom jwt configuration", func() {
			t.Log("create API with custom jwt configuration")

			testID := ctx.generatetestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, apiSecurityDisabled, true)
			ctx.setCustomJwtAuthenticationConfig(api)

			lastAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, lastAPI, t, false)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			lastAPI, err = ctx.awaitAPIChanged(kymaClient, lastAPI, true, true)
			So(err, ShouldBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)
			So(lastAPI.Spec, ctx.ShouldDeepEqual, api.Spec)

			policy, err := istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastAPI.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			expectedPolicy := ctx.policyFor(testID, api.Spec.Authentication[0].Jwt.Issuer)
			So(err, ShouldBeNil)
			So(policy.Spec, ctx.ShouldDeepEqual, expectedPolicy)
		})

		Convey("delete API and all its related resources", func() {
			t.Log("delete API and all its related resources")

			testID := ctx.generatetestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, apiSecurityEnabled, true)

			lastAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, lastAPI, t, true)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			lastAPI, err = ctx.awaitAPIChanged(kymaClient, lastAPI, true, true)
			So(err, ShouldBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)
			So(lastAPI.Spec, ctx.ShouldDeepEqual, api.Spec)
			policy, err := istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastAPI.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			So(err, ShouldBeNil)
			vs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastAPI.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			So(err, ShouldBeNil)

			err = kymaClient.GatewayV1alpha2().Apis(namespace).Delete(lastAPI.Name, &metav1.DeleteOptions{})
			So(err, ShouldBeNil)

			time.Sleep(5 * time.Second)

			_, err = kymaClient.GatewayV1alpha2().Apis(namespace).Get(lastAPI.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)

			_, err = istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(policy.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)

			_, err = istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(vs.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)
		})
	})
}

func (ctx componentTestContext) apiFor(testID, domainName string, secured APISecurity, hostWithDomain bool) *kymaApi.Api {

	return &kymaApi.Api{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("sample-app-api-%s", testID),
		},
		Spec: kymaApi.ApiSpec{
			Hostname: ctx.hostnameFor(testID, domainName, hostWithDomain),
			Service: kymaApi.Service{
				Name: fmt.Sprintf("sample-app-svc-%s", testID),
				Port: 80,
			},
			AuthenticationEnabled: (*bool)(&secured),
			Authentication:        []kymaApi.AuthenticationRule{},
		},
	}
}

func (componentTestContext) virtualServiceFor(testID, domainName string) *istioNetApi.VirtualServiceSpec {
	return &istioNetApi.VirtualServiceSpec{
		Hosts:    []string{testID + "." + domainName},
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
							Host: fmt.Sprintf("sample-app-svc-%s.%s.svc.cluster.local", testID, namespace),
						},
					},
				},
			},
		},
	}
}

func (componentTestContext) policyFor(testID, issuer string) *istioAuthApi.PolicySpec {
	return &istioAuthApi.PolicySpec{
		Targets: istioAuthApi.Targets{
			{Name: fmt.Sprintf("sample-app-svc-%s", testID)},
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
	//jwksURI := "https://www.googleapis.com/oauth2/v3/certs"

	issuer := "https://accounts.google.com"
	jwksURI := "http://dex-service.kyma-system.svc.cluster.local:5556/keys"

	rules := []kymaApi.AuthenticationRule{
		{
			Type: kymaApi.JwtType,
			Jwt: kymaApi.JwtAuthentication{
				Issuer:  issuer,
				JwksUri: jwksURI,
			},
		},
	}

	secured := true
	if api.Spec.AuthenticationEnabled != nil && !(*api.Spec.AuthenticationEnabled) { // optional property, but if set earlier to false it will force auth disabled
		api.Spec.AuthenticationEnabled = &secured
	}
	api.Spec.Authentication = rules
}

func (componentTestContext) hostnameFor(testID, domainName string, hostWithDomain bool) string {
	if hostWithDomain {
		return fmt.Sprintf("%s.%s", testID, domainName)
	}
	return testID
}

func (componentTestContext) awaitAPIChanged(iface *kyma.Clientset, api *kymaApi.Api, vsChanged, policyChanged bool) (*kymaApi.Api, error) {
	var result *kymaApi.Api
	err := retry.Do(func() error {
		lastAPI, err := iface.GatewayV1alpha2().Apis(namespace).Get(api.Name, metav1.GetOptions{})

		if err != nil {
			return err
		}
		if vsChanged && lastAPI.Status.VirtualServiceStatus.Resource.Version == api.Status.VirtualServiceStatus.Resource.Version {
			return fmt.Errorf("VirtualService not created, old: %s, new: %s",
				api.Status.VirtualServiceStatus.Resource.Version,
				lastAPI.Status.VirtualServiceStatus.Resource.Version)
		}
		if policyChanged && lastAPI.Status.AuthenticationStatus.Resource.Version == api.Status.AuthenticationStatus.Resource.Version {
			return fmt.Errorf("policy not created, old: %s, new: %s",
				api.Status.AuthenticationStatus.Resource.Version,
				lastAPI.Status.AuthenticationStatus.Resource.Version)
		}
		result = lastAPI
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

func (componentTestContext) generatetestID(n int) string {

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

func (componentTestContext) cleanUpAPI(kymaClient *kyma.Clientset, api *kymaApi.Api, t *testing.T, allowMissing bool) {
	if api == nil {
		return
	}
	err := kymaClient.GatewayV1alpha2().Apis(namespace).Delete(api.Name, &metav1.DeleteOptions{})
	if !allowMissing && err != nil {
		t.Fatalf("Cannot clean up API %s: %s", api.Name, err)
	}
}
