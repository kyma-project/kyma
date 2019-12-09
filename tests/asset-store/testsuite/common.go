package testsuite

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/rafter/tests/asset-store/pkg/retry"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func buildDeleteLeftovers(iRsc dynamic.ResourceInterface, timeout time.Duration) func(string, ...func(...interface{})) error {
	return func(testId string, callbacks ...func(...interface{})) error {
		labelSelector := fmt.Sprintf("test-id=%s", testId)
		var assetNames []string
		var initialResourceVersion = ""
		err := retry.OnTimeout(retry.DefaultBackoff, func() error {
			result, err := iRsc.List(metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if err != nil {
				return err
			}
			for _, u := range result.Items {
				assetNames = append(assetNames, u.GetName())
				if initialResourceVersion != "" {
					continue
				}
				initialResourceVersion = u.GetResourceVersion()
			}
			return nil
		})
		if err != nil {
			return err
		}
		if len(assetNames) < 1 {
			return nil
		}
		err = retry.WithIgnoreOnNotFound(retry.DefaultBackoff, func() error {
			err := iRsc.DeleteCollection(nil, metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			return err
		}, callbacks...)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		var assetDeletionCounter = len(assetNames)
		_, err = watchtools.Until(ctx, initialResourceVersion, iRsc, func(event watch.Event) (b bool, e error) {
			if event.Type != watch.Deleted {
				return false, nil
			}
			u, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				return false, ErrInvalidDataType
			}
			for _, name := range assetNames {
				if name != u.GetName() {
					continue
				}
				assetDeletionCounter--
				break
			}
			if assetDeletionCounter == 0 {
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			return errors.Wrapf(err, "while deleting resource leftovers with label selector %s", labelSelector)
		}
		return nil
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
