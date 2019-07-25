package synchronization

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConverter(t *testing.T) {

	t.Run("should convert application without services", func(t *testing.T) {
		// given
		description := "Description"

		directorApp := Application{
			ID:          "id1",
			Name:        "App1",
			Description: &description,
			Labels:      nil,
			APIs:        []APIDefinition{},
			EventAPIs:   []EventAPIDefinition{},
			Documents:   []Document{},
		}

		expected := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "App1",
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				Services:         []v1alpha1.Service{},
				AccessLabel:      "App1",
				Labels:           map[string]string{},
			},
		}

		// when
		converter := NewConverter("kyma-integration")
		application := converter.Do(directorApp)

		// then
		assert.Equal(t, expected, application)
	})

	t.Run("should convert application with services containing APIs with credentials", func(t *testing.T) {
		//// given
		//description := "Description"
		//
		//directorApp := Application{
		//	ID: "id1",
		//	Name: "App1",
		//	Description: &description,
		//	Labels: nil, // TODO? Figure out what to do with labels
		//	APIs: []APIDefinition{},
		//	EventAPIs: []EventAPIDefinition{},
		//	Documents: []Document{},
		//}
		//
		//expected := v1alpha1.Application{
		//	Spec: v1alpha1.ApplicationSpec{
		//		Description: "Description",
		//		SkipInstallation: false,
		//		Services: []v1alpha1.Service{},
		//	},
		//}
		//
		//// when
		//converter := NewConverter()
		//application := converter.Do(directorApp
	})

	t.Run("should convert application with services containing events", func(t *testing.T) {

	})

	t.Run("should convert application with services containing API and events", func(t *testing.T) {

	})
}
