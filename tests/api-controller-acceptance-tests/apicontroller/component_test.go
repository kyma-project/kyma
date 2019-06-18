package apicontroller

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/avast/retry-go"
	"github.com/go-test/deep"
	istioAuthApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/authentication.istio.io/v1alpha1"
	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	istioNetApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/networking.istio.io/v1alpha3"
	istioAuth "github.com/kyma-project/kyma/components/api-controller/pkg/clients/authentication.istio.io/clientset/versioned"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	istioNet "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned"
	. "github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type componentTestContext struct{}

func TestComponentSpec(t *testing.T) {

	domainName := os.Getenv(domainNameEnv)
	if domainName == "" {
		t.Fatal("Domain name not set.")
	}

	ctx := componentTestContext{}

	kymaClient := kyma.NewForConfigOrDie(kubeConfig)
	istioNetClient := istioNet.NewForConfigOrDie(kubeConfig)
	istioAuthClient := istioAuth.NewForConfigOrDie(kubeConfig)

	Convey("API Controller should", t, func() {

		Convey("create API with authentication disabled", func() {
			t.Log("create API with authentication disabled")

			testID := ctx.generateTestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, namespace, apiSecurityDisabled, true)

			lastAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, lastAPI, t, false, namespace)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			lastAPI, err = ctx.awaitAPIChanged(kymaClient, lastAPI, true, false, namespace)
			So(err, ShouldBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)
			So(lastAPI.Spec, ctx.ShouldDeepEqual, api.Spec)

			lastVs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastAPI.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := ctx.virtualServiceFor(testID, domainName, namespace)
			So(err, ShouldBeNil)
			So(lastVs.Spec, ctx.ShouldDeepEqual, expectedVs)
		})

		Convey("create API with hostname without domain", func() {
			t.Log("create API with hostname without domain")

			testID := ctx.generateTestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, namespace, apiSecurityDisabled, false)

			lastAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, lastAPI, t, false, namespace)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			lastAPI, err = ctx.awaitAPIChanged(kymaClient, lastAPI, true, false, namespace)
			So(err, ShouldBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)
			So(lastAPI.Spec, ctx.ShouldDeepEqual, api.Spec)

			lastVs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastAPI.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := ctx.virtualServiceFor(testID, domainName, namespace)
			So(err, ShouldBeNil)
			So(lastVs.Spec, ctx.ShouldDeepEqual, expectedVs)
		})

		Convey("not create API with wrong domain", func() {
			t.Log("not create API with wrong domain")

			testID := ctx.generateTestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName+"x", namespace, apiSecurityDisabled, true)

			_, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			So(err, ShouldNotBeNil)
		})

		Convey("create API with default jwt configuration to enable authentication", func() {
			t.Log("create API with default jwt configuration to enable authentication")

			testID := ctx.generateTestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, namespace, apiSecurityDisabled, true)
			authEnabled := true
			api.Spec.AuthenticationEnabled = &authEnabled

			lastAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, lastAPI, t, false, namespace)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			lastAPI, err = ctx.awaitAPIChanged(kymaClient, lastAPI, true, true, namespace)
			So(err, ShouldBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)
			So(lastAPI.Spec, ctx.ShouldDeepEqual, api.Spec)

			vs, err := istioNetClient.NetworkingV1alpha3().VirtualServices(namespace).Get(lastAPI.Status.VirtualServiceStatus.Resource.Name, metav1.GetOptions{})
			expectedVs := ctx.virtualServiceFor(testID, domainName, namespace)
			So(err, ShouldBeNil)
			So(vs.Spec, ctx.ShouldDeepEqual, expectedVs)

			lastPolicy, err := istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastAPI.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			expectedPolicy := ctx.policyFor(testID, fmt.Sprintf("https://dex.%s", domainName))
			So(err, ShouldBeNil)
			So(lastPolicy.Spec, ctx.ShouldDeepEqual, expectedPolicy)
		})

		Convey("update API to disable authentication", func() {
			t.Log("update API to disable authentication")

			testID := ctx.generateTestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, namespace, apiSecurityEnabled, true)

			createdAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, createdAPI, t, false, namespace)
			So(err, ShouldBeNil)
			So(createdAPI, ShouldNotBeNil)
			So(createdAPI.ResourceVersion, ShouldNotBeEmpty)

			createdAPI, err = ctx.awaitAPIChanged(kymaClient, createdAPI, true, true, namespace)
			So(err, ShouldBeNil)
			So(createdAPI.ResourceVersion, ShouldNotBeEmpty)
			So(createdAPI.Spec, ctx.ShouldDeepEqual, api.Spec)

			authEnabled := false
			createdAPI.Spec.AuthenticationEnabled = &authEnabled

			updatedAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Update(createdAPI)
			So(err, ShouldBeNil)
			So(updatedAPI, ShouldNotBeNil)
			So(updatedAPI.ResourceVersion, ShouldNotBeEmpty)

			updatedAPI, err = ctx.awaitAPIChanged(kymaClient, updatedAPI, false, true, namespace)
			So(err, ShouldBeNil)
			So(updatedAPI.ResourceVersion, ShouldNotBeEmpty)
			So(updatedAPI.Spec, ctx.ShouldDeepEqual, createdAPI.Spec)
			So(updatedAPI.Status.AuthenticationStatus.Resource.Uid, ShouldBeEmpty)

			_, err = istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(createdAPI.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)
		})

		Convey("create API with custom jwt configuration", func() {
			t.Log("create API with custom jwt configuration")

			testID := ctx.generateTestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, namespace, apiSecurityDisabled, true)
			ctx.setCustomJwtAuthenticationConfig(api)

			lastAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, lastAPI, t, false, namespace)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			lastAPI, err = ctx.awaitAPIChanged(kymaClient, lastAPI, true, true, namespace)
			So(err, ShouldBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)
			So(lastAPI.Spec, ctx.ShouldDeepEqual, api.Spec)

			policy, err := istioAuthClient.AuthenticationV1alpha1().Policies(namespace).Get(lastAPI.Status.AuthenticationStatus.Resource.Name, metav1.GetOptions{})
			expectedPolicy := ctx.policyFor(testID, api.Spec.Authentication[0].Jwt.Issuer)
			So(err, ShouldBeNil)
			So(policy.Spec, ctx.ShouldDeepEqual, expectedPolicy)
		})

		Convey("create API should not process the request if another API exists for target service", func() {
			t.Log("create API: duplicate API for a service")

			testService := "test-srv"

			id := ctx.generateTestID(testIDLength)
			t.Logf("Running test: %s", id)
			api := ctx.apiFor(id, domainName, namespace, apiSecurityEnabled, true)
			api.Spec.Service.Name = testService

			testedID := ctx.generateTestID(testIDLength)
			t.Logf("Running test: %s", testedID)
			testedApi := ctx.apiFor(testedID, domainName, namespace, apiSecurityEnabled, true)
			testedApi.Spec.Service.Name = testService

			_, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			So(err, ShouldBeNil)

			testedApi, err = kymaClient.GatewayV1alpha2().Apis(namespace).Create(testedApi)
			So(err, ShouldBeNil)
			So(testedApi, ShouldNotBeNil)
			So(testedApi.ResourceVersion, ShouldNotBeEmpty)

			err = retry.Do(func() error {

				var err error

				testedApi, err = kymaClient.GatewayV1alpha2().Apis(namespace).Get(testedApi.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}

				if !testedApi.Status.IsTargetServiceOccupied() {
					return errors.Errorf("Incorrect status: %d", testedApi.Status.ValidationStatus)
				}

				return nil

			})

			So(err, ShouldBeNil)
			So(testedApi.Status.IsTargetServiceOccupied(), ShouldBeTrue)
			So(testedApi.Status.AuthenticationStatus.IsEmpty(), ShouldBeTrue)
			So(testedApi.Status.VirtualServiceStatus.IsEmpty(), ShouldBeTrue)
			So(testedApi.Status.AuthenticationStatus.Resource.Name, ShouldBeEmpty)
			So(testedApi.Status.VirtualServiceStatus.Resource.Name, ShouldBeEmpty)
		})

		Convey("update API should not process the request if another API exists for updated target service", func() {
			t.Log("update API: duplicate API for a service")

			initialService := "unoccupiedService"
			testService := "test-srv"

			id := ctx.generateTestID(testIDLength)
			t.Logf("Running test: %s", id)
			api := ctx.apiFor(id, domainName, namespace, apiSecurityEnabled, true)
			api.Spec.Service.Name = testService

			testedID := ctx.generateTestID(testIDLength)
			t.Logf("Running test: %s", testedID)
			testedApi := ctx.apiFor(testedID, domainName, namespace, apiSecurityEnabled, true)
			testedApi.Spec.Service.Name = initialService

			_, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			So(err, ShouldBeNil)
			defer ctx.cleanUpAPI(kymaClient, api, t, false, namespace)

			originalTestedApi, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(testedApi)
			So(err, ShouldBeNil)
			defer ctx.cleanUpAPI(kymaClient, testedApi, t, false, namespace)

			originalTestedApi, err = ctx.awaitAPIChanged(kymaClient, originalTestedApi, true, true, namespace)
			So(err, ShouldBeNil)

			testedApi, err = ctx.awaitAPIChanged(kymaClient, originalTestedApi, false, false, namespace)
			So(err, ShouldBeNil)

			testedApi.Spec.Service.Name = testService

			testedApi, err = kymaClient.GatewayV1alpha2().Apis(namespace).Update(testedApi)
			So(err, ShouldBeNil)

			err = retry.Do(func() error {

				var err error

				testedApi, err = kymaClient.GatewayV1alpha2().Apis(namespace).Get(testedApi.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}

				if !testedApi.Status.IsTargetServiceOccupied() {
					return errors.Errorf("Incorrect status: %d", testedApi.Status.ValidationStatus)
				}

				return nil

			})

			So(err, ShouldBeNil)
			So(testedApi.Status.IsTargetServiceOccupied(), ShouldBeTrue)

			So(testedApi.Status.AuthenticationStatus.IsEmpty(), ShouldBeTrue)
			So(testedApi.Status.VirtualServiceStatus.IsEmpty(), ShouldBeTrue)

			So(testedApi.Status.AuthenticationStatus.Resource, ctx.ShouldDeepEqual, originalTestedApi.Status.AuthenticationStatus.Resource)
			So(testedApi.Status.VirtualServiceStatus.Resource, ctx.ShouldDeepEqual, originalTestedApi.Status.VirtualServiceStatus.Resource)

			testedApi.Spec.Service.Name = initialService

			_, err = kymaClient.GatewayV1alpha2().Apis(namespace).Update(testedApi)
			So(err, ShouldBeNil)

			err = retry.Do(func() error {

				var err error

				testedApi, err = kymaClient.GatewayV1alpha2().Apis(namespace).Get(testedApi.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}

				if !testedApi.Status.IsSuccessful() {
					return errors.Errorf("Incorrect status: %d", testedApi.Status.ValidationStatus)
				}

				return nil

			})

			So(err, ShouldBeNil)

			So(testedApi.Status.AuthenticationStatus.IsSuccessful(), ShouldBeTrue)
			So(testedApi.Status.VirtualServiceStatus.IsSuccessful(), ShouldBeTrue)

			So(testedApi.Status.AuthenticationStatus.Resource.Name, ctx.ShouldDeepEqual, originalTestedApi.Status.AuthenticationStatus.Resource.Name)
			So(testedApi.Status.AuthenticationStatus.Resource.Uid, ctx.ShouldDeepEqual, originalTestedApi.Status.AuthenticationStatus.Resource.Uid)
			So(testedApi.Status.VirtualServiceStatus.Resource.Name, ctx.ShouldDeepEqual, originalTestedApi.Status.VirtualServiceStatus.Resource.Name)
			So(testedApi.Status.VirtualServiceStatus.Resource.Uid, ctx.ShouldDeepEqual, originalTestedApi.Status.VirtualServiceStatus.Resource.Uid)

		})

		Convey("delete API and all its related resources", func() {
			t.Log("delete API and all its related resources")

			testID := ctx.generateTestID(testIDLength)
			t.Logf("Running test: %s", testID)
			api := ctx.apiFor(testID, domainName, namespace, apiSecurityEnabled, true)

			lastAPI, err := kymaClient.GatewayV1alpha2().Apis(namespace).Create(api)
			defer ctx.cleanUpAPI(kymaClient, lastAPI, t, true, namespace)
			So(err, ShouldBeNil)
			So(lastAPI, ShouldNotBeNil)
			So(lastAPI.ResourceVersion, ShouldNotBeEmpty)

			lastAPI, err = ctx.awaitAPIChanged(kymaClient, lastAPI, true, true, namespace)
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

func (ctx componentTestContext) apiFor(testID string, domainName string, namespace string, secured APISecurity, hostWithDomain bool) *kymaApi.Api {

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

func (componentTestContext) virtualServiceFor(testID string, domainName string, namespace string) *istioNetApi.VirtualServiceSpec {
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
				Route: []*istioNetApi.HTTPRouteDestination{
					{
						Destination: &istioNetApi.Destination{
							Host: fmt.Sprintf("sample-app-svc-%s.%s.svc.cluster.local", testID, namespace),
						},
					},
				},
				CorsPolicy: &istioNetApi.CorsPolicy{ //Default policy
					AllowOrigin:  []string{"*"},
					AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
					AllowHeaders: []string{"*"},
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
		Peers: istioAuthApi.Peers{
			&istioAuthApi.Peer{MTLS: struct{}{}},
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

func (componentTestContext) awaitAPIChanged(iface *kyma.Clientset, api *kymaApi.Api, vsChanged, policyChanged bool, namespace string) (*kymaApi.Api, error) {
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

func (componentTestContext) generateTestID(n int) string {

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

func (componentTestContext) cleanUpAPI(kymaClient *kyma.Clientset, api *kymaApi.Api, t *testing.T, allowMissing bool, namespace string) {
	if api == nil {
		return
	}
	err := kymaClient.GatewayV1alpha2().Apis(namespace).Delete(api.Name, &metav1.DeleteOptions{})
	if !allowMissing && err != nil {
		t.Fatalf("Cannot clean up API %s: %s", api.Name, err)
	}
}
