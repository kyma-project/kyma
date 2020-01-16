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
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"

	"github.com/google/go-cmp/cmp"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GrafanaUpgradeTest will test if Grafana contains the same dashboards after an upgrade of Kyma
type GrafanaUpgradeTest struct {
	k8sCli     kubernetes.Interface
	httpClient *http.Client
	namespace  string
	log        logrus.FieldLogger
	grafana
}

type grafana struct {
	url       string
	loginForm url.Values
}

type dashboard struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

const (
	grafanaDataField           = "dashboards"
	grafanaAdminUserSecretName = "admin-user"
	grafanaConfigMapName       = "grafana-upgrade-test"
	grafanaNS                  = "kyma-system"
	grafanaContainerName       = "grafana"
	grafanaLabelSelector       = "app=grafana"
)

var (
	dashboardsToCheck = []string{"Lambda Dashboard"}
)

// NewGrafanaUpgradeTest returns new instance of the GrafanaUpgradeTest
func NewGrafanaUpgradeTest(k8sCli kubernetes.Interface) *GrafanaUpgradeTest {
	return &GrafanaUpgradeTest{
		k8sCli:  k8sCli,
		grafana: grafana{loginForm: url.Values{}},
	}
}

// CreateResources retrieves all installed dashboards and stores this information in an config map
func (ut *GrafanaUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	ut.namespace = namespace
	ut.log = log
	if err := ut.getGrafana(); err != nil {
		return err
	}

	dashboards, err := ut.collectDashboards()
	if err != nil {
		return err
	}

	return ut.storeDashboards(dashboards)
}

// TestResources retrieves the previously stored list of installed dashboards and compares it to the current list
func (ut *GrafanaUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	ut.namespace = namespace
	ut.log = log

	err := ut.getGrafana()
	if err != nil {
		return err
	}

	return ut.compareDashboards()
}

func (ut *GrafanaUpgradeTest) getCredentials() error {
	secret, err := ut.k8sCli.CoreV1().Secrets(grafanaNS).Get(grafanaAdminUserSecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	data := secret.Data
	if val, ok := data["email"]; ok {
		ut.loginForm.Set("login", string(val))
	} else {
		return fmt.Errorf("no email found in secret '%s'", grafanaAdminUserSecretName)
	}

	if val, ok := data["password"]; ok {
		ut.loginForm.Set("password", string(val))
	} else {
		return fmt.Errorf("No password found in secret '%s'", grafanaAdminUserSecretName)
	}

	return nil
}

func getHTTPClient(skipVerify bool) (*http.Client, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &http.Client{Timeout: 15 * time.Second, Transport: tr, Jar: cookieJar}, nil
}

func (ut *GrafanaUpgradeTest) getGrafana() error {
	pods, err := ut.k8sCli.CoreV1().Pods(grafanaNS).List(metav1.ListOptions{LabelSelector: grafanaLabelSelector})
	if err != nil {
		return err
	}
	pod := pods.Items[0]

	spec := pod.Spec
	containers := spec.Containers
	for _, container := range containers {
		if container.Name == grafanaContainerName {
			envs := container.Env
			for _, envVar := range envs {
				switch envVar.Name {
				case "GF_SERVER_ROOT_URL":
					ut.url = strings.TrimSuffix(envVar.Value, "/")
				}
			}
		}
	}

	err = ut.getCredentials()
	if err != nil {
		return err
	}

	ut.httpClient, err = getHTTPClient(true)
	if err != nil {
		return err
	}

	return nil
}

func (ut *GrafanaUpgradeTest) collectDashboards() (map[string]dashboard, error) {
	dexAuthLocal, err := ut.getGrafanaAndDexAuth()
	if err != nil {
		return nil, err
	}

	domain := fmt.Sprintf("%s%s", ut.url, "/api/search")
	params := url.Values{}
	params.Set("folderIds", "0")
	cookie := dexAuthLocal.Request.Cookies()
	apiSearchFolders, err := ut.requestToGrafana(domain, "GET", params, strings.NewReader(ut.loginForm.Encode()), cookie)
	if err != nil {
		return nil, errors.Wrap(err, "could not get dashboard list")
	}

	defer apiSearchFolders.Body.Close()
	dataBody, err := ioutil.ReadAll(apiSearchFolders.Body)
	if err != nil {
		return nil, err
	}

	// Query to api
	dashboardFolders := make([]map[string]interface{}, 0)
	err = json.Unmarshal(dataBody, &dashboardFolders)
	if err != nil {
		return nil, err
	}
	dashboards := make(map[string]dashboard)

	for _, folder := range dashboardFolders {
		// http request to every dashboard
		domain = fmt.Sprintf("%s%s", ut.url, folder["url"])
		_, err := ut.requestToGrafana(domain, "GET", nil, strings.NewReader(ut.loginForm.Encode()), cookie)
		if err != nil {
			return nil, errors.Wrap(err, "could not open dashboard")
		}
		title := fmt.Sprintf("%s", folder["title"])
		dashboards[title] = dashboard{Title: title, URL: fmt.Sprintf("%s", folder["url"])}
	}

	return dashboards, nil
}

func (ut *GrafanaUpgradeTest) storeDashboards(dashboards map[string]dashboard) error {
	dashboardJSON, err := json.Marshal(dashboards)
	if err != nil {
		return err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: grafanaConfigMapName,
		},
		Data: map[string]string{
			grafanaDataField: string(dashboardJSON),
		},
	}
	_, err = ut.k8sCli.CoreV1().ConfigMaps(ut.namespace).Create(cm)
	return err
}

