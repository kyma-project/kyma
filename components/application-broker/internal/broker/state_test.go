package broker_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
	"github.com/pkg/errors"
)

func newInstanceStateServiceTestSuite(t *testing.T) *instanceStateServiceTestSuite {
	return &instanceStateServiceTestSuite{t: t}
}

type instanceStateServiceTestSuite struct {
	t   *testing.T
	Exp expAll
}

func (ts *instanceStateServiceTestSuite) SetUp() {
	ts.Exp.Populate()
}

func TestInstanceStateServiceIsProvisioned(t *testing.T) {
	for sym, tc := range map[string]struct {
		genOps func(ts *instanceStateServiceTestSuite) []*internal.InstanceOperation
		exp    bool
	}{
		"true/singleCreateSucceeded": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				return append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded))
			},
			exp: true,
		},
		"true/CreateSucceededThanRemoveInProgress": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded))
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateInProgress))
				return out
			},
			exp: true,
		},
		"false/singleCreateInProgress": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				return append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateInProgress))
			},
			exp: false,
		},
		"false/CreateSucceededThanRemoveSucceeded": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded))
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateSucceeded))
				return out
			},
			exp: false,
		},
	} {
		t.Run(fmt.Sprintf("Success/%s", sym), func(t *testing.T) {
			// GIVEN
			ts := newInstanceStateServiceTestSuite(t)
			ts.SetUp()

			ocgMock := &automock.OperationStorage{}
			defer ocgMock.AssertExpectations(t)
			ocgMock.On("GetAll", ts.Exp.InstanceID).Return(tc.genOps(ts), nil).Once()

			svc := broker.NewInstanceStateService(ocgMock)

			// WHEN
			got, err := svc.IsProvisioned(ts.Exp.InstanceID)

			// THEN
			assert.NoError(t, err)
			assert.Equal(t, tc.exp, got)
		})
	}

	t.Run("Success/false/InstanceNotFound", func(t *testing.T) {
		// GIVEN
		ts := newInstanceStateServiceTestSuite(t)
		ts.SetUp()

		ocgMock := &automock.OperationStorage{}
		defer ocgMock.AssertExpectations(t)
		ocgMock.On("GetAll", ts.Exp.InstanceID).Return(nil, notFoundError{}).Once()

		svc := broker.NewInstanceStateService(ocgMock)

		// WHEN
		got, err := svc.IsProvisioned(ts.Exp.InstanceID)

		// THEN
		assert.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("Failure/GenericStorageError", func(t *testing.T) {
		// GIVEN
		ts := newInstanceStateServiceTestSuite(t)
		ts.SetUp()

		ocgMock := &automock.OperationStorage{}
		defer ocgMock.AssertExpectations(t)
		fixErr := errors.New("fix-storage-error")
		ocgMock.On("GetAll", ts.Exp.InstanceID).Return(nil, fixErr).Once()

		svc := broker.NewInstanceStateService(ocgMock)

		// WHEN
		got, err := svc.IsProvisioned(ts.Exp.InstanceID)

		// THEN
		assert.EqualError(t, err, fmt.Sprintf("while getting operations from storage: %s", fixErr.Error()))
		assert.False(t, got)
	})
}

func TestInstanceStateServiceIsDeprovisioned(t *testing.T) {
	for sym, tc := range map[string]struct {
		genOps func(ts *instanceStateServiceTestSuite) []*internal.InstanceOperation
		exp    bool
	}{
		"true/singleRemoveSucceeded": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				return append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateSucceeded))
			},
			exp: true,
		},
		"true/CreateSucceededThanRemoveInProgress": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded))
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateInProgress))
				return out
			},
			exp: false,
		},
		"false/singleRemoveInProgress": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				return append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateInProgress))
			},
			exp: false,
		},
		"false/CreateSucceededThanRemoveSucceeded": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded))
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateSucceeded))
				return out
			},
			exp: true,
		},
	} {
		t.Run(fmt.Sprintf("Success/%s", sym), func(t *testing.T) {
			// GIVEN
			ts := newInstanceStateServiceTestSuite(t)
			ts.SetUp()

			ocgMock := &automock.OperationStorage{}
			defer ocgMock.AssertExpectations(t)
			ocgMock.On("GetAll", ts.Exp.InstanceID).Return(tc.genOps(ts), nil).Once()

			svc := broker.NewInstanceStateService(ocgMock)

			// WHEN
			got, err := svc.IsDeprovisioned(ts.Exp.InstanceID)

			// THEN
			assert.NoError(t, err)
			assert.Equal(t, tc.exp, got)
		})
	}

	t.Run("Success/false/InstanceNotFound", func(t *testing.T) {
		// GIVEN
		ts := newInstanceStateServiceTestSuite(t)
		ts.SetUp()

		ocgMock := &automock.OperationStorage{}
		defer ocgMock.AssertExpectations(t)
		ocgMock.On("GetAll", ts.Exp.InstanceID).Return(nil, notFoundError{}).Once()

		svc := broker.NewInstanceStateService(ocgMock)

		// WHEN
		got, err := svc.IsDeprovisioned(ts.Exp.InstanceID)

		// THEN
		assert.True(t, broker.IsNotFoundError(err))
		assert.False(t, got)
	})

	t.Run("Failure/GenericStorageError", func(t *testing.T) {
		// GIVEN
		ts := newInstanceStateServiceTestSuite(t)
		ts.SetUp()

		ocgMock := &automock.OperationStorage{}
		defer ocgMock.AssertExpectations(t)
		fixErr := errors.New("fix-storage-error")
		ocgMock.On("GetAll", ts.Exp.InstanceID).Return(nil, fixErr).Once()

		svc := broker.NewInstanceStateService(ocgMock)

		// WHEN
		got, err := svc.IsDeprovisioned(ts.Exp.InstanceID)

		// THEN
		assert.EqualError(t, err, fmt.Sprintf("while getting operations from storage: %s", fixErr.Error()))
		assert.False(t, got)
	})
}

