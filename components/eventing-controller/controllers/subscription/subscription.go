package subscription

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var GroupVersion = schema.GroupVersion{Group: "eventing.kyma-project.io", Version: "v1alpha1"}

var Finalizer = GroupVersion.Group
