package controller

import (
	"testing"

	"github.com/kyma-project/kyma/components/environments/internal"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testNamespacesClient struct {
	GetNamespaceCalled    bool
	UpdateNamespaceCalled bool
}

type testRolesClient struct {
	GetRoleCalled    bool
	GetListCalled    bool
	CreateRoleCalled bool
	DeleteRoleCalled bool
}

type testLimitRangeClient struct {
	CreateCalled bool
	DeleteCalled bool
}

type testResourceQuotaClient struct {
	CreateCalled bool
	DeleteCalled bool
}

var testNamespace = &v1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: "testNamespace",
	},
}

var testEnvironmentsConfig = &EnvironmentsConfig{
	Namespace: "configSecretNamespace",
	LimitRangeMemory: LimitRangeConfig{
		Max:            formattedQuantity("1024Mi"),
		Default:        formattedQuantity("96Mi"),
		DefaultRequest: formattedQuantity("32Mi"),
	},
	ResourceQuota: ResourceQuotaConfig{
		LimitsMemory:   formattedQuantity("1Gi"),
		RequestsMemory: formattedQuantity("256Mi"),
	},
}

var testLimitRange = &v1.LimitRange{
	ObjectMeta: metav1.ObjectMeta{
		Name: testNamespace.Name,
	},
	Spec: v1.LimitRangeSpec{
		Limits: []v1.LimitRangeItem{
			{
				Type: v1.LimitTypeContainer,
				Default: v1.ResourceList{
					v1.ResourceMemory: *testEnvironmentsConfig.LimitRangeMemory.Default.AsQuantity(),
				},
				DefaultRequest: v1.ResourceList{
					v1.ResourceMemory: *testEnvironmentsConfig.LimitRangeMemory.DefaultRequest.AsQuantity(),
				},
				Max: v1.ResourceList{
					v1.ResourceMemory: *testEnvironmentsConfig.LimitRangeMemory.Max.AsQuantity(),
				},
			},
		},
	},
}

var testRole = rbacv1.Role{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "kyma-admin",
		Namespace: "configSecretNamespace",
	},
}

var testRolesList = &rbacv1.RoleList{
	Items: []rbacv1.Role{testRole},
}

func GetTestSetup() (envs *environments, nc *testNamespacesClient, rc *testRolesClient, lr *testLimitRangeClient, rq *testResourceQuotaClient) {

	nc = &testNamespacesClient{GetNamespaceCalled: false, UpdateNamespaceCalled: false}
	rc = &testRolesClient{GetRoleCalled: false, GetListCalled: false, CreateRoleCalled: false, DeleteRoleCalled: false}
	lr = &testLimitRangeClient{DeleteCalled: false, CreateCalled: false}
	rq = &testResourceQuotaClient{DeleteCalled: false, CreateCalled: false}

	envs = &environments{
		Clientset:           nil,
		Config:              testEnvironmentsConfig,
		NamespacesClient:    nc,
		RolesClient:         rc,
		LimitRangeClient:    lr,
		ResourceQuotaClient: rq,
		ErrorHandlers:       &internal.ErrorHandlers{},
	}

	return envs, nc, rc, lr, rq
}

func (nc *testNamespacesClient) GetNamespace(name string) (result *v1.Namespace, err error) {
	So(testNamespace.Name, ShouldEqual, name)

	nc.GetNamespaceCalled = true

	return testNamespace, nil
}

func (nc *testNamespacesClient) UpdateNamespace(namespace *v1.Namespace) (result *v1.Namespace, err error) {
	So(testNamespace.Name, ShouldEqual, namespace.Name)
	nc.UpdateNamespaceCalled = true

	return testNamespace, nil
}

func (rc *testRolesClient) GetList(namespace string, opts metav1.ListOptions) (*rbacv1.RoleList, error) {
	So(testEnvironmentsConfig.Namespace, ShouldEqual, namespace)
	rc.GetListCalled = true

	return testRolesList, nil
}

func (rc *testRolesClient) GetRole(name string, namespace string) (*rbacv1.Role, error) {
	So(testNamespace.Name, ShouldEqual, namespace)
	So(testRole.ObjectMeta.Name, ShouldEqual, name)
	rc.GetRoleCalled = true

	return &testRole, nil
}