func TestInstanceStateServiceIsDeprovisioningInProgress(t *testing.T) {
	for sym, tc := range map[string]struct {
		genOps func(ts *instanceStateServiceTestSuite) []*internal.InstanceOperation
		exp    bool
	}{
		"false/singleRemoveSucceeded": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				return append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateSucceeded))
			},
			exp: false,
		},
		"true/CreateSucceededThanRemoveInProgress": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded))
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateInProgress))
				return out
			},
			exp: true,
		},
		"false/NoOp": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) { return out },
			exp:    false,
		},
		"false/CreateSucceededThanRemoveSucceeded": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded))
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateSucceeded))
				return out
			},
			exp: false,
		},
	} {
		t.Run(fmt.Sprintf("Success/%s", sym), func(t *testing.T) {
			// GIVEN
			ts := newInstanceStateServiceTestSuite(t)
			ts.SetUp()

			ocgMock := &automock.OperationStorage{}
			defer ocgMock.AssertExpectations(t)
			ocgMock.On("GetAll", ts.Exp.InstanceID).Return(tc.genOps(ts), nil).Once()

			svc := broker.NewInstanceStateService(ocgMock)

			// WHEN
			gotOpID, gotInProgress, err := svc.IsDeprovisioningInProgress(ts.Exp.InstanceID)

			// THEN
			assert.NoError(t, err)
			assert.Equal(t, tc.exp, gotInProgress)
			if tc.exp {
				assert.Equal(t, ts.Exp.OperationID, gotOpID)
			}
		})
	}

	t.Run("Success/false/InstanceNotFound", func(t *testing.T) {
		// GIVEN
		ts := newInstanceStateServiceTestSuite(t)
		ts.SetUp()

		ocgMock := &automock.OperationStorage{}
		defer ocgMock.AssertExpectations(t)
		ocgMock.On("GetAll", ts.Exp.InstanceID).Return(nil, notFoundError{}).Once()

		svc := broker.NewInstanceStateService(ocgMock)

		// WHEN
		gotOpID, got, err := svc.IsDeprovisioningInProgress(ts.Exp.InstanceID)

		// THEN
		assert.NoError(t, err)
		assert.False(t, got)
		assert.Zero(t, gotOpID)
	})

	t.Run("Failure/GenericStorageError", func(t *testing.T) {
		// GIVEN
		ts := newInstanceStateServiceTestSuite(t)
		ts.SetUp()

		ocgMock := &automock.OperationStorage{}
		defer ocgMock.AssertExpectations(t)
		fixErr := errors.New("fix-storage-error")
		ocgMock.On("GetAll", ts.Exp.InstanceID).Return(nil, fixErr).Once()

		svc := broker.NewInstanceStateService(ocgMock)

		// WHEN
		gotOpID, got, err := svc.IsDeprovisioningInProgress(ts.Exp.InstanceID)

		// THEN
		assert.EqualError(t, err, fmt.Sprintf("while getting operations from storage: %s", fixErr.Error()))
		assert.False(t, got)
		assert.Zero(t, gotOpID)
	})
}

