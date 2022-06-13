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

		log.Debug("old", event.ObjectOld.GetName())
		log.Debug("new:", event.ObjectNew.GetName())

		oldFn, ok := event.ObjectOld.(*serverlessv1alpha1.Function)
		if !ok {
			v := reflect.ValueOf(oldFn)
			log.Debug("Can't cast to function:", v.Type())
			return true
		}

		newFn, ok := event.ObjectNew.(*serverlessv1alpha1.Function)
		if !ok {
			v := reflect.ValueOf(newFn)
			log.Debug("Can't cast to function:", v.Type())
			return true
		}

		equalStasus := equalFunctionStatus(oldFn.Status, newFn.Status)
		log.Debug("Statuses are equal: ", equalStasus)

		return equalStasus
	}
}
