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

func (cmp *ConfigmapProber) ReadConfigMapOrEmpty(ctx context.Context, name types.NamespacedName) (string, error) {
	log := logf.FromContext(ctx)
	var cm corev1.ConfigMap
	if err := cmp.Get(ctx, name, &cm); err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info(fmt.Sprintf("Could not get  %s/%s Configmap, looks like its not present", name.Namespace, name.Name))
			return "", nil
		}
		return "", fmt.Errorf("failed to get %s/%s Configmap: %v", name.Namespace, name.Name, err)
	}
	if data, ok := cm.Data[overrideConfigFileName]; ok {
		return data, nil
	}
	return "", nil
}
