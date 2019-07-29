package addon

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
	addonChartDirName = "chart"
	addonMetaName     = "meta.yaml"
	addonDocsMetaPath = "docs/meta.yaml"
	addonPlanDirName  = "plans"

	addonPlanMetaName             = "meta.yaml"
	addonPlaSchemaCreateJSONName  = "create-instance-schema.json"
	addonPlanSchemaBindJSONName   = "bind-instance-schema.json"
	addonPlanSchemaUpdateJSONName = "update-instance-schema.json"
	addonPlanValuesFileName       = "values.yaml"
	addonPlanBindTemplateFileName = "bind.yaml"

	maxSchemaLength = 65536 // 64 k
)

// Loader provides loading of addons from repository and representing them as addons and charts.
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
		log:          log.WithField("service", "addon:loader"),
	}
}

// Load takes stream with compressed tgz archive as io.Reader, tries to unpack it to tmp directory,
// and then loads it as addon and Helm chart
func (l *Loader) Load(in io.Reader) (*internal.Addon, []*chart.Chart, error) {
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
func (l Loader) LoadDir(path string) (*internal.Addon, []*chart.Chart, error) {
	return l.loadDir(path)
}

func (l Loader) loadDir(path string) (*internal.Addon, []*chart.Chart, error) {
	c, err := l.loadChartFromDir(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while loading chart")
	}

	form, err := l.createFormFromAddonDir(path)
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
	// In current version we have only one chart per addon
	// in future version we will have some loop over each plan to load all charts
	chartPath, err := l.discoverPathToHelmChart(baseDir)
	if err != nil {
		return nil, errors.Wrapf(err, "while discovering the name of the Helm Chart under the %q addon directory", addonChartDirName)
	}

	c, err := l.loadChart(chartPath)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return nil, errors.New("addon does not contains \"chart\" directory")
	default:
		return nil, errors.Wrap(err, "while loading chart")
	}

	return c, nil
}

// discoverPathToHelmChart returns the full path to the Helm Chart directory from `addonChartDirName` folder
//
// - if more that one directory is found then error is returned
// - if additional files are found under the `addonChartDirName` directory then
//   they are ignored but logged as warning to improve transparency.
func (l Loader) discoverPathToHelmChart(baseDir string) (string, error) {
	cDir := filepath.Join(baseDir, addonChartDirName)
	rawFiles, err := ioutil.ReadDir(cDir)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return "", errors.Errorf("addon does not contains %q directory", addonChartDirName)
	default:
		return "", errors.Wrapf(err, "while reading directory %s", cDir)
	}

	directories, files := splitForDirectoriesAndFiles(rawFiles)
	if len(directories) == 0 {
		return "", fmt.Errorf("%q directory SHOULD contain one Helm Chart folder but it's empty", addonChartDirName)
	}

	if len(directories) > 1 {
		return "", fmt.Errorf("%q directory MUST contain only one Helm Chart folder but found multiple directories: [%s]", addonChartDirName, strings.Join(directories, ", "))
	}

	if len(files) != 0 { // ignoring by design
		l.log.Warningf("Found files: [%s] in %q addon directory. All are ignored.", strings.Join(files, ", "), addonChartDirName)
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

func (l Loader) createFormFromAddonDir(baseDir string) (*form, error) {
	f := &form{Plans: make(map[string]*formPlan)}

	addonMetaFile, err := ioutil.ReadFile(filepath.Join(baseDir, addonMetaName))
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return nil, fmt.Errorf("missing metadata information about addon, please check if addon contains %q file", addonMetaName)
	default:
		return nil, errors.Wrapf(err, "while reading %q file", addonMetaName)
	}

	if err := yaml.Unmarshal(addonMetaFile, &f.Meta); err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling addon %q file", addonMetaName)
	}

	addonDocsFile, err := ioutil.ReadFile(filepath.Join(baseDir, addonDocsMetaPath))
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "while reading %q file", addonDocsMetaPath)
	}

	if err := yaml.Unmarshal(addonDocsFile, &f.DocsMeta); err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling addon %q file", addonDocsMetaPath)
	}

	plansPath := filepath.Join(baseDir, addonPlanDirName)
	files, err := ioutil.ReadDir(plansPath)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return nil, fmt.Errorf("addon does not contains any plans, please check if addon contains %q directory", addonPlanDirName)
	default:
		return nil, errors.Wrapf(err, "while reading %q file", addonMetaName)
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

	if err := yamlUnmarshal(topdir, addonPlanMetaName, &plan.Meta, true); err != nil {
		return unmarshalPlanErr(err, addonPlanMetaName)
	}

	if err := yamlUnmarshal(topdir, addonPlanValuesFileName, &plan.Values, false); err != nil {
		return unmarshalPlanErr(err, addonPlanValuesFileName)
	}

	if plan.SchemasCreate, err = loadPlanSchema(topdir, addonPlaSchemaCreateJSONName, false); err != nil {
		return unmarshalPlanErr(err, addonPlaSchemaCreateJSONName)
	}

	if plan.SchemasBind, err = loadPlanSchema(topdir, addonPlanSchemaBindJSONName, false); err != nil {
		return unmarshalPlanErr(err, addonPlanSchemaBindJSONName)
	}

	if plan.SchemasUpdate, err = loadPlanSchema(topdir, addonPlanSchemaUpdateJSONName, false); err != nil {
		return unmarshalPlanErr(err, addonPlanSchemaUpdateJSONName)
	}

	if plan.BindTemplate, err = loadRaw(topdir, addonPlanBindTemplateFileName, false); err != nil {
		return errors.Wrapf(err, "while loading plan %q file", addonPlanBindTemplateFileName)
	}

	return nil
}

// unpackArchive unpack from a reader containing a compressed tar archive to tmpdir.
func (l Loader) unpackArchive(baseDir string, in io.Reader) (string, error) {
	dir, err := l.createTmpDir(baseDir, "unpacked-addon")
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
