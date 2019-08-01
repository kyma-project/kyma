// +build integration

package integration_test

import (
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

const (
	sourceHTTP = "http"
	sourceGit  = "git"
)

type repositorySource struct {
	kind    string
	ts      *testSuite
	gitRepo *gitRepo
	urls    []string
}

func newSource(kind string, ts *testSuite, repository *gitRepo, urls []string) *repositorySource {
	rs := &repositorySource{
		ts:      ts,
		gitRepo: repository,
		kind:    kind,
	}

	sourceUrls := []string{}
	for _, url := range urls {
		sourceUrls = append(sourceUrls, rs.generateURL(url))
	}
	rs.urls = sourceUrls

	return rs
}

func (rs *repositorySource) generateURL(url string) string {
	switch rs.kind {
	case sourceHTTP:
		return rs.ts.repoServer.URL + "/" + url
	case sourceGit:
		return "git::" + rs.gitRepo.path(url)
	default:
		rs.ts.t.Fatalf("Unsupported source kind: %s", rs.kind)
	}

	return ""
}

func (rs *repositorySource) removeURL(url string) {
	path := rs.generateURL(url)
	newUrls := []string{}

	for _, u := range rs.urls {
		if u == path {
			rs.ts.t.Logf("URL %q was removed from repository source", u)
			continue
		}
		newUrls = append(newUrls, u)
	}

	rs.urls = newUrls
}

func (rs *repositorySource) replaceURL(url string) {
	rs.urls = []string{rs.generateURL(url)}
}

func (rs *repositorySource) generateAddonRepositories() []v1alpha1.SpecRepository {
	var repositories []v1alpha1.SpecRepository

	// v1alpha1.SpecRepository cannot be null, needs to be empty array
	if len(rs.urls) == 0 {
		repositories = append(repositories, v1alpha1.SpecRepository{})
		return repositories
	}

	for _, url := range rs.urls {
		repositories = append(repositories, v1alpha1.SpecRepository{URL: url})
	}

	return repositories
}
