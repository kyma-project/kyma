package kubernetes

import (
	"testing"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/kubernetes/automock"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/model"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestClient_EnsureUserPermissions(t *testing.T) {
	// GIVEN
	testEmail := "test@example.com"
	testError := errors.New("test error")

	testCases := []struct {
		Name            string
		User            model.User
		FixCRBInterface func() *automock.CRBInterface
		ExpectedError   error
	}{
		{
			Name: "Success when Admin and Dev and permissions already granted",
			User: model.User{
				Email:       testEmail,
				IsDeveloper: true,
				IsAdmin:     true,
			},
			FixCRBInterface: func() *automock.CRBInterface {
				m := &automock.CRBInterface{}
				m.On("Get", mock.Anything, rbacConfig[model.UserPermissionAdmin].clusterRoleBindingName, metav1.GetOptions{}).
					Return(fixCRB(model.UserPermissionAdmin, []string{testEmail}), nil).Once()
				m.On("Get", mock.Anything, rbacConfig[model.UserPermissionDeveloper].clusterRoleBindingName, metav1.GetOptions{}).
					Return(fixCRB(model.UserPermissionDeveloper, []string{testEmail}), nil).Once()
				return m
			},
			ExpectedError: nil,
		},
		{
			Name: "Success when CRB exists but subject has to be added",
			User: model.User{
				Email:       testEmail,
				IsDeveloper: false,
				IsAdmin:     true,
			},
			FixCRBInterface: func() *automock.CRBInterface {
				m := &automock.CRBInterface{}
				m.On("Get", mock.Anything, rbacConfig[model.UserPermissionAdmin].clusterRoleBindingName, metav1.GetOptions{}).
					Return(fixCRB(model.UserPermissionAdmin, []string{}), nil).Once()
				m.On("Update", mock.Anything, fixCRB(model.UserPermissionAdmin, []string{testEmail}), metav1.UpdateOptions{}).
					Return(fixCRB(model.UserPermissionAdmin, []string{}), nil).Once()
				return m
			},
			ExpectedError: nil,
		},
		{
			Name: "Success when CRB doesn't exist",
			User: model.User{
				Email:       testEmail,
				IsDeveloper: false,
				IsAdmin:     true,
			},
			FixCRBInterface: func() *automock.CRBInterface {
				m := &automock.CRBInterface{}
				m.On("Get", mock.Anything, rbacConfig[model.UserPermissionAdmin].clusterRoleBindingName, metav1.GetOptions{}).
					Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "")).Once()
				m.On("Create", mock.Anything, fixCRB(model.UserPermissionAdmin, []string{}), metav1.CreateOptions{}).
					Return(fixCRB(model.UserPermissionAdmin, []string{}), nil).Once()
				m.On("Update", mock.Anything, fixCRB(model.UserPermissionAdmin, []string{testEmail}), metav1.UpdateOptions{}).
					Return(fixCRB(model.UserPermissionAdmin, []string{}), nil).Once()
				return m
			},
			ExpectedError: nil,
		},
		{
			Name: "Error while getting CRB",
			User: model.User{
				Email:       testEmail,
				IsDeveloper: false,
				IsAdmin:     true,
			},
			FixCRBInterface: func() *automock.CRBInterface {
				m := &automock.CRBInterface{}
				m.On("Get", mock.Anything, rbacConfig[model.UserPermissionAdmin].clusterRoleBindingName, metav1.GetOptions{}).
					Return(nil, testError).Once()
				return m
			},
			ExpectedError: testError,
		},
		{
			Name: "Error while getting CRB as Developer",
			User: model.User{
				Email:       testEmail,
				IsDeveloper: true,
				IsAdmin:     false,
			},
			FixCRBInterface: func() *automock.CRBInterface {
				m := &automock.CRBInterface{}
				m.On("Get", mock.Anything, rbacConfig[model.UserPermissionDeveloper].clusterRoleBindingName, metav1.GetOptions{}).
					Return(nil, testError).Once()
				return m
			},
			ExpectedError: testError,
		},
		{
			Name: "Error while creating CRB",
			User: model.User{
				Email:       testEmail,
				IsDeveloper: false,
				IsAdmin:     true,
			},
			FixCRBInterface: func() *automock.CRBInterface {
				m := &automock.CRBInterface{}
				m.On("Get", mock.Anything, rbacConfig[model.UserPermissionAdmin].clusterRoleBindingName, metav1.GetOptions{}).
					Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "")).Once()
				m.On("Create", mock.Anything, fixCRB(model.UserPermissionAdmin, []string{}), metav1.CreateOptions{}).
					Return(nil, testError).Once()
				return m
			},
			ExpectedError: testError,
		},
		{
			Name: "Error while updating CRB",
			User: model.User{
				Email:       testEmail,
				IsDeveloper: false,
				IsAdmin:     true,
			},
			FixCRBInterface: func() *automock.CRBInterface {
				m := &automock.CRBInterface{}
				m.On("Get", mock.Anything, rbacConfig[model.UserPermissionAdmin].clusterRoleBindingName, metav1.GetOptions{}).
					Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "")).Once()
				m.On("Create", mock.Anything, fixCRB(model.UserPermissionAdmin, []string{}), metav1.CreateOptions{}).
					Return(fixCRB(model.UserPermissionAdmin, []string{}), nil).Once()
				m.On("Update", mock.Anything, fixCRB(model.UserPermissionAdmin, []string{testEmail}), metav1.UpdateOptions{}).
					Return(nil, testError).Once()
				return m
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			crbInterfaceMock := testCase.FixCRBInterface()
			client := Client{clusterRoleBindingsClient: crbInterfaceMock}

			// WHEN
			err := client.EnsureUserPermissions(testCase.User)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, crbInterfaceMock)
		})
	}
}

func fixCRB(permission model.UserPermission, subjects []string) *rbacv1.ClusterRoleBinding {
	crb := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: rbacConfig[permission].clusterRoleBindingName,
			Labels: map[string]string{
				"kyma-project.io/uaa": "migrated",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacAPIGroup,
			Kind:     roleRefKind,
			Name:     rbacConfig[permission].clusterRoleName,
		},
	}
	for _, subject := range subjects {
		crb.Subjects = append(crb.Subjects, rbacv1.Subject{
			Kind:     subjectKind,
			APIGroup: rbacAPIGroup,
			Name:     subject,
		})
	}
	return &crb
}
