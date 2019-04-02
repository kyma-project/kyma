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

	"github.com/google/go-cmp/cmp"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type grafanaUpgradeTest struct {
	k8sCli    kubernetes.Interface
	namespace string
	log       logrus.FieldLogger
	grafana
}

type grafana struct {
	url        string
	oauthUrl   string
	loginForm  url.Values
	httpClient *http.Client
}
type dashboard struct {
	Title string `json:title`
	URL   string `json:url`
}

func NewGrafanaUpgradeTest(k8sCli kubernetes.Interface) *grafanaUpgradeTest {
	return &grafanaUpgradeTest{
		k8sCli:  k8sCli,
		grafana: grafana{loginForm: url.Values{}},
	}

}

func (ut *grafanaUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	ut.namespace = namespace
	ut.log = log
	err := ut.getGrafana()
	if err != nil {
		return err
	}

	dashboards, err := ut.collectDashboards()
	if err != nil {
		return err
	}

	err = ut.storeDashboards(dashboards)
	return err
}

func (ut *grafanaUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	ut.namespace = namespace
	ut.log = log
	err := ut.getGrafana()
	if err != nil {
		return err
	}
	return ut.compareDashboards()
}

const (
	grafanaDataField           = "dashboards"
	grafanaAdminUserSecretName = "admin-user"
	grafanaConfigMapName       = "grafana-upgrade-test"
	grafanaNS                  = "kyma-system"
	grafanaPodName             = "monitoring-grafana-0"
	grafanaServiceName         = "monitoring-grafana"
	grafanaStatefulsetName     = "monitoring-grafana"
	grafanaPvcName             = "monitoring-grafana"
	grafanaContainerName       = "grafana"
	grafanaLabelSelector       = "app=monitoring-grafana"
)

func (ut *grafanaUpgradeTest) getCredentials() error {
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

func (ut *grafanaUpgradeTest) getGrafana() error {

	pod, err := ut.k8sCli.CoreV1().Pods(grafanaNS).Get(grafanaPodName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	spec := pod.Spec
	containers := spec.Containers
	for _, container := range containers {
		if container.Name == grafanaContainerName {
			envs := container.Env
			for _, envVar := range envs {
				switch envVar.Name {
				case "GF_AUTH_GENERIC_OAUTH_AUTH_URL":
					ut.oauthUrl = envVar.Value
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

func (ut *grafanaUpgradeTest) collectDashboards() (map[string]dashboard, error) {

	dexAuthLocal, err := ut.getGrafanaAndDexAuth()
	if err != nil {
		return nil, err
	}
	domain := fmt.Sprintf("%s%s", dexAuthLocal.Request.URL.String(), "api/search")
	params := url.Values{}
	params.Set("folderIds", "0")
	cookie := dexAuthLocal.Request.Cookies()
	apiSearchFolders, err := ut.requestToGrafana(domain, "GET", params, nil, cookie)
	if err != nil {
		return nil, err
	}
	if apiSearchFolders.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed %v, %v", domain, apiSearchFolders.StatusCode)
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
		dashResp, err := ut.requestToGrafana(domain, "GET", nil, strings.NewReader(ut.loginForm.Encode()), cookie)
		if err != nil {

		}
		if dashResp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("request failed %v, %v", domain, dashResp.StatusCode)
		}
		title := fmt.Sprintf("%s", folder["title"])
		dashboards[title] = dashboard{Title: title, URL: fmt.Sprintf("%s", folder["url"])}
	}

	return dashboards, nil
}

func (ut *grafanaUpgradeTest) storeDashboards(dashboards map[string]dashboard) error {
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

func (ut *grafanaUpgradeTest) retrievePreviousDashboards() (map[string]dashboard, error) {

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

func (ut *grafanaUpgradeTest) compareDashboards() error {
	previous, err := ut.retrievePreviousDashboards()
	if err != nil {
		return err
	}
	current, err := ut.collectDashboards()

	ut.log.Debugln(current)
	if !cmp.Equal(previous, current) {
		return fmt.Errorf("retrieved data not equal: before: %+v, after: %+v", previous, current)
	}
	return nil
}

func (ut *grafanaUpgradeTest) getGrafanaAndDexAuth() (*http.Response, error) {

	//  /login
	domain := fmt.Sprintf("%s%s", ut.url, "/login")
	grafLogin, err := ut.requestToGrafana(domain, "GET", nil, nil, nil)
	if grafLogin.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login to grafana failed: StatusCode: %v", grafLogin.StatusCode)
	}

	//  /login/generic_oauth
	domain = fmt.Sprintf("%s%s", domain, "/generic_oauth")
	genericOauth, err := ut.requestToGrafana(domain, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if genericOauth.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login to grafana failed (oauth): StatusCode: %v", genericOauth.StatusCode)
	}

	// /auth
	domain = genericOauth.Request.Referer()
	dexAuth, err := ut.requestToGrafana(domain, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if dexAuth.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Dex auth failed: StatusCode %v", dexAuth.StatusCode)
	}

	// /auth/local
	domain = dexAuth.Request.URL.String()
	dexAuthLocal, err := ut.requestToGrafana(domain, "POST", nil, strings.NewReader(ut.loginForm.Encode()), nil)
	if err != nil {
		return nil, err
	}
	if dexAuthLocal.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %v failed: StatusCode %v", domain, dexAuthLocal.StatusCode)
	}
	return dexAuthLocal, nil
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
