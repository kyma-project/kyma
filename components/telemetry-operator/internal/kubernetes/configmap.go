package kubernetes

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ConfigmapProber struct {
	client.Client
}

func (cmp *ConfigmapProber) IsPresent(ctx context.Context, name types.NamespacedName) (map[string]interface{}, error) {
	log := logf.FromContext(ctx)
	config := make(map[string]interface{})
	var cm corev1.ConfigMap
	if err := cmp.Get(ctx, name, &cm); err != nil {
		log.V(1).Info(fmt.Sprintf("failed to get  %s/%s Configmap", name.Namespace, name.Name))
		if apierrors.IsNotFound(err) {
			return config, nil
		}
		return config, fmt.Errorf("failed to get %s/%s Configmap: %v", name.Namespace, name.Name, err)
	}
	if _, ok := cm.Data["override-config"]; ok {
		if err := yaml.Unmarshal([]byte(cm.Data["override-config"]), &config); err != nil {
			return config, err
		}
		return config, nil
	}
	return config, nil
}
