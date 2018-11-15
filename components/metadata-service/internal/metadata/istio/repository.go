package istio

import (
	"fmt"

	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/metadata-service/pkg/apis/istio/v1alpha2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	matchTemplateFormat = `(destination.service.host == "%s.%s.svc.cluster.local") && (source.labels["%s"] != "true")`
)

// RuleInterface allows to perform operations for Rules in kubernetes
type RuleInterface interface {
	Create(*v1alpha2.Rule) (*v1alpha2.Rule, error)
	Delete(name string, options *v1.DeleteOptions) error
}

// ChecknothingInterface allows to perform operations for CheckNothings in kubernetes
type ChecknothingInterface interface {
	Create(*v1alpha2.Checknothing) (*v1alpha2.Checknothing, error)
	Delete(name string, options *v1.DeleteOptions) error
}

// DenierInterface allows to perform operations for Deniers in kubernetes
type DenierInterface interface {
	Create(*v1alpha2.Denier) (*v1alpha2.Denier, error)
	Delete(name string, options *v1.DeleteOptions) error
}

// Repository allows to perform various operations for Istio resources
type Repository interface {
	// CreateDenier creates Denier
	CreateDenier(remoteEnvironment, serviceId, name string) apperrors.AppError
	// CreateCheckNothing creates CheckNothing
	CreateCheckNothing(remoteEnvironment, serviceId, name string) apperrors.AppError
	// CreateRule creates Rule
	CreateRule(remoteEnvironment, serviceId, name string) apperrors.AppError
	// UpserDenier creates or updates Denier
	UpsertDenier(remoteEnvironment, serviceId, name string) apperrors.AppError
	// UpsertCheckNothing creates or updates CheckNothing
	UpsertCheckNothing(remoteEnvironment, serviceId, name string) apperrors.AppError
	// UpsertRule creates or updates Rule
	UpsertRule(remoteEnvironment, serviceId, name string) apperrors.AppError
	// DeleteDenier deletes Denier
	DeleteDenier(name string) apperrors.AppError
	// DeleteCheckNothing deletes CheckNothing
	DeleteCheckNothing(name string) apperrors.AppError
	// DeleteRule deletes Rule
	DeleteRule(name string) apperrors.AppError
}

type RepositoryConfig struct {
	Namespace string
}

type repository struct {
	ruleInterface         RuleInterface
	checknothingInterface ChecknothingInterface
	denierInterface       DenierInterface
	config                RepositoryConfig
}

// NewRepository creates new repository with provided interfaces
func NewRepository(ruleInterface RuleInterface, checknothingInterface ChecknothingInterface, denierInterface DenierInterface, config RepositoryConfig) Repository {
	return &repository{
		ruleInterface:         ruleInterface,
		checknothingInterface: checknothingInterface,
		denierInterface:       denierInterface,
		config:                config,
	}
}

// CreateDenier creates Denier
func (repo *repository) CreateDenier(remoteEnvironment, serviceId, name string) apperrors.AppError {
	denier := repo.makeDenierObject(remoteEnvironment, serviceId, name)

	_, err := repo.denierInterface.Create(denier)
	if err != nil {
		return apperrors.Internal("Creating %s denier failed, %s", name, err.Error())
	}

	return nil
}

// CreateCheckNothing creates CheckNothing
func (repo *repository) CreateCheckNothing(remoteEnvironment, serviceId, name string) apperrors.AppError {
	checkNothing := repo.makeCheckNothingObject(remoteEnvironment, serviceId, name)

	_, err := repo.checknothingInterface.Create(checkNothing)
	if err != nil {
		return apperrors.Internal("Creating %s checknothing failed, %s", name, err.Error())
	}
	return nil
}

// CreateRule creates Rule
func (repo *repository) CreateRule(remoteEnvironment, serviceId, name string) apperrors.AppError {
	rule := repo.makeRuleObject(remoteEnvironment, serviceId, name)

	_, err := repo.ruleInterface.Create(rule)
	if err != nil {
		return apperrors.Internal("Creating %s rule failed, %s", name, err.Error())
	}
	return nil
}

// UpserDenier creates or updates Denier
func (repo *repository) UpsertDenier(remoteEnvironment, serviceId, name string) apperrors.AppError {
	denier := repo.makeDenierObject(remoteEnvironment, serviceId, name)

	_, err := repo.denierInterface.Create(denier)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return apperrors.Internal("Updating %s denier failed, %s", name, err.Error())
	}
	return nil
}

// UpsertCheckNothing creates or updates CheckNothing
func (repo *repository) UpsertCheckNothing(remoteEnvironment, serviceId, name string) apperrors.AppError {
	checkNothing := repo.makeCheckNothingObject(remoteEnvironment, serviceId, name)

	_, err := repo.checknothingInterface.Create(checkNothing)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return apperrors.Internal("Updating %s checknothing failed, %s", name, err.Error())
	}
	return nil
}

// UpsertRule creates or updates Rule
func (repo *repository) UpsertRule(remoteEnvironment, serviceId, name string) apperrors.AppError {
	rule := repo.makeRuleObject(remoteEnvironment, serviceId, name)

	_, err := repo.ruleInterface.Create(rule)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return apperrors.Internal("Updating %s rule failed, %s", name, err.Error())
	}
	return nil
}

// DeleteDenier deletes Denier
func (repo *repository) DeleteDenier(name string) apperrors.AppError {
	err := repo.denierInterface.Delete(name, nil)
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("Deleting %s denier failed, %s", name, err.Error())
	}
	return nil
}

// DeleteCheckNothing deletes CheckNothing
func (repo *repository) DeleteCheckNothing(name string) apperrors.AppError {
	err := repo.checknothingInterface.Delete(name, nil)
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("Deleting %s checknothing failed, %s", name, err.Error())
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

func (repo *repository) makeDenierObject(remoteEnvironment, serviceId, name string) *v1alpha2.Denier {
	return &v1alpha2.Denier{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelRemoteEnvironment: remoteEnvironment,
				k8sconsts.LabelServiceId:         serviceId,
			},
		},
		Spec: &v1alpha2.DenierSpec{
			Status: &v1alpha2.DenierStatus{
				Code:    7,
				Message: "Not allowed",
			},
		},
	}
}

func (repo *repository) makeCheckNothingObject(remoteEnvironment, serviceId, name string) *v1alpha2.Checknothing {
	return &v1alpha2.Checknothing{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelRemoteEnvironment: remoteEnvironment,
				k8sconsts.LabelServiceId:         serviceId,
			},
		},
	}
}

func (repo *repository) makeRuleObject(remoteEnvironment, serviceId, name string) *v1alpha2.Rule {
	match := repo.matchExpression(name, repo.config.Namespace, name)
	handlerName := name + ".denier"
	instanceName := name + ".checknothing"

	return &v1alpha2.Rule{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelRemoteEnvironment: remoteEnvironment,
				k8sconsts.LabelServiceId:         serviceId,
			},
		},
		Spec: &v1alpha2.RuleSpec{
			Match: match,
			Actions: []v1alpha2.RuleAction{{
				Handler:   handlerName,
				Instances: []string{instanceName},
			}},
		},
	}
}

func (repo *repository) matchExpression(serviceName, namespace, accessLabel string) string {
	return fmt.Sprintf(matchTemplateFormat, serviceName, namespace, accessLabel)
}
