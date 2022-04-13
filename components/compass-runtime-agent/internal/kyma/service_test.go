package kyma

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	appMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	appSecrets "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/secrets/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestKymaUpsertCredentialsSecrets(t *testing.T) {
	type upsert struct {
		bundleID    string
		credentials *model.Credentials
	}

	tests := []struct {
		name        string
		application model.Application
		upserts     []upsert
	}{
		{
			name: "DefaultInstanceAuth is null",
			application: model.Application{
				Name: "",
				ApiBundles: []model.APIBundle{
					{
						DefaultInstanceAuth: nil,
					},
				},
			},
		},
		{
			name: "Credentials are nil",
			application: model.Application{
				ApiBundles: []model.APIBundle{
					{
						DefaultInstanceAuth: &model.Auth{
							Credentials: nil,
						},
					},
				},
			},
		},
		{
			name: "Basic auth",
			application: model.Application{
				ApiBundles: []model.APIBundle{
					{
						ID:                  "bundle-1",
						DefaultInstanceAuth: fixAuthBasic(),
					},
				},
			},
			upserts: []upsert{{
				bundleID:    "bundle-1",
				credentials: fixAuthBasic().Credentials,
			}},
		},
		{
			name: "Oauths",
			application: model.Application{
				ApiBundles: []model.APIBundle{
					{
						ID:                  "bundle-1",
						DefaultInstanceAuth: fixAuthOauth(),
					},
					{
						ID: "bundle-2",
						DefaultInstanceAuth: &model.Auth{
							Credentials: &model.Credentials{
								Oauth: &model.Oauth{
									URL:          "https://auth.expamle.com",
									ClientID:     "my-client-2",
									ClientSecret: "my-secret-2",
								},
							},
						},
					},
				},
			},
			upserts: []upsert{
				{
					bundleID:    "bundle-1",
					credentials: fixAuthOauth().Credentials,
				},
				{
					bundleID: "bundle-2",
					credentials: &model.Credentials{
						Oauth: &model.Oauth{
							URL:          "https://auth.expamle.com",
							ClientID:     "my-client-2",
							ClientSecret: "my-secret-2",
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			const UID = "f940c082-be4f-11eb-8529-0242ac130003"
			tc.application.Name = "my-app"

			repositoryMock := &appMocks.Repository{}
			repositoryMock.On("Get", tc.application.Name, metav1.GetOptions{}).Return(&v1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					UID: UID,
				},
			}, nil)
			credentialsServiceMock := &appSecrets.CredentialsService{}
			for _, upsert := range tc.upserts {
				credentialsServiceMock.On("Upsert", tc.application.Name, types.UID(UID), upsert.bundleID, upsert.credentials).
					Return(applications.Credentials{}, nil).Once()
			}

			service := &service{
				applicationRepository: repositoryMock,
				credentialsService:    credentialsServiceMock,
			}
			err := service.upsertCredentialsSecrets(tc.application)
			assert.NoError(t, err)

			credentialsServiceMock.AssertExpectations(t)
		})
	}
}

