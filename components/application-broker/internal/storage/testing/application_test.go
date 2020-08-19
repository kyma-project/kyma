package testing

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
)

func TestApplicationGet(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newAppTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// when
		got, err := ts.appStorage.Get(internal.ApplicationName(exp.Name))

		// then
		assert.NoError(t, err)
		assert.EqualValues(t, exp, got)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newAppTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// when
		got, err := ts.appStorage.Get(internal.ApplicationName(exp.Name))

		// then
		ts.AssertNotFoundError(err)
		assert.Nil(t, got)
	})
}

func TestApplicationFindAll(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newAppTestSuite(t, sf)
		ts.PopulateStorage()

		// when
		got, err := ts.appStorage.FindAll()

		// then
		assert.NoError(t, err)
		ts.AssertContainsAllFixtures(got)
	})
}

func TestApplicationFindOneByServiceID(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newAppTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// when
		got, err := ts.appStorage.FindOneByServiceID(exp.Services[0].ID)

		// then
		assert.NoError(t, err)
		assert.EqualValues(t, exp, got)
	})
	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newAppTestSuite(t, sf)
		ts.PopulateStorage()

		// when
		got, err := ts.appStorage.FindOneByServiceID(internal.ApplicationServiceID("apud"))

		// then
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestApplicationUpsert(t *testing.T) {
	tRunDrivers(t, "Success/New", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newAppTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")

		// when
		replace, err := ts.appStorage.Upsert(fix)

		// then
		assert.NoError(t, err)
		assert.False(t, replace)
	})

	tRunDrivers(t, "Success/Replace", func(t *testing.T, sf storage.Factory) {
		// given
		expDesc := "updated description"
		ts := newAppTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		ts.appStorage.Upsert(fix)

		// when
		fixNew := ts.MustCopyFixture(fix)
		fixNew.Description = expDesc
		replace, err := ts.appStorage.Upsert(fixNew)

		// then
		assert.NoError(t, err)
		assert.True(t, replace)

		got, err := ts.appStorage.Get(internal.ApplicationName(fixNew.Name))
		assert.NoError(t, err)
		assert.EqualValues(t, fixNew, got)
	})

	tRunDrivers(t, "Failure/EmptyName", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newAppTestSuite(t, sf)
		fix := ts.MustGetFixture("A1")
		fix.Name = ""

		// when
		_, err := ts.appStorage.Upsert(fix)

		// then
		assert.EqualError(t, err, "name must be set")
	})
}

func TestApplicationRemove(t *testing.T) {
	tRunDrivers(t, "Found", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newAppTestSuite(t, sf)
		ts.PopulateStorage()
		exp := ts.MustGetFixture("A1")

		// when
		err := ts.appStorage.Remove(internal.ApplicationName(exp.Name))

		// then
		assert.NoError(t, err)
		ts.AssertChartDoesNotExist(exp)
	})

	tRunDrivers(t, "NotFound", func(t *testing.T, sf storage.Factory) {
		// given
		ts := newAppTestSuite(t, sf)
		exp := ts.MustGetFixture("A1")

		// when
		err := ts.appStorage.Remove(internal.ApplicationName(exp.Name))

		// then
		ts.AssertNotFoundError(err)
	})
}

func newAppTestSuite(t *testing.T, sf storage.Factory) *appTestSuite {
	ts := appTestSuite{
		t:          t,
		appStorage: sf.Application(),
		fixtures:   make(map[string]*internal.Application),
	}

	ts.generateFixtures()

	return &ts
}

type appTestSuite struct {
	t          *testing.T
	appStorage storage.Application
	fixtures   map[string]*internal.Application
}

func (ts *appTestSuite) generateFixtures() {
	for fs, ft := range map[string]struct{ name, id, desc string }{
		"A1": {"name-A1", "service-id-A1", "desc-A1"},
		"A2": {"name-A2", "service-id-A2", "desc-A2"},
		"B1": {"name-B1", "service-id-B1", "desc-B1"},
		"B2": {"name-B2", "service-id-B2", "desc-B2"},
	} {
		app := &internal.Application{
			Name:        internal.ApplicationName(ft.name),
			Description: ft.desc,
			Services: []internal.Service{
				{
					ID:          internal.ApplicationServiceID(ft.id),
					DisplayName: "displayName",
					Entries: []internal.Entry{
						{
							Type: "API",
							APIEntry: &internal.APIEntry{
								GatewayURL:  "http://gateway.io",
								TargetURL:   "http://target.io",
								Name:        "api-mock",
								AccessLabel: "access-label",
							},
						},
						{
							Type: "Events",
						},
					},
					Tags: []string{"tag1", "tag2"},
				},
			},
		}

		ts.fixtures[fs] = app
	}
}

func (ts *appTestSuite) PopulateStorage() {
	for _, fix := range ts.fixtures {
		ts.appStorage.Upsert(fix)
	}
}

func (ts *appTestSuite) MustGetFixture(name string) *internal.Application {
	fix, found := ts.fixtures[name]
	if !found {
		panic(fmt.Sprintf("fixture with name %q not found", name))
	}

	return ts.MustCopyFixture(fix)
}

func (ts *appTestSuite) MustCopyFixture(in *internal.Application) *internal.Application {
	m, err := json.Marshal(in)
	if err != nil {
		panic(fmt.Sprintf("input application marchaling failed, err: %s", err))
	}

	var out internal.Application
	if err := json.Unmarshal(m, &out); err != nil {
		panic(fmt.Sprintf("input application unmarchaling failed, err: %s", err))
	}

	return &out
}
func (ts *appTestSuite) AssertNotFoundError(err error) bool {
	ts.t.Helper()
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}

func (ts *appTestSuite) AssertChartDoesNotExist(exp *internal.Application) bool {
	ts.t.Helper()
	_, err := ts.appStorage.Get(internal.ApplicationName(exp.Name))
	return assert.True(ts.t, storage.IsNotFoundError(err), "NotFound error expected")
}

func (ts *appTestSuite) AssertContainsAllFixtures(got []*internal.Application) {
	ts.t.Helper()

	assert.Len(ts.t, got, len(ts.fixtures))

	for _, fix := range ts.fixtures {
		assert.Contains(ts.t, got, fix)
	}
}
