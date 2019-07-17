/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package compassconnection

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/stretchr/testify/require"

	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//var c client.Client
//
var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}

//var depKey = types.NamespacedName{Name: "foo-deployment", Namespace: "default"}

const timeout = time.Second * 5

const (
	appName = "compass"

	compassConnectionName = "compass-connection"
)

var (
	compassConnectionNamespacedName = types.NamespacedName{
		Name: compassConnectionName,
	}
)

func TestReconcile(t *testing.T) {

	g := gomega.NewGomegaWithT(t)
	instance := &v1alpha1.CompassConnection{ObjectMeta: v1.ObjectMeta{Name: "foo", Namespace: "default"}}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(c))
	//g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())

	err = startController(mgr, recFn)
	require.NoError(t, err)

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the Greeting object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	//if apierrors.IsInvalid(err) {
	//	t.Logf("failed to create object, got an invalid object error: %v", err)
	//	return
	//}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	//
	//deploy := &appsv1.Deployment{}
	//g.Eventually(func() error { return c.Get(context.TODO(), depKey, deploy) }, timeout).
	//	Should(gomega.Succeed())
	//
	//Delete the Deployment and expect Reconcile to be called for Deployment deletion
	//g.Expect(c.Delete(context.TODO(), deploy)).NotTo(gomega.HaveOccurred())
	//g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	//g.Eventually(func() error { return c.Get(context.TODO(), depKey, deploy) }, timeout).
	//	Should(gomega.Succeed())
	//
	//Manually delete Deployment since GC isn't enabled in the test control plane
	//g.Eventually(func() error { return c.Delete(context.TODO(), deploy) }, timeout).
	//	Should(gomega.MatchError("deployments.apps \"foo-deployment\" not found"))

}

//
//func TestReconcile(t *testing.T) {
//
//	t.Run("should reconcile", func(t *testing.T) {
//		// given
//		compassConnection := &v1alpha1.CompassConnection{
//			ObjectMeta: v1.ObjectMeta{
//				Name: compassConnectionName,
//			},
//		}
//
//		client := NewFakeClient(compassConnection)
//
//		reconciler := newReconciler(client)
//
//		request := reconcile.Request{
//			NamespacedName: compassConnectionNamespacedName,
//		}
//
//		// when
//		result, err := reconciler.Reconcile(request)
//
//		// then
//		require.NoError(t, err)
//		require.Empty(t, result)
//
//	})
//}
