/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RolebindingName = "namespace-admin"
	RoleRefKind     = "ClusterRole"
	RoleRefName     = "kyma-admin"

	subjectKindGroup  = "Group"
	subjectKindUser   = "User"
	SubjectStaticUser = "admin@kyma.cx"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	ExcludedNamespaces []string
	SubjectGroups      []string
	UseStaticConnector bool
}

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces/status,verbs=get;update;patch;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=create;update;patch;delete

func (r *NamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("namespace", req.NamespacedName)

	var namespace corev1.Namespace

	if err := r.Client.Get(ctx, req.NamespacedName, &namespace); err != nil {

		if apierrs.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !namespace.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if r.isProtected(namespace.Name) {
		log.Info(fmt.Sprintf("%s is a system namespace. Skipping...", namespace.Name))
		return ctrl.Result{}, nil
	}

	rb := &rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RolebindingName,
			Namespace: namespace.Name,
		},
		Subjects: r.generateSubjects(),
		RoleRef: rbac.RoleRef{
			Kind:     RoleRefKind,
			Name:     RoleRefName,
			APIGroup: rbac.GroupName,
		},
	}

	if err := r.Create(ctx, rb); err != nil {
		if apierrs.IsAlreadyExists(err) {
			return ctrl.Result{}, r.Update(ctx, rb)
		}

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) isProtected(namespaceName string) bool {
	for _, name := range r.ExcludedNamespaces {
		if name == namespaceName {
			return true
		}
	}

	return false
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}

func (r *NamespaceReconciler) generateSubjects() []rbac.Subject {

	var subjectGroups []rbac.Subject

	for _, groupName := range r.SubjectGroups {
		subjectGroups = append(subjectGroups, rbac.Subject{
			Kind:     subjectKindGroup,
			Name:     groupName,
			APIGroup: rbac.GroupName,
		})
	}

	if r.UseStaticConnector {
		subjectGroups = append(subjectGroups, rbac.Subject{
			Kind:     subjectKindUser,
			Name:     SubjectStaticUser,
			APIGroup: rbac.GroupName,
		})
	}

	return subjectGroups
}
