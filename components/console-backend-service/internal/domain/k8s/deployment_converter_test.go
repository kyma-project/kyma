package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsApi "k8s.io/api/apps/v1"
	coreApi "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploymentConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		var zeroTimeStamp time.Time

		deployment := fixDeployment()

		expected := &gqlschema.Deployment{
			Name:              "name",
			CreationTimestamp: zeroTimeStamp,
			Namespace:         "namespace",
			Labels:            gqlschema.Labels{"test": "ok", "ok": "test"},
			Status: gqlschema.DeploymentStatus{
				Replicas:          1,
				AvailableReplicas: 1,
				ReadyReplicas:     1,
				UpdatedReplicas:   1,
				Conditions: []gqlschema.DeploymentCondition{
					{
						Status:  "True",
						Type:    "Available",
						Message: "message",
						Reason:  "reason",
					},
				},
			},
			Containers: []gqlschema.Container{
				{
					Name:  "test",
					Image: "image",
				},
			},
		}

		converter := &deploymentConverter{}
		result := converter.ToGQL(deployment)

		require.NotNil(t, result)
		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &deploymentConverter{}
		converter.ToGQL(&appsApi.Deployment{})
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &deploymentConverter{}
		result := converter.ToGQL(nil)

		assert.Nil(t, result)
	})
}

func TestDeploymentConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deployments := []*appsApi.Deployment{
			fixDeployment(),
			fixDeployment(),
		}

		converter := deploymentConverter{}
		result := converter.ToGQLs(deployments)

		assert.Len(t, result, 2)
		assert.Equal(t, "name", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var deployments []*appsApi.Deployment

		converter := deploymentConverter{}
		result := converter.ToGQLs(deployments)

		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		deployments := []*appsApi.Deployment{
			nil,
			fixDeployment(),
			nil,
		}

		converter := deploymentConverter{}
		result := converter.ToGQLs(deployments)

		assert.Len(t, result, 1)
		assert.Equal(t, "name", result[0].Name)
	})
}

func fixDeployment() *appsApi.Deployment {
	var mockTimeStamp v1.Time

	return &appsApi.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:              "name",
			Namespace:         "namespace",
			CreationTimestamp: mockTimeStamp,
			Labels: map[string]string{
				"test": "ok",
				"ok":   "test",
			},
		},
		Status: appsApi.DeploymentStatus{
			UpdatedReplicas:   1,
			Replicas:          1,
			ReadyReplicas:     1,
			AvailableReplicas: 1,
			Conditions: []appsApi.DeploymentCondition{
				{
					Reason:             "reason",
					Message:            "message",
					LastUpdateTime:     mockTimeStamp,
					LastTransitionTime: mockTimeStamp,
					Type:               appsApi.DeploymentAvailable,
					Status:             coreApi.ConditionTrue,
				},
			},
		},
		Spec: appsApi.DeploymentSpec{
			Template: coreApi.PodTemplateSpec{
				Spec: coreApi.PodSpec{
					Containers: []coreApi.Container{
						{
							Name:  "test",
							Image: "image",
						},
					},
				},
			},
		},
	}
}
