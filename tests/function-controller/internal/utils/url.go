package utils

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

func GetSvcURL(name, namespace string, useProxy bool) (*url.URL, error) {
	var svcURL = fmt.Sprintf("http://%s.%s.svc.cluster.local", name, namespace)
	if useProxy {
		svcURL = fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/services/%s:80/proxy/", namespace, name)
	}
	parsedURL, err := url.Parse(svcURL)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing function access URL")
	}
	return parsedURL, nil
}

func GetGitURL(name, namespace, repoName string, useProxy bool) (*url.URL, error) {
	svcURL, err := GetSvcURL(name, namespace, useProxy)
	if err != nil {
		return nil, errors.Wrap(err, "while calculating svc url")
	}
	svcURL = svcURL.JoinPath("", fmt.Sprintf("%s.git", repoName))
	return svcURL, nil
}
