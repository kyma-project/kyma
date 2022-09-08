package testing

import (
	"net/url"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_GetRestAPIObject(t *testing.T) {
	g := NewGomegaWithT(t)

	urlString := "/messaging/events/subscriptions/my-subscription"
	urlObject, err := url.Parse(urlString)
	g.Expect(err).ShouldNot(HaveOccurred())
	restObject := GetRestAPIObject(urlObject)
	g.Expect(restObject).To(Equal("my-subscription"))
}
