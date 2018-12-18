package testing

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
)

func TestRemoteEnvironmentGet(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newRETestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// when
		got, err := ts.reStorage.Get(internal.RemoteEnvironmentName(exp.Name))

		// then
		assert.NoError(t, err)
		assert.EqualValues(t, exp, got)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newRETestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// when
		got, err := ts.reStorage.Get(internal.RemoteEnvironmentName(exp.Name))

		// then
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestRemoteEnvironmentFindAll(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newRETestSuite(t, sf)
		ts.PopulateStorage()

		// when
		got, err := ts.reStorage.FindAll()

		// then
		assert.NoError(t, err)
		ts.AssertContainsAllFixtures(got)
	})
}

func TestRemoteEnvironmentFindOneByServiceID(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newRETestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// when
		got, err := ts.reStorage.FindOneByServiceID(exp.Services[0].ID)

		// then
		assert.NoError(t, err)
		assert.EqualValues(t, exp, got)
	})
	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newRETestSuite(t, sf)
		ts.PopulateStorage()

		// when
		got, err := ts.reStorage.FindOneByServiceID(internal.RemoteServiceID("apud"))

		// then
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestRemoteEnvironmentUpsert(t *testing.T) {
	tRunDrivers(t, "Success/New", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newRETestSuite(t, sf)
		fix := ts.MustGetFixture("A1")

		// when
		replace, err := ts.reStorage.Upsert(fix)

		// then
		assert.NoError(t, err)
		assert.False(t, replace)
	})

	tRunDrivers(t, "Success/Replace", func(t *testing.T, sf storage.Factory) {
		// given
		expDesc := "updated description"
		ts := newRETestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		ts.reStorage.Upsert(fix)

		// when
		fixNew := ts.MustCopyFixture(fix)
		fixNew.Description = expDesc
		replace, err := ts.reStorage.Upsert(fixNew)

		// then
		assert.NoError(t, err)
		assert.True(t, replace)

		got, err := ts.reStorage.Get(internal.RemoteEnvironmentName(fixNew.Name))
		assert.NoError(t, err)
		assert.EqualValues(t, fixNew, got)
	})

	tRunDrivers(t, "Failure/EmptyName", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newRETestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		fix.Name = ""

		// when
		_, err := ts.reStorage.Upsert(fix)

		// then
		assert.EqualError(t, err, "name must be set")
	})
}

func TestRemoteEnvironmentRemove(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newRETestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// when
		err := ts.reStorage.Remove(internal.RemoteEnvironmentName(exp.Name))

		// then
		assert.NoError(t, err)
		ts.AssertChartDoesNotExist(exp)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newRETestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// when
		err := ts.reStorage.Remove(internal.RemoteEnvironmentName(exp.Name))

		// then
		ts.AssertNotFoundError(err)
	})
}

func newRETestSuite(t *testing.T, sf storage.Factory) *reTestSuite {
	ts := reTestSuite{
		t:         t,
		reStorage: sf.RemoteEnvironment(),
		fixtures:  make(map[string]*internal.RemoteEnvironment),
	}

	ts.generateFixtures()

	return &ts
}

type reTestSuite struct {
	t         *testing.T
	reStorage storage.RemoteEnvironment
	fixtures  map[string]*internal.RemoteEnvironment
}

func (ts *reTestSuite) generateFixtures() {
	for fs, ft := range map[string]struct{ name, id, desc string }{
		"A1": {"name-A1", "service-id-A1", "desc-A1"},
		"A2": {"name-A2", "service-id-A2", "desc-A2"},
		"B1": {"name-B1", "service-id-B1", "desc-B1"},
		"B2": {"name-B2", "service-id-B2", "desc-B2"},
	} {
		re := &internal.RemoteEnvironment{
			Name:        internal.RemoteEnvironmentName(ft.name),
			Description: ft.desc,
			Services: []internal.Service{
				{
					ID:          internal.RemoteServiceID(ft.id),
					DisplayName: "displayName",
					APIEntry: &internal.APIEntry{
						AccessLabel: "access-label",
						GatewayURL:  "http://gateway.io",
						Entry: internal.Entry{
							Type: "API",
						},
					},
					Tags: []string{"tag1", "tag2"},
				},
			},
		}

		ts.fixtures[fs] = re
	}
}

func (ts *reTestSuite) PopulateStorage() {
	for _, fix := range ts.fixtures {
		ts.reStorage.Upsert(fix)
	}
}

func (ts *reTestSuite) MustGetFixture(name string) *internal.RemoteEnvironment {
	fix, found := ts.fixtures[name]
	if !found {
		panic(fmt.Sprintf("fixture with name %q not found", name))
	}

	return ts.MustCopyFixture(fix)
}

func (ts *reTestSuite) MustCopyFixture(in *internal.RemoteEnvironment) *internal.RemoteEnvironment {
	m, err := json.Marshal(in)
	if err != nil {
		panic(fmt.Sprintf("input remote environemnt marchaling failed, err: %s", err))
	}

	var out internal.RemoteEnvironment
	if err := json.Unmarshal(m, &out); err != nil {
		panic(fmt.Sprintf("input remote environemnt unmarchaling failed, err: %s", err))
	}

	return &out
}
func (ts *reTestSuite) AssertNotFoundError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}

func (ts *reTestSuite) AssertChartDoesNotExist(exp *internal.RemoteEnvironment) bool {
	ts.t.Helper()
	_, err := ts.reStorage.Get(internal.RemoteEnvironmentName(exp.Name))
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}

func (ts *reTestSuite) AssertContainsAllFixtures(got []*internal.RemoteEnvironment) {
	ts.t.Helper()

	assert.Len(ts.t, got, len(ts.fixtures))

	for _, fix := range ts.fixtures {
		assert.Contains(ts.t, got, fix)
	}
}
