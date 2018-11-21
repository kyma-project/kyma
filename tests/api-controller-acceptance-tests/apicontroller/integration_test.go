package apicontroller

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type integrationTestContext struct{}

func TestIntegrationSpec(t *testing.T) {

	domainName := os.Getenv(domainNameEnv)
	if domainName == "" {
		t.Fatal("Domain name not set.")
	}

	ctx := integrationTestContext{}
	testId := ctx.generateTestId(testIdLength)

	log.Infof("Running test: %s", testId)

	httpClient, err := ctx.newHttpClient(testId, domainName)
	if err != nil {
		t.Fatalf("Error while creating HTTP client. Root cause: %v", err)
	}

	kubeConfig := ctx.defaultConfigOrExit()
	k8sInterface := ctx.k8sInterfaceOrExit(kubeConfig)

	t.Logf("Set up...")
	fixture := setUpOrExit(k8sInterface, namespace, testId)

	var lastApi *kymaApi.Api

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

			api := ctx.apiFor(testId, domainName, fixture.SampleAppService, apiSecurityDisabled, true)

			lastApi, err = kymaInterface.GatewayV1alpha2().Apis(namespace).Create(api)
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			ctx.validateApiNotSecured(httpClient, lastApi.Spec.Hostname)
			lastApi, err = kymaInterface.GatewayV1alpha2().Apis(namespace).Get(lastApi.Name, metav1.GetOptions{})
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
		})

		Convey("update API with custom jwt configuration", func() {

			api := *lastApi
			ctx.setCustomJwtAuthenticationConfig(&api)

			lastApi, err = kymaInterface.GatewayV1alpha2().Apis(namespace).Update(&api)
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)

			ctx.validateApiSecured(httpClient, lastApi)
			lastApi, err = kymaInterface.GatewayV1alpha2().Apis(namespace).Get(lastApi.Name, metav1.GetOptions{})
			So(err, ShouldBeNil)
			So(lastApi, ShouldNotBeNil)
			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
		})

		Convey("delete API", func() {

			suiteFinished = true
			ctx.checkPreconditions(lastApi, t)

			err := kymaInterface.GatewayV1alpha2().Apis(namespace).Delete(lastApi.Name, &metav1.DeleteOptions{})
			So(err, ShouldBeNil)

			_, err = kymaInterface.GatewayV1alpha2().Apis(namespace).Get(lastApi.Name, metav1.GetOptions{})
			So(err, ShouldNotBeNil)
		})
	})
}

func (ctx integrationTestContext) apiFor(testId, domainName string, svc *apiv1.Service, secured ApiSecurity, hostWithDomain bool) *kymaApi.Api {

	return &kymaApi.Api{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("sample-app-api-%s", testId),
		},
		Spec: kymaApi.ApiSpec{
			Hostname: ctx.hostnameFor(testId, domainName, hostWithDomain),
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

func (integrationTestContext) checkPreconditions(lastApi *kymaApi.Api, t *testing.T) {
	if lastApi == nil {
		t.Fatal("Precondition failed - last API not set")
	}
}

func (integrationTestContext) hostnameFor(testId, domainName string, hostWithDomain bool) string {
	if hostWithDomain {
		return fmt.Sprintf("%s.%s", testId, domainName)
	}
	return testId
}

func (ctx integrationTestContext) validateApiSecured(httpClient *http.Client, api *kymaApi.Api) {

	response, err := ctx.withRetries(maxRetries, minimalNumberOfCorrectResults, func() (*http.Response, error) {
		return httpClient.Get(fmt.Sprintf("https://%s", api.Spec.Hostname))
	}, ctx.httpUnauthorizedPredicate)

	So(err, ShouldBeNil)
	So(response.StatusCode, ShouldEqual, http.StatusUnauthorized)
}

func (ctx integrationTestContext) validateApiNotSecured(httpClient *http.Client, hostname string) {

	response, err := ctx.withRetries(maxRetries, minimalNumberOfCorrectResults, func() (*http.Response, error) {
		return httpClient.Get(fmt.Sprintf("https://%s", hostname))
	}, ctx.httpOkPredicate)

	So(err, ShouldBeNil)
	So(response.StatusCode, ShouldEqual, http.StatusOK)
}

func (integrationTestContext) withRetries(maxRetries, minCorrect int, httpCall func() (*http.Response, error), shouldRetryPredicate func(*http.Response) bool) (*http.Response, error) {

	var response *http.Response
	var err error

	count := 0
	retry := true
	for retryNo := 0; retry; retryNo++ {

		log.Debugf("[%d / %d] Retrying...", retryNo, maxRetries)
		response, err = httpCall()

		if err != nil {
			log.Errorf("[%d / %d] Got error: %s", retryNo, maxRetries, err)
			count = 0
		} else if shouldRetryPredicate(response) {
			log.Errorf("[%d / %d] Got response: %s", retryNo, maxRetries, response.Status)
			count = 0
		} else {
			log.Infof("Got expected response %d in a row.", count+1)
			if count++; count == minCorrect {
				log.Infof("Reached minimal number of expected responses in a row. Do not need to retry anymore.")
				retry = false
			}
		}

		if retry {

			if retryNo >= maxRetries {
				// do not retry anymore
				log.Infof("No more retries (max retries exceeded).")
				retry = false
			} else {
				time.Sleep(retrySleep)
			}
		}
	}

	return response, err
}

func (integrationTestContext) httpOkPredicate(response *http.Response) bool {
	return response.StatusCode < 200 || response.StatusCode > 299
}

func (integrationTestContext) httpUnauthorizedPredicate(response *http.Response) bool {
	return response.StatusCode != 401
}

func (integrationTestContext) defaultConfigOrExit() *rest.Config {

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

func (integrationTestContext) k8sInterfaceOrExit(kubeConfig *rest.Config) kubernetes.Interface {

	k8sInterface, k8sErr := kubernetes.NewForConfig(kubeConfig)
	if k8sErr != nil {
		log.Fatalf("can create k8s clientset. Root cause: %v", k8sErr)
	}
	return k8sInterface
}

func (integrationTestContext) generateTestId(n int) string {

	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (tctx integrationTestContext) newHttpClient(testId, domainName string) (*http.Client, error) {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	ingressGatewayControllerAddr, err := net.LookupHost(ingressGatewayControllerServiceURL)
	if err != nil {
		log.Warnf("Unable to resolve host '%s' (if you are running this test from outside of Kyma ignore this log). Root cause: %v", ingressGatewayControllerServiceURL, err)
		minikubeIp := tctx.tryToGetMinikubeIp()
		if minikubeIp == "" {
			return nil, err
		}
		ingressGatewayControllerAddr = []string{minikubeIp}
	}
	log.Infof("Ingress controller address: '%s'", ingressGatewayControllerAddr)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Changes request destination address to ingressGateway internal cluster address for requests to sample app service.
			hostname := tctx.hostnameFor(testId, domainName, true)
			if strings.HasPrefix(addr, hostname) {
				addr = strings.Replace(addr, hostname, ingressGatewayControllerAddr[0], 1)
			}
			dialer := net.Dialer{}
			return dialer.DialContext(ctx, network, addr)
		},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}

	return client, nil
}

func (integrationTestContext) tryToGetMinikubeIp() string {
	mipCmd := exec.Command("minikube", "ip")
	if mipOut, err := mipCmd.Output(); err != nil {
		log.Warnf("Error while getting minikube IP (ignore this message if you are running this test inside Kyma). Root cause: %s", err)
		return ""
	} else {
		return strings.Trim(string(mipOut), "\n")
	}
}
