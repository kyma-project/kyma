package bundle_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
)

func TestHTTPRepository_IndexReader(t *testing.T) {
	// GIVEN:
	expContentGen := func() string { return "expected content - index" }

	mux := http.NewServeMux()
	mux.HandleFunc("/index.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expContentGen())
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	rep := bundle.NewHTTPRepository(bundle.RepositoryConfig{URL: ts.URL})

	// WHEN:
	r, clo, err := rep.IndexReader()

	// THEN:
	require.NoError(t, err)
	defer clo()

	got, err := ioutil.ReadAll(r)
	require.NoError(t, err)

	assert.EqualValues(t, expContentGen(), string(got))
}

func TestHTTPRepository_BundleReader(t *testing.T) {
	const (
		expBundleName bundle.Name    = "bundle_name"
		expBundleVer  bundle.Version = "1.2.3"
	)
	expContentGen := func() string { return "expected content - bundle" }

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/%s-%s.tgz", expBundleName, expBundleVer), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expContentGen())
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	rep := bundle.NewHTTPRepository(bundle.RepositoryConfig{URL: ts.URL})

	r, clo, err := rep.BundleReader(expBundleName, expBundleVer)
	require.NoError(t, err)
	defer clo()

	got, err := ioutil.ReadAll(r)
	require.NoError(t, err)

	assert.EqualValues(t, expContentGen(), string(got))
}
