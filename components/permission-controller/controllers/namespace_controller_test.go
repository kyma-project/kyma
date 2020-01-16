package controllers_test

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kyma-project/kyma/components/permission-controller/controllers"
)

const (
	timeout         = 5 * time.Second
	systemNamespace = "test-system"
	adminGroup      = "namespace-admins"
)

type testSetup struct {
	c          client.Client
	requests   chan reconcile.Request
	stopMgr    chan struct{}
	mgrStopped *sync.WaitGroup
}

var _ = Describe("Permissions Controller", func() {
	Context("should watch for namespace creation and", func() {
		Context("create rolebinding if the namespace is not a system one and", func() {
			It("add static user to subjects if UseStaticConnector is set to true", func() {

				testRolebindingName, testNamespaceName := controllers.RolebindingName, "dev"
				expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: testNamespaceName}}

				t := setupTest(true)

				testNamespace := &corev1.Namespace{
					ObjectMeta: v1.ObjectMeta{
						Name: testNamespaceName,
					},
				}
				err := t.c.Create(context.TODO(), testNamespace)
				Expect(err).NotTo(HaveOccurred())
				Eventually(t.requests, timeout).Should(Receive(Equal(*expectedRequest)))

				// Verify if the rolebinding is created
				expectedSubjects := []rbac.Subject{
					{
						Kind:     "Group",
						Name:     adminGroup,
						APIGroup: rbac.GroupName,
					},
					{
						Kind:     "User",
						Name:     controllers.SubjectStaticUser,
						APIGroup: rbac.GroupName,
					},
				}
				var retrieved rbac.RoleBinding
				ok := client.ObjectKey{Name: testRolebindingName, Namespace: testNamespaceName}
				err = t.c.Get(context.TODO(), ok, &retrieved)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.RoleRef.Name).To(Equal(controllers.RoleRefName))
				Expect(retrieved.RoleRef.Kind).To(Equal(controllers.RoleRefKind))
				Expect(retrieved.Subjects).To(Equal(expectedSubjects))

				t.c.Delete(context.TODO(), testNamespace)
				close(t.stopMgr)
				t.mgrStopped.Wait()
			})

			It("not add static user to subjects if UseStaticConnector is set to false", func() {

				testRolebindingName, testNamespaceName := controllers.RolebindingName, "dev-static"
				expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: testNamespaceName}}

				t := setupTest(false)

				testNamespace := &corev1.Namespace{
					ObjectMeta: v1.ObjectMeta{
						Name: testNamespaceName,
					},
				}
				err := t.c.Create(context.TODO(), testNamespace)
				Expect(err).NotTo(HaveOccurred())
				Eventually(t.requests, timeout).Should(Receive(Equal(*expectedRequest)))

				// Verify if the rolebinding is created
				expectedSubjects := []rbac.Subject{
					{
						Kind:     "Group",
						Name:     adminGroup,
						APIGroup: rbac.GroupName,
					},
				}
				var retrieved rbac.RoleBinding
				ok := client.ObjectKey{Name: testRolebindingName, Namespace: testNamespaceName}
				err = t.c.Get(context.TODO(), ok, &retrieved)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.RoleRef.Name).To(Equal(controllers.RoleRefName))
				Expect(retrieved.RoleRef.Kind).To(Equal(controllers.RoleRefKind))
				Expect(retrieved.Subjects).To(Equal(expectedSubjects))

				t.c.Delete(context.TODO(), testNamespace)
				close(t.stopMgr)
				t.mgrStopped.Wait()
			})
		})

		It("not create a rolebinding if the namespace is a system one", func() {

			testRolebindingName, testNamespaceName := controllers.RolebindingName, systemNamespace
			expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: testNamespaceName}}

			t := setupTest(true)

			testNamespace := &corev1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: testNamespaceName,
				},
			}

			err := t.c.Create(context.TODO(), testNamespace)
			Expect(err).NotTo(HaveOccurred())
			Eventually(t.requests, timeout).Should(Receive(Equal(*expectedRequest)))

			// Verify if the rolebinding is not crated
			var retrieved rbac.RoleBinding
			ok := client.ObjectKey{Name: testRolebindingName, Namespace: testNamespaceName}
			err = t.c.Get(context.TODO(), ok, &retrieved)
			Expect(err).To(HaveOccurred())

			t.c.Delete(context.TODO(), testNamespace)
			close(t.stopMgr)
			t.mgrStopped.Wait()
		})
	})
})

func getAPIReconciler(mgr ctrl.Manager, useStaticConnector bool) reconcile.Reconciler {
	return &controllers.NamespaceReconciler{
		Client:             mgr.GetClient(),
		Log:                ctrl.Log.WithName("controllers").WithName("NamespaceReconciler"),
		ExcludedNamespaces: []string{systemNamespace},
		SubjectGroups:      []string{adminGroup},
		UseStaticConnector: useStaticConnector,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("permission_controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Api
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

func setupTest(useStaticConnector bool) testSetup {
	s := runtime.NewScheme()
	err := rbac.AddToScheme(s)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(s)
	Expect(err).NotTo(HaveOccurred())

	// Setup Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{Scheme: s})
	Expect(err).NotTo(HaveOccurred())
	c := mgr.GetClient()

	reconcileFn, requests := SetupTestReconcile(getAPIReconciler(mgr, useStaticConnector))
	Expect(add(mgr, reconcileFn)).To(Succeed())

	// Start the manager and the controller
	stopMgr, mgrStopped := StartTestManager(mgr)

	return testSetup{
		c:          c,
		requests:   requests,
		stopMgr:    stopMgr,
		mgrStopped: mgrStopped,
	}
}
