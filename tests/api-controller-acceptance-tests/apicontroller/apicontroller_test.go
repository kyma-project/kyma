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

	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/v1alpha2"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma.cx/clientset/versioned"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	namespace                               = "kyma-system"
	ingressControllerServiceURL             = "istio-ingress.istio-system.svc.cluster.local"
	testIdLength                            = 8
	maxRetries                              = 100
	retrySleep                              = 3 * time.Second
	domainNameEnv                           = "DOMAIN_NAME"
	apiSecurityEnabled          ApiSecurity = true
	apiSecurityDisabled         ApiSecurity = false
)

type ApiSecurity bool

func TestSpec(t *testing.T) {

	domainName := os.Getenv(domainNameEnv)
	if domainName == "" {
		t.Fatal("Domain name not set.")
	}

	testId := generateTestId(testIdLength)

	log.Infof("Running test: %s", testId)

	httpClient, err := newHttpClient(testId, domainName)
	if err != nil {
		t.Fatalf("Error while creating HTTP client. Root cause: %v", err)
	}

	kubeConfig := defaultConfigOrExit()
	k8sInterface := k8sInterfaceOrExit(kubeConfig)

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

			t.Log("Create...")

			api := apiFor(testId, domainName, fixture.SampleAppService, apiSecurityDisabled)

			createdApi, err := kymaInterface.GatewayV1alpha2().Apis(namespace).Create(api)

			So(err, ShouldBeNil)
			So(createdApi, ShouldNotBeNil)

			lastApi = createdApi

			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
		})

		Convey("validate created API", func() {

			t.Log("Validate created...")

			checkPreconditions(lastApi, t)
			validateApiNotSecured(httpClient, lastApi)
		})

		Convey("get API", func() {

			t.Log("Get...")

			gotApi, err := kymaInterface.GatewayV1alpha2().Apis(namespace).Get(lastApi.Name, metav1.GetOptions{})

			So(err, ShouldBeNil)
			So(gotApi, ShouldNotBeNil)

			lastApi = gotApi

			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
		})

		Convey("update API to enable authentication", func() {

			t.Log("Update...")

			api := apiFor(testId, domainName, fixture.SampleAppService, apiSecurityEnabled)
			api.ResourceVersion = lastApi.ResourceVersion

			updatedApi, err := kymaInterface.GatewayV1alpha2().Apis(namespace).Update(api)

			So(err, ShouldBeNil)
			So(updatedApi, ShouldNotBeNil)

			lastApi = updatedApi

			So(lastApi.ResourceVersion, ShouldNotBeEmpty)
		})

		Convey("validate updated API", func() {

			t.Log("Validate updated...")

			checkPreconditions(lastApi, t)
			validateApiSecured(httpClient, lastApi)
		})

		Convey("delete API", func() {

			t.Log("Delete...")
			suiteFinished = true

			checkPreconditions(lastApi, t)
			err := kymaInterface.GatewayV1alpha2().Apis(namespace).Delete(lastApi.Name, &metav1.DeleteOptions{})

			So(err, ShouldBeNil)
		})

		Convey("validate if API is deleted properly", func() {

			t.Log("Validate deleted...")

			checkPreconditions(lastApi, t)
			_, err := kymaInterface.GatewayV1alpha2().Apis(namespace).Get(lastApi.Name, metav1.GetOptions{})

			So(err, ShouldNotBeNil)
		})

	})
}

func apiFor(testId, domainName string, svc *apiv1.Service, secured ApiSecurity) *kymaApi.Api {

	auth := kymaApi.Authentication{}
	if secured {

		issuer := "https://accounts.google.com"
		jwksUri := "https://www.googleapis.com/oauth2/v3/certs"

		auth = kymaApi.Authentication{
			kymaApi.AuthenticationRule{
				Type: kymaApi.JwtType,
				Jwt: kymaApi.JwtAuthentication{
					Issuer:  issuer,
					JwksUri: jwksUri,
				},
			},
		}
	}

	return &kymaApi.Api{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("sample-app-api-%s", testId),
		},
		Spec: kymaApi.ApiSpec{
			Hostname: hostnameFor(testId, domainName),
			Service: kymaApi.Service{
				Name: svc.Name,
				Port: int(svc.Spec.Ports[0].Port),
			},
			Authentication: auth,
		},
	}
}

