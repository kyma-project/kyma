package testing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
)

func TestInstanceGet(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.Get(exp.ID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertInstanceEqual(exp, got)
	})

	tRunDrivers(t, "Failure/NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceTestSuite(t, sf)
		ts.PopulateStorage()

		// WHEN:
		got, err := ts.s.Get(internal.InstanceID("non-existing-iID"))

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestInstanceInsert(t *testing.T) {
	tRunDrivers(t, "Success/New", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Insert(fix)

		// THEN:
		assert.NoError(t, err)
	})

	tRunDrivers(t, "Failure/AlreadyExist", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		ts.s.Insert(fix)

		// WHEN:
		fixNew := ts.MustCopyFixture(fix)
		err := ts.s.Insert(fixNew)

		// THEN:
		ts.AssertAlreadyExistsError(err)
	})

	tRunDrivers(t, "Failure/EmptyID", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		fix.ID = internal.InstanceID("")

		// WHEN:
		err := ts.s.Insert(fix)

		// THEN:
		assert.EqualError(t, err, "instance id must be set")
	})
}

func TestInstanceRemove(t *testing.T) {
	tRunDrivers(t, "Success", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Remove(exp.ID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertInstanceDoesNotExist(exp)
	})

	tRunDrivers(t, "Failure/NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Remove(exp.ID)

		// THEN:
		ts.AssertNotFoundError(err)
	})
}

func newInstanceTestSuite(t *testing.T, sf storage.Factory) *instanceTestSuite {
	ts := instanceTestSuite{
		t:                   t,
		s:                   sf.Instance(),
		fixtures:            make(map[internal.InstanceID]*internal.Instance),
		fixturesSymToKeyMap: make(map[string]internal.InstanceID),
	}

	ts.generateFixtures()

	return &ts
}

type instanceTestSuite struct {
	t                   *testing.T
	s                   storage.Instance
	fixtures            map[internal.InstanceID]*internal.Instance
	fixturesSymToKeyMap map[string]internal.InstanceID
}

func (ts *instanceTestSuite) generateFixtures() {
	for fs, ft := range map[string]struct{ id, sID, spID, rName, pHash string }{
		"A1": {"id-01", "sID-01", "spID-01", "rName-01-01", "pHash-01"},
		"A2": {"id-02", "sID-01", "spID-01", "rName-01-02", "pHash-02"},
		"A3": {"id-03", "sID-03", "spID-03", "rName-03", "pHash-03"},
	} {
		i := &internal.Instance{
			ID:            internal.InstanceID(ft.id),
			ServiceID:     internal.ServiceID(ft.sID),
			ServicePlanID: internal.ServicePlanID(ft.spID),
			ReleaseName:   internal.ReleaseName(ft.rName),
			ParamsHash:    ft.pHash,
		}

		ts.fixtures[i.ID] = i
		ts.fixturesSymToKeyMap[fs] = i.ID
	}
}

func (ts *instanceTestSuite) PopulateStorage() {
	for _, b := range ts.fixtures {
		ts.s.Insert(ts.MustCopyFixture(b))
	}
}

func (ts *instanceTestSuite) MustGetFixture(sym string) *internal.Instance {
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
func (ts *instanceTestSuite) MustCopyFixture(in *internal.Instance) *internal.Instance {
	return &internal.Instance{
		ID:            in.ID,
		ServiceID:     in.ServiceID,
		ServicePlanID: in.ServicePlanID,
		ReleaseName:   in.ReleaseName,
		ParamsHash:    in.ParamsHash,
	}
}

func (ts *instanceTestSuite) AssertInstanceEqual(exp, got *internal.Instance) bool {
	ts.t.Helper()
	return assert.EqualValues(ts.t, exp, got)
}

func (ts *instanceTestSuite) AssertNotFoundError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}

func (ts *instanceTestSuite) AssertAlreadyExistsError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsAlreadyExistsError(err), "AlreadyExists error expected")
}

func (ts *instanceTestSuite) AssertInstanceDoesNotExist(i *internal.Instance) bool {
	ts.t.Helper()
	_, err := ts.s.Get(i.ID)
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}
