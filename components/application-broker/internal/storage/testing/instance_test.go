package testing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
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

func TestInstanceFindOne(t *testing.T) {
	tRunDrivers(t, "Success/Found/MatchNamespaceAndClass", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("B3")

		// WHEN:
		got, err := ts.s.FindOne(func(i *internal.Instance) bool {
			if i.Namespace != exp.Namespace {
				return false
			}
			if i.ServiceID != exp.ServiceID {
				return false
			}
			return true
		})

		// THEN:
		assert.NoError(t, err)
		ts.AssertInstanceEqual(exp, got)
	})

	tRunDrivers(t, "Success/NotFound/MatchNamespaceButNotClass", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("B3")

		// WHEN:
		got, err := ts.s.FindOne(func(i *internal.Instance) bool {
			if i.Namespace != exp.Namespace {
				return false
			}
			if i.ServiceID != internal.ServiceID("not-existing-service-id") {
				return false
			}
			return true
		})

		// THEN:
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestInstanceFindAll(t *testing.T) {
	tRunDrivers(t, "Success/Found/MatchID", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceTestSuite(t, sf)
		ts.PopulateStorage()
		fixA1 := ts.MustGetFixture("A1")
		fixA2 := ts.MustGetFixture("A2")
		fixB1 := ts.MustGetFixture("B3")

		// WHEN:
		got, err := ts.s.FindAll(func(i *internal.Instance) bool {
			switch i.ID {
			case fixA1.ID, fixA2.ID, fixB1.ID:
				return true
			default:
				return false
			}
		})

		// THEN:
		assert.NoError(t, err)
		ts.AssertContainsAll(got, fixA1, fixA2, fixB1)
	})

	tRunDrivers(t, "Success/NotFound/MatchIDButNotNamespace", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceTestSuite(t, sf)
		ts.PopulateStorage()

		fixA1 := ts.MustGetFixture("A1")
		fixB1 := ts.MustGetFixture("B3")

		// WHEN:
		got, err := ts.s.FindAll(func(i *internal.Instance) bool {
			if !(i.ID == fixA1.ID || i.ID == fixB1.ID) {
				return false
			}
			if i.Namespace != "not-existing-namespace" {
				return false
			}
			return true
		})

		// THEN:
		assert.NoError(t, err)
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
	for fs, ft := range map[string]struct{ ns, id, sID, spID, rName string }{
		"A1": {"ns-a", "id-01", "sID-01", "spID-01", "rName-01-01"},
		"A2": {"ns-a", "id-02", "sID-01", "spID-01", "rName-01-02"},
		"A3": {"ns-a", "id-03", "sID-03", "spID-03", "rName-03"},
		"B3": {"ns-b", "id-04", "sID-04", "spID-04", "rName-04"},
	} {
		i := &internal.Instance{
			Namespace:     internal.Namespace(ft.ns),
			ID:            internal.InstanceID(ft.id),
			ServiceID:     internal.ServiceID(ft.sID),
			ServicePlanID: internal.ServicePlanID(ft.spID),
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
		Namespace:     in.Namespace,
		ID:            in.ID,
		ServiceID:     in.ServiceID,
		ServicePlanID: in.ServicePlanID,
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

func (ts *instanceTestSuite) AssertContainsAll(got []*internal.Instance, exp ...*internal.Instance) {
	ts.t.Helper()

	assert.Len(ts.t, got, len(exp))

	for _, fix := range exp {
		assert.Contains(ts.t, got, fix)
	}
}
