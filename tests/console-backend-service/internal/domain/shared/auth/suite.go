package auth

import (
	"log"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	rbacError string = "graphql: access denied"
	authError string = "graphql: server returned a non-200 status code: 401"
)

type TestSuite struct {
	client *graphql.Client
}

type OperationType int

const (
	Get OperationType = iota
	List
	Create
	Update
	Delete
)

type OperationsInput map[OperationType][]*graphql.Request

type testCases map[userType]struct {
	user          graphql.User
	hasPermission bool
}

type userType int

const (
	noUser userType = iota
	noRightsUser
	readOnlyUser
	readWriteUser
)

func (u userType) String() string {
	switch u {
	case noUser:
		return "NoUser"
	case noRightsUser:
		return "NoRightsUser"
	case readOnlyUser:
		return "ReadOnlyUser"
	case readWriteUser:
		return "ReadWriteUser"
	default:
		return ""
	}
}

func New() *TestSuite {
	c, err := graphql.New()
	if err != nil {
		log.Fatal(errors.Wrap(err, "while GraphQL client setup"))
	}

	return &TestSuite{
		client: c,
	}
}

func (a *TestSuite) Run(t *testing.T, ops *OperationsInput) {
	if ops == nil || len(*ops) == 0 {
		return
	}

	for op, reqs := range *ops {
		if reqs == nil {
			continue
		}

		for _, req := range reqs {
			if req == nil {
				continue
			}

			switch op {
			case Get:
				a.testGet(t, req)
			case List:
				a.testList(t, req)
			case Create:
				a.testCreate(t, req)
			case Update:
				a.testUpdate(t, req)
			case Delete:
				a.testDelete(t, req)
			default:
				log.Println("unknown operation type")
			}
		}
	}
}

func (a *TestSuite) testGet(t *testing.T, request *graphql.Request) {
	tc := testCases{
		noUser:        {graphql.NoUser, false},
		noRightsUser:  {graphql.NoRightsUser, false},
		readOnlyUser:  {graphql.ReadOnlyUser, true},
		readWriteUser: {graphql.AdminUser, true},
	}

	t.Run("Get", func(t *testing.T) {
		a.runTests(t, &tc, request)
	})
}

func (a *TestSuite) testList(t *testing.T, request *graphql.Request) {
	tc := testCases{
		noUser:        {graphql.NoUser, false},
		noRightsUser:  {graphql.NoRightsUser, false},
		readOnlyUser:  {graphql.ReadOnlyUser, true},
		readWriteUser: {graphql.AdminUser, true},
	}

	t.Run("List", func(t *testing.T) {
		a.runTests(t, &tc, request)
	})
}

func (a *TestSuite) testCreate(t *testing.T, request *graphql.Request) {
	tc := testCases{
		noUser:        {graphql.NoUser, false},
		noRightsUser:  {graphql.NoRightsUser, false},
		readOnlyUser:  {graphql.ReadOnlyUser, false},
		readWriteUser: {graphql.AdminUser, true},
	}

	t.Run("Create", func(t *testing.T) {
		a.runTests(t, &tc, request)
	})
}

func (a *TestSuite) testUpdate(t *testing.T, request *graphql.Request) {
	tc := testCases{
		noUser:        {graphql.NoUser, false},
		noRightsUser:  {graphql.NoRightsUser, false},
		readOnlyUser:  {graphql.ReadOnlyUser, false},
		readWriteUser: {graphql.AdminUser, true},
	}

	t.Run("Update", func(t *testing.T) {
		a.runTests(t, &tc, request)
	})
}

func (a *TestSuite) testDelete(t *testing.T, request *graphql.Request) {
	tc := testCases{
		noUser:        {graphql.NoUser, false},
		noRightsUser:  {graphql.NoRightsUser, false},
		readOnlyUser:  {graphql.ReadOnlyUser, false},
		readWriteUser: {graphql.AdminUser, true},
	}

	t.Run("Delete", func(t *testing.T) {
		a.runTests(t, &tc, request)
	})
}

func (a *TestSuite) runTests(t *testing.T, testCases *testCases, request *graphql.Request) {
	for testName, testCase := range *testCases {
		t.Run(testName.String(), func(t *testing.T) {
			a.changeUser(t, testCase.user)

			var res interface{}
			err := a.client.Do(request, res)

			if testCase.user == graphql.NoUser {
				assert.EqualError(t, err, authError)
				return
			}

			if testCase.hasPermission {
				assert.NotEqual(t, rbacError, err)
			} else {
				assert.EqualError(t, err, rbacError)
			}
		})
	}
}

func (a *TestSuite) changeUser(t *testing.T, user graphql.User) {
	err := a.client.ChangeUser(user)
	require.NoError(t, err)
}
