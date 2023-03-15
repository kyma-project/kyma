package watcher

import corev1 "k8s.io/api/core/v1"

type UpdateNotifiable interface {
	NotifyUpdate(*corev1.ConfigMap)
}
