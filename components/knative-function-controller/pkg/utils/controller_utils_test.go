package utils_test

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/kyma-project/kyma/components/knative-function-controller/pkg/utils"

	corev1 "k8s.io/api/core/v1"
)

func TestNewRuntimeInfo(t *testing.T) {

	g := gomega.NewGomegaWithT(t)
	cm := &corev1.ConfigMap{
		Data: map[string]string{
			"serviceAccountName": "test",
			"dockerRegistry":     "foo",
		},
	}
	ri, err := utils.New(cm)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(ri.ServiceAccount).To(gomega.Equal("test"))
	g.Expect(ri.RegistryInfo).To(gomega.Equal("foo"))

	cmBroken := &corev1.ConfigMap{
		Data: map[string]string{
			"serviceAccountName": "test",
		},
	}
	_, err = utils.New(cmBroken)
	g.Expect(err.Error()).To(gomega.Equal("Error while fetching docker registry info from configmap"))

	cmBroken = &corev1.ConfigMap{
		Data: map[string]string{
			"dockerRegistry": "foo",
		},
	}
	_, err = utils.New(cmBroken)
	g.Expect(err.Error()).To(gomega.Equal("Error while fetching serviceAccountName"))

	cmBroken = &corev1.ConfigMap{
		Data: map[string]string{
			"serviceAccountName": "test",
			"dockerRegistry":     "foo",
			"runtimes":           "foo",
		},
	}
	_, err = utils.New(cmBroken)
	g.Expect(err.Error()).To(gomega.ContainSubstring("unmarshal"))

}

func TestDockerFileConfigMapName(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	runtime := "nodejs8"
	cm := &corev1.ConfigMap{
		Data: map[string]string{
			"serviceAccountName": "test",
			"dockerRegistry":     "foo",
			"runtimes": `[
				{
					"ID": "nodejs8",
					"DockerFileName": "dockerfile-nodejs8",
				}
			]`,
		},
	}
	ri, err := utils.New(cm)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	dockerFileCMName := ri.DockerFileConfigMapName(runtime)
	g.Expect(dockerFileCMName).To(gomega.Equal("dockerfile-nodejs8"))
	dockerFileCMName = ri.DockerFileConfigMapName("foo")
	g.Expect(dockerFileCMName).To(gomega.Equal(""))
}
