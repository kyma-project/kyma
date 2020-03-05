package controllers

import (
	"strconv"
	"testing"

	"github.com/onsi/gomega"
)

func TestRegistryHelper(t *testing.T) {
	testCases := []struct {
		rh                       RegistryHelper
		name                     string
		namespace                string
		tag                      string
		expectedImageName        string
		expectedBuildImageName   string
		expectedServiceImageName string
	}{
		{
			rh:                       DefaultRegistryHelper,
			name:                     "test",
			namespace:                "default",
			tag:                      "123",
			expectedImageName:        "test-default:123",
			expectedBuildImageName:   "function-controller-docker-registry.kyma-system.svc.cluster.local:5000/test-default:123",
			expectedServiceImageName: "https://registry.kyma.local/test-default:123",
		},
		{
			rh: &registryHelper{
				dockerRegistryFQDN:            "FQDN",
				dockerRegistryPort:            1,
				dockerRegistryName:            "testregistryname",
				dockerRegistryExternalAddress: "https://test.me",
			},
			name:                     "t1",
			namespace:                "test",
			tag:                      "456",
			expectedImageName:        "testregistryname/t1-test:456",
			expectedBuildImageName:   "FQDN:1/testregistryname/t1-test:456",
			expectedServiceImageName: "https://test.me/testregistryname/t1-test:456",
		},
	}
	for i, tC := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := gomega.NewWithT(t)

			actualImageName := tC.rh.ImageName(tC.name, tC.namespace, tC.tag)
			g.Expect(actualImageName).Should(gomega.Equal(tC.expectedImageName))

			actualBuildImageName := tC.rh.BuildImageName(tC.name, tC.namespace, tC.tag)
			g.Expect(actualBuildImageName).Should(gomega.Equal(tC.expectedBuildImageName))

			actualServiceImageName := tC.rh.ServiceImageName(tC.name, tC.namespace, tC.tag)
			g.Expect(actualServiceImageName).Should(gomega.Equal(tC.expectedServiceImageName))
		})
	}
}
