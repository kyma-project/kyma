package object

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ObjectOption is a functional option for API objects builders.
type ObjectOption func(metav1.Object)
