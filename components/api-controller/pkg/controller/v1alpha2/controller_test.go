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

type fakeAPILister struct {
	name string
}

func (l *fakeAPILister) List(selector labels.Selector) (ret []*apis.Api, err error) {

	fakeAPI := &apis.Api{}
	fakeAPI.Spec.Service.Name = "occupied-service"
	fakeAPI.SetUID("0")

	return []*apis.Api{fakeAPI}, nil
}

func (l *fakeAPILister) Apis(namespace string) listers.ApiNamespaceLister { return nil }

type failingFakeAPILister struct{}

func (fl *failingFakeAPILister) List(selector labels.Selector) (ret []*apis.Api, err error) {
	return nil, errors.New("unable to list existing APIs")
}

func (fl *failingFakeAPILister) Apis(namespace string) listers.ApiNamespaceLister { return nil }

func TestValidateApi(t *testing.T) {

	Convey("If validateApi", t, func() {

		c := &Controller{}
		c.apisLister = &fakeAPILister{}
		testAPI := &apis.Api{}
		testAPI.Status.SetInProgress()
		testAPI.SetUID("1")

		Convey("is fed with an API for a non-occupied service", func() {

			//given
			testAPI.Spec.Service.Name = "non-occupied-service"

			statusHelper := NewApiStatusHelper(nil, testAPI)

			Convey("it should update the helper with the \"Done\" status code and return this code", func() {

				//when
				statusCode := c.validateAPI(testAPI, statusHelper)

				//then
				So(statusHelper.hasChanged, ShouldBeTrue)
				So(statusHelper.apiCopy.Status.ValidationStatus.IsDone(), ShouldBeTrue)
				So(statusHelper.apiCopy.Status.IsTargetServiceOccupied(), ShouldBeFalse)
				So(statusCode, ShouldEqual, kymaMeta.Done)
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
				So(statusHelper.apiCopy.Status.ValidationStatus.IsDone(), ShouldBeFalse)
				So(statusHelper.apiCopy.Status.IsTargetServiceOccupied(), ShouldBeTrue)
				So(statusCode, ShouldEqual, kymaMeta.TargetServiceOccupied)
			})
		})

		Convey("is unable to list existing APIs", func() {

			//given
			c.apisLister = &failingFakeAPILister{}

			testAPI.Spec.Service.Name = "any-service"

			statusHelper := NewApiStatusHelper(nil, testAPI)

			Convey("it should update the helper with error and return the \"Error\" status code", func() {

				//when
				statusCode := c.validateAPI(testAPI, statusHelper)

				//then
				So(statusHelper.hasChanged, ShouldBeTrue)
				So(statusHelper.apiCopy.Status.ValidationStatus.IsDone(), ShouldBeFalse)
				So(statusHelper.apiCopy.Status.IsError(), ShouldBeTrue)
				So(statusCode, ShouldEqual, kymaMeta.Error)
			})
		})

		Convey("is called in OnUpdate context", func() {

			//given
			testAPI.Spec.Service.Name = "occupied-service"
			testAPI.SetUID("0")

			statusHelper := NewApiStatusHelper(nil, testAPI)

			Convey("it should return the \"InProgress\" status code if the service has not been changed", func() {

				//when
				statusCode := c.validateAPI(testAPI, statusHelper)

				//then
				So(statusHelper.hasChanged, ShouldBeTrue)
				So(statusHelper.apiCopy.Status.ValidationStatus.IsDone(), ShouldBeTrue)
				So(statusHelper.apiCopy.Status.IsTargetServiceOccupied(), ShouldBeFalse)
				So(statusCode, ShouldEqual, kymaMeta.Done)
			})
		})
	})
}
