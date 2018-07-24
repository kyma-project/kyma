package testing

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
)

func TestBundleGet(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.Get(exp.Name, exp.Version)

		// THEN:
		assert.NoError(t, err)
		ts.AssertBundleEqual(exp, got)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.Get(exp.Name, exp.Version)

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestBundleGetByID(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.GetByID(exp.ID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertBundleEqual(exp, got)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.GetByID(exp.ID)

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestBundleUpsert(t *testing.T) {
	tRunDrivers(t, "Success/New", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)
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
		ts := newBundleTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		ts.s.Upsert(fix)

		// WHEN:
		fixNew := ts.MustCopyFixture(fix)
		fixNew.Description = expDesc
		replace, err := ts.s.Upsert(fixNew)

		// THEN:
		assert.NoError(t, err)
		assert.True(t, replace)

		got, err := ts.s.GetByID(fixNew.ID)
		assert.NoError(t, err)
		ts.AssertBundleEqual(fixNew, got)
	})

	tRunDrivers(t, "Failure/EmptyVersion", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		fix.Version = semver.Version{}

		// WHEN:
		_, err := ts.s.Upsert(fix)

		// THEN:
		assert.EqualError(t, err, "both name and version must be set")
	})
}

func TestBundleRemove(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Remove(exp.Name, exp.Version)

		// THEN:
		assert.NoError(t, err)
		ts.AssertBundleDoesNotExist(exp)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Remove(exp.Name, exp.Version)

		// THEN:
		ts.AssertNotFoundError(err)
	})
}

func TestBundleRemoveByID(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.RemoveByID(exp.ID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertBundleDoesNotExist(exp)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.RemoveByID(exp.ID)

		// THEN:
		ts.AssertNotFoundError(err)
	})
}

func TestBundleFindAll(t *testing.T) {

	tRunDrivers(t, "NonEmpty", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)
		ts.PopulateStorage()

		// WHEN:
		got, err := ts.s.FindAll()

		// THEN:
		assert.NoError(t, err)
		ts.AssertBundlesReturned(got)
	})

	tRunDrivers(t, "Empty", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newBundleTestSuite(t, sf)

		// WHEN:
		got, err := ts.s.FindAll()

		// THEN:
		assert.Empty(t, got)
		assert.NoError(t, err)
	})
}

func newBundleTestSuite(t *testing.T, sf storage.Factory) *bundleTestSuite {
	ts := bundleTestSuite{
		t:                  t,
		s:                  sf.Bundle(),
		fixtures:           make(map[internal.BundleID]*internal.Bundle),
		fixturesSymToIDMap: make(map[string]internal.BundleID),
	}

	ts.generateFixtures()

	return &ts
}

type bundleTestSuite struct {
	t                  *testing.T
	s                  storage.Bundle
	fixtures           map[internal.BundleID]*internal.Bundle
	fixturesSymToIDMap map[string]internal.BundleID
}

func (ts *bundleTestSuite) generateFixtures() {
	for fs, ft := range map[string]struct{ id, name, version, desc string }{
		"A1": {"id-A-001", "name-A", "0.0.1", "desc-A-001"},
		"A2": {"id-A-002", "name-A", "0.0.2", "desc-A-002"},
		"B1": {"id-B-001", "name-B", "0.0.1", "desc-B-001"},
		"B2": {"id-B-002", "name-B", "0.0.2", "desc-B-002"},
	} {
		b := &internal.Bundle{
			ID:          internal.BundleID(ft.id),
			Name:        internal.BundleName(ft.name),
			Version:     *semver.MustParse(ft.version),
			Description: ft.desc,
		}

		ts.fixtures[b.ID] = b
		ts.fixturesSymToIDMap[fs] = b.ID
	}
}

func (ts *bundleTestSuite) PopulateStorage() {
	for _, b := range ts.fixtures {
		ts.s.Upsert(ts.MustCopyFixture(b))
	}
}

func (ts *bundleTestSuite) MustGetFixture(sym string) *internal.Bundle {
	id, found := ts.fixturesSymToIDMap[sym]
	if !found {
		panic(fmt.Sprintf("fixture symbol not found, sym: %s", sym))
	}

	b, found := ts.fixtures[id]
	if !found {
		panic(fmt.Sprintf("fixture not found, sym: %s, id: %s", sym, id))
	}

	return b
}

// CopyFixture is copying fixture
// BEWARE: not all fields are copied, only those currently used in this test suite scope
func (ts *bundleTestSuite) MustCopyFixture(in *internal.Bundle) *internal.Bundle {
	return &internal.Bundle{
		ID:          in.ID,
		Name:        in.Name,
		Version:     *semver.MustParse(in.Version.String()),
		Description: in.Description,
	}
}

// AssertBundleEqual performs partial match for bundle.
// It's suitable only for tests as match is PARTIAL.
func (ts *bundleTestSuite) AssertBundleEqual(exp, got *internal.Bundle) bool {
	ts.t.Helper()

	result := assert.Equal(ts.t, exp.ID, got.ID, "mismatch on ID")
	result = assert.Equal(ts.t, exp.Name, got.Name, "mismatch on Name") && result
	result = assert.True(ts.t, exp.Version.Equal(&got.Version), "mismatch on Version") && result
	result = assert.Equal(ts.t, exp.Description, got.Description, "mismatch on Description") && result

	return result
}

func (ts *bundleTestSuite) AssertBundlesReturned(got []*internal.Bundle) bool {
	ts.t.Helper()

	result := true

	fixturesToMatch := make(map[internal.BundleID]struct{})
	for id := range ts.fixtures {
		fixturesToMatch[id] = struct{}{}
	}

	for _, bGot := range got {
		if _, found := fixturesToMatch[bGot.ID]; !found {
			assert.Fail(ts.t, "unexpected fixtures returned, ID: %s", bGot.ID)
			result = false
			continue
		}

		delete(fixturesToMatch, bGot.ID)
		bExp, _ := ts.fixtures[bGot.ID]
		result = result && ts.AssertBundleEqual(bExp, bGot)
	}

	result = result && assert.Empty(ts.t, fixturesToMatch, "not all expected fixtures matched")

	return result
}

func (ts *bundleTestSuite) AssertNotFoundError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}

func (ts *bundleTestSuite) AssertBundleDoesNotExist(b *internal.Bundle) bool {
	ts.t.Helper()

	_, err := ts.s.GetByID(b.ID)
	result := assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected on GetByID")

	_, err = ts.s.Get(b.Name, b.Version)
	result = assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected on Get") && result

	return result
}
