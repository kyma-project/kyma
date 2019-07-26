package compassconnection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compassconnection/mocks"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/fake"
	clientset "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	compassConnectionName = "compass-connection"

	minimalConfigSyncTime = 30 * time.Second
)

var (
	compassConnectionNamespacedName = types.NamespacedName{
		Name: compassConnectionName,
	}
)

func TestReconcile(t *testing.T) {

	t.Run("should reconcile request", func(t *testing.T) {
		// given
		compassConnection := &v1alpha1.CompassConnection{
			ObjectMeta: v1.ObjectMeta{
				Name: compassConnectionName,
			},
		}

		fakeClientset := fake.NewSimpleClientset(compassConnection).Compass().CompassConnections()
		objClient := NewObjectClientWrapper(fakeClientset)
		supervisor := &mocks.Supervisor{}
		supervisor.On("SynchronizeWithCompass", compassConnection).Return(compassConnection, nil)

		reconciler := newReconciler(objClient, supervisor, minimalConfigSyncTime)

		request := reconcile.Request{
			NamespacedName: compassConnectionNamespacedName,
		}

		// when
		result, err := reconciler.Reconcile(request)

		// then
		require.NoError(t, err)
		require.Empty(t, result)

	})

	t.Run("should reconcile delete request", func(t *testing.T) {
		// given
		compassConnection := &v1alpha1.CompassConnection{
			ObjectMeta: v1.ObjectMeta{
				Name: compassConnectionName,
			},
			Status: v1alpha1.CompassConnectionStatus{
				State: v1alpha1.Connected,
			},
		}

		fakeClientset := fake.NewSimpleClientset().Compass().CompassConnections()
		objClient := NewObjectClientWrapper(fakeClientset)
		supervisor := &mocks.Supervisor{}
		supervisor.On("InitializeCompassConnection").
			Run(func(args mock.Arguments) {
				_, err := fakeClientset.Create(compassConnection)
				require.NoError(t, err)
			}).
			Return(compassConnection, nil)

		reconciler := newReconciler(objClient, supervisor, minimalConfigSyncTime)

		request := reconcile.Request{
			NamespacedName: compassConnectionNamespacedName,
		}

		// when
		result, err := reconciler.Reconcile(request)

		// then
		require.NoError(t, err)
		require.Empty(t, result)

	})
}

func Test_shouldResyncConfig(t *testing.T) {

	minimalResyncTime := 300 * time.Second

	for _, testCase := range []struct {
		description  string
		syncStatus   *v1alpha1.SynchronizationStatus
		shouldResync bool
	}{
		{
			description:  "resync if sync status not present",
			syncStatus:   nil,
			shouldResync: true,
		},
		{
			description: "resync if sync minimal time passed",
			syncStatus: &v1alpha1.SynchronizationStatus{
				LastAttempt: metav1.Unix(time.Now().Unix()-600, 0),
			},
			shouldResync: true,
		},
		{
			description: "not resync if sync minimal time did not pass",
			syncStatus: &v1alpha1.SynchronizationStatus{
				LastAttempt: metav1.Now(),
			},
			shouldResync: false,
		},
	} {
		t.Run("should "+testCase.description, func(t *testing.T) {
			// given
			connection := &v1alpha1.CompassConnection{
				ObjectMeta: v1.ObjectMeta{Name: "connection"},
				Status: v1alpha1.CompassConnectionStatus{
					SynchronizationStatus: testCase.syncStatus,
				},
			}

			// when
			resync := shouldResyncConfig(connection, minimalResyncTime)

			// then
			assert.Equal(t, testCase.shouldResync, resync)
		})
	}

}

func NewObjectClientWrapper(client clientset.CompassConnectionInterface) Client {
	return &objectClientWrapper{
		fakeClient: client,
	}
}

type objectClientWrapper struct {
	fakeClient clientset.CompassConnectionInterface
}

func (f *objectClientWrapper) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	compassConnection, ok := obj.(*v1alpha1.CompassConnection)
	if !ok {
		return errors.New("object is not Compass Connection")
	}

	cc, err := f.fakeClient.Get(key.Name, v1.GetOptions{})
	if err != nil {
		return err
	}

	cc.DeepCopyInto(compassConnection)
	return nil
}

func (f *objectClientWrapper) Update(ctx context.Context, obj runtime.Object) error {
	panic("implement me")
}

func (f *objectClientWrapper) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error {
	panic("implement me")
}
