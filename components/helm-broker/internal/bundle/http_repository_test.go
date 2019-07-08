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
	// given
	const expContentGen = "expected content - index"

	mux := http.NewServeMux()
	mux.HandleFunc("/index.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expContentGen)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	rep := bundle.NewHTTPRepository()

	// when
	r, err := rep.IndexReader(ts.URL + "/index.yaml")

	// then
	require.NoError(t, err)
	defer r.Close()

	got, err := ioutil.ReadAll(r)
	require.NoError(t, err)

	assert.EqualValues(t, expContentGen, string(got))
}

func TestHTTPRepository_BundleReader(t *testing.T) {
	// given
	const (
		expBundleName bundle.Name    = "bundle_name"
		expBundleVer  bundle.Version = "1.2.3"
		expContentGen string         = "expected content - bundle"
	)

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/%s-%s.tgz", expBundleName, expBundleVer), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expContentGen)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	rep := bundle.NewHTTPRepository()

	// when
	_, err := rep.IndexReader(ts.URL + fmt.Sprintf("/%s-%s.tgz", expBundleName, expBundleVer))
	require.NoError(t, err)
	r, err := rep.BundleReader(expBundleName, expBundleVer)

	// then
	require.NoError(t, err)
	defer r.Close()

	got, err := ioutil.ReadAll(r)
	require.NoError(t, err)

	assert.EqualValues(t, expContentGen, string(got))
}

func TestHTTPRepository_URLForBundle(t *testing.T) {
	// given
	const (
		bundleName bundle.Name    = "bundle_name"
		bundleVer  bundle.Version = "1.2.3"
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/index.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	rep := bundle.NewHTTPRepository()

	// when
	_, err := rep.IndexReader(ts.URL + "/index.yaml")
	require.NoError(t, err)
	gotURL := rep.URLForBundle(bundleName, bundleVer)

	// then
	assert.Equal(t, fmt.Sprintf("%s/%s-%s.tgz", ts.URL, bundleName, bundleVer), gotURL)
}