func TestKymaRequestParametersSecrets(t *testing.T) {
	type upsert struct {
		bundleID          string
		requestParameters *model.RequestParameters
	}

	tests := []struct {
		name        string
		application model.Application
		upserts     []upsert
	}{
		{
			name: "DefaultInstanceAuth is null",
			application: model.Application{
				Name: "",
				ApiBundles: []model.APIBundle{
					{
						DefaultInstanceAuth: nil,
					},
				},
			},
		},
		{
			name: "Credentials are nil",
			application: model.Application{
				ApiBundles: []model.APIBundle{
					{
						DefaultInstanceAuth: &model.Auth{
							Credentials: nil,
						},
					},
				},
			},
		},
		{
			name: "Request params are empty",
			application: model.Application{
				ApiBundles: []model.APIBundle{
					{
						DefaultInstanceAuth: &model.Auth{
							Credentials: &model.Credentials{
								Oauth: &model.Oauth{
									URL:          "https://auth.expamle.com",
									ClientID:     "my-client-2",
									ClientSecret: "my-secret-2",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Request params once",
			application: model.Application{
				ApiBundles: []model.APIBundle{
					{
						ID:                  "bundle-1",
						DefaultInstanceAuth: fixAuthBasic(),
					},
				},
			},
			upserts: []upsert{
				{
					bundleID:          "bundle-1",
					requestParameters: fixAuthBasic().RequestParameters,
				},
			},
		},
		{
			name: "Request params twice",
			application: model.Application{
				ApiBundles: []model.APIBundle{
					{
						ID:                  "bundle-1",
						DefaultInstanceAuth: fixAuthBasic(),
					},
					{
						ID:                  "bundle-2",
						DefaultInstanceAuth: fixAuthOauth(),
					},
				},
			},
			upserts: []upsert{
				{
					bundleID:          "bundle-1",
					requestParameters: fixAuthBasic().RequestParameters,
				},
				{
					bundleID:          "bundle-2",
					requestParameters: fixAuthOauth().RequestParameters,
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			const UID = "f940c082-be4f-11eb-8529-0242ac130003"
			tc.application.Name = "my-app"

			repositoryMock := &appMocks.Repository{}
			repositoryMock.On("Get", tc.application.Name, metav1.GetOptions{}).Return(&v1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					UID: UID,
				},
			}, nil)
			requestParametersServiceMock := &appSecrets.RequestParametersService{}
			for _, upsert := range tc.upserts {
				requestParametersServiceMock.On("Upsert", tc.application.Name, types.UID(UID), upsert.bundleID, upsert.requestParameters).
					Return("", nil).Once()
			}

			service := &service{
				applicationRepository:    repositoryMock,
				requestParametersService: requestParametersServiceMock,
			}
			err := service.upsertRequestParametersSecrets(tc.application)
			assert.NoError(t, err)

			requestParametersServiceMock.AssertExpectations(t)
		})
	}
}

func TestKymaService(t *testing.T) {
	t.Run("should return error in case failed to determine differences between current and desired runtime state", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		credentialsServiceMock := &appSecrets.CredentialsService{}
		requestParametersServiceMock := &appSecrets.RequestParametersService{}
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(nil, apperrors.Internal("some error"))

		directorApplication := getTestDirectorApplication("id1", "name1", []model.APIDefinition{}, []model.EventAPIDefinition{})

		directorApplications := []model.Application{
			directorApplication,
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, credentialsServiceMock, requestParametersServiceMock)
		_, err := kymaService.Apply(directorApplications)

		// then
		assert.Error(t, err)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
	})

	t.Run("should apply Create operation", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		credentialsServiceMock := &appSecrets.CredentialsService{}
		requestParametersServiceMock := &appSecrets.RequestParametersService{}

		api := fixDirectorAPiDefinition("API1", "name", "API description", fixAPISpec())
		eventAPI := fixDirectorEventAPIDefinition("EventAPI1", "name", "Event API 1 description", fixEventAPISpec())

		apiBundle1 := fixAPIBundle("bundle1", []model.APIDefinition{api}, nil, nil)
		apiBundle2 := fixAPIBundle("bundle2", nil, []model.EventAPIDefinition{eventAPI}, nil)
		apiBundle3 := fixAPIBundle("bundle3", []model.APIDefinition{api}, []model.EventAPIDefinition{eventAPI}, nil)
		directorApplication := fixDirectorApplication("id1", "name1", apiBundle1, apiBundle2, apiBundle3)

		entry1 := fixAPIEntry("API1", "api1")
		entry2 := fixEventAPIEntry("EventAPI1", "eventapi1")

		newRuntimeService1 := fixService("bundle1", entry1)
		newRuntimeService2 := fixService("bundle2", entry2)
		newRuntimeService3 := fixService("bundle3", entry1, entry2)

		newRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{newRuntimeService1, newRuntimeService2, newRuntimeService3})

		directorApplications := []model.Application{
			directorApplication,
		}

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{},
		}

		converterMock.On("Do", directorApplication).Return(newRuntimeApplication)
		applicationsManagerMock.On("Create", &newRuntimeApplication).Return(&newRuntimeApplication, nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		expectedResult := []Result{
			{
				ApplicationName: "name1",
				ApplicationID:   "id1",
				Operation:       Create,
				Error:           nil,
			},
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, credentialsServiceMock, requestParametersServiceMock)
		result, err := kymaService.Apply(directorApplications)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
	})

	t.Run("should apply Create operation and create credentials", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		credentialsServiceMock := &appSecrets.CredentialsService{}
		requestParametersServiceMock := &appSecrets.RequestParametersService{}

		api := fixDirectorAPiDefinition("API1", "name", "API description", fixAPISpec())
		eventAPI := fixDirectorEventAPIDefinition("EventAPI1", "name", "Event API 1 description", fixEventAPISpec())

		authBundle1 := fixAuthOauth()
		authBundle1.RequestParameters = nil
		authBundle2 := fixAuthBasic()
		authBundle4 := fixAuthRequestParameters()

		apiBundle1 := fixAPIBundle("bundle1", []model.APIDefinition{api}, nil, authBundle1)
		apiBundle2 := fixAPIBundle("bundle2", nil, []model.EventAPIDefinition{eventAPI}, authBundle2)
		apiBundle3 := fixAPIBundle("bundle3", []model.APIDefinition{api}, []model.EventAPIDefinition{eventAPI}, nil)
		apiBundle4 := fixAPIBundle("bundle4", []model.APIDefinition{api}, nil, authBundle4)
		directorApplication := fixDirectorApplication("id1", "name1", apiBundle1, apiBundle2, apiBundle3, apiBundle4)

		entry1 := fixAPIEntry("API1", "api1")
		entry2 := fixEventAPIEntry("EventAPI1", "eventapi1")

		newRuntimeService1 := fixService("bundle1", entry1)
		newRuntimeService2 := fixService("bundle2", entry2)
		newRuntimeService3 := fixService("bundle3", entry1, entry2)
		newRuntimeService4 := fixService("bundle4", entry1)

		newRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{newRuntimeService1, newRuntimeService2, newRuntimeService3, newRuntimeService4})

		directorApplications := []model.Application{
			directorApplication,
		}

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{},
		}

		converterMock.On("Do", directorApplication).Return(newRuntimeApplication)
		applicationsManagerMock.On("Get", "name1", metav1.GetOptions{}).Return(&newRuntimeApplication, nil)
		applicationsManagerMock.On("Create", &newRuntimeApplication).Return(&newRuntimeApplication, nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		credentialsServiceMock.On("Upsert", "name1", newRuntimeApplication.UID, "bundle1", authBundle1.Credentials).Return(applications.Credentials{}, nil)
		credentialsServiceMock.On("Upsert", "name1", newRuntimeApplication.UID, "bundle2", authBundle2.Credentials).Return(applications.Credentials{}, nil)
		requestParametersServiceMock.On("Upsert", "name1", newRuntimeApplication.UID, "bundle2", authBundle2.RequestParameters).Return("", nil)
		requestParametersServiceMock.On("Upsert", "name1", newRuntimeApplication.UID, "bundle4", authBundle4.RequestParameters).Return("", nil)

		expectedResult := []Result{
			{
				ApplicationName: "name1",
				ApplicationID:   "id1",
				Operation:       Create,
				Error:           nil,
			},
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, credentialsServiceMock, requestParametersServiceMock)
		result, err := kymaService.Apply(directorApplications)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
	})

	t.Run("should apply Update operation", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		credentialsServiceMock := &appSecrets.CredentialsService{}
		requestParametersServiceMock := &appSecrets.RequestParametersService{}

		api1 := fixDirectorAPiDefinition("API1", "Name", "API 1 description", fixAPISpec())
		eventAPI1 := fixDirectorEventAPIDefinition("EventAPI1", "Name", "Event API 1 description", fixEventAPISpec())
		apiBundle1 := fixAPIBundle("bundle1", []model.APIDefinition{api1}, []model.EventAPIDefinition{eventAPI1}, nil)

		api2 := fixDirectorAPiDefinition("API2", "Name", "API 2 description", fixAPISpec())
		eventAPI2 := fixDirectorEventAPIDefinition("EventAPI2", "Name", "Event API 2 description", fixEventAPISpec())
		apiBundle2 := fixAPIBundle("bundle2", []model.APIDefinition{api2}, []model.EventAPIDefinition{eventAPI2}, nil)

		api3 := fixDirectorAPiDefinition("API3", "Name", "API 3 description", nil)
		eventAPI3 := fixDirectorEventAPIDefinition("EventAPI2", "Name", "Event API 3 description", nil)
		apiBundle3 := fixAPIBundle("bundle3", []model.APIDefinition{api3}, []model.EventAPIDefinition{eventAPI3}, nil)

		directorApplication := fixDirectorApplication("id1", "name1", apiBundle1, apiBundle2, apiBundle3)

		runtimeServiceToCreate := fixService("bundle1", fixServiceAPIEntry("API1"), fixEventAPIEntry("EventAPI1", "EventAPI1Name"))
		runtimeServiceToUpdate1 := fixService("bundle2", fixServiceAPIEntry("API2"), fixEventAPIEntry("EventAPI2", "EventAPI2Name"))
		runtimeServiceToUpdate2 := fixService("bundle3", fixServiceAPIEntry("API3"), fixEventAPIEntry("EventAPI3", "EventAPI3Name"))
		runtimeServiceToDelete := fixService("bundle4", fixServiceAPIEntry("API4"), fixEventAPIEntry("EventAPI4", "EventAPI4Name"))

		newRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeServiceToCreate, runtimeServiceToUpdate1, runtimeServiceToUpdate2})

		directorApplications := []model.Application{
			directorApplication,
		}

		existingRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeServiceToUpdate1, runtimeServiceToUpdate2, runtimeServiceToDelete})
		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{existingRuntimeApplication},
		}

		converterMock.On("Do", directorApplication).Return(newRuntimeApplication)
		applicationsManagerMock.On("Update", &newRuntimeApplication).Return(&newRuntimeApplication, nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		expectedResult := []Result{
			{
				ApplicationName: "name1",
				ApplicationID:   "id1",
				Operation:       Update,
				Error:           nil,
			},
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, credentialsServiceMock, requestParametersServiceMock)
		result, err := kymaService.Apply(directorApplications)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
	})

	t.Run("should apply Update operation and update credentials", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		credentialsServiceMock := &appSecrets.CredentialsService{}
		requestParametersServiceMock := &appSecrets.RequestParametersService{}

		authBundle1 := fixAuthOauth()
		authBundle3 := fixAuthBasic()

		api1 := fixDirectorAPiDefinition("API1", "Name", "API 1 description", fixAPISpec())
		eventAPI1 := fixDirectorEventAPIDefinition("EventAPI1", "Name", "Event API 1 description", fixEventAPISpec())
		apiBundle1 := fixAPIBundle("bundle1", []model.APIDefinition{api1}, []model.EventAPIDefinition{eventAPI1}, authBundle1)

		api2 := fixDirectorAPiDefinition("API2", "Name", "API 2 description", fixAPISpec())
		eventAPI2 := fixDirectorEventAPIDefinition("EventAPI2", "Name", "Event API 2 description", fixEventAPISpec())
		apiBundle2 := fixAPIBundle("bundle2", []model.APIDefinition{api2}, []model.EventAPIDefinition{eventAPI2}, nil)

		api3 := fixDirectorAPiDefinition("API3", "Name", "API 3 description", nil)
		eventAPI3 := fixDirectorEventAPIDefinition("EventAPI2", "Name", "Event API 3 description", nil)
		apiBundle3 := fixAPIBundle("bundle3", []model.APIDefinition{api3}, []model.EventAPIDefinition{eventAPI3}, authBundle3)

		directorApplication := fixDirectorApplication("id1", "name1", apiBundle1, apiBundle2, apiBundle3)

		runtimeServiceToCreate := fixService("bundle1", fixServiceAPIEntryWithOauth("API1", "bundle1"), fixEventAPIEntry("EventAPI1", "EventAPI1Name"))
		existingServiceToUpdate1 := fixService("bundle2", fixServiceAPIEntryWithOauth("API2", "bundle2"), fixEventAPIEntry("EventAPI2", "EventAPI2Name"))
		runtimeServiceToUpdate1 := fixService("bundle2", fixServiceAPIEntry("API2"), fixEventAPIEntry("EventAPI2", "EventAPI2Name"))
		existingServiceToUpdate2 := fixService("bundle3", fixServiceAPIEntry("API2"), fixEventAPIEntry("EventAPI3", "EventAPI3Name"))
		runtimeServiceToUpdate2 := fixService("bundle3", fixServiceAPIEntryWithBasic("API3", "bundle3"), fixEventAPIEntry("EventAPI3", "EventAPI3Name"))
		runtimeServiceToDelete1 := fixService("bundle4", fixServiceAPIEntry("API4"), fixEventAPIEntry("EventAPI4", "EventAPI4Name"))
		runtimeServiceToDelete2 := fixService("bundle5", fixServiceAPIEntryWithBasic("API5", "bundle5"), fixEventAPIEntry("EventAPI5", "EventAPI5Name"))

		newRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeServiceToCreate, runtimeServiceToUpdate1, runtimeServiceToUpdate2})

		directorApplications := []model.Application{
			directorApplication,
		}

		existingRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{existingServiceToUpdate1, existingServiceToUpdate2, runtimeServiceToDelete1, runtimeServiceToDelete2})
		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{existingRuntimeApplication},
		}

		converterMock.On("Do", directorApplication).Return(newRuntimeApplication)
		applicationsManagerMock.On("Update", &newRuntimeApplication).Return(&newRuntimeApplication, nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
		applicationsManagerMock.On("Get", "name1", metav1.GetOptions{}).Return(&existingRuntimeApplication, nil)

		credentialsServiceMock.On("Upsert", "name1", newRuntimeApplication.UID, "bundle1", authBundle1.Credentials).Return(applications.Credentials{}, nil)
		credentialsServiceMock.On("Delete", "name1-bundle2").Return(nil)
		credentialsServiceMock.On("Upsert", "name1", newRuntimeApplication.UID, "bundle3", authBundle3.Credentials).Return(applications.Credentials{}, nil)
		credentialsServiceMock.On("Delete", "name1-bundle5").Return(nil)

		requestParametersServiceMock.On("Upsert", "name1", newRuntimeApplication.UID, "bundle1", authBundle1.RequestParameters).Return("", nil)
		requestParametersServiceMock.On("Delete", "params-name1-bundle2").Return(nil)
		requestParametersServiceMock.On("Upsert", "name1", newRuntimeApplication.UID, "bundle3", authBundle3.RequestParameters).Return("", nil)
		requestParametersServiceMock.On("Delete", "params-name1-bundle5").Return(nil)

		expectedResult := []Result{
			{
				ApplicationName: "name1",
				ApplicationID:   "id1",
				Operation:       Update,
				Error:           nil,
			},
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, credentialsServiceMock, requestParametersServiceMock)
		result, err := kymaService.Apply(directorApplications)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
	})

	t.Run("should apply Delete operation", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		credentialsServiceMock := &appSecrets.CredentialsService{}
		requestParametersServiceMock := &appSecrets.RequestParametersService{}

		runtimeServiceToDelete := fixService("bundle1", fixServiceAPIEntry("API1"), fixEventAPIEntry("EventAPI1", "EventAPI1Name"))
		runtimeApplicationToDelete := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeServiceToDelete})

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				runtimeApplicationToDelete,
			},
		}

		applicationsManagerMock.On("Delete", runtimeApplicationToDelete.Name, &metav1.DeleteOptions{}).Return(nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		expectedResult := []Result{
			{
				ApplicationName: "name1",
				ApplicationID:   "",
				Operation:       Delete,
				Error:           nil,
			},
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, credentialsServiceMock, requestParametersServiceMock)
		result, err := kymaService.Apply([]model.Application{})

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
	})

	t.Run("should apply Delete operation and delete credentials", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		credentialsServiceMock := &appSecrets.CredentialsService{}
		requestParametersServiceMock := &appSecrets.RequestParametersService{}

		runtimeServiceToDelete := fixService("bundle1", fixServiceAPIEntryWithBasic("API1", "bundle1"), fixEventAPIEntry("EventAPI1", "EventAPI1Name"))
		runtimeApplicationToDelete := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeServiceToDelete})

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				runtimeApplicationToDelete,
			},
		}

		applicationsManagerMock.On("Delete", runtimeApplicationToDelete.Name, &metav1.DeleteOptions{}).Return(nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		credentialsServiceMock.On("Delete", "name1-bundle1").Return(nil)
		requestParametersServiceMock.On("Delete", "params-name1-bundle1").Return(nil)

		expectedResult := []Result{
			{
				ApplicationName: "name1",
				ApplicationID:   "",
				Operation:       Delete,
				Error:           nil,
			},
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, credentialsServiceMock, requestParametersServiceMock)
		result, err := kymaService.Apply([]model.Application{})

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
	})

	t.Run("should manage only Applications with CompassMetadata in the Spec", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		credentialsServiceMock := &appSecrets.CredentialsService{}
		requestParametersServiceMock := &appSecrets.RequestParametersService{}

		runtimeServiceToDelete := fixService("bundle1", fixServiceAPIEntry("API1"), fixEventAPIEntry("EventAPI1", "EventAPI1Name"))
		notManagedRuntimeService := fixService("bundle2", fixServiceAPIEntry("API2"), fixEventAPIEntry("EventAPI2", "EventAPI2Name"))

		runtimeApplicationToDelete := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeServiceToDelete})
		notManagedRuntimeApplication := getTestApplicationNotManagedByCompass("id2", []v1alpha1.Service{notManagedRuntimeService})

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				runtimeApplicationToDelete,
				notManagedRuntimeApplication,
			},
		}

		applicationsManagerMock.On("Delete", runtimeApplicationToDelete.Name, &metav1.DeleteOptions{}).Return(nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		expectedResult := []Result{
			{
				ApplicationName: "name1",
				ApplicationID:   "",
				Operation:       Delete,
				Error:           nil,
			},
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, credentialsServiceMock, requestParametersServiceMock)
		result, err := kymaService.Apply([]model.Application{})

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
	})

	t.Run("should not break execution when error occurred when applying Application CR", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		credentialsServiceMock := &appSecrets.CredentialsService{}
		requestParametersServiceMock := &appSecrets.RequestParametersService{}

		newRuntimeService1 := fixService("bundle1", fixServiceAPIEntry("API1"), fixEventAPIEntry("EventAPI1", "EventAPI1Name"))
		newRuntimeService2 := fixService("bundle2", fixServiceAPIEntry("API2"), fixEventAPIEntry("EventAPI2", "EventAPI2Name"))

		existingRuntimeService1 := fixService("bundle3", fixServiceAPIEntry("API3"), fixEventAPIEntry("EventAPI3", "EventAPI1Name"))
		existingRuntimeService2 := fixService("bundle4", fixServiceAPIEntry("API4"), fixEventAPIEntry("EventAPI4", "EventAPI2Name"))

		runtimeServiceToBeDeleted1 := v1alpha1.Service{
			ID: "bundle5",
			Entries: []v1alpha1.Entry{
				fixServiceAPIEntry("API1"),
				fixServiceEventAPIEntry("EventAPI1"),
			},
		}

		api := fixDirectorAPiDefinition("API1", "name", "API description", fixAPISpec())
		eventAPI := fixDirectorEventAPIDefinition("EventAPI1", "name", "Event API 1 description", fixEventAPISpec())

		apiBundle1 := fixAPIBundle("bundle1", []model.APIDefinition{api}, nil, nil)
		apiBundle2 := fixAPIBundle("bundle2", nil, []model.EventAPIDefinition{eventAPI}, nil)
		newDirectorApplication := fixDirectorApplication("id1", "name1", apiBundle1, apiBundle2)

		newRuntimeApplication1 := getTestApplication("name1", "id1", []v1alpha1.Service{newRuntimeService1, newRuntimeService2})

		apiBundle3 := fixAPIBundle("bundle3", []model.APIDefinition{api}, []model.EventAPIDefinition{eventAPI}, nil)

		existingDirectorApplication := fixDirectorApplication("id2", "name2", apiBundle3)
		newRuntimeApplication2 := getTestApplication("name2", "id2", []v1alpha1.Service{newRuntimeService1, newRuntimeService2, existingRuntimeService1, existingRuntimeService2})

		runtimeApplicationToBeDeleted := getTestApplication("name3", "id3", []v1alpha1.Service{runtimeServiceToBeDeleted1})

		directorApplications := []model.Application{
			newDirectorApplication,
			existingDirectorApplication,
		}

		existingRuntimeApplication := getTestApplication("name2", "id2", []v1alpha1.Service{existingRuntimeService1, existingRuntimeService2, runtimeServiceToBeDeleted1})

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				existingRuntimeApplication,
				runtimeApplicationToBeDeleted,
			},
		}

		converterMock.On("Do", newDirectorApplication).Return(newRuntimeApplication1)
		converterMock.On("Do", existingDirectorApplication).Return(newRuntimeApplication2)
		applicationsManagerMock.On("Create", &newRuntimeApplication1).Return(nil, apperrors.Internal("some error"))
		applicationsManagerMock.On("Update", &newRuntimeApplication2).Return(nil, apperrors.Internal("some error"))
		applicationsManagerMock.On("Delete", runtimeApplicationToBeDeleted.Name, &metav1.DeleteOptions{}).Return(apperrors.Internal("some error"))
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, credentialsServiceMock, requestParametersServiceMock)
		result, err := kymaService.Apply(directorApplications)

		// then
		require.NoError(t, err)
		require.Equal(t, 3, len(result))
		assert.NotNil(t, result[0].Error)
		assert.NotNil(t, result[1].Error)
		assert.NotNil(t, result[2].Error)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
	})
}

