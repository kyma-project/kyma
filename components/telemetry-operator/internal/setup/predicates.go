package setup

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func OnlyUpdate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		UpdateFunc:  func(e event.UpdateEvent) bool { return true },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
}

func CreateOrUpdate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return true },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		UpdateFunc:  func(e event.UpdateEvent) bool { return true },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
}

func CreateOrUpdateorDelete() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			//cm, ok := e.Object.(*corev1.ConfigMap)
			//if !ok {
			//	return false
			//}
			//if cm.Name == "override-config-tracepipeline" && cm.Namespace == "kyma-system" {
			//	return true
			//}
			//return false
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool { return true },
		UpdateFunc: func(e event.UpdateEvent) bool {
			return true
			//cmNew, ok := e.ObjectNew.(*corev1.ConfigMap)
			//if !ok {
			//	return false
			//}
			//if cmNew.Name == "override-config-tracepipeline" && cmNew.Namespace == "kyma-system" {
			//	return true
			//}
			//return false
		},
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
}
