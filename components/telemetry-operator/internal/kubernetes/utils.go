package kubernetes

import (
	"context"
	"fmt"
	v1 "k8s.io/api/rbac/v1"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateOrUpdateClusterRoleBinding(ctx context.Context, c client.Client, desired *v1.ClusterRoleBinding) error {
	var existing v1.ClusterRoleBinding
	err := c.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &existing)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return c.Create(ctx, desired)
	}
	mutated := existing.DeepCopy()
	mergeMetadata(&desired.ObjectMeta, mutated.ObjectMeta)
	if apiequality.Semantic.DeepEqual(mutated, desired) {
		return nil
	}
	return c.Update(ctx, desired)
}

func CreateOrUpdateClusterRole(ctx context.Context, c client.Client, desired *v1.ClusterRole) error {
	var existing v1.ClusterRole
	err := c.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &existing)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return c.Create(ctx, desired)
	}
	mutated := existing.DeepCopy()
	mergeMetadata(&desired.ObjectMeta, mutated.ObjectMeta)
	if apiequality.Semantic.DeepEqual(mutated, desired) {
		return nil
	}
	return c.Update(ctx, desired)
}

func CreateOrUpdateServiceAccount(ctx context.Context, c client.Client, desired *corev1.ServiceAccount) error {
	var existing corev1.ServiceAccount
	err := c.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &existing)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return c.Create(ctx, desired)
	}
	mutated := existing.DeepCopy()
	mergeMetadata(&desired.ObjectMeta, mutated.ObjectMeta)
	if apiequality.Semantic.DeepEqual(mutated, desired) {
		return nil
	}
	return c.Update(ctx, desired)
}

func CreateIfNotExistsConfigMap(ctx context.Context, c client.Client, desired *corev1.ConfigMap) error {
	err := c.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &corev1.ConfigMap{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		return c.Create(ctx, desired)
	}
	return nil
}

func CreateOrUpdateConfigMap(ctx context.Context, c client.Client, desired *corev1.ConfigMap) error {
	var existing corev1.ConfigMap
	err := c.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &existing)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		return c.Create(ctx, desired)
	}

	mutated := existing.DeepCopy()
	mergeMetadata(&desired.ObjectMeta, mutated.ObjectMeta)
	if apiequality.Semantic.DeepEqual(mutated, desired) {
		return nil
	}
	return c.Update(ctx, desired)
}

func CreateOrUpdateSecret(ctx context.Context, c client.Client, desired *corev1.Secret) error {
	var existing corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &existing)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		return c.Create(ctx, desired)
	}

	mutated := existing.DeepCopy()
	mergeMetadata(&desired.ObjectMeta, mutated.ObjectMeta)
	if apiequality.Semantic.DeepEqual(mutated, desired) {
		return nil
	}
	return c.Update(ctx, desired)
}

func CreateOrUpdateDeployment(ctx context.Context, c client.Client, desired *appsv1.Deployment) error {
	var existing appsv1.Deployment
	err := c.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &existing)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		return c.Create(ctx, desired)
	}

	mergeMetadata(&desired.ObjectMeta, existing.ObjectMeta)
	mergeKubectlAnnotations(&desired.Spec.Template.ObjectMeta, existing.Spec.Template.ObjectMeta)
	return c.Update(ctx, desired)
}

func CreateOrUpdateDaemonSet(ctx context.Context, c client.Client, desired *appsv1.DaemonSet) error {
	var existing appsv1.DaemonSet
	err := c.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &existing)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		return c.Create(ctx, desired)
	}

	mergeMetadata(&desired.ObjectMeta, existing.ObjectMeta)
	mergeKubectlAnnotations(&desired.Spec.Template.ObjectMeta, existing.Spec.Template.ObjectMeta)
	mergeChecksumAnnotations(&desired.Spec.Template.ObjectMeta, existing.Spec.Template.ObjectMeta)
	return c.Update(ctx, desired)
}

func CreateOrUpdateService(ctx context.Context, c client.Client, desired *corev1.Service) error {
	var existing corev1.Service
	err := c.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &existing)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		return c.Create(ctx, desired)
	}

	// Apply immutable fields from the existing service.
	desired.Spec.IPFamilies = existing.Spec.IPFamilies
	desired.Spec.IPFamilyPolicy = existing.Spec.IPFamilyPolicy
	desired.Spec.ClusterIP = existing.Spec.ClusterIP
	desired.Spec.ClusterIPs = existing.Spec.ClusterIPs

	mergeMetadata(&desired.ObjectMeta, existing.ObjectMeta)

	return c.Update(ctx, desired)
}