func getTestApplication(name, id string, services []v1alpha1.Service) v1alpha1.Application {
	testApplication := getTestApplicationNotManagedByCompass(name, services)
	testApplication.Spec.CompassMetadata = &v1alpha1.CompassMetadata{Authentication: v1alpha1.Authentication{ClientIds: []string{id}}}

	return testApplication
}

func getTestDirectorApplication(id, name string, apiDefinitions []model.APIDefinition, eventApiDefinitions []model.EventAPIDefinition) model.Application {
	return model.Application{
		ID:   id,
		Name: name,
	}
}

func getTestApplicationNotManagedByCompass(id string, services []v1alpha1.Service) v1alpha1.Application {
	return v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
			UID:  "11e912a4-b489-11eb-8529-0242ac130003",
		},
		Spec: v1alpha1.ApplicationSpec{
			Description: "Description",
			Services:    services,
		},
	}
}

func fixDirectorAPiDefinition(id, name, description string, spec *model.APISpec) model.APIDefinition {
	return model.APIDefinition{
		ID:          id,
		Name:        name,
		Description: description,
		TargetUrl:   "www.example.com",
		APISpec:     spec,
	}
}

func fixDirectorEventAPIDefinition(id, name, description string, spec *model.EventAPISpec) model.EventAPIDefinition {
	return model.EventAPIDefinition{
		ID:           id,
		Name:         name,
		Description:  description,
		EventAPISpec: spec,
	}
}