func checkPreconditions(lastApi *kymaApi.Api, t *testing.T) {
	if lastApi == nil {
		t.Fatal("Precondition failed - last API not set")
	}
}

func hostnameFor(testId, domainName string) string {
	return fmt.Sprintf("%s.%s", testId, domainName)
}

func validateApiSecured(httpClient *http.Client, api *kymaApi.Api) {

	response, err := withRetries(maxRetries, func() (*http.Response, error) {
		return httpClient.Get(fmt.Sprintf("https://%s/api.yaml", api.Spec.Hostname))
	}, httpUnauthorizedPredicate)

	So(err, ShouldBeNil)
	So(response.StatusCode, ShouldEqual, http.StatusUnauthorized)
}

func validateApiNotSecured(httpClient *http.Client, api *kymaApi.Api) {

	response, err := withRetries(maxRetries, func() (*http.Response, error) {
		return httpClient.Get(fmt.Sprintf("https://%s/api.yaml", api.Spec.Hostname))
	}, httpOkPredicate)

	So(err, ShouldBeNil)
	So(response.StatusCode, ShouldEqual, http.StatusOK)
}

func withRetries(maxRetries int, httpCall func() (*http.Response, error), shouldRetryPredicate func(*http.Response) bool) (*http.Response, error) {

	var response *http.Response
	var err error

	retry := true
	for retryNo := 0; retry; retryNo++ {

		log.Debugf("[%d / %d] Retrying...", retryNo, maxRetries)
		response, err = httpCall()

		if err != nil {
			log.Errorf("[%d / %d] Got error: %s", retryNo, maxRetries, err)
		} else if shouldRetryPredicate(response) {
			log.Errorf("[%d / %d] Got response: %s", retryNo, maxRetries, response.Status)
		} else {
			log.Infof("Got expected response - do not need retry anymore.")
			retry = false
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

func httpOkPredicate(response *http.Response) bool {
	return response.StatusCode < 200 || response.StatusCode > 299
}

func httpUnauthorizedPredicate(response *http.Response) bool {
	return response.StatusCode != 401
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

func k8sInterfaceOrExit(kubeConfig *rest.Config) kubernetes.Interface {

	k8sInterface, k8sErr := kubernetes.NewForConfig(kubeConfig)
	if k8sErr != nil {
		log.Fatalf("can create k8s clientset. Root cause: %v", k8sErr)
	}
	return k8sInterface
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

func newHttpClient(testId, domainName string) (*http.Client, error) {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	ingressControllerAddr, err := net.LookupHost(ingressControllerServiceURL)
	if err != nil {
		log.Warnf("Unable to resolve host '%s' (if you are running this test from outside of Kyma ignore this log). Root cause: %v", ingressControllerServiceURL, err)
		minikubeIp := tryToGetMinikubeIp()
		if minikubeIp == "" {
			return nil, err
		}
		ingressControllerAddr = []string{minikubeIp}
	}
	log.Infof("Ingress controller address: '%s'", ingressControllerAddr)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Changes request destination address to ingress internal cluster address for requests to sample app service.
			hostname := hostnameFor(testId, domainName)
			if strings.HasPrefix(addr, hostname) {
				addr = strings.Replace(addr, hostname, ingressControllerAddr[0], 1)
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

func tryToGetMinikubeIp() string {
	mipCmd := exec.Command("minikube", "ip")
	if mipOut, err := mipCmd.Output(); err != nil {
		log.Warnf("Error while getting minikube IP (ignore this message if you are running this test inside Kyma). Root cause: %s", err)
		return ""
	} else {
		return strings.Trim(string(mipOut), "\n")
	}
}
