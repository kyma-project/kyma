package testsuite

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	watchtools "k8s.io/client-go/tools/watch"
)

var (
	ErrInvalidDataType = errors.New("invalid data type")
)

func buildWaitForStatusesReady(iRsc dynamic.ResourceInterface, timeout time.Duration, rscNames ...string) func(string, ...func(...interface{})) error {
	return func(initialResourceVersion string, callbacks ...func(...interface{})) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		var conditions []watchtools.ConditionFunc
		for _, rsName := range rscNames {
			conditions = append(conditions, isPhaseReady(rsName, callbacks...))
		}
		_, err := watchtools.Until(ctx, initialResourceVersion, iRsc, conditions...)
		if err != nil {
			return err
		}
		return err
	}
}

var ready = "Ready"

func isPhaseReady(name string, callbacks ...func(...interface{})) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type != watch.Modified {
			return false, nil
		}
		u, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, ErrInvalidDataType
		}
		if u.GetName() != name {
			return false, nil
		}
		var bucketLike struct {
			Status struct {
				Phase string
			}
		}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &bucketLike)
		if err != nil || bucketLike.Status.Phase != ready {
			return false, err
		}
		for _, callback := range callbacks {
			callback(fmt.Sprintf("%s is ready:\n%v", name, u))
		}
		return true, nil
	}
}
