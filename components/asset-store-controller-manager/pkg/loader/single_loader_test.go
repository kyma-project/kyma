package loader

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"testing"
)

func TestLoader_Load_Single(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		loader := &loader{
			temporaryDir:    "/tmp",
			osRemoveAllFunc: os.RemoveAll,
			osCreateFunc:    os.Create,
			httpGetFunc:     get,
			ioutilTempDir:   ioutil.TempDir,
		}

		// When
		_, files, err := loader.Load("test", "asset", v1alpha2.AssetSingle, "")

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(files).To(gomega.HaveLen(1))
	})

	t.Run("SuccessNoFileInPath", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		loader := &loader{
			temporaryDir:    "/tmp",
			osRemoveAllFunc: os.RemoveAll,
			osCreateFunc:    os.Create,
			httpGetFunc:     get,
			ioutilTempDir:   ioutil.TempDir,
		}

		// When
		_, files, err := loader.Load("https://ala.ma/", "asset", v1alpha2.AssetSingle, "")

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(files).To(gomega.HaveLen(1))
	})

	t.Run("FailTemp", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		loader := &loader{
			temporaryDir:    "/tmp",
			osRemoveAllFunc: os.RemoveAll,
			osCreateFunc:    os.Create,
			httpGetFunc:     get,
			ioutilTempDir:   tempDirError,
		}

		// When
		_, _, err := loader.Load("test", "asset", v1alpha2.AssetSingle, "")

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("FailCreate", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		loader := &loader{
			temporaryDir:    "/tmp",
			osRemoveAllFunc: os.RemoveAll,
			osCreateFunc:    createError,
			httpGetFunc:     get,
			ioutilTempDir:   ioutil.TempDir,
		}

		// When
		_, _, err := loader.Load("test", "asset", v1alpha2.AssetSingle, "")

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("FailDownload", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		loader := &loader{
			temporaryDir:    "/tmp",
			osRemoveAllFunc: os.RemoveAll,
			osCreateFunc:    os.Create,
			httpGetFunc:     get,
			ioutilTempDir:   ioutil.TempDir,
		}

		// When
		_, _, err := loader.Load("error3", "asset", v1alpha2.AssetSingle, "")

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}
