package kubernetes

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ConfigmapProber struct {
	client.Client
}

const overrideConfigFileName = "override-config"

func (cmp *ConfigmapProber) IsPresent(ctx context.Context, name types.NamespacedName) (string, error) {
	log := logf.FromContext(ctx)
	var cm corev1.ConfigMap
	if err := cmp.Get(ctx, name, &cm); err != nil {
		log.V(1).Info(fmt.Sprintf("failed to get  %s/%s Configmap", name.Namespace, name.Name))
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get %s/%s Configmap: %v", name.Namespace, name.Name, err)
	}
	if _, ok := cm.Data[overrideConfigFileName]; ok {
		return cm.Data[overrideConfigFileName], nil
	}
	return "", nil
}
