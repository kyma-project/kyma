package v1alpha2

import (
	"errors"
	"testing"

	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/meta/v1"
	apis "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	listers "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/listers/gateway.kyma-project.io/v1alpha2"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/labels"
)

var fixHostnameTestCases = []struct {
	hostname   string
	domainName string
	expected   string
}{
	{"my-service", "kyma.local", "my-service.kyma.local"},
	{"my-service.kyma.local", "kyma.local", "my-service.kyma.local"},
}

func TestFixHostname(t *testing.T) {
	for _, tc := range fixHostnameTestCases {
		if fixHostname(tc.domainName, tc.hostname) != tc.expected {
			t.Errorf("Fixed hostname %s should be: %s", tc.hostname, tc.expected)
		}
	}
}

type fakeAPINamespaceLister struct{}

func (fnl *fakeAPINamespaceLister) Get(name string) (*apis.Api, error) { return nil, nil }

func (fnl *fakeAPINamespaceLister) List(selector labels.Selector) (ret []*apis.Api, err error) {
	fakeAPI := &apis.Api{}
	fakeAPI.SetName("existing-api")
	fakeAPI.SetNamespace("test-ns")
	fakeAPI.Spec.Service.Name = "occupied-service"

	return []*apis.Api{fakeAPI}, nil
}

type failingFakeAPINamespaceLister struct{}

func (ffnl *failingFakeAPINamespaceLister) Get(name string) (*apis.Api, error) { return nil, nil }

func (ffnl *failingFakeAPINamespaceLister) List(selector labels.Selector) (ret []*apis.Api, err error) {
	return nil, errors.New("unable to list existing APIs")
}

type fakeAPILister struct{}

func (fal *fakeAPILister) List(selector labels.Selector) (ret []*apis.Api, err error) { return nil, nil }

func (fal *fakeAPILister) Apis(namespace string) listers.ApiNamespaceLister {
	return &fakeAPINamespaceLister{}
}

type failingFakeAPILister struct{}

func (ffal *failingFakeAPILister) List(selector labels.Selector) (ret []*apis.Api, err error) {
	return nil, nil
}

func (ffal *failingFakeAPILister) Apis(namespace string) listers.ApiNamespaceLister {
	return &failingFakeAPINamespaceLister{}
}