func (rc *testRolesClient) CreateRole(role *rbacv1.Role, namespace string) (*rbacv1.Role, error) {
	So(testNamespace.Name, ShouldEqual, namespace)
	So(testRole.ObjectMeta.Name, ShouldEqual, role.ObjectMeta.Name)
	rc.CreateRoleCalled = true

	return &testRole, nil
}

func (rc *testRolesClient) DeleteRole(name string, namespace string) error {
	So(testNamespace.Name, ShouldEqual, namespace)
	So(testRole.ObjectMeta.Name, ShouldEqual, name)
	rc.DeleteRoleCalled = true

	return nil
}

func (lr *testLimitRangeClient) CreateLimitRange(namespace string, limitRange *v1.LimitRange) error {
	So(testNamespace.Name, ShouldEqual, namespace)
	So(limitRange.Name, ShouldEqual, "kyma-default")
	So(limitRange.Spec.Limits[0].Default.Memory().Value(), ShouldEqual, testEnvironmentsConfig.LimitRangeMemory.Default.AsQuantity().Value())
	So(limitRange.Spec.Limits[0].DefaultRequest.Memory().Value(), ShouldEqual, testEnvironmentsConfig.LimitRangeMemory.DefaultRequest.AsQuantity().Value())
	So(limitRange.Spec.Limits[0].Max.Memory().Value(), ShouldEqual, testEnvironmentsConfig.LimitRangeMemory.Max.AsQuantity().Value())
	lr.CreateCalled = true

	return nil
}

func (lr *testLimitRangeClient) DeleteLimitRange(namespace string) error {
	So(testNamespace.Name, ShouldEqual, namespace)
	lr.DeleteCalled = true

	return nil
}

func (lr *testResourceQuotaClient) CreateResourceQuota(namespace string, resourceQuota *v1.ResourceQuota) error {
	rm := resourceQuota.Spec.Hard[v1.ResourceRequestsMemory]
	lm := resourceQuota.Spec.Hard[v1.ResourceLimitsMemory]

	So(testNamespace.Name, ShouldEqual, namespace)
	So(resourceQuota.Name, ShouldEqual, "kyma-default")
	So(rm.String(), ShouldEqual, "256Mi")
	So(lm.String(), ShouldEqual, "1Gi")
	lr.CreateCalled = true
	return nil
}

func (lr *testResourceQuotaClient) DeleteResourceQuota(namespace string) error {
	So(testNamespace.Name, ShouldEqual, namespace)
	lr.DeleteCalled = true
	return nil
}

