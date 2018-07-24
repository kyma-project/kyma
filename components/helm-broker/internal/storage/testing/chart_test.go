package testing

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
)

func TestChartGet(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newChartTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.Get(internal.ChartName(exp.Metadata.Name), *semver.MustParse(exp.Metadata.Version))

		// THEN:
		assert.NoError(t, err)
		ts.AssertChartEqual(exp, got)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newChartTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.Get(internal.ChartName(exp.Metadata.Name), *semver.MustParse(exp.Metadata.Version))

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestChartUpsert(t *testing.T) {
	tRunDrivers(t, "Success/New", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newChartTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")

		// WHEN:
		replace, err := ts.s.Upsert(fix)

		// THEN:
		assert.NoError(t, err)
		assert.False(t, replace)
	})

	tRunDrivers(t, "Success/Replace", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		expDesc := "updated description"
		ts := newChartTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		ts.s.Upsert(fix)

		// WHEN:
		fixNew := ts.MustCopyFixture(fix)
		fixNew.Metadata.Description = expDesc
		replace, err := ts.s.Upsert(fixNew)

		// THEN:
		assert.NoError(t, err)
		assert.True(t, replace)

		got, err := ts.s.Get(internal.ChartName(fixNew.Metadata.Name), *semver.MustParse(fixNew.Metadata.Version))
		assert.NoError(t, err)
		ts.AssertChartEqual(fixNew, got)
	})

	tRunDrivers(t, "Failure/EmptyVersion", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newChartTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		fix.Metadata.Version = ""

		// WHEN:
		_, err := ts.s.Upsert(fix)

		// THEN:
		assert.EqualError(t, err, "both name and version must be set")
	})
}

func TestChartRemove(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newChartTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Remove(internal.ChartName(exp.Metadata.Name), *semver.MustParse(exp.Metadata.Version))

		// THEN:
		assert.NoError(t, err)
		ts.AssertChartDoesNotExist(exp)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newChartTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Remove(internal.ChartName(exp.Metadata.Name), *semver.MustParse(exp.Metadata.Version))

		// THEN:
		ts.AssertNotFoundError(err)
	})
}

func newChartTestSuite(t *testing.T, sf storage.Factory) *chartTestSuite {
	ts := chartTestSuite{
		t:                   t,
		s:                   sf.Chart(),
		fixtures:            make(map[chartNameVersion]*chart.Chart),
		fixturesSymToKeyMap: make(map[string]chartNameVersion),
	}

	ts.generateFixtures()

	return &ts
}

type chartNameVersion string

type chartTestSuite struct {
	t                   *testing.T
	s                   storage.Chart
	fixtures            map[chartNameVersion]*chart.Chart
	fixturesSymToKeyMap map[string]chartNameVersion
}

func (chartTestSuite) key(name internal.ChartName, ver semver.Version) chartNameVersion {
	return chartNameVersion(fmt.Sprintf("%s|%s", name, ver.String()))
}

func (chartTestSuite) mustKeyFromChart(c *chart.Chart) chartNameVersion {
	if c.Metadata == nil {
		panic("metadata must not be nil")
	}

	return chartNameVersion(fmt.Sprintf("%s|%s", c.Metadata.Name, c.Metadata.Version))
}

func (ts *chartTestSuite) generateFixtures() {
	for fs, ft := range map[string]struct{ id, name, version, desc string }{
		"A1": {"id-A-001", "name-A", "0.0.1", "desc-A-001"},
		"A2": {"id-A-002", "name-A", "0.0.2", "desc-A-002"},
		"B1": {"id-B-001", "name-B", "0.0.1", "desc-B-001"},
		"B2": {"id-B-002", "name-B", "0.0.2", "desc-B-002"},
	} {
		c := &chart.Chart{
			Metadata: &chart.Metadata{
				Name:        ft.name,
				Version:     ft.version,
				Description: ft.desc,
			},
		}

		k := ts.mustKeyFromChart(c)
		ts.fixtures[k] = c
		ts.fixturesSymToKeyMap[fs] = k
	}
}

func (ts *chartTestSuite) PopulateStorage() {
	for _, b := range ts.fixtures {
		ts.s.Upsert(ts.MustCopyFixture(b))
	}
}

func (ts *chartTestSuite) MustGetFixture(sym string) *chart.Chart {
	k, found := ts.fixturesSymToKeyMap[sym]
	if !found {
		panic(fmt.Sprintf("fixture symbol not found, sym: %s", sym))
	}

	b, found := ts.fixtures[k]
	if !found {
		panic(fmt.Sprintf("fixture not found, sym: %s, nameVersion: %s", sym, k))
	}

	return b
}

// CopyFixture is copying fixture
// BEWARE: not all fields are copied, only those currently used in this test suite scope
func (ts *chartTestSuite) MustCopyFixture(in *chart.Chart) *chart.Chart {
	m, err := json.Marshal(in)
	if err != nil {
		panic(fmt.Sprintf("input chart marchaling failed, err: %s", err))
	}

	var out chart.Chart
	if err := json.Unmarshal(m, &out); err != nil {
		panic(fmt.Sprintf("input chart unmarchaling failed, err: %s", err))
	}

	return &out
}

// AssertBundleEqual performs partial match for bundle.
// It's suitable only for tests as match is PARTIAL.
func (ts *chartTestSuite) AssertChartEqual(exp, got *chart.Chart) bool {
	ts.t.Helper()

	expSet := exp == nil
	gotSet := got == nil

	if expSet != gotSet {
		assert.Fail(ts.t, fmt.Sprintf("mismatch on charts existence, exp set: %t, got set: %t", expSet, gotSet))
		return false
	}

	expMetaSet := exp.Metadata == nil
	gotMetaSet := got.Metadata == nil

	if expMetaSet != gotMetaSet {
		assert.Fail(ts.t, fmt.Sprintf("mismatch on metadata status, exp set: %t, got set: %t", expMetaSet, gotMetaSet))
		return false
	}

	result := assert.Equal(ts.t, exp.Metadata.Name, got.Metadata.Name, "mismatch on Name")
	result = assert.Equal(ts.t, exp.Metadata.Version, got.Metadata.Version, "mismatch on Version") && result
	result = assert.Equal(ts.t, exp.Metadata.Description, got.Metadata.Description, "mismatch on Description") && result

	return result
}

func (ts *chartTestSuite) AssertNotFoundError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}

func (ts *chartTestSuite) AssertChartDoesNotExist(exp *chart.Chart) bool {
	ts.t.Helper()
	_, err := ts.s.Get(internal.ChartName(exp.Metadata.Name), *semver.MustParse(exp.Metadata.Version))
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}