func fixDirectorApplication(id, name string, apiBundles ...model.APIBundle) model.Application {
	return model.Application{
		ID:         id,
		Name:       name,
		ApiBundles: apiBundles,
	}
}

func fixAPIBundle(id string, apiDefinitions []model.APIDefinition, eventAPIDefinitions []model.EventAPIDefinition, defaultInstanceAuth *model.Auth) model.APIBundle {
	return model.APIBundle{
		ID:                  id,
		APIDefinitions:      apiDefinitions,
		EventDefinitions:    eventAPIDefinitions,
		DefaultInstanceAuth: defaultInstanceAuth,
	}
}

func fixAPIEntry(id, name string) v1alpha1.Entry {
	return v1alpha1.Entry{
		ID:        id,
		Name:      name,
		Type:      applications.SpecAPIType,
		TargetUrl: "www.example.com/1",
	}
}

func fixEventAPIEntry(id, name string) v1alpha1.Entry {
	return v1alpha1.Entry{
		ID:   id,
		Name: name,
		Type: applications.SpecEventsType,
	}
}

func fixAPISpec() *model.APISpec {
	return &model.APISpec{
		Data:   []byte("spec"),
		Type:   model.APISpecTypeOpenAPI,
		Format: model.SpecFormatJSON,
	}
}

