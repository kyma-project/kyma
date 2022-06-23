package serverless

import (
	"reflect"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func IsNotFunctionStatusUpdate(log *zap.SugaredLogger) func(event.UpdateEvent) bool {
	return func(event event.UpdateEvent) bool {
		if event.ObjectOld == nil || event.ObjectNew == nil {
			return true
		}

		log.Debug("old: ", event.ObjectOld.GetName())
		log.Debug("new: ", event.ObjectNew.GetName())

		oldFn, ok := event.ObjectOld.(*serverlessv1alpha1.Function)
		if !ok {
			v := reflect.ValueOf(event.ObjectOld)
			log.Debug("Can't cast to function from type: ", v.Type())
			return true
		}

		newFn, ok := event.ObjectNew.(*serverlessv1alpha1.Function)
		if !ok {
			v := reflect.ValueOf(event.ObjectNew)
			log.Debug("Can't cast to function from type: ", v.Type())
			return true
		}

		equalStatus := equalFunctionStatus(oldFn.Status, newFn.Status)
		log.Debug("Statuses are equal: ", equalStatus)

		return equalStatus
	}
}
