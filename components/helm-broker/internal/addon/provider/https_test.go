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

func TestHTTPRepositoryAddonLoadInfo(t *testing.T) {
	// given
	const (
		expAddonName  addon.Name    = "addon_name"
		expAddonVer   addon.Version = "1.2.3"
		expContentGen string        = "expected content - addon"
	)

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/%s-%s.tgz", expAddonName, expAddonVer), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expContentGen)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	idxURL := ts.URL + "/index.yaml"
	dst := "../../../tmp/TestHTTPRepositoryAddonLoadInfo"
	httpGetter, err := provider.NewHTTP(idxURL, dst)
	require.NoError(t, err)

	// when
	loadType, path, err := httpGetter.AddonLoadInfo(expAddonName, expAddonVer)

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

func TestHTTPRepositoryAddonDocURL(t *testing.T) {
	// given
	const (
		addonName addon.Name    = "addon_name"
		addonVer  addon.Version = "1.2.3"
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
	gotURL, err := httpGetter.AddonDocURL(addonName, addonVer)
	require.NoError(t, err)

	// then
	assert.Equal(t, fmt.Sprintf("%s/%s-%s.tgz", ts.URL, addonName, addonVer), gotURL)
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
