package istio

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-registry/pkg/apis/istio/v1alpha2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// AuthorizationPolicyInterface allows to perform operations for Istio AuthorizationPolicies in kubernetes
//go:generate mockery --name AuthorizationPolicyInterface
type AuthorizationPolicyInterface interface {
	Create(policy *v1alpha2.AuthorizationPolicy) (*v1alpha2.AuthorizationPolicy, error)
	Delete(name string, options *v1.DeleteOptions) error
}

// Repository allows to perform various operations for Istio resources
//go:generate mockery --name Repository
type Repository interface {
	// CreateAuthorizationPolicy creates AuthorizationPolicy
	CreateAuthorizationPolicy(application string, appUID types.UID, serviceId, name string) apperrors.AppError
	// UpsertAuthorizationPolicy creates or updates AuthorizationPolicy
	UpsertAuthorizationPolicy(application string, appUID types.UID, serviceId, name string) apperrors.AppError
	// DeleteAuthorizationPolicy deletes AuthorizationPolicy
	DeleteAuthorizationPolicy(name string) apperrors.AppError
}

type RepositoryConfig struct {
	Namespace string
}

type repository struct {
	authorizationPolicyInterface AuthorizationPolicyInterface
	config                       RepositoryConfig
}

// NewRepository creates new repository with provided interfaces
func NewRepository(authorizationPolicyInterface AuthorizationPolicyInterface, config RepositoryConfig) Repository {
	return &repository{
		authorizationPolicyInterface: authorizationPolicyInterface,
		config:                       config,
	}
}

// CreateAuthorizationPolicy creates AuthorizationPolicy
func (repo *repository) CreateAuthorizationPolicy(application string, appUID types.UID, serviceId, name string) apperrors.AppError {
	authorizationPolicy := repo.makeAuthorizationPolicyObject(application, appUID, serviceId, name)

	if _, err := repo.authorizationPolicyInterface.Create(authorizationPolicy); err != nil {
		return apperrors.Internal("Creating %s authorization policy failed, %v", name, err)
	}
	return nil
}

// UpsertAuthorizationPolicy creates or updates AuthorizationPolicy
func (repo *repository) UpsertAuthorizationPolicy(application string, appUID types.UID, serviceId, name string) apperrors.AppError {
	authorizationPolicy := repo.makeAuthorizationPolicyObject(application, appUID, serviceId, name)

	_, err := repo.authorizationPolicyInterface.Create(authorizationPolicy)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return apperrors.Internal("Updating %s authorization policy failed, %v", name, err)
	}
	return nil
}

// DeleteAuthorizationPolicy deletes AuthorizationPolicy
func (repo *repository) DeleteAuthorizationPolicy(name string) apperrors.AppError {
	err := repo.authorizationPolicyInterface.Delete(name, nil)
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("Deleting %s authorization policy failed, %v", name, err)
	}
	return nil
}

func (repo *repository) makeAuthorizationPolicyObject(application string, appUID types.UID, serviceId, name string) *v1alpha2.AuthorizationPolicy {
	return &v1alpha2.AuthorizationPolicy{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelApplication: application,
				k8sconsts.LabelServiceId:   serviceId,
			},
			OwnerReferences: k8sconsts.CreateOwnerReferenceForApplication(application, appUID),
		},
		Spec: &v1alpha2.AuthorizationPolicySpec{
			//TODO: Make it work!
			Selector: nil,
			Action:   v1alpha2.Deny,
			Rules:    nil,
		},
	}
}

/*
const (
	matchTemplateFormat = `(destination.service.host == "%s.%s.svc.cluster.local") && (source.labels["%s"] != "true")`

	denierAdapterName        = "denier"
	checkNothingTemplateName = "checknothing"
)

func (repo *repository) makeHandlerObject(application string, appUID types.UID, serviceId, name string) *v1alpha2.Handler {
	return &v1alpha2.Handler{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelApplication: application,
				k8sconsts.LabelServiceId:   serviceId,
			},
			OwnerReferences: k8sconsts.CreateOwnerReferenceForApplication(application, appUID),
		},
		Spec: &v1alpha2.HandlerSpec{
			CompiledAdapter: denierAdapterName,
			Params: &v1alpha2.DenierHandlerParams{
				Status: &v1alpha2.DenierStatus{
					Code:    7,
					Message: "Not allowed",
				},
			},
		},
	}
}

func (repo *repository) makeInstanceObject(application string, appUID types.UID, serviceId, name string) *v1alpha2.Instance {
	return &v1alpha2.Instance{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelApplication: application,
				k8sconsts.LabelServiceId:   serviceId,
			},
			OwnerReferences: k8sconsts.CreateOwnerReferenceForApplication(application, appUID),
		},
		Spec: &v1alpha2.InstanceSpec{
			CompiledTemplate: checkNothingTemplateName,
		},
	}
}

func (repo *repository) makeRuleObject(application string, appUID types.UID, serviceId, name string) *v1alpha2.Rule {
	match := repo.matchExpression(name, repo.config.Namespace, name)

	return &v1alpha2.Rule{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelApplication: application,
				k8sconsts.LabelServiceId:   serviceId,
			},
			OwnerReferences: k8sconsts.CreateOwnerReferenceForApplication(application, appUID),
		},
		Spec: &v1alpha2.RuleSpec{
			Match: match,
			Actions: []v1alpha2.RuleAction{{
				Handler:   name,
				Instances: []string{name},
			}},
		},
	}
}

func (repo *repository) matchExpression(serviceName, namespace, accessLabel string) string {
	return fmt.Sprintf(matchTemplateFormat, serviceName, namespace, accessLabel)
}
*/
