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

		deployment := fixDeployment("image")

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
			fixDeployment("image"),
			fixDeployment("image"),
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
			fixDeployment("image"),
			nil,
		}

		converter := deploymentConverter{}
		result := converter.ToGQLs(deployments)

		assert.Len(t, result, 1)
		assert.Equal(t, "name", result[0].Name)
	})
}

func fixDeployment(image string) *appsApi.Deployment {
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
							Image: image,
						},
					},
				},
			},
		},
	}
}

func TestDeploymentConverter_ToKymaVersion(t *testing.T) {
	t.Run("Non eu.gcr.io version", func(t *testing.T) {
		image := "test-repo/test-image"

		converter := &deploymentConverter{}
		result := converter.ToKymaVersion(image)

		assert.Equal(t, image, result)
	})

	t.Run("Scemantic version", func(t *testing.T) {
		image := "eu.gcr.io/test/1.2.3"
		converter := &deploymentConverter{}
		result := converter.ToKymaVersion(image)

		assert.Equal(t, "1.2.3", result)
	})

	t.Run("PR version", func(t *testing.T) {
		image := "eu.gcr.io/test/PR-1234"
		converter := &deploymentConverter{}
		result := converter.ToKymaVersion(image)

		assert.Equal(t, "pull request PR-1234", result)
	})

	t.Run("Master version", func(t *testing.T) {
		image := "eu.gcr.io/test/12345678"
		converter := &deploymentConverter{}
		result := converter.ToKymaVersion(image)

		assert.Equal(t, "master 12345678", result)
	})
}
