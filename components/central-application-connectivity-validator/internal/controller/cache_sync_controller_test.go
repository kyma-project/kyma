package controller_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"
	"time"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apis/applicationconnector/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testResult struct {
	err        error
	statusCode int
}

var _ = Describe("Cache synchronization controller", func() {
	Context("When controller is running", func() {

		appCount := 1000

		It("should not fail with cache miss", func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
			defer cancel()

			for i := 0; i < appCount; i++ {
				testApplication := fmt.Sprintf("ta%d", i)
				app := application(testApplication)
				Expect(k8sClient.Create(ctx, &app)).To(BeNil())

				Eventually(func() bool {
					_, found := idCache.Get(testApplication)
					return found
				}, 5*time.Second).Should(BeTrue())
			}

			counter := 0
			client := http.Client{}

			Consistently(func() testResult {
				if counter == appCount {
					counter = 0
				}
				ta := fmt.Sprintf("ta%d", counter)
				URL := fmt.Sprintf("http://localhost:%s/%s/v2/events", testProxyServerPort, ta)
				req, err := http.NewRequest(http.MethodGet, URL, nil)
				Expect(err).Should(BeNil())

				req.Header.Add("X-Forwarded-Client-Cert", fmt.Sprintf(`Subject="CN=%s"`, ta))

				resp, err := client.Do(req)
				counter++
				return testResult{
					err:        err,
					statusCode: resp.StatusCode,
				}
			}, time.Second*10, time.Millisecond*25).Should(Equal(testResult{
				statusCode: http.StatusOK,
			}))
		})

		It("should delete application cache when cr is deleted", func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
			defer cancel()

			appName := "deleteme"
			app := application(appName)
			Expect(k8sClient.Create(ctx, &app)).To(BeNil())

			Eventually(func() bool {
				_, found := idCache.Get(appName)
				return found
			}).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, &app)).Should(BeNil())

			Eventually(func() bool {
				_, found := idCache.Get(appName)
				return found
			}).Should(BeFalse())
		})

	})
})

func application(name string) v1alpha1.Application {
	return v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ApplicationSpec{},
	}
}
