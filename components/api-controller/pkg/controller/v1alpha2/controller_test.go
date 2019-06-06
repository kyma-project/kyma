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

	Convey("If validateApi", t, func() {

		c := &Controller{}
		c.apisLister = &fakeAPILister{}

		testAPI := &apis.Api{}
		testAPI.SetName("test-api-1")
		testAPI.SetNamespace("test-ns")
		testAPI.Status.SetInProgress()

		Convey("is fed with an API for a non-occupied service", func() {

			//given
			testAPI.Spec.Service.Name = "non-occupied-service"

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

			//given
			testAPI.Spec.Service.Name = "occupied-service"

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

			//given
			c.apisLister = &failingFakeAPILister{}

			testAPI.Spec.Service.Name = "any-service"

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

			//given
			testAPI.SetName("existing-api")
			testAPI.Spec.Service.Name = "occupied-service"

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
			c.blacklistedServices = []string{"forbidden-service.test-ns"}

			//given
			testAPI.Spec.Service.Name = "forbidden-service"

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
	})
}
