package addon_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/helm/pkg/chartutil"

	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
)

// TestLoaderLoad processes given test case addon and compares it to the
// corresponding files from ./testdata/addon-redis-0.0.1.golden dir
func TestLoaderLoadSuccess(t *testing.T) {
	for tn, tc := range map[string]struct {
		tgzPath string
	}{
		"additional files in chart directory should be ignored": {
			tgzPath: "./testdata/addon-ignore-file-in-chart-dir.input.tgz",
		},
		"simple": {
			tgzPath: "testdata/addon-redis-0.0.1.input.tgz",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			fixBaseDir := "../../tmp"
			var fixDirName string
			createTmpDirFake := func(dir, prefix string) (string, error) {
				assert.Equal(t, fixBaseDir, dir)

				name, err := ioutil.TempDir(dir, prefix)
				fixDirName = name
				return name, err
			}

			addonLoader := addon.NewLoader(fixBaseDir, spy.NewLogDummy())
			addonLoader.SetCreateTmpDir(createTmpDirFake)

			expChart, err := chartutil.Load("testdata/addon-redis-0.0.1.golden/chart/redis")
			require.NoError(t, err)

			fd, err := os.Open(tc.tgzPath)
			require.NoError(t, err)
			defer fd.Close()

			// when
			yb, c, err := addonLoader.Load(fd)

			// then
			require.NoError(t, err)

			require.Len(t, c, 1)
			assert.Equal(t, expChart, c[0])

			require.NotNil(t, yb)
			assert.Equal(t, fixtureAddon(t, "./testdata/addon-redis-0.0.1.golden/"), *yb)

			assertDirNotExits(t, filepath.Join("../tmp/", fixDirName))
		})
	}
}

func TestLoaderLoadFailure(t *testing.T) {
	for tn, tc := range map[string]struct {
		tgzPath string
		errMsg  string
	}{
		"missing plans": {
			tgzPath: "./testdata/addon-missing-plans.input.tgz",
			errMsg:  "while mapping buffered files to form: addon does not contains any plans, please check if addon contains \"plans\" directory",
		},
		"missing chart directory": {
			tgzPath: "./testdata/addon-missing-chart-dir.input.tgz",
			errMsg:  "while loading chart: while discovering the name of the Helm Chart under the \"chart\" addon directory: addon does not contains \"chart\" directory",
		},
		"multiple folders in chart directory": {
			tgzPath: "./testdata/addon-multiple-charts-in-chart-dir.input.tgz",
			errMsg:  "while loading chart: while discovering the name of the Helm Chart under the \"chart\" addon directory: \"chart\" directory MUST contain only one Helm Chart folder but found multiple directories: [redis, redis-v2]",
		},
		"missing chart in chart directory": {
			tgzPath: "./testdata/addon-no-chart-in-chart-dir.input.tgz",
			errMsg:  "while loading chart: while discovering the name of the Helm Chart under the \"chart\" addon directory: \"chart\" directory SHOULD contain one Helm Chart folder but it's empty",
		},
		"missing meta.yaml file for micro plan": {
			tgzPath: "./testdata/addon-missing-plan-meta-file.input.tgz",
			errMsg:  "while mapping buffered files to form: while unmarshaling plan \"meta.yaml\" file: \"meta.yaml\" is required but is not present",
		},
		"missing bind.yaml file for micro plan which is marked as bindable": {
			tgzPath: "./testdata/addon-missing-bind-yaml-when-bindable.input.tgz",
			errMsg:  "while validating form: while validating \"micro\" plan: plans is marked as bindable but bind.yaml file was not found in plan micro",
		},
		"incorrect create schema": {
			tgzPath: "./testdata/addon-incorrect-create-schema.input.tgz",
			errMsg:  "while mapping buffered files to form: while unmarshaling plan \"create-instance-schema.json\" file: while loading plan shcema: invalid character '}' after object key",
		},
		"big schema": {
			tgzPath: "./testdata/addon-big-schema.input.tgz",
			errMsg:  "while mapping buffered files to form: while unmarshaling plan \"update-instance-schema.json\" file: schema update-instance-schema.json is larger than 64 kB",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			addonLoader := addon.NewLoader("", spy.NewLogDummy())

			fd, err := os.Open(tc.tgzPath)
			require.NoError(t, err)
			defer fd.Close()

			// when
			yb, c, err := addonLoader.Load(fd)

			// then
			require.EqualError(t, err, tc.errMsg)
			assert.Nil(t, yb)
			assert.Nil(t, c)
		})
	}

}
