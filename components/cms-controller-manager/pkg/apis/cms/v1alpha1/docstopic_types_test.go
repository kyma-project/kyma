package v1alpha1

import (
	"testing"
)

func TestStorageDocsTopic(t *testing.T) {
	//key := types.NamespacedName{
	//	Name:      "foo",
	//	Namespace: "default",
	//}
	//created := &DocsTopic{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "foo",
	//		Namespace: "default",
	//	}}
	//g := gomega.NewGomegaWithT(t)
	//
	//// Test Create
	//fetched := &DocsTopic{}
	//g.Expect(c.Create(context.TODO(), created)).NotTo(gomega.HaveOccurred())
	//
	//g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
	//g.Expect(fetched).To(gomega.Equal(created))
	//
	//// Test Updating the Labels
	//updated := fetched.DeepCopy()
	//updated.Labels = map[string]string{"hello": "world"}
	//g.Expect(c.Update(context.TODO(), updated)).NotTo(gomega.HaveOccurred())
	//
	//g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
	//g.Expect(fetched).To(gomega.Equal(updated))
	//
	//// Test Delete
	//g.Expect(c.Delete(context.TODO(), fetched)).NotTo(gomega.HaveOccurred())
	//g.Expect(c.Get(context.TODO(), key, fetched)).To(gomega.HaveOccurred())
}