func TestInstanceStateServiceIsProvisioningInProgress(t *testing.T) {
	for sym, tc := range map[string]struct {
		genOps        func(ts *instanceStateServiceTestSuite) []*internal.InstanceOperation
		expInProgress bool
	}{
		"true/singleCreateInProgress": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				return append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateInProgress))
			},
			expInProgress: true,
		},
		"false/singleCreateSucceeded": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				return append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded))
			},
			expInProgress: false,
		},
		"false/NoOp": {
			genOps:        func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) { return out },
			expInProgress: false,
		},
		"false/CreateSucceededThanRemoveInProgress": {
			genOps: func(ts *instanceStateServiceTestSuite) (out []*internal.InstanceOperation) {
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded))
				out = append(out, ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateInProgress))
				return out
			},
			expInProgress: false,
		},
	} {
		t.Run(fmt.Sprintf("Success/%s", sym), func(t *testing.T) {
			// GIVEN
			ts := newInstanceStateServiceTestSuite(t)
			ts.SetUp()

			ocgMock := &automock.OperationStorage{}
			defer ocgMock.AssertExpectations(t)
			ocgMock.On("GetAll", ts.Exp.InstanceID).Return(tc.genOps(ts), nil).Once()

			svc := broker.NewInstanceStateService(ocgMock)

			// WHEN
			gotOpID, gotInProgress, err := svc.IsProvisioningInProgress(ts.Exp.InstanceID)

			// THEN
			assert.NoError(t, err)
			assert.Equal(t, tc.expInProgress, gotInProgress)
			if tc.expInProgress {
				assert.Equal(t, ts.Exp.OperationID, gotOpID)
			}
		})
	}

	t.Run("Success/false/InstanceNotFound", func(t *testing.T) {
		// GIVEN
		ts := newInstanceStateServiceTestSuite(t)
		ts.SetUp()

		ocgMock := &automock.OperationStorage{}
		defer ocgMock.AssertExpectations(t)
		ocgMock.On("GetAll", ts.Exp.InstanceID).Return(nil, notFoundError{}).Once()

		svc := broker.NewInstanceStateService(ocgMock)

		// WHEN
		gotOpID, got, err := svc.IsProvisioningInProgress(ts.Exp.InstanceID)

		// THEN
		assert.NoError(t, err)
		assert.False(t, got)
		assert.Zero(t, gotOpID)
	})

	t.Run("Failure/GenericStorageError", func(t *testing.T) {
		// GIVEN
		ts := newInstanceStateServiceTestSuite(t)
		ts.SetUp()

		ocgMock := &automock.OperationStorage{}
		defer ocgMock.AssertExpectations(t)
		fixErr := errors.New("fix-storage-error")
		ocgMock.On("GetAll", ts.Exp.InstanceID).Return(nil, fixErr).Once()

		svc := broker.NewInstanceStateService(ocgMock)

		// WHEN
		gotOpID, got, err := svc.IsProvisioningInProgress(ts.Exp.InstanceID)

		// THEN
		assert.EqualError(t, err, fmt.Sprintf("while getting operations from storage: %s", fixErr.Error()))
		assert.False(t, got)
		assert.Zero(t, gotOpID)
	})
}
