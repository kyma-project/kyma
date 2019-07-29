package addon_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
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

	rep := addon.NewHTTPRepository()

	// when
	r, err := rep.IndexReader(ts.URL + "/index.yaml")

	// then
	require.NoError(t, err)
	defer r.Close()

	got, err := ioutil.ReadAll(r)
	require.NoError(t, err)

	assert.EqualValues(t, expContentGen, string(got))
}

func TestHTTPRepository_AddonReader(t *testing.T) {
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

	rep := addon.NewHTTPRepository()

	// when
	_, err := rep.IndexReader(ts.URL + fmt.Sprintf("/%s-%s.tgz", expAddonName, expAddonVer))
	require.NoError(t, err)
	r, err := rep.AddonReader(expAddonName, expAddonVer)

	// then
	require.NoError(t, err)
	defer r.Close()

	got, err := ioutil.ReadAll(r)
	require.NoError(t, err)

	assert.EqualValues(t, expContentGen, string(got))
}

func TestHTTPRepository_URLForAddon(t *testing.T) {
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

	rep := addon.NewHTTPRepository()

	// when
	_, err := rep.IndexReader(ts.URL + "/index.yaml")
	require.NoError(t, err)
	gotURL := rep.URLForAddon(addonName, addonVer)

	// then
	assert.Equal(t, fmt.Sprintf("%s/%s-%s.tgz", ts.URL, addonName, addonVer), gotURL)
}
