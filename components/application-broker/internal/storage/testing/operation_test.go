package testing

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
)

func TestInstanceOperationGet(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("i01/o01/Create/h1/InProgress")

		// WHEN:
		got, err := ts.s.Get(exp.InstanceID, exp.OperationID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertInstanceOperationEqualWithoutCreatedAt(exp, got)
	})

	tRunDrivers(t, "NotFound/Instance", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("i01/o01/Create/h1/InProgress")

		// WHEN:
		got, err := ts.s.Get(internal.InstanceID("non-existing-iID"), exp.OperationID)

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})

	tRunDrivers(t, "NotFound/Operation", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("i01/o01/Create/h1/InProgress")

		// WHEN:
		got, err := ts.s.Get(exp.InstanceID, internal.OperationID("non-existing-opID"))

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})

	tRunDrivers(t, "NotFound/InstanceAndOperation", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		ts.PopulateStorage()

		// WHEN:
		got, err := ts.s.Get(internal.InstanceID("non-existing-iID"), internal.OperationID("non-existing-opID"))

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestInstanceOperationInsert(t *testing.T) {
	tRunDrivers(t, "Success", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		fix := ts.MustGetFixture("i01/o01/Create/h1/InProgress")

		// WHEN:
		err := ts.s.Insert(fix)

		// THEN:
		assert.NoError(t, err)
	})

	tRunDrivers(t, "Failure/AlreadyExist", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		fix := ts.MustGetFixture("i01/o01/Create/h1/InProgress")
		ts.s.Insert(fix)

		// WHEN:
		fixNew := ts.MustCopyFixture(fix)
		err := ts.s.Insert(fixNew)

		// THEN:
		ts.AssertAlreadyExistsError(err)
	})

	tRunDrivers(t, "Failure/EmptyInstanceID", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		fix := ts.MustGetFixture("i01/o01/Create/h1/InProgress")
		fix.InstanceID = internal.InstanceID("")

		// WHEN:
		err := ts.s.Insert(fix)

		// THEN:
		assert.EqualError(t, err, "both instance and operation id must be set")
	})

	tRunDrivers(t, "Failure/EmptyOperationID", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		fix := ts.MustGetFixture("i01/o01/Create/h1/InProgress")
		fix.OperationID = internal.OperationID("")

		// WHEN:
		err := ts.s.Insert(fix)

		// THEN:
		assert.EqualError(t, err, "both instance and operation id must be set")
	})

}

func TestInstanceOperationInsertGet(t *testing.T) {
	tRunDrivers(t, "Success/TimeSetInInstance", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)

		fixTime := time.Date(2020, 1, 1, 0, 0, 1, 0, time.UTC)
		ts.s.WithTimeProvider(func() time.Time { return fixTime })

		fixOp := ts.MustGetFixture("i01/o01/Create/h1/InProgress")

		// WHEN:
		err := ts.s.Insert(fixOp)

		// THEN:
		assert.NoError(t, err)
		assert.True(t, fixOp.CreatedAt.Equal(fixTime))
	})

	tRunDrivers(t, "Success/TimeSetInStorage", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)

		fixTime := time.Date(2020, 1, 1, 0, 0, 1, 0, time.UTC)
		ts.s.WithTimeProvider(func() time.Time { return fixTime })

		fixOp := ts.MustGetFixture("i01/o01/Create/h1/InProgress")

		// WHEN:
		err := ts.s.Insert(fixOp)

		// THEN:
		assert.NoError(t, err)

		gotOp, err := ts.s.Get(fixOp.InstanceID, fixOp.OperationID)
		require.NoError(t, err)

		assert.True(t, gotOp.CreatedAt.Equal(fixTime))
	})
}