func TestValidateApi(t *testing.T) {

	var initController = func() Controller{
		c := &Controller{}
		c.apisLister = &fakeAPILister{}

		return *c
	}

	var getTestAPI = func(name, serviceName, namespace  string) *apis.Api {
		testAPI := &apis.Api{}
		testAPI.SetName(name)
		testAPI.SetNamespace(namespace)
		testAPI.Spec.Service.Name = serviceName
		testAPI.Status.SetInProgress()

		return testAPI
	}

	Convey("If validateApi", t, func() {

		Convey("is fed with an API for a non-occupied service", func() {
			c := initController()

			//given
			testAPI := getTestAPI("test-api-1", "non-occupied-service", "test-ns")
			statusHelper := NewApiStatusHelper(nil, testAPI)

			Convey("it should update the helper with the \"Successful\" status code and return this code", func() {

				//when
				statusCode := c.validateAPI(testAPI, statusHelper)

				//then
				So(statusHelper.hasChanged, ShouldBeTrue)
				So(statusHelper.apiCopy.Status.ValidationStatus.IsSuccessful(), ShouldBeTrue)
				So(statusHelper.apiCopy.Status.IsTargetServiceOccupied(), ShouldBeFalse)
				So(statusCode, ShouldEqual, kymaMeta.Successful)
			})
		})

		Convey("is fed with an API for an occupied service", func() {
			c := initController()

			//given
			testAPI := getTestAPI("test-api-1","occupied-service", "test-ns")
			statusHelper := NewApiStatusHelper(nil, testAPI)

			Convey("it should update the helper with the \"TargetServiceOccupied\" status code and return this code", func() {

				//when
				statusCode := c.validateAPI(testAPI, statusHelper)

				//then
				So(statusHelper.hasChanged, ShouldBeTrue)
				So(statusHelper.apiCopy.Status.ValidationStatus.IsSuccessful(), ShouldBeFalse)
				So(statusHelper.apiCopy.Status.IsTargetServiceOccupied(), ShouldBeTrue)
				So(statusCode, ShouldEqual, kymaMeta.TargetServiceOccupied)
			})
		})

		Convey("is unable to list existing APIs", func() {
			c := initController()
			c.apisLister = &failingFakeAPILister{}

			//given
			testAPI := getTestAPI("test-api-1","any-service", "test-ns")
			statusHelper := NewApiStatusHelper(nil, testAPI)

			Convey("it should update the helper with an error and return the \"Error\" status code", func() {

				//when
				statusCode := c.validateAPI(testAPI, statusHelper)

				//then
				So(statusHelper.hasChanged, ShouldBeTrue)
				So(statusHelper.apiCopy.Status.ValidationStatus.IsSuccessful(), ShouldBeFalse)
				So(statusHelper.apiCopy.Status.IsError(), ShouldBeTrue)
				So(statusCode, ShouldEqual, kymaMeta.Error)
			})
		})

		Convey("is called in OnUpdate context", func() {
			c := initController()

			//given
			testAPI := getTestAPI("existing-api","occupied-service", "test-ns")
			statusHelper := NewApiStatusHelper(nil, testAPI)

			Convey("it should return the \"Successful\" status code if the service has not been changed", func() {

				//when
				statusCode := c.validateAPI(testAPI, statusHelper)

				//then
				So(statusHelper.hasChanged, ShouldBeTrue)
				So(statusHelper.apiCopy.Status.ValidationStatus.IsSuccessful(), ShouldBeTrue)
				So(statusHelper.apiCopy.Status.IsTargetServiceOccupied(), ShouldBeFalse)
				So(statusCode, ShouldEqual, kymaMeta.Successful)
			})
		})

		Convey("is fed with an API for a forbidden service", func() {
			c := initController()
			c.blacklistedServices = []string{"forbidden-service.test-ns"}

			//given
			testAPI := getTestAPI("test-api-1","forbidden-service", "test-ns")
			statusHelper := NewApiStatusHelper(nil, testAPI)

			Convey("it should update the helper with the \"Error\" status code and return this code", func() {

				//when
				statusCode := c.validateAPI(testAPI, statusHelper)

				//then
				So(statusHelper.hasChanged, ShouldBeTrue)
				So(statusHelper.apiCopy.Status.ValidationStatus.IsSuccessful(), ShouldBeFalse)
				So(statusHelper.apiCopy.Status.IsTargetServiceOccupied(), ShouldBeFalse)
				So(statusCode, ShouldEqual, kymaMeta.Error)
			})
		})

		Convey("is fed with an API for a forbidden services", func() {
			c := initController()
			c.blacklistedServices = []string{"forbidden-service-1.test-ns", "forbidden-service-2.test-ns"}

			//given
			testAPI := getTestAPI("test-api-1","forbidden-service-1", "test-ns")
			statusHelper := NewApiStatusHelper(nil, testAPI)

			testAPI2 := getTestAPI("test-api-2","forbidden-service-2", "test-ns")
			statusHelper2 := NewApiStatusHelper(nil, testAPI2)

			Convey("it should update the helper with the \"Error\" status code and return this code for both APIs", func() {

				//when
				statusCode := c.validateAPI(testAPI, statusHelper)
				statusCode2 := c.validateAPI(testAPI2, statusHelper2)

				//then
				So(statusHelper.hasChanged, ShouldBeTrue)
				So(statusHelper.apiCopy.Status.ValidationStatus.IsSuccessful(), ShouldBeFalse)
				So(statusHelper.apiCopy.Status.IsTargetServiceOccupied(), ShouldBeFalse)
				So(statusCode, ShouldEqual, kymaMeta.Error)

				So(statusHelper2.hasChanged, ShouldBeTrue)
				So(statusHelper2.apiCopy.Status.ValidationStatus.IsSuccessful(), ShouldBeFalse)
				So(statusHelper2.apiCopy.Status.IsTargetServiceOccupied(), ShouldBeFalse)
				So(statusCode2, ShouldEqual, kymaMeta.Error)
			})
		})
	})
}
