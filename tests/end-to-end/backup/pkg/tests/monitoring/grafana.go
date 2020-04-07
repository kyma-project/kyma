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

package monitoring

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
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/config"
)

const (
	grafanaNS            = "kyma-system"
	adminUserSecretName  = "admin-user"
	containerName        = "grafana"
	grafanaLabelSelector = "app=grafana"
)

var (
	dashboards = make(map[string]dashboard)
)

type grafanaTest struct {
	grafanaName string
	coreClient  *kubernetes.Clientset
	httpClient  *http.Client
	grafana
	log logrus.FieldLogger
}

type grafana struct {
	url       string
	loginForm url.Values
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
		coreClient:  coreClient,
		grafanaName: "grafana",
		grafana:     grafana{loginForm: url.Values{}},
		log:         logrus.WithField("test", "grafana"),
	}, nil
}

func (gt *grafanaTest) CreateResources(_ *testing.T, namespace string) {
	// There is no need to be implemented for this test.
}

func (gt *grafanaTest) TestResources(t *testing.T, namespace string) {
	err := gt.waitForPodGrafana(5 * time.Minute)
	require.NoError(t, err)

	err = gt.getGrafana(t)
	require.NoError(t, err)

	dexAuthLocal := gt.getGrafanaAndDexAuth(t)

	// request for retrieving folders and dashboards of the general folder
	// /api/search?folderIds=0
	domain := fmt.Sprintf("%s%s", gt.url, "/api/search")
	params := url.Values{}
	params.Set("folderIds", "0")
	cookie := dexAuthLocal.Request.Cookies()
	apiSearchFolders, err := gt.requestToGrafana(domain, "GET", params, strings.NewReader(gt.loginForm.Encode()), cookie)
	require.NoError(t, err)
	require.Equal(t, apiSearchFolders.StatusCode, http.StatusOK)

	defer func() {
		err := apiSearchFolders.Body.Close()
		require.NoError(t, err)
	}()
	dataBody, err := ioutil.ReadAll(apiSearchFolders.Body)
	require.NoError(t, err)

	// Query to api
	dashboardFolders := make([]map[string]interface{}, 0)
	err = json.Unmarshal(dataBody, &dashboardFolders)
	require.NoError(t, err)
	require.Len(t, dashboardFolders, 0)

	for _, folder := range dashboardFolders {
		// http request to every dashboard
		domain = fmt.Sprintf("%s%s", gt.url, folder["url"])
		dashResp, err := gt.requestToGrafana(domain, "GET", nil, strings.NewReader(gt.loginForm.Encode()), cookie)
		require.NoError(t, err)
		require.Equal(t, dashResp.StatusCode, http.StatusOK)
		title := fmt.Sprintf("%s", folder["title"])
		dashboards[title] = dashboard{title: title, url: fmt.Sprintf("%s", folder["url"])}
	}
}

func (gt *grafanaTest) getGrafanaAndDexAuth(t *testing.T) *http.Response {
	//  /login
	domain := fmt.Sprintf("%s%s", gt.url, "/login")
	grafLogin, err := gt.requestToGrafana(domain, "GET", nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, grafLogin.StatusCode, http.StatusOK)

	//  /login/generic_oauth
	domain = fmt.Sprintf("%s%s", domain, "/generic_oauth")
	genericOauth, err := gt.requestToGrafana(domain, "GET", nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, genericOauth.StatusCode, http.StatusOK)

	// /auth
	domain = genericOauth.Request.Referer()
	dexAuth, err := gt.requestToGrafana(domain, "GET", nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, dexAuth.StatusCode, http.StatusOK)

	// /auth/local
	domain = dexAuth.Request.URL.String()
	dexAuthLocal, err := gt.requestToGrafana(domain, "POST", nil, strings.NewReader(gt.loginForm.Encode()), nil)
	require.NoError(t, err)
	require.Equal(t, dexAuthLocal.StatusCode, http.StatusOK)

	return dexAuthLocal
}

func (gt *grafanaTest) requestToGrafana(domain, method string, params url.Values, formData io.Reader, cookies []*http.Cookie) (*http.Response, error) {
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

	var resp *http.Response

	err = retry.Do(func() error {
		resp, err = gt.httpClient.Do(req)
		if err != nil {
			return err
		}

		if err := verifyStatusCode(resp, 200); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		gt.log.Errorf("http request to the url (%s) failed with '%s'", u, err)
		return nil, fmt.Errorf("http request to the url (%s) failed with '%s'", u, err)
	}

	return resp, err
}

func (gt *grafanaTest) getGrafana(t *testing.T) error {
	pods, err := gt.coreClient.CoreV1().Pods(grafanaNS).List(metav1.ListOptions{LabelSelector: grafanaLabelSelector})
	require.NoError(t, err)
	require.Len(t, pods.Items, 1)
	pod := pods.Items[0]
	require.Equal(t, strings.TrimSpace(string(pod.Status.Phase)), corev1.PodRunning)

	spec := pod.Spec
	containers := spec.Containers
	for _, container := range containers {
		if container.Name == containerName {
			envs := container.Env
			for _, envVar := range envs {
				switch envVar.Name {
				case "GF_SERVER_ROOT_URL":
					gt.url = strings.TrimSuffix(envVar.Value, "/")
				}
			}
		}
	}

	err = gt.getCredentials()
	if err != nil {
		return err
	}

	gt.httpClient = getHTTPClient(t, true)

	return nil
}

func (gt *grafanaTest) getCredentials() error {
	secret, err := gt.coreClient.CoreV1().Secrets(grafanaNS).Get(adminUserSecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	data := secret.Data
	for key, value := range data {
		switch key {
		case "email":
			if string(value) == "" {
				return fmt.Errorf("No email found in secret '%s'", adminUserSecretName)
			}
			gt.loginForm.Set("login", string(value))
		case "password":
			if string(value) == "" {
				return fmt.Errorf("No password found in secret '%s'", adminUserSecretName)
			}
			gt.loginForm.Set("password", string(value))
		}

	}

	return nil

}

func getHTTPClient(t *testing.T, skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	cookieJar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &http.Client{Timeout: 15 * time.Second, Transport: tr, Jar: cookieJar}
}

func (gt *grafanaTest) waitForPodGrafana(waitmax time.Duration) error {
	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-timeout:
			pods, err := gt.coreClient.CoreV1().Pods(grafanaNS).List(metav1.ListOptions{LabelSelector: grafanaLabelSelector})
			if err != nil {
				return err
			}
			if len(pods.Items) < 1 {
				return fmt.Errorf("Grafana pod could not be found")
			}
			return fmt.Errorf("Pod did not start within given time  %v: %+v", waitmax, pods.Items[0])
		case <-tick:
			pods, err := gt.coreClient.CoreV1().Pods(grafanaNS).List(metav1.ListOptions{LabelSelector: grafanaLabelSelector})
			if err != nil {
				return err
			}
			if len(pods.Items) < 1 {
				return fmt.Errorf("Grafana pod could not be found")
			}
			pod := pods.Items[0]
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

func verifyStatusCode(res *http.Response, expectedStatusCode int) error {
	if res.StatusCode != expectedStatusCode {
		return fmt.Errorf("status code is wrong, have: %d, want: %d", res.StatusCode, expectedStatusCode)
	}
	return nil
}
