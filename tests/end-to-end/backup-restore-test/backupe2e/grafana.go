// Before backup and after restore test.
// This Test:
// - Verify if the grafana pod is running.
// - Login in the grafana ui.
// - Gets Dex's credentials from Kubernetes Secret.
// - Authenticates through Dex form.
// - Execute http request to the grafana api to get the list of folders and dashboards.
// - Execute a http request to every dashboard through the grafana api , expecting a 200 http status code.
// Login and Auth (Grafana and Dex)
// 1. https://grafana.kyma.local/login (GET)
// 2. https://grafana.kyma.local/login/generic_oauth (GET)
// 3. https://dex.kyma.local/auth (GET)
// 4. https://dex.kyma.local/auth/local (POST)
// Api:
// 1. https://grafana.kyma.local/api/search?folderIds=0
// 2. https://grafana.kyma.local/d/opc3b8Tyik/lambda-dashboard

package backupe2e

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"fmt"
	"net/url"
	"net/http"
	"net/http/cookiejar"
	"time"
	"strings"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	grafanaNS              string = "kyma-system"
	grafanaPodName         string = "monitoring-grafana-0"
	grafanaServiceName     string = "monitoring-grafana"
	grafanaStatefulsetName string = "monitoring-grafana"
	adminUserSecretName    string = "admin-user"
	containerName          string = "grafana"
	grafanaLabelSelector   string = "app=monitoring-grafana"
)

var (
	apiRequestsStatus = map[string]int{
		"Lambda Dashboard":        0,
		"Etcd":                    0,
		"Connector-Service":       0,
		"Istio Mesh Dashboard":    0,
		"Istio Service Dashboard": 0,
	}
)

type grafanaTest struct {
	grafanaName string
	uuid        string
	coreClient  *kubernetes.Clientset
	grafana
}

type grafana struct {
	url        string
	oauthUrl   string
	user       string
	password   string
	httpClient *http.Client
}

func NewGrafanaTest() (*grafanaTest, error) {

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return &grafanaTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return &grafanaTest{}, err
	}

	return &grafanaTest{
		coreClient:  coreClient,
		grafanaName: "grafana",
		uuid:        uuid.New().String(),
		grafana:     grafana{},
	}, nil
}

func (t *grafanaTest) CreateResources(namespace string) {
	// There is not need to be implemented for this test.
}

func (t *grafanaTest) DeleteResources() {
	// It needs to be implemented for this test.
	err := t.waitForPodGrafana(1 * time.Minute)
	So(err, ShouldBeNil)

	deletePolicy := metav1.DeletePropagationForeground

	serviceList, err := t.coreClient.CoreV1().Services(grafanaNS).List(metav1.ListOptions{LabelSelector: grafanaLabelSelector,})
	So(err, ShouldBeNil)

	for _, service := range serviceList.Items {
		if service.Name == grafanaServiceName {
			// Delete Service
			err = t.coreClient.CoreV1().Services(grafanaNS).Delete(grafanaServiceName, &metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			})
			So(err, ShouldBeNil)
		}
	}

	collection := t.coreClient.AppsV1().StatefulSets(grafanaNS)
	err = collection.Delete(grafanaStatefulsetName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	So(err, ShouldBeNil)

	podList, err := t.coreClient.CoreV1().Pods(grafanaNS).List(metav1.ListOptions{LabelSelector: grafanaLabelSelector,})
	So(err, ShouldBeNil)

	for _, pod := range podList.Items {
		if pod.Name == grafanaPodName {
			// Delete Pod
			err = t.coreClient.CoreV1().Pods(grafanaNS).Delete(grafanaPodName, &metav1.DeleteOptions{})
			So(err, ShouldBeNil)
		}
	}

	err = t.waitForPodGrafana(2 * time.Minute)
	So(err, ShouldBeError)

}

func (t *grafanaTest) TestResources(namespace string) {
	err := t.waitForPodGrafana(5 * time.Minute)
	So(err, ShouldBeNil)

	err = t.getGrafana()
	So(err, ShouldBeNil)

	dexAuthLocal := t.getGrafanaAndDexAuth()

	// request for retrieving folders and dashboards of the general folder
	// /api/search?folderIds=0
	domain := fmt.Sprintf("%s%s", dexAuthLocal.Request.URL.String(), "api/search")
	params := url.Values{}
	params.Set("folderIds", "0")
	cookie := dexAuthLocal.Request.Cookies()
	apiSearchFolders, err := t.requestToGrafana(domain, "GET", params, nil, cookie)
	So(err, ShouldBeNil)
	So(apiSearchFolders.StatusCode, ShouldEqual, http.StatusOK)

	defer apiSearchFolders.Body.Close()
	dataBody, err := ioutil.ReadAll(apiSearchFolders.Body)
	So(err, ShouldBeNil)

	// Query to api
	dashboardFolders := make([]map[string]interface{}, 0)
	err = json.Unmarshal(dataBody, &dashboardFolders)
	So(err, ShouldBeNil)
	So(len(dashboardFolders), ShouldNotEqual, 0)

	for _, folder := range dashboardFolders {
		// http request to every dashboard
		domain = fmt.Sprintf("%s%s", t.url, folder["url"])
		formData := url.Values{}
		formData.Set("login", t.user)
		formData.Set("password", t.password)
		dashboard, err := t.requestToGrafana(domain, "GET", nil, strings.NewReader(formData.Encode()), cookie)
		So(err, ShouldBeNil)
		So(dashboard.StatusCode, ShouldEqual, http.StatusOK)
		title := folder["title"].(string)
		switch title {
		case "Lambda Dashboard":
			apiRequestsStatus[title] = dashboard.StatusCode
		case "Etcd":
			apiRequestsStatus[title] = dashboard.StatusCode
		case "Connector-Service":
			apiRequestsStatus[title] = dashboard.StatusCode
		case "Istio Mesh Dashboard":
			apiRequestsStatus[title] = dashboard.StatusCode
		case "Istio Service Dashboard":

		}
	}

	t.testResult()

}

