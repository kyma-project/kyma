package provider_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon/provider"
)

func TestHTTPRepositoryIndexReader(t *testing.T) {
	// given
	const expContentGen = "expected content - index"

	mux := http.NewServeMux()
	mux.HandleFunc("/index.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expContentGen)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	idxURL := ts.URL + "/index.yaml"
	dst := "../../../tmp/TestHTTPRepositoryIndexReader"
	httpGetter, err := provider.NewHTTP(idxURL, dst)
	require.NoError(t, err)

	// when
	idxReader, err := httpGetter.IndexReader()

	// then
	require.NoError(t, err)
	defer idxReader.Close()

	got, err := ioutil.ReadAll(idxReader)
	require.NoError(t, err)

	assert.EqualValues(t, expContentGen, string(got))

	// when
	err = httpGetter.Cleanup()

	// then
	assertDirIsEmpty(t, dst)
}

func TestHTTPRepositoryBundleLoadInfo(t *testing.T) {
	// given
	const (
		expBundleName addon.Name    = "bundle_name"
		expBundleVer  addon.Version = "1.2.3"
		expContentGen string        = "expected content - bundle"
	)

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/%s-%s.tgz", expBundleName, expBundleVer), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expContentGen)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	idxURL := ts.URL + "/index.yaml"
	dst := "../../../tmp/TestHTTPRepositoryBundleLoadInfo"
	httpGetter, err := provider.NewHTTP(idxURL, dst)
	require.NoError(t, err)

	// when
	loadType, path, err := httpGetter.BundleLoadInfo(expBundleName, expBundleVer)

	// then
	require.NoError(t, err)
	require.Equal(t, provider.ArchiveLoadType, loadType)

	got, err := ioutil.ReadFile(path)
	require.NoError(t, err)

	assert.EqualValues(t, expContentGen, string(got))

	// when
	err = httpGetter.Cleanup()

	// then
	assertDirIsEmpty(t, dst)
}

func TestHTTPRepositoryBundleDocURL(t *testing.T) {
	// given
	const (
		bundleName addon.Name    = "bundle_name"
		bundleVer  addon.Version = "1.2.3"
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/index.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	idxURL := ts.URL + "/index.yaml"
	httpGetter, err := provider.NewHTTP(idxURL, "../../../tmp")
	require.NoError(t, err)

	// when
	gotURL := httpGetter.BundleDocURL(bundleName, bundleVer)

	// then
	assert.Equal(t, fmt.Sprintf("%s/%s-%s.tgz", ts.URL, bundleName, bundleVer), gotURL)
}

func assertDirIsEmpty(t *testing.T, name string) {
	f, err := os.Open(name)
	require.NoError(t, err)
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err != io.EOF {
		t.Fatalf("Directory %s is not empty, got error: %v", name, err)
	}
}
