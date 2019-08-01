package testing

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
)

func TestAddonGet(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.Get(internal.ClusterWide, exp.Name, exp.Version)

		// THEN:
		assert.NoError(t, err)
		ts.AssertAddonEqual(exp, got)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.Get(internal.ClusterWide, exp.Name, exp.Version)

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestAddonGetByID(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.GetByID(internal.ClusterWide, exp.ID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertAddonEqual(exp, got)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.GetByID(internal.ClusterWide, exp.ID)

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestAddonUpsert(t *testing.T) {
	tRunDrivers(t, "Success/New", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")

		// WHEN:
		replace, err := ts.s.Upsert(internal.ClusterWide, fix)

		// THEN:
		assert.NoError(t, err)
		assert.False(t, replace)
	})

	tRunDrivers(t, "Success/Replace", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		expDesc := "updated description"
		ts := newAddonTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		ts.s.Upsert(internal.ClusterWide, fix)

		// WHEN:
		fixNew := ts.MustCopyFixture(fix)
		fixNew.Description = expDesc
		replace, err := ts.s.Upsert(internal.ClusterWide, fixNew)

		// THEN:
		assert.NoError(t, err)
		assert.True(t, replace)

		got, err := ts.s.GetByID(internal.ClusterWide, fixNew.ID)
		assert.NoError(t, err)
		ts.AssertAddonEqual(fixNew, got)
	})

	tRunDrivers(t, "Failure/EmptyVersion", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		fix.Version = semver.Version{}

		// WHEN:
		_, err := ts.s.Upsert(internal.ClusterWide, fix)

		// THEN:
		assert.EqualError(t, err, "both name and version must be set")
	})
}

func TestAddonRemove(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Remove(internal.ClusterWide, exp.Name, exp.Version)

		// THEN:
		assert.NoError(t, err)
		ts.AssertAddonDoesNotExist(exp)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Remove(internal.ClusterWide, exp.Name, exp.Version)

		// THEN:
		ts.AssertNotFoundError(err)
	})
}

func TestAddonRemoveByID(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.RemoveByID(internal.ClusterWide, exp.ID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertAddonDoesNotExist(exp)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.RemoveByID(internal.ClusterWide, exp.ID)

		// THEN:
		ts.AssertNotFoundError(err)
	})
}

func TestAddonFindAll(t *testing.T) {

	tRunDrivers(t, "NonEmpty", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)
		ts.PopulateStorage()
		ts.s.Upsert(internal.Namespace("stage"), &internal.Addon{
			ID:          internal.AddonID("id-0000"),
			Name:        internal.AddonName("other-addon"),
			Version:     *semver.MustParse("1.1.1"),
			Description: "",
		})

		// WHEN:
		got, err := ts.s.FindAll(internal.ClusterWide)

		// THEN:
		assert.NoError(t, err)
		ts.AssertAddonsReturned(got)
	})

	tRunDrivers(t, "Empty", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newAddonTestSuite(t, sf)

		// WHEN:
		got, err := ts.s.FindAll(internal.ClusterWide)

		// THEN:
		assert.Empty(t, got)
		assert.NoError(t, err)
	})
}

func newAddonTestSuite(t *testing.T, sf storage.Factory) *addonTestSuite {
	ts := addonTestSuite{
		t:                  t,
		s:                  sf.Addon(),
		fixtures:           make(map[internal.AddonID]*internal.Addon),
		fixturesSymToIDMap: make(map[string]internal.AddonID),
	}

	ts.generateFixtures()

	return &ts
}

type addonTestSuite struct {
	t                  *testing.T
	s                  storage.Addon
	fixtures           map[internal.AddonID]*internal.Addon
	fixturesSymToIDMap map[string]internal.AddonID
}

func (ts *addonTestSuite) generateFixtures() {
	for fs, ft := range map[string]struct{ id, name, version, desc string }{
		"A1": {"id-A-001", "name-A", "0.0.1", "desc-A-001"},
		"A2": {"id-A-002", "name-A", "0.0.2", "desc-A-002"},
		"B1": {"id-B-001", "name-B", "0.0.1", "desc-B-001"},
		"B2": {"id-B-002", "name-B", "0.0.2", "desc-B-002"},
	} {
		b := &internal.Addon{
			ID:          internal.AddonID(ft.id),
			Name:        internal.AddonName(ft.name),
			Version:     *semver.MustParse(ft.version),
			Description: ft.desc,
		}

		ts.fixtures[b.ID] = b
		ts.fixturesSymToIDMap[fs] = b.ID
	}
}

func (ts *addonTestSuite) PopulateStorage() {
	for _, b := range ts.fixtures {
		ts.s.Upsert(internal.ClusterWide, ts.MustCopyFixture(b))
	}
}

func (ts *addonTestSuite) MustGetFixture(sym string) *internal.Addon {
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
func (ts *addonTestSuite) MustCopyFixture(in *internal.Addon) *internal.Addon {
	return &internal.Addon{
		ID:          in.ID,
		Name:        in.Name,
		Version:     *semver.MustParse(in.Version.String()),
		Description: in.Description,
	}
}

// AssertAddonEqual performs partial match for addon.
// It's suitable only for tests as match is PARTIAL.
func (ts *addonTestSuite) AssertAddonEqual(exp, got *internal.Addon) bool {
	ts.t.Helper()

	result := assert.Equal(ts.t, exp.ID, got.ID, "mismatch on ID")
	result = assert.Equal(ts.t, exp.Name, got.Name, "mismatch on Name") && result
	result = assert.True(ts.t, exp.Version.Equal(&got.Version), "mismatch on Version") && result
	result = assert.Equal(ts.t, exp.Description, got.Description, "mismatch on Description") && result

	return result
}

func (ts *addonTestSuite) AssertAddonsReturned(got []*internal.Addon) bool {
	ts.t.Helper()

	result := true

	fixturesToMatch := make(map[internal.AddonID]struct{})
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
		result = result && ts.AssertAddonEqual(bExp, bGot)
	}

	result = result && assert.Empty(ts.t, fixturesToMatch, "not all expected fixtures matched")

	return result
}

func (ts *addonTestSuite) AssertNotFoundError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}

func (ts *addonTestSuite) AssertAddonDoesNotExist(b *internal.Addon) bool {
	ts.t.Helper()

	_, err := ts.s.GetByID(internal.ClusterWide, b.ID)
	result := assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected on GetByID")

	_, err = ts.s.Get(internal.ClusterWide, b.Name, b.Version)
	result = assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected on Get") && result

	return result
}