func TestInstanceOperationGetAll(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		ts.PopulateStorage()
		iID := internal.InstanceID("iID-002")
		expOpIDs := []internal.OperationID{"oID-002-002", "oID-002-003"}

		// WHEN:
		got, err := ts.s.GetAll(iID)

		// THEN:
		assert.NoError(t, err)

		require.Len(t, got, len(expOpIDs))
		for _, op := range got {
			assert.Contains(t, expOpIDs, op.OperationID)
		}
	})

	tRunDrivers(t, "NotFound/Instance", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		ts.PopulateStorage()

		// WHEN:
		got, err := ts.s.GetAll(internal.InstanceID("iID-non-existent"))

		// THEN:
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestInstanceOperationFindLast(t *testing.T) {
	tRunDrivers(t, "FindLast", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		ts.PopulateStorage()
		iID := internal.InstanceID("iID-002")
		fixOp := ts.MustGetFixture("i02/o03/Remove/h3/InProgress")

		// WHEN:
		got, err := ts.s.GetLast(iID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertInstanceOperationEqualWithoutCreatedAt(got, fixOp)
	})

}

func TestInstanceOperationGetAllSortOrder(t *testing.T) {
	tRunDrivers(t, "GetAllSorted", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		ts.PopulateStorage()
		iID := internal.InstanceID("iID-002")
		fixOp := ts.MustGetFixture("i02/o03/Remove/h3/InProgress")

		// WHEN:
		got, err := ts.s.GetAll(iID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertInstanceOperationEqualWithoutCreatedAt(fixOp, got[0])
	})

}

func TestInstanceOperationUpdateState(t *testing.T) {
	tRunDrivers(t, "Success", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		fix := ts.MustGetFixture("i01/o01/Create/h1/InProgress")
		ts.s.Insert(fix)

		exp := ts.MustCopyFixture(fix)
		exp.State = internal.OperationStateSucceeded
		exp.StateDescription = nil

		// WHEN:
		err := ts.s.UpdateState(fix.InstanceID, fix.OperationID, internal.OperationStateSucceeded)

		// THEN:
		assert.NoError(t, err)

		got, err := ts.s.Get(fix.InstanceID, fix.OperationID)
		require.NoError(t, err)

		ts.AssertInstanceOperationEqualWithoutCreatedAt(exp, got)
	})
}

func TestInstanceOperationUpdateStateDesc(t *testing.T) {
	tRunDrivers(t, "Success/DescSet", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		fix := ts.MustGetFixture("i01/o01/Create/h1/InProgress")
		ts.s.Insert(fix)

		expDesc := "updated desc"
		exp := ts.MustCopyFixture(fix)
		exp.State = internal.OperationStateSucceeded
		exp.StateDescription = &expDesc

		// WHEN:
		expDescCpy := expDesc
		err := ts.s.UpdateStateDesc(fix.InstanceID, fix.OperationID, internal.OperationStateSucceeded, &expDescCpy)

		// THEN:
		assert.NoError(t, err)

		got, err := ts.s.Get(fix.InstanceID, fix.OperationID)
		require.NoError(t, err)

		ts.AssertInstanceOperationEqualWithoutCreatedAt(exp, got)
	})

	tRunDrivers(t, "Success/DescRemoved", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		fix := ts.MustGetFixture("i01/o01/Create/h1/InProgress")
		ts.s.Insert(fix)

		exp := ts.MustCopyFixture(fix)
		exp.State = internal.OperationStateSucceeded
		exp.StateDescription = nil

		// WHEN:
		err := ts.s.UpdateStateDesc(fix.InstanceID, fix.OperationID, internal.OperationStateSucceeded, nil)

		// THEN:
		assert.NoError(t, err)

		got, err := ts.s.Get(fix.InstanceID, fix.OperationID)
		require.NoError(t, err)

		ts.AssertInstanceOperationEqualWithoutCreatedAt(exp, got)
	})
}

func TestInstanceOperationRemove(t *testing.T) {
	tRunDrivers(t, "Success", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("i01/o01/Create/h1/InProgress")

		// WHEN:
		err := ts.s.Remove(exp.InstanceID, exp.OperationID)

		// THEN:
		assert.NoError(t, err)
		ts.AssertOperationDoesNotExist(exp)
	})

	tRunDrivers(t, "Failure/NotFound", func(t *testing.T, sf storage.Factory) {
		// GIVEN:
		ts := newInMemoryOperationTestSuite(t, sf)
		exp := ts.MustGetFixture("i01/o01/Create/h1/InProgress")

		// WHEN:
		err := ts.s.Remove(exp.InstanceID, exp.OperationID)

		// THEN:
		ts.AssertNotFoundError(err)
	})
}

func newInMemoryOperationTestSuite(t *testing.T, sf storage.Factory) *operationTestSuite {
	ts := operationTestSuite{
		t:                   t,
		s:                   sf.InstanceOperation(),
		fixtures:            make(map[instanceOperationRef]*internal.InstanceOperation),
		fixturesSymToRefMap: make(map[string]instanceOperationRef),
	}

	ts.generateFixtures()

	return &ts
}

