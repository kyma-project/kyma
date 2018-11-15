package bundle

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

const (
	bundleChartDirName = "chart"
	bundleMetaName     = "meta.yaml"
	bundlePlanDirName  = "plans"

	bundlePlanMetaName             = "meta.yaml"
	bundlePlaSchemaCreateJSONName  = "create-instance-schema.json"
	bundlePlanSchemaUpdateJSONName = "update-instance-schema.json"
	bundlePlanValuesFileName       = "values.yaml"
	bundlePlanBindTemplateFileName = "bind.yaml"

	maxSchemaLength = 65536 // 64 k
)

// Loader provides loading of bundles from repository and representing them as bundles and charts.
type Loader struct {
	tmpDir       string
	loadChart    func(name string) (*chart.Chart, error)
	createTmpDir func(dir, prefix string) (name string, err error)
	log          logrus.FieldLogger
}

// NewLoader returns new instance of Loader.
func NewLoader(tmpDir string, log logrus.FieldLogger) *Loader {
	return &Loader{
		tmpDir:       tmpDir,
		loadChart:    chartutil.Load,
		createTmpDir: ioutil.TempDir,
		log:          log.WithField("service", "bundle:loader"),
	}
}

// Load takes stream with compressed tgz archive as io.Reader, tries to unpack it to tmp directory,
// and then loads it as bundle and Helm chart
func (l *Loader) Load(in io.Reader) (*internal.Bundle, []*chart.Chart, error) {
	unpackedDir, err := l.unpackArchive(l.tmpDir, in)
	if err != nil {
		return nil, nil, err
	}
	cleanPath := filepath.Clean(unpackedDir)
	if strings.HasPrefix(cleanPath, l.tmpDir) {
		defer os.RemoveAll(unpackedDir)
	} else {
		defer l.log.Warnf("directory %s was not deleted because its name does not resolve to expected path", unpackedDir)
	}

	return l.loadDir(unpackedDir)
}

// LoadDir takes uncompressed chart in specified directory and loads it.
func (l Loader) LoadDir(path string) (*internal.Bundle, []*chart.Chart, error) {
	return l.loadDir(path)
}

func (l Loader) loadDir(path string) (*internal.Bundle, []*chart.Chart, error) {
	c, err := l.loadChartFromDir(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while loading chart")
	}

	form, err := l.createFormFromBundleDir(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while mapping buffered files to form")
	}

	if err := form.Validate(); err != nil {
		return nil, nil, errors.Wrap(err, "while validating form")
	}

	yb, err := form.ToModel(c)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while mapping form to model")
	}

	return &yb, []*chart.Chart{c}, nil
}

func (l Loader) loadChartFromDir(baseDir string) (*chart.Chart, error) {
	// In current version we have only one chart per bundle
	// in future version we will have some loop over each plan to load all charts
	chartPath, err := l.discoverPathToHelmChart(baseDir)
	if err != nil {
		return nil, errors.Wrapf(err, "while discovering the name of the Helm Chart under the %q bundle directory", bundleChartDirName)
	}

	c, err := l.loadChart(chartPath)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return nil, errors.New("bundle does not contains \"chart\" directory")
	default:
		return nil, errors.Wrap(err, "while loading chart")
	}

	return c, nil
}

// discoverPathToHelmChart returns the full path to the Helm Chart directory from `bundleChartDirName` folder
//
// - if more that one directory is found then error is returned
// - if additional files are found under the `bundleChartDirName` directory then
//   they are ignored but logged as warning to improve transparency.
func (l Loader) discoverPathToHelmChart(baseDir string) (string, error) {
	cDir := filepath.Join(baseDir, bundleChartDirName)
	rawFiles, err := ioutil.ReadDir(cDir)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return "", errors.Errorf("bundle does not contains %q directory", bundleChartDirName)
	default:
		return "", errors.Wrapf(err, "while reading directory %s", cDir)
	}

	directories, files := splitForDirectoriesAndFiles(rawFiles)
	if len(directories) == 0 {
		return "", fmt.Errorf("%q directory SHOULD contain one Helm Chart folder but it's empty", bundleChartDirName)
	}

	if len(directories) > 1 {
		return "", fmt.Errorf("%q directory MUST contain only one Helm Chart folder but found multiple directories: [%s]", bundleChartDirName, strings.Join(directories, ", "))
	}

	if len(files) != 0 { // ignoring by design
		l.log.Warningf("Found files: [%s] in %q bundle directory. All are ignored.", strings.Join(files, ", "), bundleChartDirName)
	}

	chartFullPath := filepath.Join(cDir, directories[0])
	return chartFullPath, nil
}

func splitForDirectoriesAndFiles(rawFiles []os.FileInfo) (dirs []string, files []string) {
	for _, f := range rawFiles {
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		} else {
			files = append(files, f.Name())
		}
	}

	return dirs, files
}

