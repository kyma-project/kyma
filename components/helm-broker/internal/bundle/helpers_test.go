package bundle_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"

	"github.com/Masterminds/semver"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"
)

func fixtureBundle(t *testing.T, testdataBasePath string) internal.Bundle {
	meta := bundle.FormMeta{}
	unmarshalYamlTestdata(t, testdataBasePath+"meta.yaml", &meta)
	bVer, err := semver.NewVersion(meta.Version)
	require.NoError(t, err)

	charRef := fixChartRef(t, testdataBasePath)
	micro := fixturePlan(t, testdataBasePath, "micro", charRef)
	enterprise := fixturePlan(t, testdataBasePath, "enterprise", charRef)

	return internal.Bundle{
		ID:          internal.BundleID(meta.ID),
		Version:     *bVer,
		Name:        internal.BundleName(meta.Name),
		Description: meta.Description,
		Bindable:    meta.Bindable,
		Metadata: internal.BundleMetadata{
			DisplayName:         meta.DisplayName,
			DocumentationURL:    meta.DocumentationURL,
			ImageURL:            meta.ImageURL,
			LongDescription:     meta.LongDescription,
			ProviderDisplayName: meta.ProviderDisplayName,
			SupportURL:          meta.SupportURL,
			Labels:              meta.MapLabelsToModel(),
		},
		Tags: meta.MapTagsToModel(),
		Plans: map[internal.BundlePlanID]internal.BundlePlan{
			micro.ID:      micro,
			enterprise.ID: enterprise,
		},
		Requires:            meta.Requires,
		PlanUpdatable:       meta.PlanUpdatable,
		BindingsRetrievable: meta.BindingsRetrievable,
	}
}

func fixChartRef(t *testing.T, testdataBasePath string) internal.ChartRef {
	var chart struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}
	unmarshalYamlTestdata(t, testdataBasePath+"/chart/redis/Chart.yaml", &chart)
	cVer, err := semver.NewVersion(chart.Version)
	require.NoError(t, err)

	return internal.ChartRef{
		Name:    internal.ChartName(chart.Name),
		Version: *cVer,
	}
}

func fixturePlan(t *testing.T, testdataBasePath string, planName string, cRef internal.ChartRef) internal.BundlePlan {
	fullbasePath := fmt.Sprintf("%s/plans/%s/", testdataBasePath, planName)
	var meta struct {
		ID          string `yaml:"id"`
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		DisplayName string `yaml:"displayName"`
		Bindable    *bool  `yaml:"bindable"`
		Free        *bool  `yaml:"free"`
	}
	unmarshalYamlTestdata(t, fullbasePath+"meta.yaml", &meta)

	var chartVal map[string]interface{}
	unmarshalYamlTestdata(t, fullbasePath+"values.yaml", &chartVal)

	schemaCreate := unmarshalJSONTestdata(t, fullbasePath+"create-instance-schema.json")

	schemaUpdate := unmarshalJSONTestdata(t, fullbasePath+"update-instance-schema.json")

	schemaBind := unmarshalJSONTestdata(t, fullbasePath+"bind-instance-schema.json")

	bindTemplate := loadRawTestdata(t, fullbasePath+"bind.yaml")

	return internal.BundlePlan{
		ID:          internal.BundlePlanID(meta.ID),
		Description: meta.Description,
		Metadata: internal.BundlePlanMetadata{
			DisplayName: meta.DisplayName,
		},
		Bindable: meta.Bindable,
		Name:     internal.BundlePlanName(meta.Name),
		ChartRef: cRef,
		Schemas: map[internal.PlanSchemaType]internal.PlanSchema{
			internal.SchemaTypeProvision: schemaCreate,
			internal.SchemaTypeUpdate:    schemaUpdate,
			internal.SchemaTypeBind:      schemaBind,
		},
		ChartValues:  internal.ChartValues(chartVal),
		BindTemplate: bindTemplate,
	}
}

func unmarshalYamlTestdata(t *testing.T, filename string, unmarshalTo interface{}) {
	b, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	err = yaml.Unmarshal(b, unmarshalTo)
	require.NoError(t, err)
}

func unmarshalJSONTestdata(t *testing.T, filename string) internal.PlanSchema {
	b, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	schema := new(internal.PlanSchema)
	err = json.Unmarshal(b, schema)
	require.NoError(t, err)
	return *schema
}

func loadRawTestdata(t *testing.T, filename string) []byte {
	b, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	return b
}
func assertDirNotExits(t *testing.T, path string) {
	_, err := os.Stat(path)
	if err == nil {
		t.Errorf("Directory %q stil exists", path)
	}
	if !os.IsNotExist(err) {
		t.Errorf("Got error while checking if dir %q exits: %v", path, err)
	}
}

// fakeRepository provide access to bundles repository
type fakeRepository struct {
	path string
}

// IndexReader returns index.yaml file from fake repository
func (p *fakeRepository) IndexReader(URL string) (io.ReadCloser, error) {
	p.path = URL
	fName := fmt.Sprintf("%s/%s", p.path, "index.yaml")
	return os.Open(fName)
}

// BundleReader returns body reader for the given bundle
func (p *fakeRepository) BundleReader(name bundle.Name, version bundle.Version) (io.ReadCloser, error) {
	return os.Open(p.URLForBundle(name, version))
}

// URLForBundle returns download url for given bundle
func (p *fakeRepository) URLForBundle(name bundle.Name, version bundle.Version) string {
	return fmt.Sprintf("%s/%s-%s.tgz", p.path, name, version)
}