func (t *grafanaTest) testResult() {

	for _, status := range apiRequestsStatus {
		So(status, ShouldEqual, http.StatusOK)
	}

}

func (t *grafanaTest) getGrafanaAndDexAuth() (*http.Response) {
	//  /login
	domain := fmt.Sprintf("%s%s", t.url, "/login")
	grafLogin, err := t.requestToGrafana(domain, "GET", nil, nil, nil)
	So(err, ShouldBeNil)
	fmt.Println(grafLogin.StatusCode)
	So(grafLogin.StatusCode, ShouldEqual, http.StatusOK)

	//  /login/generic_oauth
	domain = fmt.Sprintf("%s%s", domain, "/generic_oauth")
	genericOauth, err := t.requestToGrafana(domain, "GET", nil, nil, nil)
	So(err, ShouldBeNil)
	So(genericOauth.StatusCode, ShouldEqual, http.StatusOK)

	// /auth
	domain = genericOauth.Request.Referer()
	dexAuth, err := t.requestToGrafana(domain, "GET", nil, nil, nil)
	So(err, ShouldBeNil)
	So(dexAuth.StatusCode, ShouldEqual, http.StatusOK)

	// /auth/local
	domain = dexAuth.Request.URL.String()
	formData := url.Values{}
	formData.Set("login", t.user)
	formData.Set("password", t.password)
	dexAuthLocal, err := t.requestToGrafana(domain, "POST", nil, strings.NewReader(formData.Encode()), nil)
	So(err, ShouldBeNil)
	So(dexAuthLocal.StatusCode, ShouldEqual, http.StatusOK)

	return dexAuthLocal
}

func (g *grafana) requestToGrafana(domain, method string, params url.Values, formData io.Reader, cookies []*http.Cookie) (*http.Response, error) {
	u, _ := url.Parse(domain)

	if params != nil {
		u.RawQuery = params.Encode()
	}
	req, err := http.NewRequest(method, u.String(), formData)

	// Cookies
	if len(cookies) > 0 {
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}

	// Header
	req.Header.Set("Accept", "application/json")
	req.Header.Add("Accept", "text/html")
	req.Header.Add("Accept", "text/plain")
	req.Header.Add("Accept", "application/xhtml+xml")
	req.Header.Add("Accept", "application/xml")
	req.Header.Set("User-Agent", "autograf")
	req.Header.Set("Connection", "keep-alive")

	switch method {
	case "POST":
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	case "GET":
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request to the url (%s) failed with '%s'\n", u, err)
	}
	return resp, err
}

func (t *grafanaTest) getGrafana() error {

	pod, err := t.coreClient.CoreV1().Pods(grafanaNS).Get(grafanaPodName, metav1.GetOptions{})
	So(strings.TrimSpace(string(pod.Status.Phase)), ShouldEqual, corev1.PodRunning)

	spec := pod.Spec
	containers := spec.Containers
	for _, container := range containers {
		if container.Name == containerName {
			envs := container.Env
			for _, envVar := range envs {
				switch envVar.Name {
				case "GF_AUTH_GENERIC_OAUTH_AUTH_URL":
					t.oauthUrl = envVar.Value
				case "GF_SERVER_ROOT_URL":
					t.url = strings.TrimSuffix(envVar.Value, "/")
				}
			}
		}
	}

	err = t.getCredentials()
	if err != nil {
		return err
	}

	t.httpClient = getHttpClient(true)

	return nil
}

func (t *grafanaTest) getCredentials() error {

	secret, err := t.coreClient.CoreV1().Secrets(grafanaNS).Get(adminUserSecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	data := secret.Data
	for key, value := range data {
		switch key {
		case "email":
			if string(value) == "" {
				return fmt.Errorf("No email found in secret '%s'\n", adminUserSecretName)
			}
			t.user = string(value)
		case "password":
			if string(value) == "" {
				return fmt.Errorf("No password found in secret '%s'\n", adminUserSecretName)
			}
			t.password = string(value)
		}

	}

	return nil

}

func getHttpClient(skipVerify bool) (*http.Client) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	cookieJar, _ := cookiejar.New(nil)

	return &http.Client{Timeout: 15 * time.Second, Transport: tr, Jar: cookieJar,}
}

func (t *grafanaTest) waitForPodGrafana(waitmax time.Duration) error {
	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-timeout:
			pod, err := t.coreClient.CoreV1().Pods(grafanaNS).Get(grafanaPodName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			return fmt.Errorf("Pod did not start within given time  %v: %+v", waitmax, pod)
		case <-tick:
			pod, err := t.coreClient.CoreV1().Pods(grafanaNS).Get(grafanaPodName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if pod.Status.Phase == corev1.PodRunning {
				return nil
			}
			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
				return fmt.Errorf("Grafana in state %v: \n%+v", pod.Status.Phase, pod)
			}
		}
	}
}
