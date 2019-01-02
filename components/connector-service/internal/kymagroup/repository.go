package kymagroup

import (
	"fmt"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/pkg/apis/applicationconnector/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/util/retry"
)

// TODO - add comments

// Manager contains operations for managing Kyma Group CR
type Manager interface {
	Create(kymaGroup *v1alpha1.KymaGroup) (*v1alpha1.KymaGroup, error)
	Update(kymaGroup *v1alpha1.KymaGroup) (*v1alpha1.KymaGroup, error)
	Get(name string, options v1.GetOptions) (*v1alpha1.KymaGroup, error)
}

type KymaGroupsRepository interface {
	Create(application *v1alpha1.KymaGroup) apperrors.AppError
	UpdateClusterData(group string, cluster *v1alpha1.Cluster) apperrors.AppError
	AddApplication(group string, app *v1alpha1.Application) apperrors.AppError
	RemoveApplication(group string, appId string) apperrors.AppError
	Get(name string) (*v1alpha1.KymaGroup, apperrors.AppError)
}

type repository struct {
	kymaGroupManager Manager
}

func NewKymaGroupRepository(kymaGroupManager Manager) KymaGroupsRepository {
	return &repository{
		kymaGroupManager: kymaGroupManager,
	}
}

func (r *repository) Create(kymaGroup *v1alpha1.KymaGroup) apperrors.AppError {
	_, err := r.kymaGroupManager.Create(kymaGroup)
	if err != nil {
		return apperrors.Internal("Failed to create %s Kyma Group, %s", kymaGroup.Name, err.Error())
	}

	return nil
}

func (r *repository) UpdateClusterData(group string, cluster *v1alpha1.Cluster) apperrors.AppError {
	return r.updateKymaGroup(group, func(kg *v1alpha1.KymaGroup) apperrors.AppError {
		kg.Spec.Cluster = *cluster
		return nil
	})
}

func (r *repository) AddApplication(group string, app *v1alpha1.Application) apperrors.AppError {
	return r.updateKymaGroup(group, func(kg *v1alpha1.KymaGroup) apperrors.AppError {
		if applicationInGroup(kg, app.ID) == -1 {
			kg.Spec.Applications = append(kg.Spec.Applications, *app)
		} else {
			return apperrors.AlreadyExists("Application with id %s already exists", group)
		}

		return nil
	})
}

func (r *repository) RemoveApplication(group string, appID string) apperrors.AppError {
	return r.updateKymaGroup(group, func(kg *v1alpha1.KymaGroup) apperrors.AppError {
		return removeAppFromGroup(kg, appID)
	})
}

func (r *repository) updateKymaGroup(group string, modification func(kymaGroup *v1alpha1.KymaGroup) apperrors.AppError) apperrors.AppError {
	kymaGroup, appErr := r.getKymaGroup(group)
	if appErr != nil {
		return appErr.Append("Failed to update %s Kyma Group", group)
	}

	appErr = modification(kymaGroup)
	if appErr != nil {
		return appErr.Append("Failed to update %s Kyma Group", group)
	}

	err := r.updateWithRetries(kymaGroup)
	if err != nil {
		return apperrors.Internal("Failed to update %s Kyma Group, %s", group, err.Error())
	}

	return nil
}

func (r *repository) Get(name string) (*v1alpha1.KymaGroup, apperrors.AppError) {
	return r.getKymaGroup(name)
}

func (r *repository) getKymaGroup(group string) (*v1alpha1.KymaGroup, apperrors.AppError) {
	re, err := r.kymaGroupManager.Get(group, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			message := fmt.Sprintf("Kyma Group %s not found", group)
			return nil, apperrors.NotFound(message)
		}

		message := fmt.Sprintf("Getting Kyma Group %s failed, %s", group, err.Error())
		return nil, apperrors.Internal(message)
	}

	return re, nil
}

func (r *repository) updateWithRetries(kymaGroup *v1alpha1.KymaGroup) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, e := r.kymaGroupManager.Update(kymaGroup)
		return e
	})
}

func applicationInGroup(kymaGroup *v1alpha1.KymaGroup, appID string) int {
	for i, app := range kymaGroup.Spec.Applications {
		if app.ID == appID {
			return i
		}
	}

	return -1
}

func removeAppFromGroup(kymaGroup *v1alpha1.KymaGroup, appID string) apperrors.AppError {
	index := applicationInGroup(kymaGroup, appID)
	if index == -1 {
		return apperrors.NotFound("Application with id %s not found", appID)
	}

	kymaGroup.Spec.Applications = append(kymaGroup.Spec.Applications[:index], kymaGroup.Spec.Applications[index+1:]...)

	return nil
}
