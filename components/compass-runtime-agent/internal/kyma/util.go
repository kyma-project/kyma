package kyma

import (
	memoize "github.com/kofalt/go-memoize"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
)

func newResult(application v1alpha1.Application, applicationID string, operation Operation, appError apperrors.AppError) Result {
	return Result{
		ApplicationName: application.Name,
		ApplicationID:   applicationID,
		Operation:       operation,
		Error:           appError,
	}
}

func ApplicationExists(applicationName string, applicationList []v1alpha1.Application) bool {
	if applicationList == nil {
		return false
	}

	for _, runtimeApplication := range applicationList {
		if runtimeApplication.Name == applicationName {
			return true
		}
	}

	return false
}

func GetApplication(applicationName string, applicationList []v1alpha1.Application) v1alpha1.Application {
	if applicationList == nil {
		return v1alpha1.Application{}
	}

	for _, runtimeApplication := range applicationList {
		if runtimeApplication.Name == applicationName {
			return runtimeApplication
		}
	}

	return v1alpha1.Application{}
}

type getApplicationUIDResult struct {
	AppUID   types.UID
	AppError apperrors.AppError
}

// cachingGetApplicationUIDFunc provides a function which prevents duplicate function calls to input function for the same application parameter.
// Subsequent invocations return the cached result.
func cachingGetApplicationUIDFunc(f func(application string) (types.UID, apperrors.AppError)) func(application string) (getApplicationUIDResult, error) {
	cache := memoize.NewMemoizer(0, 0)
	return func(application string) (getApplicationUIDResult, error) {
		v, err, _ := cache.Memoize(application, func() (interface{}, error) {
			appUID, apperr := f(application)
			return getApplicationUIDResult{
				AppUID:   appUID,
				AppError: apperr,
			}, nil
		})
		return v.(getApplicationUIDResult), err
	}
}
