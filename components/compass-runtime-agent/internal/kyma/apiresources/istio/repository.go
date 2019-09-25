package istio

import (
	"fmt"

	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/k8sconsts"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/application-registry/pkg/apis/istio/v1alpha2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	matchTemplateFormat = `(destination.service.host == "%s.%s.svc.cluster.local") && (source.labels["%s"] != "true")`

	denierAdapterName        = "denier"
	checkNothingTemplateName = "checknothing"
)

// RuleInterface allows to perform operations for Istio Rules in kubernetes
//go:generate mockery -name RuleInterface
type RuleInterface interface {
	Create(*v1alpha2.Rule) (*v1alpha2.Rule, error)
	Delete(name string, options *v1.DeleteOptions) error
}

// InstanceInterface allows to perform operations for Istio Instances in kubernetes
//go:generate mockery -name InstanceInterface
type InstanceInterface interface {
	Create(*v1alpha2.Instance) (*v1alpha2.Instance, error)
	Delete(name string, options *v1.DeleteOptions) error
}

// HandlerInterface allows to perform operations for Istio Handlers in kubernetes
//go:generate mockery -name HandlerInterface
type HandlerInterface interface {
	Create(*v1alpha2.Handler) (*v1alpha2.Handler, error)
	Delete(name string, options *v1.DeleteOptions) error
}

// Repository allows to perform various operations for Istio resources
//go:generate mockery -name Repository
type Repository interface {
	// CreateHandler creates Handler
	CreateHandler(application string, appUID types.UID, serviceId, name string) apperrors.AppError
	// CreateInstance creates Instance
	CreateInstance(application string, appUID types.UID, serviceId, name string) apperrors.AppError
	// CreateRule creates Rule
	CreateRule(application string, appUID types.UID, serviceId, name string) apperrors.AppError
	// UpsertHandler creates or updates Handler
	UpsertHandler(application string, appUID types.UID, serviceId, name string) apperrors.AppError
	// UpsertInstance creates or updates Instance
	UpsertInstance(application string, appUID types.UID, serviceId, name string) apperrors.AppError
	// UpsertRule creates or updates Rule
	UpsertRule(application string, appUID types.UID, serviceId, name string) apperrors.AppError
	// DeleteHandler deletes Handler
	DeleteHandler(name string) apperrors.AppError
	// DeleteInstance deletes Instance
	DeleteInstance(name string) apperrors.AppError
	// DeleteRule deletes Rule
	DeleteRule(name string) apperrors.AppError
}

type RepositoryConfig struct {
	Namespace string
}

type repository struct {
	ruleInterface     RuleInterface
	instanceInterface InstanceInterface
	handlerInterface  HandlerInterface
	config            RepositoryConfig
}

// NewRepository creates new repository with provided interfaces
func NewRepository(ruleInterface RuleInterface, instanceInterface InstanceInterface, handlerInterface HandlerInterface, config RepositoryConfig) Repository {
	return &repository{
		ruleInterface:     ruleInterface,
		instanceInterface: instanceInterface,
		handlerInterface:  handlerInterface,
		config:            config,
	}
}

// CreateHandler creates Handler
func (repo *repository) CreateHandler(application string, appUID types.UID, serviceId, name string) apperrors.AppError {
	handler := repo.makeHandlerObject(application, appUID, serviceId, name)

	_, err := repo.handlerInterface.Create(handler)
	if err != nil {
		return apperrors.Internal("Creating %s handler failed, %s", name, err.Error())
	}

	return nil
}

// CreateInstance creates Instance
func (repo *repository) CreateInstance(application string, appUID types.UID, serviceId, name string) apperrors.AppError {
	checkNothing := repo.makeInstanceObject(application, appUID, serviceId, name)

	_, err := repo.instanceInterface.Create(checkNothing)
	if err != nil {
		return apperrors.Internal("Creating %s instance failed, %s", name, err.Error())
	}
	return nil
}

// CreateRule creates Rule
func (repo *repository) CreateRule(application string, appUID types.UID, serviceId, name string) apperrors.AppError {
	rule := repo.makeRuleObject(application, appUID, serviceId, name)

	_, err := repo.ruleInterface.Create(rule)
	if err != nil {
		return apperrors.Internal("Creating %s rule failed, %s", name, err.Error())
	}
	return nil
}

// UpserHandler creates or updates Handler
func (repo *repository) UpsertHandler(application string, appUID types.UID, serviceId, name string) apperrors.AppError {
	handler := repo.makeHandlerObject(application, appUID, serviceId, name)

	_, err := repo.handlerInterface.Create(handler)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return apperrors.Internal("Updating %s handler failed, %s", name, err.Error())
	}
	return nil
}

// UpsertInstance creates or updates Instance
func (repo *repository) UpsertInstance(application string, appUID types.UID, serviceId, name string) apperrors.AppError {
	checkNothing := repo.makeInstanceObject(application, appUID, serviceId, name)

	_, err := repo.instanceInterface.Create(checkNothing)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return apperrors.Internal("Updating %s instance failed, %s", name, err.Error())
	}
	return nil
}

// UpsertRule creates or updates Rule
func (repo *repository) UpsertRule(application string, appUID types.UID, serviceId, name string) apperrors.AppError {
	rule := repo.makeRuleObject(application, appUID, serviceId, name)

	_, err := repo.ruleInterface.Create(rule)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return apperrors.Internal("Updating %s rule failed, %s", name, err.Error())
	}
	return nil
}

// DeleteHandler deletes Handler
func (repo *repository) DeleteHandler(name string) apperrors.AppError {
	err := repo.handlerInterface.Delete(name, nil)
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("Deleting %s handler failed, %s", name, err.Error())
	}
	return nil
}

// DeleteInstance deletes Instance
func (repo *repository) DeleteInstance(name string) apperrors.AppError {
	err := repo.instanceInterface.Delete(name, nil)
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("Deleting %s instance failed, %s", name, err.Error())
	}
	return nil
}

// DeleteRule deletes Rule
func (repo *repository) DeleteRule(name string) apperrors.AppError {
	err := repo.ruleInterface.Delete(name, nil)
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("Deleting %s rule failed, %s", name, err.Error())
	}
	return nil
}

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
