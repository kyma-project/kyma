package bucket

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/buckethandler/automock"
	"testing"
	"time"

	assetstorev1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
	instance := &assetstorev1alpha1.Bucket{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"}}

	bucketHandler := &automock.BucketHandler{}
	bucketHandler.On("CreateIfDoesntExist").Return(true, nil)


	cfg := prepareReconcilerTest(t, bucketHandler)
	g := cfg.g
	c := cfg.c

	defer cfg.finishTest()


	err := c.Create(context.TODO(), instance)

	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}

	g.Expect(err).NotTo(gomega.HaveOccurred())

	defer c.Delete(context.TODO(), instance)

	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	

}


