package compassconnection

import (
	"context"
	"testing"

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
)

const (
	compassConnectionName = "compass-connection"
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

		objClient := NewFakeClient(compassConnection)

		reconciler := newReconciler(objClient)

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
		objClient := NewFakeClient()

		reconciler := newReconciler(objClient)

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

func NewFakeClient(objects ...runtime.Object) Client {
	return &fakeClientWrapper{
		fakeClient: fake.NewSimpleClientset(objects...).CompassV1alpha1().CompassConnections(),
	}
}

type fakeClientWrapper struct {
	fakeClient clientset.CompassConnectionInterface
}

func (f *fakeClientWrapper) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
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

func (f *fakeClientWrapper) Update(ctx context.Context, obj runtime.Object) error {
	panic("implement me")
}

func (f *fakeClientWrapper) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error {
	panic("implement me")
}