func TestYfenvironments(t *testing.T) {
	Convey("Adding roles for environment shouldn't return error", t, func() {

		envs, nc, rc, lr, _ := GetTestSetup()

		err := envs.AddRolesForEnvironment(testNamespace)

		So(err, ShouldBeNil)
		So(nc.GetNamespaceCalled, ShouldBeTrue)
		So(nc.UpdateNamespaceCalled, ShouldBeTrue)

		allLimitRangeClientMethodsShouldNotBeCalled(lr)

		So(rc.CreateRoleCalled, ShouldBeTrue)
		So(rc.GetListCalled, ShouldBeTrue)
		So(rc.DeleteRoleCalled, ShouldBeFalse)
		So(rc.GetRoleCalled, ShouldBeFalse)
	})

	Convey("Should not add roles for environment with existing roles", t, func() {

		origNamespace := testNamespace.DeepCopy()
		envs, nc, rc, lr, _ := GetTestSetup()

		annotations := make(map[string]string)
		annotations[rolesAnnotName] = "true"
		testNamespace.SetAnnotations(annotations)

		err := envs.AddRolesForEnvironment(testNamespace)

		So(err, ShouldBeNil)
		So(nc.GetNamespaceCalled, ShouldBeTrue)
		So(nc.UpdateNamespaceCalled, ShouldBeFalse)

		allRolesClientMethodsShuldNotBeCalled(rc)
		allLimitRangeClientMethodsShouldNotBeCalled(lr)

		Reset(func() {
			testNamespace = origNamespace
		})
	})

	Convey("Removing roles from environment shouldn't return error", t, func() {

		origNamespace := testNamespace.DeepCopy()
		envs, nc, rc, lr, _ := GetTestSetup()

		annotations := make(map[string]string)
		annotations[rolesAnnotName] = "true"
		testNamespace.SetAnnotations(annotations)

		err := envs.RemoveRolesFromEnvironment(testNamespace)

		So(err, ShouldBeNil)
		So(nc.GetNamespaceCalled, ShouldBeTrue)
		So(nc.UpdateNamespaceCalled, ShouldBeTrue)

		allLimitRangeClientMethodsShouldNotBeCalled(lr)

		So(rc.GetListCalled, ShouldBeTrue)
		So(rc.DeleteRoleCalled, ShouldBeTrue)
		So(rc.CreateRoleCalled, ShouldBeFalse)
		So(rc.GetRoleCalled, ShouldBeFalse)

		Reset(func() {
			testNamespace = origNamespace
		})
	})

	Convey("istio-inject label", t, func() {

		origNamespace := testNamespace.DeepCopy()
		envs, nc, _, _, _ := GetTestSetup()

		Convey("should be added with no error", func() {

			labels := make(map[string]string)
			testNamespace.SetLabels(labels)

			err := envs.LabelWithIstioInjection(testNamespace)

			So(err, ShouldBeNil)
			So(nc.GetNamespaceCalled, ShouldBeTrue)
			So(nc.UpdateNamespaceCalled, ShouldBeTrue)
		})

		Convey("and removing with no error as well", func() {

			err := envs.RemoveIstioInjectionLabel(testNamespace)

			So(err, ShouldBeNil)
			So(nc.GetNamespaceCalled, ShouldBeTrue)
			So(nc.UpdateNamespaceCalled, ShouldBeTrue)
		})

		Reset(func() {
			testNamespace = origNamespace
		})
	})

	Convey("should create limit range", t, func() {

		origNamespace := testNamespace.DeepCopy()
		envs, _, _, lr, _ := GetTestSetup()

		err := envs.CreateLimitRangeForEnv(testNamespace)

		So(err, ShouldBeNil)
		So(lr.CreateCalled, ShouldBeTrue)

		Reset(func() {
			testNamespace = origNamespace
		})
	})

	Convey("should delete limit range", t, func() {

		origNamespace := testNamespace.DeepCopy()
		envs, _, _, lr, _ := GetTestSetup()

		err := envs.DeleteLimitRange(testNamespace)

		So(err, ShouldBeNil)
		So(lr.DeleteCalled, ShouldBeTrue)

		Reset(func() {
			testNamespace = origNamespace
		})
	})

	Convey("should create resource quota", t, func() {

		origNamespace := testNamespace.DeepCopy()
		envs, _, _, _, rq := GetTestSetup()

		err := envs.CreateResourceQuota(testNamespace)

		So(err, ShouldBeNil)
		So(rq.CreateCalled, ShouldBeTrue)

		Reset(func() {
			testNamespace = origNamespace
		})
	})

	Convey("should delete resource quota", t, func() {

		origNamespace := testNamespace.DeepCopy()
		envs, _, _, _, rq := GetTestSetup()

		err := envs.DeleteResourceQuota(testNamespace)

		So(err, ShouldBeNil)
		So(rq.DeleteCalled, ShouldBeTrue)

		Reset(func() {
			testNamespace = origNamespace
		})
	})
}

func allRolesClientMethodsShuldNotBeCalled(rc *testRolesClient) {
	So(rc.CreateRoleCalled, ShouldBeFalse)
	So(rc.GetListCalled, ShouldBeFalse)
	So(rc.GetRoleCalled, ShouldBeFalse)
}

func allLimitRangeClientMethodsShouldNotBeCalled(lr *testLimitRangeClient) {
	So(lr.CreateCalled, ShouldBeFalse)
	So(lr.DeleteCalled, ShouldBeFalse)
}

func formattedQuantity(v string) FormattedQuantity {
	q := resource.MustParse(v)
	return FormattedQuantity(q.Value())
}