func fixEventAPISpec() *model.EventAPISpec {
	return &model.EventAPISpec{
		Data:   []byte("spec"),
		Type:   model.EventAPISpecTypeAsyncAPI,
		Format: model.SpecFormatJSON,
	}
}

func fixServiceAPIEntry(id string) v1alpha1.Entry {
	return v1alpha1.Entry{
		ID:        id,
		Name:      "Name",
		Type:      applications.SpecAPIType,
		TargetUrl: "www.example.com/1",
	}
}

func fixServiceAPIEntryWithOauth(id, bundleID string) v1alpha1.Entry {
	application := "name1"
	return v1alpha1.Entry{
		ID:        id,
		Name:      "Name",
		Type:      applications.SpecAPIType,
		TargetUrl: "www.example.com/1",
		Credentials: v1alpha1.Credentials{
			Type:              "OAuth",
			SecretName:        fmt.Sprintf("%s-%s", application, bundleID),
			AuthenticationUrl: "https://dev-name.eu.auth0.com/oauth/token",
			CSRFInfo:          nil,
		},
		RequestParametersSecretName: fmt.Sprintf("params-%s-%s", application, bundleID),
	}
}

func fixServiceAPIEntryWithBasic(id, bundleID string) v1alpha1.Entry {
	application := "name1"
	return v1alpha1.Entry{
		ID:        id,
		Name:      "Name",
		Type:      applications.SpecAPIType,
		TargetUrl: "www.example.com/1",
		Credentials: v1alpha1.Credentials{
			Type:       "Basic",
			SecretName: fmt.Sprintf("%s-%s", application, bundleID),
		},
		RequestParametersSecretName: fmt.Sprintf("params-%s-%s", application, bundleID),
	}
}

