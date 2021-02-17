package object

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Option is a functional option for API objects builders.
type Option func(metav1.Object)
