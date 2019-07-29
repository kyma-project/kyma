package sync

import (
	"testing"
)

func TestReconciler(t *testing.T) {

	t.Run("should handle application creation without apis", func(t *testing.T) {
		//// given
		//mockApplicationsInterface := &mocks.Applications{}
		//reconciler := NewReconciler(mockApplicationsInterface)
		//
		//mockApplicationsInterface.On("List", v1.ListOptions{}).Return(&v1alpha1.ApplicationList{
		//	Items: []v1alpha1.Application{
		//		{
		//			TypeMeta: metav1.TypeMeta{
		//				Kind:       "Application",
		//				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		//			},
		//			ObjectMeta: metav1.ObjectMeta{
		//				Name: "id2",
		//			},
		//		},
		//	},
		//})
		//
		//application1 := model.Application{
		//	ID:        "id1",
		//	Name:      "First App",
		//	APIs:      []model.APIDefinition{},
		//	EventAPIs: []model.EventAPIDefinition{},
		//}
		//
		//directorApplications := []model.Application{
		//	application1,
		//}
		//
		//expectedResult := []ApplicationAction{
		//	{
		//		Operation:   Create,
		//		Application: application1,
		//	},
		//}
		//
		//// when
		//result, err := reconciler.Do(directorApplications)
		//
		//// then
		//assert.NoError(t, err)
		//assert.Equal(t, expectedResult, result)
	})

	t.Run("should handle application deletion", func(t *testing.T) {

	})

	t.Run("should handle application update", func(t *testing.T) {

	})

}