type instanceOperationRef struct {
	InstanceID  internal.InstanceID
	OperationID internal.OperationID
}

type operationTestSuite struct {
	t                       *testing.T
	s                       storage.InstanceOperation
	fixtures                map[instanceOperationRef]*internal.InstanceOperation
	fixturesSymToRefMap     map[string]instanceOperationRef
	fixturesPopulationOrder []instanceOperationRef
}

func (ts *operationTestSuite) generateFixtures() {
	for _, ft := range []struct {
		sym          string
		iID, opID    string
		opType       internal.OperationType
		opState      internal.OperationState
		sDesc, pHash string
	}{
		// Order is important as storage will reject insert if there is in ptogress operation for given instance.
		{"i01/o01/Create/h1/InProgress", "iID-001", "oID-001-001", internal.OperationTypeCreate, internal.OperationStateInProgress, "state desc 001", "pHash-001"},
		{"i02/o02/Create/h2/Succeeded", "iID-002", "oID-002-002", internal.OperationTypeCreate, internal.OperationStateSucceeded, "state desc 002", "pHash-002"},
		{"i02/o03/Remove/h3/InProgress", "iID-002", "oID-002-003", internal.OperationTypeRemove, internal.OperationStateInProgress, "state desc 003", "pHash-003"},
	} {
		iID := internal.InstanceID(ft.iID)
		opID := internal.OperationID(ft.opID)

		ir := instanceOperationRef{
			InstanceID:  iID,
			OperationID: opID,
		}

		io := internal.InstanceOperation{
			InstanceID:       iID,
			OperationID:      opID,
			Type:             ft.opType,
			State:            ft.opState,
			StateDescription: &ft.sDesc,
		}

		ts.fixtures[ir] = &io
		ts.fixturesSymToRefMap[ft.sym] = ir
		ts.fixturesPopulationOrder = append(ts.fixturesPopulationOrder, ir)
	}
}

func (ts *operationTestSuite) PopulateStorage() {
	for _, ir := range ts.fixturesPopulationOrder {
		io := ts.fixtures[ir]
		err := ts.s.Insert(ts.MustCopyFixture(io))
		if err != nil {
			ts.t.Fatalf("populate storage failed, io: %v, err: %s", io, err)
		}
	}
}

func (ts *operationTestSuite) MustGetFixture(sym string) *internal.InstanceOperation {
	ref, found := ts.fixturesSymToRefMap[sym]
	if !found {
		panic(fmt.Sprintf("fixture symbol not found, sym: %s", sym))
	}

	b, found := ts.fixtures[ref]
	if !found {
		panic(fmt.Sprintf("fixture not found, sym: %s, ref: %v", sym, ref))
	}

	return b
}

func (ts *operationTestSuite) MustCopyFixture(in *internal.InstanceOperation) *internal.InstanceOperation {
	out := internal.InstanceOperation{
		InstanceID:  in.InstanceID,
		OperationID: in.OperationID,
		Type:        in.Type,
		State:       in.State,
		CreatedAt:   in.CreatedAt,
	}

	if in.StateDescription != nil {
		var sDescCpy string
		sDescCpy = *in.StateDescription
		out.StateDescription = &sDescCpy
	}

	return &out
}

// AssertInstanceOperationEqualWithoutCreatedAt performs partial match for bundle.
// It's suitable only for tests as match is PARTIAL.
func (ts *operationTestSuite) AssertInstanceOperationEqualWithoutCreatedAt(exp, got *internal.InstanceOperation) bool {
	ts.t.Helper()

	expSet := exp == nil
	gotSet := got == nil

	if expSet != gotSet {
		assert.Fail(ts.t, fmt.Sprintf("mismatch on operations existence, exp set: %t, got set: %t", expSet, gotSet))
		return false
	}

	expCpy := ts.MustCopyFixture(exp)
	expCpy.CreatedAt = time.Time{}
	gotCpy := ts.MustCopyFixture(got)
	gotCpy.CreatedAt = time.Time{}

	return assert.EqualValues(ts.t, expCpy, gotCpy)
}

func (ts *operationTestSuite) AssertNotFoundError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}

func (ts *operationTestSuite) AssertAlreadyExistsError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsAlreadyExistsError(err), "AlreadyExists error expected")
}

func (ts *operationTestSuite) AssertOperationDoesNotExist(op *internal.InstanceOperation) bool {
	ts.t.Helper()
	_, err := ts.s.Get(op.InstanceID, op.OperationID)
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}