func (ut *GrafanaUpgradeTest) retrievePreviousDashboards() (map[string]dashboard, error) {
	cm, err := ut.k8sCli.CoreV1().ConfigMaps(ut.namespace).Get(grafanaConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	dashboards := make(map[string]dashboard)
	err = json.Unmarshal([]byte(cm.Data[grafanaDataField]), &dashboards)
	if err != nil {
		return nil, err
	}

	return dashboards, nil
}

func filter(input map[string]dashboard, f func(string) bool) map[string]dashboard {
	dashboards := make(map[string]dashboard)
	for title, dashboard := range input {
		if f(title) {
			dashboards[title] = dashboard
		}
	}
	return dashboards
}

func (ut *GrafanaUpgradeTest) compareDashboards() error {
	previous, err := ut.retrievePreviousDashboards()
	if err != nil {
		return err
	}

	previousFiltered := filter(previous, func(title string) bool {
		for _, t := range dashboardsToCheck {
			if t == title {
				return true
			}
		}
		return false
	})
	ut.log.Println(previousFiltered)

	current, err := ut.collectDashboards()
	if err != nil {
		return err
	}
	currentFiltered := filter(current, func(title string) bool {
		for _, t := range dashboardsToCheck {
			if t == title {
				return true
			}
		}
		return false
	})
	ut.log.Println(currentFiltered)

	if !cmp.Equal(previousFiltered, currentFiltered) {
		return fmt.Errorf("retrieved data not equal: before: \n%+v \nafter: \n%+v", previousFiltered, currentFiltered)
	}
	return nil
}

func (ut *GrafanaUpgradeTest) getGrafanaAndDexAuth() (*http.Response, error) {
	//  /login
	domain := fmt.Sprintf("%s%s", ut.url, "/login")
	_, err := ut.requestToGrafana(domain, "GET", nil, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "login failed")
	}

	//  /login/generic_oauth
	domain = fmt.Sprintf("%s%s", domain, "/generic_oauth")
	genericOauth, err := ut.requestToGrafana(domain, "GET", nil, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "oauth failed")
	}

	// /auth
	domain = genericOauth.Request.Referer()
	dexAuth, err := ut.requestToGrafana(domain, "GET", nil, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "oauth referrer failed")
	}
	// /auth/local
	domain = dexAuth.Request.URL.String()
	dexAuthLocal, err := ut.requestToGrafana(domain, "POST", nil, strings.NewReader(ut.loginForm.Encode()), nil)
	if err != nil {
		return nil, errors.Wrap(err, "during dex auth")
	}
	return dexAuthLocal, nil
}

func (ut *GrafanaUpgradeTest) requestToGrafana(domain, method string, params url.Values, formData io.Reader, cookies []*http.Cookie) (*http.Response, error) {
	u, err := url.Parse(domain)
	if err != nil {
		return nil, fmt.Errorf("parsing url (%s) failed with '%s'", domain, err)
	}

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

	var resp *http.Response

	err = retry.Do(func() error {
		resp, err = ut.httpClient.Do(req)
		if err != nil {
			return err
		}

		if err := verifyStatusCode(resp, http.StatusOK); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("http request to the url (%s) failed with '%s'", u, err)
	}

	return resp, err
}

func verifyStatusCode(res *http.Response, expectedStatusCode int) error {
	if res.StatusCode != expectedStatusCode {
		return fmt.Errorf("status code is wrong, have: %d, want: %d", res.StatusCode, expectedStatusCode)
	}
	return nil
}
