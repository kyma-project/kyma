package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ admission.Handler = &PodHandler{}
var _ admission.DecoderInjector = &PodHandler{}

type PodHandler struct {
	decoder *admission.Decoder
	log     log.FieldLogger
	client  client.Client
}

func NewPodHandler(client client.Client, log log.FieldLogger) *PodHandler {
	return &PodHandler{
		client: client,
		log:    log,
	}
}

func (h *PodHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	h.log.Infof("start handling pod: %s", req.UID)

	pod := &corev1.Pod{}
	if err := matchKinds(pod, req.Kind); err != nil {
		h.log.Errorf("kind does not match: %s", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := h.decoder.Decode(req, pod); err != nil {
		h.log.Errorf("cannot decode Pod: %s", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	bindingsName := h.findAssignedBindings(pod)
	if len(bindingsName) == 0 {
		h.log.Infof("finish handling pod: %s. action not taken", req.UID)
		return admission.Allowed("pod has no any assigned bindings. action not taken.")
	}

	bindings, err := h.findBindings(ctx, bindingsName, req.Namespace)
	if err != nil {
		h.log.Errorf("cannot find Bindings: %s", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if err := h.mutatePod(ctx, pod, bindings); err != nil {
		h.log.Errorf("cannot mutate Pod: %s", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	rawPod, err := json.Marshal(pod)
	if err != nil {
		h.log.Errorf("cannot marshal mutated pod: %s", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	h.log.Infof("finish handling pod: %s", req.UID)
	return admission.PatchResponseFromRaw(req.Object.Raw, rawPod)
}

func (h *PodHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// findAssignedBindings checks if pod has any label with an assigned Binding; if yes returns Bindings name
func (h *PodHandler) findAssignedBindings(pod *corev1.Pod) []string {
	bindingsName := make([]string, 0)
	if pod.ObjectMeta.Labels == nil {
		return bindingsName
	}

	for label := range pod.ObjectMeta.Labels {
		if !strings.Contains(label, v1alpha1.BindingLabelKey) {
			continue
		}
		bindingsName = append(bindingsName, strings.TrimPrefix(label, fmt.Sprintf("%s-", v1alpha1.BindingLabelKey)))
	}

	return bindingsName
}

// findBindings fetches all Bindings based on Bindings name and request namespace
func (h *PodHandler) findBindings(ctx context.Context, bindingsName []string, namespace string) ([]*v1alpha1.Binding, error) {
	bindings := make([]*v1alpha1.Binding, 0)

	for _, bindingName := range bindingsName {
		var binding = &v1alpha1.Binding{}
		var lastError error
		err := wait.PollImmediate(500*time.Millisecond, 3*time.Second, func() (bool, error) {
			err := h.client.Get(ctx, client.ObjectKey{Name: bindingName, Namespace: namespace}, binding)
			if err != nil {
				lastError = err
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			return bindings, errors.Wrapf(lastError, "while getting Binding %s/%s", bindingName, namespace)
		}
		bindings = append(bindings, binding)
	}

	return bindings, nil
}

// mutatePod injects to the Pod envFromSource reference coming from Secret/ConfigMap based on Bindings
func (h *PodHandler) mutatePod(ctx context.Context, pod *corev1.Pod, bindings []*v1alpha1.Binding) error {
	for _, binding := range bindings {
		switch binding.Spec.Source.Kind {
		case v1alpha1.SourceKindSecret:
			secret, err := h.findSecret(ctx, binding)
			if err != nil {
				return errors.Wrapf(err, "while finding Secrets for %s/%s Binding", binding.Namespace, binding.Name)
			}
			h.addSecretReference(pod, secret)
		case v1alpha1.SourceKindConfigMap:
			cm, err := h.findConfigMap(ctx, binding)
			if err != nil {
				return errors.Wrapf(err, "while finding ConfigMap for %s/%s Binding", binding.Namespace, binding.Name)
			}
			h.addConfigMapReference(pod, cm)
		default:
			h.log.Warnf("source kind %s not supported, skip source binding", binding.Spec.Source.Kind)
			continue
		}
	}

	return nil
}

func (h *PodHandler) findSecret(ctx context.Context, binding *v1alpha1.Binding) (*corev1.Secret, error) {
	secret := &corev1.Secret{}

	var lastErrror error
	err := wait.PollImmediate(500*time.Millisecond, 3*time.Second, func() (bool, error) {
		err := h.client.Get(ctx, client.ObjectKey{Name: binding.Spec.Source.Name, Namespace: binding.Namespace}, secret)
		if err != nil {
			lastErrror = err
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return secret, errors.Wrapf(lastErrror, "while getting Secret %s/%s", binding.Namespace, binding.Spec.Source.Name)
	}

	return secret, nil
}

func (h *PodHandler) findConfigMap(ctx context.Context, binding *v1alpha1.Binding) (*corev1.ConfigMap, error) {
	configmap := &corev1.ConfigMap{}

	var lastErrror error
	err := wait.PollImmediate(500*time.Millisecond, 3*time.Second, func() (bool, error) {
		err := h.client.Get(ctx, client.ObjectKey{Name: binding.Spec.Source.Name, Namespace: binding.Namespace}, configmap)
		if err != nil {
			lastErrror = err
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return configmap, errors.Wrapf(lastErrror, "while getting ConfigMap %s/%s", binding.Namespace, binding.Spec.Source.Name)
	}

	return configmap, nil
}

func (h *PodHandler) addSecretReference(pod *corev1.Pod, secret *corev1.Secret) {
	for i, ctr := range pod.Spec.Containers {
		origEnv := map[string]corev1.EnvVar{}
		for _, v := range ctr.Env {
			origEnv[v.Name] = v
		}

		mergedEnv := make([]corev1.EnvVar, len(ctr.Env))
		copy(mergedEnv, ctr.Env)

		for key := range secret.Data {
			_, ok := origEnv[key]
			if ok {
				h.log.Warnf("key %s from Secret already exist in container. Environment will not be injected. skip env.", key)
				continue
			}
			mergedEnv = append(mergedEnv, corev1.EnvVar{
				Name: key,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
						Key:                  key,
					},
				},
			})
		}

		ctr.Env = mergedEnv
		pod.Spec.Containers[i] = ctr
	}
}

func (h *PodHandler) addConfigMapReference(pod *corev1.Pod, configmap *corev1.ConfigMap) {
	for i, ctr := range pod.Spec.Containers {
		origEnv := map[string]corev1.EnvVar{}
		for _, v := range ctr.Env {
			origEnv[v.Name] = v
		}

		mergedEnv := make([]corev1.EnvVar, len(ctr.Env))
		copy(mergedEnv, ctr.Env)

		for key := range configmap.Data {
			_, ok := origEnv[key]
			if ok {
				h.log.Warnf("key %s from ConfigMap already exist in container. Environment will not be injected. skip env.", key)
				continue
			}
			mergedEnv = append(mergedEnv, corev1.EnvVar{
				Name: key,
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: configmap.Name},
						Key:                  key,
					},
				},
			})
		}

		ctr.Env = mergedEnv
		pod.Spec.Containers[i] = ctr
	}
}