func DeleteFluentBit(ctx context.Context, c client.Client, name types.NamespacedName) error {

	if err := deleteResource(c, ctx, name, &appsv1.DaemonSet{}); err != nil {
		return fmt.Errorf("unable to delete daemonset %s: %v", name.Name, err)
	}

	if err := deleteResource(c, ctx, name, &corev1.Service{}); err != nil {
		return fmt.Errorf("unable to delete service %s: %v", name.Name, err)
	}

	if err := deleteResource(c, ctx, name, &corev1.ConfigMap{}); err != nil {
		return fmt.Errorf("unable to delete configmap %s: %v", name.Name, err)
	}

	if err := deleteResource(c, ctx, name, &corev1.ServiceAccount{}); err != nil {
		return fmt.Errorf("unable to delete service account %s: %v", name.Name, err)
	}

	if err := deleteResource(c, ctx, name, &v1.ClusterRoleBinding{}); err != nil {
		return fmt.Errorf("unable to delete cluster role binding %s: %v", name.Name, err)
	}

	if err := deleteResource(c, ctx, name, &v1.ClusterRole{}); err != nil {
		return fmt.Errorf("unable to delete cluster role %s: %v", name.Name, err)
	}

	name.Name = fmt.Sprintf("%s-luascripts", name.Name)
	if err := deleteResource(c, ctx, name, &corev1.ConfigMap{}); err != nil {
		return fmt.Errorf("unable to delete configmap %s: %v", name.Name, err)
	}

	return nil
}

func deleteResource(c client.Client, ctx context.Context, name client.ObjectKey, obj client.Object) error {
	err := c.Get(ctx, name, obj)
	if err == nil {
		if err = c.Delete(ctx, obj); err != nil && !errors.IsNotFound(err) {
			return err
		}
	} else if !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func mergeMetadata(new *metav1.ObjectMeta, old metav1.ObjectMeta) {
	new.ResourceVersion = old.ResourceVersion

	new.SetLabels(mergeMaps(new.Labels, old.Labels))
	new.SetAnnotations(mergeMaps(new.Annotations, old.Annotations))
}

func mergeMaps(new map[string]string, old map[string]string) map[string]string {
	return mergeMapsByPrefix(new, old, "")
}

func mergeKubectlAnnotations(new *metav1.ObjectMeta, old metav1.ObjectMeta) {
	new.SetAnnotations(mergeMapsByPrefix(new.Annotations, old.Annotations, "kubectl.kubernetes.io/"))
}

func mergeChecksumAnnotations(new *metav1.ObjectMeta, old metav1.ObjectMeta) {
	new.SetAnnotations(mergeMapsByPrefix(new.Annotations, old.Annotations, "checksum/"))
}

func mergeMapsByPrefix(new map[string]string, old map[string]string, prefix string) map[string]string {
	if old == nil {
		old = make(map[string]string)
	}

	if new == nil {
		new = make(map[string]string)
	}

	for k, v := range old {
		_, hasValue := new[k]
		if strings.HasPrefix(k, prefix) && !hasValue {
			new[k] = v
		}
	}

	return new
}

func GetOrCreateConfigMap(ctx context.Context, c client.Client, name types.NamespacedName) (corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name.Name, Namespace: name.Namespace}}
	err := c.Get(ctx, client.ObjectKeyFromObject(&cm), &cm)
	if err == nil {
		return cm, nil
	}
	if errors.IsNotFound(err) {
		err = c.Create(ctx, &cm)
		if err == nil {
			return cm, nil
		}
	}
	return corev1.ConfigMap{}, err
}

func GetOrCreateSecret(ctx context.Context, c client.Client, name types.NamespacedName) (corev1.Secret, error) {
	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name.Name, Namespace: name.Namespace}}
	err := c.Get(ctx, client.ObjectKeyFromObject(&secret), &secret)
	if err == nil {
		return secret, nil
	}
	if errors.IsNotFound(err) {
		err = c.Create(ctx, &secret)
		if err == nil {
			return secret, nil
		}
	}
	return corev1.Secret{}, err
}
