package testing

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
)

func TestOperationGet(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newOperationTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		got, err := ts.s.Get(exp.InstanceID, exp.OperationID)
		got.CreatedAt = exp.CreatedAt
		// THEN:
		assert.NoError(t, err)
		ts.AssertOperationEqual(exp, got)
	})

	tRunDrivers(t, "Failure/NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newOperationTestSuite(t, sf)
		ts.PopulateStorage()

		// WHEN:
		got, err := ts.s.Get(internal.InstanceID("non-existin iID"), internal.OperationID("non-existing-oID"))

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestFindLast(t *testing.T) {
	tRunDrivers(t, "FindLast", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newOperationTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("B3")

		// WHEN:
		got, err := ts.s.GetLast(exp.InstanceID)
		got.CreatedAt = exp.CreatedAt

		// THEN:
		assert.NoError(t, err)
		ts.AssertOperationEqual(exp, got)
	})

}

func TestGetAllSortOrder(t *testing.T) {
	tRunDrivers(t, "GetAll", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newOperationTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("B3")

		// WHEN:
		got, err := ts.s.GetAll(exp.InstanceID)
		got[0].CreatedAt = exp.CreatedAt

		// THEN:
		assert.NoError(t, err)
		ts.AssertOperationEqual(exp, got[0])
	})

}

func TestOperationInsert(t *testing.T) {
	tRunDrivers(t, "Success/New", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newOperationTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Insert(fix)

		// THEN:
		assert.NoError(t, err)
	})

	tRunDrivers(t, "Failure/AlreadyExist", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newOperationTestSuite(t, sf)
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
		ts := newOperationTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		fix.OperationID = internal.OperationID("")
		fix.InstanceID = internal.InstanceID("")

		// WHEN:
		err := ts.s.Insert(fix)

		// THEN:
		assert.EqualError(t, err, "both instance and operation id must be set")
	})
}

func TestOperationRemove(t *testing.T) {
	tRunDrivers(t, "Success", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newOperationTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Remove(exp.InstanceID, exp.OperationID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertOperationDoesNotExist(exp)
	})

	tRunDrivers(t, "Failure/NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newOperationTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// WHEN:
		err := ts.s.Remove(exp.InstanceID, exp.OperationID)

		// THEN:
		ts.AssertNotFoundError(err)
	})
}

func newOperationTestSuite(t *testing.T, sf storage.Factory) *operationTestSuite {
	ts := operationTestSuite{
		t:                   t,
		s:                   sf.InstanceOperation(),
		fixtures:            make(map[internal.OperationID]*internal.InstanceOperation),
		fixturesSymToKeyMap: make(map[string]internal.OperationID),
	}

	ts.generateFixtures()

	return &ts
}

type operationTestSuite struct {
	t                   *testing.T
	s                   storage.InstanceOperation
	fixtures            map[internal.OperationID]*internal.InstanceOperation
	fixturesSymToKeyMap map[string]internal.OperationID
}

func (ts *operationTestSuite) generateFixtures() {
	for fs, ft := range map[string]struct{ iID, oID, oState, pHash  string }{
		"A1": {"iID-01", "oID-01", "state01",  "pHash-01"},
		"A2": {"iID-03", "oID-02", "state02", "pHash-02"},
		"A3": {"iID-03", "oID-03", "state03",  "pHash-03"},
		"B3": {"iID-03", "oID-04", "state04", "pHash-04"},
	} {
		i := &internal.InstanceOperation{
			OperationID:     internal.OperationID(ft.oID),
			InstanceID:            internal.InstanceID(ft.iID),
			State: internal.OperationState(ft.oState),
			ParamsHash:    ft.pHash,
		}
		ts.fixtures[i.OperationID] = i
		ts.fixturesSymToKeyMap[fs] = i.OperationID

		// CreatedAt field is populated based on current time.
		// For some tests we need these values to be different
		time.Sleep(5)
	}
}

func (ts *operationTestSuite) PopulateStorage() {
	for _, b := range ts.fixtures {
		ts.s.Insert(ts.MustCopyFixture(b))
	}
}

func (ts *operationTestSuite) MustGetFixture(sym string) *internal.InstanceOperation {
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
func (ts *operationTestSuite) MustCopyFixture(in *internal.InstanceOperation) *internal.InstanceOperation {
	return &internal.InstanceOperation{
		OperationID:     in.OperationID,
		InstanceID:            in.InstanceID,
		State:     in.State,
		CreatedAt: in.CreatedAt,
		ParamsHash:    in.ParamsHash,
	}
}

func (ts *operationTestSuite) AssertOperationEqual(exp, got *internal.InstanceOperation) bool {
	ts.t.Helper()
	return assert.EqualValues(ts.t, exp, got)
}

func (ts *operationTestSuite) AssertNotFoundError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}

func (ts *operationTestSuite) AssertAlreadyExistsError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsAlreadyExistsError(err), "AlreadyExists error expected")
}

func (ts *operationTestSuite) AssertOperationDoesNotExist(i *internal.InstanceOperation) bool {
	ts.t.Helper()
	_, err := ts.s.Get(i.InstanceID, i.OperationID)
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}