func (l Loader) createFormFromBundleDir(baseDir string) (*form, error) {
	f := &form{Plans: make(map[string]*formPlan)}

	bundleMetaFile, err := ioutil.ReadFile(filepath.Join(baseDir, bundleMetaName))
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return nil, fmt.Errorf("missing metadata information about bundle, please check if bundle contains %q file", bundleMetaName)
	default:
		return nil, errors.Wrapf(err, "while reading %q file", bundleMetaName)
	}

	if err := yaml.Unmarshal(bundleMetaFile, &f.Meta); err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling bundle %q file", bundleMetaName)
	}

	plansPath := filepath.Join(baseDir, bundlePlanDirName)
	files, err := ioutil.ReadDir(plansPath)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return nil, fmt.Errorf("bundle does not contains any plans, please check if bundle contains %q directory", bundlePlanDirName)
	default:
		return nil, errors.Wrapf(err, "while reading %q file", bundleMetaName)
	}

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			planName := fileInfo.Name()
			f.Plans[planName] = &formPlan{}

			if err := l.loadPlanDefinition(filepath.Join(plansPath, planName), f.Plans[planName]); err != nil {
				return nil, err
			}
		}
	}

	return f, nil
}

func (Loader) loadPlanDefinition(path string, plan *formPlan) error {
	topdir, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	unmarshalPlanErr := func(err error, filename string) error {
		return errors.Wrapf(err, "while unmarshaling plan %q file", filename)
	}

	if err := yamlUnmarshal(topdir, bundlePlanMetaName, &plan.Meta, true); err != nil {
		return unmarshalPlanErr(err, bundlePlanMetaName)
	}

	if err := yamlUnmarshal(topdir, bundlePlanValuesFileName, &plan.Values, false); err != nil {
		return unmarshalPlanErr(err, bundlePlanValuesFileName)
	}

	if plan.SchemasCreate, err = loadPlanSchema(topdir, bundlePlaSchemaCreateJSONName, false); err != nil {
		return unmarshalPlanErr(err, bundlePlaSchemaCreateJSONName)
	}

	if plan.SchemasUpdate, err = loadPlanSchema(topdir, bundlePlanSchemaUpdateJSONName, false); err != nil {
		return unmarshalPlanErr(err, bundlePlanSchemaUpdateJSONName)
	}

	if plan.BindTemplate, err = loadRaw(topdir, bundlePlanBindTemplateFileName, false); err != nil {
		return errors.Wrapf(err, "while loading plan %q file", bundlePlanBindTemplateFileName)
	}

	return nil
}

// unpackArchive unpack from a reader containing a compressed tar archive to tmpdir.
func (l Loader) unpackArchive(baseDir string, in io.Reader) (string, error) {
	dir, err := l.createTmpDir(baseDir, "unpacked-bundle")
	if err != nil {
		return "", err
	}

	unzipped, err := gzip.NewReader(in)
	if err != nil {
		return "", err
	}
	defer unzipped.Close()

	tr := tar.NewReader(unzipped)

Unpack:
	for {
		header, err := tr.Next()
		switch err {
		case nil:
		case io.EOF:
			break Unpack
		default:
			return "", err
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dir, header.Name)

		// check the file type
		switch header.Typeflag {
		// its a dir and if it doesn't exist - create it
		case tar.TypeDir:
			if _, err := os.Stat(target); os.IsNotExist(err) {
				if err := os.MkdirAll(target, 0755); err != nil {
					return "", err
				}
			}
			// it's a file - create it
		case tar.TypeReg:
			if err := l.createFile(target, header.Mode, tr); err != nil {
				return "", err
			}
		}
	}

	return dir, nil
}

func (Loader) createFile(target string, mode int64, r io.Reader) error {
	f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(mode))
	if err != nil {
		return err
	}
	defer f.Close()

	// copy over contents
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	return nil
}

func yamlUnmarshal(basePath, fileName string, unmarshalTo interface{}, required bool) error {
	b, err := ioutil.ReadFile(filepath.Join(basePath, fileName))
	switch {
	case err == nil:
	case os.IsNotExist(err) && !required:
		return nil
	case os.IsNotExist(err) && required:
		return fmt.Errorf("%q is required but is not present", fileName)
	default:
		return err
	}

	return yaml.Unmarshal(b, unmarshalTo)
}

func loadPlanSchema(basePath, fileName string, required bool) (*internal.PlanSchema, error) {
	b, err := ioutil.ReadFile(filepath.Join(basePath, fileName))
	switch {
	case err == nil:
	case os.IsNotExist(err) && !required:
		return nil, nil
	case os.IsNotExist(err) && required:
		return nil, fmt.Errorf("%q is required but is not present", fileName)
	default:
		return nil, errors.Wrap(err, "while loading plan schema")
	}

	// OSB API defines: Schemas MUST NOT be larger than 64kB.
	// See: https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#schema-object
	if len(b) >= maxSchemaLength {
		return nil, fmt.Errorf("schema %s is larger than 64 kB", fileName)
	}

	var schema internal.PlanSchema
	err = json.Unmarshal(b, &schema)
	return &schema, errors.Wrap(err, "while loading plan shcema")
}

func loadRaw(basePath, fileName string, required bool) ([]byte, error) {
	b, err := ioutil.ReadFile(filepath.Join(basePath, fileName))
	switch {
	case err == nil:
	case os.IsNotExist(err) && !required:
		return nil, nil
	case os.IsNotExist(err) && required:
		return nil, fmt.Errorf("%q is required but is not present", fileName)
	default:
		return nil, err
	}

	return b, nil
}
