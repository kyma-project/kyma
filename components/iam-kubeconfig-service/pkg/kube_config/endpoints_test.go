package kube_config

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestGetKubeConfig_ShouldReturnKubeConfig(t *testing.T) {

	// given
	clusterName := "testname"
	ca := "testca"
	url := "testurl"
	namespace := "testnamespace"
	token := "testtoken"

	kubeConfig := NewKubeConfig(clusterName, url, ca, namespace)
	endpoints := NewEndpoints(kubeConfig)

	resp := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "/kube-config", nil)
	req.Header.Add("Authorization", "Bearer "+token)

	// when
	h := http.HandlerFunc(endpoints.GetKubeConfig)
	h.ServeHTTP(resp, req)

	// then
	if resp.Code != http.StatusOK {
		t.Errorf("Invalid response status code (actual: %d, expected: %d)", resp.Code, http.StatusOK)
	}

	contentType := resp.Header().Get("Content-Type")
	if contentType != MimeTypeYaml {
		t.Errorf("Invalid response content type (actual: '%s', expected: '%s')", contentType, MimeTypeYaml)
	}

	respBody := resp.Body.String()
	if respBody == "" {
		t.Error("Response body should not be empty.")
	}

	shouldContain(t, respBody, clusterName, "Response should contain cluster name.")
	shouldContain(t, respBody, ca, "Response should contain CA.")
	shouldContain(t, respBody, url, "Response should contain URL.")
	shouldContain(t, respBody, namespace, "Response should contain namespace.")
	shouldContain(t, respBody, "token: "+token, "Response should contain user token.")
}

func TestGetKubeConfig_ShouldReturnBadRequest_IfMissingAuthorization(t *testing.T) {

	// given
	clusterName := "testname"
	ca := "testca"
	url := "testurl"
	namespace := "testnamespace"

	kubeConfig := NewKubeConfig(clusterName, url, ca, namespace)
	endpoints := NewEndpoints(kubeConfig)

	resp := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "/kube-config", nil)

	// when
	h := http.HandlerFunc(endpoints.GetKubeConfig)
	h.ServeHTTP(resp, req)

	// then
	if resp.Code != http.StatusBadRequest {
		t.Errorf("Invalid response status code (actual: %d, expected: %d)", resp.Code, http.StatusBadRequest)
	}
}

func TestGetKubeConfig_ShouldReturnBadRequest_IfInvalidAuthorization(t *testing.T) {

	// given
	clusterName := "testname"
	ca := "testca"
	url := "testurl"
	namespace := "testnamespace"

	kubeConfig := NewKubeConfig(clusterName, url, ca, namespace)
	endpoints := NewEndpoints(kubeConfig)

	resp := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "/kube-config", nil)
	req.SetBasicAuth("u", "p")

	// when
	h := http.HandlerFunc(endpoints.GetKubeConfig)
	h.ServeHTTP(resp, req)

	// then
	if resp.Code != http.StatusBadRequest {
		t.Errorf("Invalid response status code (actual: %d, expected: %d)", resp.Code, http.StatusBadRequest)
	}
}

func shouldContain(t *testing.T, s, pattern, message string) {
	matches, matchErr := regexp.MatchString(pattern, s)
	if matchErr != nil {
		t.Fatal(matchErr)
	}
	if !matches {
		t.Errorf(message)
	}
}
