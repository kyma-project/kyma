package ybundle_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/helm/pkg/chartutil"

	"github.com/kyma-project/kyma/components/helm-broker/internal/ybundle"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
)

// TestLoaderLoad processes given test case bundle and compares it to the
// corresponding files from ./testdata/bundle-redis-0.0.1.golden dir
func TestLoaderLoadSuccess(t *testing.T) {
	// given
	fixBaseDir := "../../tmp"
	var fixDirName string
	createTmpDirFake := func(dir, prefix string) (string, error) {
		assert.Equal(t, fixBaseDir, dir)

		name, err := ioutil.TempDir(dir, prefix)
		fixDirName = name
		return name, err
	}

	bundleLoader := ybundle.NewLoader(fixBaseDir, spy.NewLogDummy())
	bundleLoader.SetCreateTmpDir(createTmpDirFake)

	expChart, err := chartutil.Load("testdata/bundle-redis-0.0.1.golden/chart/redis")
	require.NoError(t, err)

	fd, err := os.Open("testdata/bundle-redis-0.0.1.input.tgz")
	require.NoError(t, err)
	defer fd.Close()

	// when
	yb, c, err := bundleLoader.Load(fd)

	// then
	require.NoError(t, err)

	require.Len(t, c, 1)
	assert.Equal(t, expChart, c[0])

	require.NotNil(t, yb)
	assert.Equal(t, fixtureBundle(t, "./testdata/bundle-redis-0.0.1.golden/"), *yb)

	assertDirNotExits(t, filepath.Join("../tmp/", fixDirName))
}

func TestLoaderLoadFailure(t *testing.T) {
	for tn, tc := range map[string]struct {
		tgzPath string
		errMsg  string
	}{
		"missing plans": {
			tgzPath: "./testdata/bundle-missing-plans.input.tgz",
			errMsg:  "while mapping buffered files to form: bundle does not contains any plans, please check if bundle contains \"plans\" directory",
		},
		"missing chart directory": {
			tgzPath: "./testdata/bundle-missing-chart-dir.input.tgz",
			errMsg:  "while loading chart: bundle does not contains \"chart\" directory",
		},
		"multiple charts in chart directory": {
			tgzPath: "./testdata/bundle-multiple-charts-in-chart-dir.input.tgz",
			errMsg:  "while loading chart: \"chart\" directory MUST contain one folder",
		},
		"missing chart in chart directory": {
			tgzPath: "./testdata/bundle-no-chart-in-chart-dir.input.tgz",
			errMsg:  "while loading chart: \"chart\" directory MUST contain one folder",
		},
		"missing meta.yaml file for micro plan": {
			tgzPath: "./testdata/bundle-missing-plan-meta-file.input.tgz",
			errMsg:  "while mapping buffered files to form: while unmarshaling plan \"meta.yaml\" file: \"meta.yaml\" is required but is not present",
		},
		"missing bind.yaml file for micro plan which is marked as bindable": {
			tgzPath: "./testdata/bundle-missing-bind-yaml-when-bindable.input.tgz",
			errMsg:  "while validating form: while validating \"micro\" plan: plans is marked as bindable but bind.yaml file was not found in plan micro",
		},
		"incorrect create schema": {
			tgzPath: "./testdata/bundle-incorrect-create-schema.input.tgz",
			errMsg:  "while mapping buffered files to form: while unmarshaling plan \"create-instance-schema.json\" file: while loading plan shcema: invalid character '}' after object key",
		},
		"big schema": {
			tgzPath: "./testdata/bundle-big-schema.input.tgz",
			errMsg:  "while mapping buffered files to form: while unmarshaling plan \"update-instance-schema.json\" file: schema update-instance-schema.json is larger than 64 kB",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			bundleLoader := ybundle.NewLoader("", spy.NewLogDummy())

			fd, err := os.Open(tc.tgzPath)
			require.NoError(t, err)
			defer fd.Close()

			// when
			yb, c, err := bundleLoader.Load(fd)

			// then
			require.EqualError(t, err, tc.errMsg)
			assert.Nil(t, yb)
			assert.Nil(t, c)
		})
	}

}