func fixServiceEventAPIEntry(id string) v1alpha1.Entry {
	return v1alpha1.Entry{
		ID:   id,
		Name: "Name",
		Type: applications.SpecEventsType,
	}
}

func fixService(serviceID string, entries ...v1alpha1.Entry) v1alpha1.Service {
	return v1alpha1.Service{
		ID:      serviceID,
		Entries: entries,
	}
}

func fixAuthOauth() *model.Auth {
	return &model.Auth{
		Credentials: &model.Credentials{
			Oauth: &model.Oauth{
				URL:          "https://auth.example.com",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			},
			CSRFInfo: nil,
		},
		RequestParameters: &model.RequestParameters{
			Headers: &map[string][]string{"header1": {"header-value1"}},
		},
	}
}

func fixAuthBasic() *model.Auth {
	return &model.Auth{
		Credentials: &model.Credentials{
			Basic: &model.Basic{
				Username: "my-user",
				Password: "my-password",
			},
		},
		RequestParameters: &model.RequestParameters{
			Headers: &map[string][]string{"header2": {"header-value2"}},
		},
	}
}

func fixAuthRequestParameters() *model.Auth {
	return &model.Auth{
		RequestParameters: &model.RequestParameters{
			Headers: &map[string][]string{"header3": {"header-value3"}},
		},
	}
}
