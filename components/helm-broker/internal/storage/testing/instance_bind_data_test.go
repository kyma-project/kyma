package testing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
)

func TestInstanceBindDataGet(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceBindDataTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("single")

		// WHEN:
		got, err := ts.s.Get(exp.InstanceID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertInstanceEqual(exp, got)
	})

	tRunDrivers(t, "Failure/NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceBindDataTestSuite(t, sf)
		ts.PopulateStorage()

		// WHEN:
		got, err := ts.s.Get(internal.InstanceID("non-existing-iID"))

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestInstanceBindDataInsert(t *testing.T) {
	tRunDrivers(t, "Success/New", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceBindDataTestSuite(t, sf)
		fix := ts.MustGetFixture("single")

		// WHEN:
		err := ts.s.Insert(fix)

		// THEN:
		assert.NoError(t, err)
	})

	tRunDrivers(t, "Failure/AlreadyExist", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceBindDataTestSuite(t, sf)
		fix := ts.MustGetFixture("single")
		ts.s.Insert(fix)

		// WHEN:
		fixNew := ts.MustCopyFixture(fix)
		err := ts.s.Insert(fixNew)

		// THEN:
		ts.AssertAlreadyExistsError(err)
	})

	tRunDrivers(t, "Failure/EmptyID", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceBindDataTestSuite(t, sf)
		fix := ts.MustGetFixture("single")
		fix.InstanceID = internal.InstanceID("")

		// WHEN:
		err := ts.s.Insert(fix)

		// THEN:
		assert.EqualError(t, err, "instance id must be set")
	})
}

func TestInstanceBindDataRemove(t *testing.T) {
	tRunDrivers(t, "Success", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceBindDataTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("single")

		// WHEN:
		err := ts.s.Remove(exp.InstanceID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertInstanceBindDataDoesNotExist(exp)
	})

	tRunDrivers(t, "Failure/NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInstanceBindDataTestSuite(t, sf)
		ts.PopulateStorage()

		// WHEN:
		err := ts.s.Remove(internal.InstanceID("non-existing-iID"))

		// THEN:
		ts.AssertNotFoundError(err)
	})
}

func newInstanceBindDataTestSuite(t *testing.T, sf storage.Factory) *instanceBindDataTestSuite {
	ts := instanceBindDataTestSuite{
		t:                   t,
		s:                   sf.InstanceBindData(),
		fixtures:            make(map[internal.InstanceID]*internal.InstanceBindData),
		fixturesSymToKeyMap: make(map[string]internal.InstanceID),
	}

	ts.generateFixtures()

	return &ts
}

type instanceBindDataTestSuite struct {
	t                   *testing.T
	s                   storage.InstanceBindData
	fixtures            map[internal.InstanceID]*internal.InstanceBindData
	fixturesSymToKeyMap map[string]internal.InstanceID
}

func (ts *instanceBindDataTestSuite) generateFixtures() {
	for fs, ft := range map[string]struct {
		id   string
		cred map[string]string
	}{
		"single":   {"id-01", map[string]string{"c1": "v1"}},
		"multiple": {"id-02", map[string]string{"c1": "v1", "c2": "v2"}},
		"empty":    {"id-03", map[string]string{}},
	} {
		cred := make(internal.InstanceCredentials)

		i := &internal.InstanceBindData{
			InstanceID:  internal.InstanceID(ft.id),
			Credentials: cred,
		}

		ts.fixtures[i.InstanceID] = i
		ts.fixturesSymToKeyMap[fs] = i.InstanceID
	}
}

func (ts *instanceBindDataTestSuite) PopulateStorage() {
	for _, b := range ts.fixtures {
		ts.s.Insert(ts.MustCopyFixture(b))
	}
}

func (ts *instanceBindDataTestSuite) MustGetFixture(sym string) *internal.InstanceBindData {
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
func (ts *instanceBindDataTestSuite) MustCopyFixture(in *internal.InstanceBindData) *internal.InstanceBindData {
	out := &internal.InstanceBindData{
		InstanceID:  in.InstanceID,
		Credentials: make(internal.InstanceCredentials),
	}

	for i := range in.Credentials {
		val := in.Credentials[i]
		out.Credentials[i] = val
	}

	return out
}

func (ts *instanceBindDataTestSuite) AssertInstanceEqual(exp, got *internal.InstanceBindData) bool {
	ts.t.Helper()
	return assert.EqualValues(ts.t, exp, got)
}

func (ts *instanceBindDataTestSuite) AssertNotFoundError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}

func (ts *instanceBindDataTestSuite) AssertAlreadyExistsError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsAlreadyExistsError(err), "AlreadyExists error expected")
}

func (ts *instanceBindDataTestSuite) AssertInstanceBindDataDoesNotExist(i *internal.InstanceBindData) bool {
	ts.t.Helper()
	_, err := ts.s.Get(i.InstanceID)
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}
