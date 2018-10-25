package component

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/v1alpha2"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma.cx/clientset/versioned"
	istioNetApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/networking.istio.io/v1alpha3"
	istioNet "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned"
	istioAuthApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/authentication.istio.io/v1alpha1"
	istioAuth "github.com/kyma-project/kyma/components/api-controller/pkg/clients/authentication.istio.io/clientset/versioned"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/go-test/deep"
	"github.com/avast/retry-go"
	"errors"
)

const (
	namespace                                      = "kyma-system"
	testIdLength                                   = 8
	domainNameEnv                                  = "DOMAIN_NAME"
	apiSecurityDisabled                ApiSecurity = false
)

type ApiSecurity bool

func TestSpec(t *testing.T) {

	domainName := os.Getenv(domainNameEnv)
	if domainName == "" {
		t.Fatal("Domain name not set.")
	}

	testId := generateTestId(testIdLength)

	log.Infof("Running test: %s", testId)

	kubeConfig := defaultConfigOrExit()

	var lastApi *kymaApi.Api
	var lastVs *istioNetApi.VirtualService
	var lastPolicy *istioAuthApi.Policy

	suiteFinished := false

	Convey("API Controller should", t, func() {

		kymaClient, err := kyma.NewForConfig(kubeConfig)
		if err != nil {
			log.Fatalf("can create kymaClient clientset. Root cause: %v", err)
		}
		istioNetClient, err := istioNet.NewForConfig(kubeConfig)
		if err != nil {
			log.Fatalf("can create istioNet clientset. Root cause: %v", err)
		}
		istioAuthClient, err := istioAuth.NewForConfig(kubeConfig)
		if err != nil {
			log.Fatalf("can create istioAuth clientset. Root cause: %v", err)
		}

		suiteFinished = false

		Convey("create API with authentication disabled", func() {

			api := apiFor(testId, domainName, apiSecurityDisabled, true)

			lastApi, err = kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			lastApi, err = awaitApiUpdated(kymaClient, lastApi, true, false)
			So(err, ShouldBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
			So(lastApi.Spec, ShouldDeepEqual, api.Spec)

			lastVs, err = istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastApi.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := virtualServiceFor(testId, domainName)
			So(err, ShouldBeNil)
			So(lastVs.Spec, ShouldDeepEqual, expectedVs)
		})

		Convey("update API with hostname without domain", func() {

			api := *lastApi
			api.Spec.Hostname = hostnameFor(testId, domainName, false)

			lastApi, err = kymaClient.GatewayV1alpha2().Apis(namespace).Update(&api)
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			time.Sleep(5 * time.Second)

			lastApi, err = awaitApiUpdated(kymaClient, lastApi, false, false)
			So(err, ShouldBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
			So(lastApi.Spec, ShouldDeepEqual, api.Spec)

			lastVs, err = istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastApi.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := virtualServiceFor(testId, domainName)
			So(err, ShouldBeNil)
			So(lastVs.Spec, ShouldDeepEqual, expectedVs)
		})

		Convey("do not update API with wrong domain", func() {

			api := apiFor(testId, domainName+"x", apiSecurityDisabled, true)
			api.ResourceVersion = lastApi.ResourceVersion

			_, err = kymaClient.GatewayV1alpha2().Apis(namespace).Update(api)
			So(err, ShouldNotBeNil)
		})

		Convey("update API with default jwt configuration to enable authentication", func() {

			api := *lastApi
			authEnabled := true
			api.Spec.AuthenticationEnabled = &authEnabled
			api.Spec.Hostname = hostnameFor(testId, domainName, true)

			lastApi, err = kymaClient.GatewayV1alpha2().Apis(namespace).Update(&api)

			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			lastApi, err = awaitApiUpdated(kymaClient, lastApi, false, true)
			So(err, ShouldBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
			So(lastApi.Spec, ShouldDeepEqual, api.Spec)

			vs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastApi.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := virtualServiceFor(testId, domainName)
			So(err, ShouldBeNil)
			So(vs.Spec, ShouldDeepEqual, expectedVs)

			lastPolicy, err = istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastApi.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			expectedPolicy := policyFor(testId, fmt.Sprintf("https://dex.%s", domainName))
			So(err, ShouldBeNil)
			So(lastPolicy.Spec, ShouldDeepEqual, expectedPolicy)
		})

		Convey("update API to disable authentication", func() {

			api := *lastApi
			authEnabled := false
			api.Spec.AuthenticationEnabled = &authEnabled

			lastApi, err = kymaClient.GatewayV1alpha2().Apis(namespace).Update(&api)
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			lastApi, err = awaitApiUpdated(kymaClient, lastApi, false, true)
			So(err, ShouldBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
			So(lastApi.Spec, ShouldDeepEqual, api.Spec)
			So(lastApi.Status.AuthenticationStatus.Resource.Uid, ShouldBeEmpty)

			_, err := istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastPolicy.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)
		})

		Convey("update API with custom jwt configuration", func() {

			api := *lastApi
			setCustomJwtAuthenticationConfig(&api)

			lastApi, err = kymaClient.GatewayV1alpha2().Apis(namespace).Update(&api)
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			lastApi, err = awaitApiUpdated(kymaClient, lastApi, false, true)
			So(err, ShouldBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
			So(lastApi.Spec, ShouldDeepEqual, api.Spec)

			lastPolicy, err = istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastApi.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			expectedPolicy := policyFor(testId, api.Spec.Authentication[0].Jwt.Issuer)
			So(err, ShouldBeNil)
			So(lastPolicy.Spec, ShouldDeepEqual, expectedPolicy)
		})

		Convey("delete API", func() {

			suiteFinished = true
			if lastApi == nil {
				t.Fatal("Precondition failed - last API not set")
			}

			err := kymaClient.GatewayV1alpha2().Apis(namespace).Delete(lastApi.Name, &metav1.DeleteOptions{})
			So(err, ShouldBeNil)

			time.Sleep(5 * time.Second)

			_, err = kymaClient.GatewayV1alpha2().Apis(namespace).Get(lastApi.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)

			_, err = istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastPolicy.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)

			_, err = istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastVs.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)
		})
	})
}

func apiFor(testId, domainName string, secured ApiSecurity, hostWithDomain bool) *kymaApi.Api {

	return &kymaApi.Api{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("sample-app-api-%s", testId),
		},
		Spec: kymaApi.ApiSpec{
			Hostname: hostnameFor(testId, domainName, hostWithDomain),
			Service: kymaApi.Service{
				Name: fmt.Sprintf("sample-app-svc-%s", testId),
				Port: 80,
			},
			AuthenticationEnabled: (*bool)(&secured),
			Authentication:        []kymaApi.AuthenticationRule{},
		},
	}
}

func virtualServiceFor(testId, domainName string) *istioNetApi.VirtualServiceSpec {
	return &istioNetApi.VirtualServiceSpec{
		Hosts: []string{testId+"."+domainName},
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

func policyFor(testId, issuer string) *istioAuthApi.PolicySpec {
	return &istioAuthApi.PolicySpec{
		Targets: istioAuthApi.Targets{
			{Name: fmt.Sprintf("sample-app-svc-%s", testId)},
		},
		PrincipalBinding: istioAuthApi.UseOrigin,
		Origins: istioAuthApi.Origins{
			{
				Jwt: &istioAuthApi.Jwt{
					Issuer: issuer,
					JwksUri: "http://dex-service.kyma-system.svc.cluster.local:5556/keys",
				},
			},
		},
	}
}

func setCustomJwtAuthenticationConfig(api *kymaApi.Api) {
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

func hostnameFor(testId, domainName string, hostWithDomain bool) string {
	if hostWithDomain {
		return fmt.Sprintf("%s.%s", testId, domainName)
	}
	return testId
}

func awaitApiUpdated(iface *kyma.Clientset, api *kymaApi.Api, vsChanged, policyChanged bool) (*kymaApi.Api, error) {
	var result *kymaApi.Api
	err := retry.Do(func() error {
		lastApi, err := iface.GatewayV1alpha2().Apis(namespace).Get(api.Name, metav1.GetOptions{})

		if err != nil {
			return err
		}
		if vsChanged && (lastApi.Status.VirtualServiceStatus.Code != 2 || lastApi.Status.VirtualServiceStatus.Resource.Version == api.Status.VirtualServiceStatus.Resource.Version) {
			return errors.New("virtual service not created")
		}
		if policyChanged && (lastApi.Status.AuthenticationStatus.Code != 2 || lastApi.Status.AuthenticationStatus.Resource.Version == api.Status.AuthenticationStatus.Resource.Version) {
			return errors.New("virtual service not created")
		}
		result = lastApi
		return nil
	}, retry.Attempts(10))
	return result, err
}

func defaultConfigOrExit() *rest.Config {

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

func generateTestId(n int) string {

	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func ShouldDeepEqual(actual interface{}, expected ...interface{}) string {
	return strings.Join(deep.Equal(actual, expected[0]), "\n")
}
