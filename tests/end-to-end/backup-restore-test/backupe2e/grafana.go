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
// 2. https://grafana.kyma.local/api/search?query=Lambda%20Dashboard

package backupe2e

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	grafanaNS              = "kyma-system"
	grafanaPodName         = "monitoring-grafana-0"
	grafanaServiceName     = "monitoring-grafana"
	grafanaStatefulsetName = "monitoring-grafana"
	grafanaPvcName         = "monitoring-grafana"
	adminUserSecretName    = "admin-user"
	containerName          = "grafana"
	grafanaLabelSelector   = "app=monitoring-grafana"
)

var (
	dashboards = make(map[string]dashboard)
)

type grafanaTest struct {
	grafanaName  string
	uuid         string
	coreClient   *kubernetes.Clientset
	beforeBackup bool
	grafana
}

type grafana struct {
	url        string
	oauthUrl   string
	loginForm  url.Values
	httpClient *http.Client
}

type dashboard struct {
	title string
	url   string
}

func NewGrafanaTest() (*grafanaTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return &grafanaTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return &grafanaTest{}, err
	}

	return &grafanaTest{
		coreClient:   coreClient,
		grafanaName:  "grafana",
		uuid:         uuid.New().String(),
		beforeBackup: true,
		grafana:      grafana{loginForm: url.Values{}},
	}, nil
}

func (t *grafanaTest) CreateResources(namespace string) {
	// There is not need to be implemented for this test.
}

func (t *grafanaTest) DeleteResources(namespace string) {
	// It needs to be implemented for this test.
	err := t.waitForPodGrafana(1 * time.Minute)
	So(err, ShouldBeNil)

	err = t.deleteServices(grafanaNS, grafanaServiceName, grafanaLabelSelector)
	So(err, ShouldBeNil)

	err = t.deleteStatefulset(grafanaNS, grafanaStatefulsetName)
	So(err, ShouldBeNil)

	err = t.deletePod(grafanaNS, grafanaPodName, grafanaLabelSelector)
	So(err, ShouldBeNil)

	err = t.deletePVC(grafanaNS, grafanaPvcName, grafanaLabelSelector)
	So(err, ShouldBeNil)

	//err1 := t.waitForPodGrafana(2 * time.Minute)
	//So(err1, ShouldBeError) // An error is expected.

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

	defer func() {
		err := apiSearchFolders.Body.Close()
		So(err, ShouldBeNil)
	}()
	dataBody, err := ioutil.ReadAll(apiSearchFolders.Body)
	So(err, ShouldBeNil)

	// Query to api
	if t.beforeBackup {
		t.beforeBackup = false
		dashboardFolders := make([]map[string]interface{}, 0)
		err = json.Unmarshal(dataBody, &dashboardFolders)
		So(err, ShouldBeNil)
		So(len(dashboardFolders), ShouldNotEqual, 0)

		for _, folder := range dashboardFolders {
			// http request to every dashboard
			domain = fmt.Sprintf("%s%s", t.url, folder["url"])
			dashResp, err := t.requestToGrafana(domain, "GET", nil, strings.NewReader(t.loginForm.Encode()), cookie)
			So(err, ShouldBeNil)
			So(dashResp.StatusCode, ShouldEqual, http.StatusOK)
			title := fmt.Sprintf("%s", folder["title"])
			dashboards[title] = dashboard{title: title, url: fmt.Sprintf("%s", folder["url"])}
		}
	} else {
		// iterate over the list of dashboards found before the backup (first time the test runs)
		for _, dash := range dashboards {
			domain = fmt.Sprintf("%s%s", t.url, dash.url)
			resp, err := t.requestToGrafana(domain, "GET", nil, strings.NewReader(t.loginForm.Encode()), cookie)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
		}
	}

}

func (t *grafanaTest) deleteServices(namespace, serviceName, labelSelector string) error {
	deletePolicy := metav1.DeletePropagationForeground

	serviceList, err := t.coreClient.CoreV1().Services(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return err
	}

	for _, service := range serviceList.Items {
		if service.Name == serviceName {
			err := t.coreClient.CoreV1().Services(namespace).Delete(serviceName, &metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil

}

func (t *grafanaTest) deleteStatefulset(namespace, statefulsetName string) error {
	deletePolicy := metav1.DeletePropagationForeground

	collection := t.coreClient.AppsV1().StatefulSets(namespace)
	err := collection.Delete(statefulsetName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return err
	}

	return nil
}

func (t *grafanaTest) deletePod(namespace, podName, labelSelector string) error {
	podList, err := t.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		if pod.Name == podName {
			err = t.coreClient.CoreV1().Pods(namespace).Delete(podName, &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}

	return nil

}

func (t *grafanaTest) deletePVC(namespace, pvcName, labelSelector string) error {
	pvcList, err := t.coreClient.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return err
	}

	for _, pvc := range pvcList.Items {
		if pvc.Name == pvcName {
			err = t.coreClient.CoreV1().PersistentVolumeClaims(namespace).Delete(pvcName, &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}

	return nil

}

func (t *grafanaTest) getGrafanaAndDexAuth() *http.Response {
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
	dexAuthLocal, err := t.requestToGrafana(domain, "POST", nil, strings.NewReader(t.loginForm.Encode()), nil)
	So(err, ShouldBeNil)
	So(dexAuthLocal.StatusCode, ShouldEqual, http.StatusOK)

	return dexAuthLocal
}

func (g *grafana) requestToGrafana(domain, method string, params url.Values, formData io.Reader, cookies []*http.Cookie) (*http.Response, error) {
	u, err := url.Parse(domain)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing domain: %s", domain)
	}

	if params != nil {
		u.RawQuery = params.Encode()
	}
	req, err := http.NewRequest(method, u.String(), formData)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating new request")
	}

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
			t.loginForm.Set("login", string(value))
		case "password":
			if string(value) == "" {
				return fmt.Errorf("No password found in secret '%s'\n", adminUserSecretName)
			}
			t.loginForm.Set("password", string(value))
		}

	}

	return nil

}

func getHttpClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	cookieJar, err := cookiejar.New(nil)
	So(err, ShouldBeNil)

	return &http.Client{Timeout: 15 * time.Second, Transport: tr, Jar: cookieJar}
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

			// If Pod condition is not ready the for will continue until timeout
			if len(pod.Status.Conditions) > 0 {
				conditions := pod.Status.Conditions
				for _, cond := range conditions {
					if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
						return nil
					}
				}
			}

			// Succeeded or Failed or Unknoen are taken as a error
			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
				return fmt.Errorf("Grafana in state %v: \n%+v", pod.Status.Phase, pod)
			}
		}
	}
}
