package kymasources

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSrcDirFor(t *testing.T) {
	Convey("SrcDirFor", t, func() {
		Convey("When component Source.URL is not defined", func() {
			Convey("returns default (bundled) sources", func() {

				//given
				mockKymaPackages := mockKymaPackagesForBundled{}
				fsWrapper := fsWrapperMockedForExistingSources{}
				rootDir := "Not used in this test case"
				sgCreator := NewSourceGetterCreator(mockKymaPackages, fsWrapper, rootDir)
				sg := sgCreator.NewGetterFor(LegacyKymaSourceConfig{"KymaURL not used", "KymaVersion not used"})
				component := v1alpha1.KymaComponent{
					Name: "testComponent",
				}

				//when
				componentSrcDir, err := sg.SrcDirFor(component)

				//then
				So(err, ShouldBeNil)
				So(componentSrcDir, ShouldEqual, "/kyma/injected/resources/testComponent")
			})
		})

		Convey("When component Source.URL is defined", func() {
			Convey("and local copy does not exist", func() {
				Convey("fetches the sources to a local directory and returns it's path", func() {
					//given

					cwd, err := os.Getwd()
					So(err, ShouldBeNil)
					//Sources of the test component
					sourceDir := path.Join(cwd, "sampleSources/components-remote-src/testComponent")

					//Create temp directory for the target location
					targetDirRoot, err := ioutil.TempDir("", "prefix")
					So(err, ShouldBeNil)
					defer os.RemoveAll(targetDirRoot)

					rootDir := targetDirRoot
					fsWrapper := NewFilesystemWrapper()
					sgCreator := NewSourceGetterCreator(nil, fsWrapper, rootDir)
					sg := sgCreator.NewGetterFor(LegacyKymaSourceConfig{"KymaURL not used", "KymaVersion not used"})

					component := v1alpha1.KymaComponent{
						Name: "testComponent",
						Source: &v1alpha1.ComponentSource{
							URL: "file://" + sourceDir,
							//URL: "https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/dex",
						},
					}

					//when
					componentSrcDir, err := sg.SrcDirFor(component)

					//then
					So(err, ShouldBeNil)
					So(fsWrapper.Exists(path.Join(componentSrcDir, "Chart.yaml")), ShouldBeTrue)
				})
			})

			Convey("and local copy does exist", func() {
				Convey("returns local copy path", func() {
					//given
					cwd, err := os.Getwd()
					So(err, ShouldBeNil)
					//Existing sources of the test component
					rootDir := path.Join(cwd, "sampleSources")

					fsWrapper := NewFilesystemWrapper()
					sgCreator := NewSourceGetterCreator(nil, fsWrapper, rootDir)
					sg := sgCreator.NewGetterFor(LegacyKymaSourceConfig{"KymaURL not used", "KymaVersion not used"})

					component := v1alpha1.KymaComponent{
						Name: "testComponent",
						Source: &v1alpha1.ComponentSource{
							URL: "since sources exists, url is not used in this case",
						},
					}

					//when
					componentSrcDir, err := sg.SrcDirFor(component)

					//then
					So(err, ShouldBeNil)
					expectedChartFile := path.Join(componentSrcDir, "Chart.yaml")
					So(fsWrapper.Exists(expectedChartFile), ShouldBeTrue)
				})
			})
		})
	})
}
