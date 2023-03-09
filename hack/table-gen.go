package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"
)

type FunctionSpecGenerator struct {
	elementsToSkip map[string]bool
}

// how to deal with comment tag? +
// what to do with required label? // there is a required label, detect it and extract info // if existing tool does not work, keep minimal, no need for required
// which kind of elements should we include by default? +
// how to start this script automatically? move to hack, do the prow job
// add auto crd name comment insertion

// try to remove keep-this tag if it is easy +

const (
	FunctionSpecIdentifier      = `FUNCTION-SPEC`
	REFunctionSpecPattern       = `(?s)<!--\s*` + FunctionSpecIdentifier + `-START\s* -->.*<!--\s*` + FunctionSpecIdentifier + `-END\s*-->`
	SkipIdentifier              = `SKIP-ELEMENT`
	RESkipPattern               = `<!--\s*` + SkipIdentifier + `\s*([^\s]+)\s*-->`
	SkipWithAncestorsIdentifier = `SKIP-WITH-ANCESTORS`
	RESkipWithAncestorsPattern  = `<!--\s*` + SkipWithAncestorsIdentifier + `\s*([^\s-]+)\s*-->`
)

var (
	CRDFilename string
	MDFilename  string
	APIVersion  string
	CRDTitle    string
)

func main() {
	//remove default value
	flag.StringVar(&CRDFilename, "crd-filename", "/Users/I572465/telemetry-manager/config/crd/bases/telemetry.kyma-project.io_logpipelines.yaml", "Full or relative path to the .yaml file containing crd")
	flag.StringVar(&MDFilename, "md-filename", "/Users/I572465/Go/src/github.com/kyma-project/kyma/docs/01-overview/main-areas/telemetry/telemetry-02-logs.md", "Full or relative path to the .md file containing the file where we should insert table rows")
	flag.StringVar(&APIVersion, "api-version", "v1alpha1", "API version your operattor uses")
	flag.StringVar(&CRDTitle, "crd-title", "", "The name of the CRD which was passed in crd-filename")
	flag.Parse()

	toSkip := getElementsToSkip()
	generator := CreateFunctionSpecGenerator(toSkip)
	doc := generator.generateDocFromCRD()
	replaceDocInMD(doc)
	print(doc)
}

func getElementsToSkip() map[string]bool {
	inDoc, err := os.ReadFile(MDFilename)
	if err != nil {
		panic(err)
	}

	doc := string(inDoc)
	reSkip := regexp.MustCompile(RESkipPattern)
	toSkip := map[string]bool{}

	pairsToParamsToSkip(toSkip, reSkip.FindAllStringSubmatch(doc, -1), false)

	reSkipWithAncestors := regexp.MustCompile(RESkipWithAncestorsPattern)
	pairsToParamsToSkip(toSkip, reSkipWithAncestors.FindAllStringSubmatch(doc, -1), true)

	return toSkip
}

func replaceDocInMD(doc string) {
	inDoc, err := os.ReadFile(MDFilename)
	if err != nil {
		panic(err)
	}

	newContent := strings.Join([]string{
		"<!-- " + FunctionSpecIdentifier + "-START -->",
		doc + "<!-- " + FunctionSpecIdentifier + "-END -->",
	}, "\n")
	re := regexp.MustCompile(REFunctionSpecPattern)
	outDoc := re.ReplaceAll(inDoc, []byte(newContent))

	outFile, err := os.OpenFile(MDFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()
	outFile.Write(outDoc)
}

func CreateFunctionSpecGenerator(toSkip map[string]bool) FunctionSpecGenerator {
	return FunctionSpecGenerator{
		elementsToSkip: toSkip,
	}
}

func (generator *FunctionSpecGenerator) generateDocFromCRD() string {
	input, err := os.ReadFile(CRDFilename)
	if err != nil {
		panic(err)
	}

	var obj interface{}
	if err := yaml.Unmarshal(input, &obj); err != nil {
		panic(err)
	}

	docElements := map[string]string{}
	versions := getElement(obj, "spec", "versions")

	for _, version := range versions.([]interface{}) {
		name := getElement(version, "name")
		if name.(string) != APIVersion {
			continue
		}

		functionSpec := getElement(version, "schema", "openAPIV3Schema", "properties", "spec")
		mergeMaps(docElements, generator.generateElementDoc(functionSpec, "spec", ""))

		functionStatus := getElement(version, "schema", "openAPIV3Schema", "properties", "status")
		mergeMaps(docElements, generator.generateElementDoc(functionStatus, "status", ""))
	}

	var doc []string
	for _, propName := range sortKeys(docElements) {
		doc = append(doc, docElements[propName])
	}

	doc = append([]string{
		"<!-- " + CRDTitle + " -->",
		"| Parameter         | Description                                   |",
		"| ---------------------------------------- | ---------|",
	}, doc...)

	return strings.Join(doc, "\n")
}

func (generator *FunctionSpecGenerator) generateElementDoc(obj interface{}, name string, parentPath string) map[string]string {
	result := map[string]string{}
	element := obj.(map[string]interface{})
	elementType := element["type"].(string)
	description := ""
	if d := element["description"]; d != nil {
		description = d.(string)
	}

	fullName := fmt.Sprintf("%s%s", parentPath, name)
	skipWithAncestors, shouldBeSkipped := generator.elementsToSkip[fullName]
	if shouldBeSkipped && skipWithAncestors {
		return result
	}

	if !shouldBeSkipped {
		result[fullName] = generateTableRow(fullName, description, name)
	}

	if elementType == "object" {
		mergeMaps(result, generator.generateObjectDoc(element, name, parentPath))
	}
	return result
}

func (generator *FunctionSpecGenerator) generateObjectDoc(element map[string]interface{}, name string, parentPath string) map[string]string {
	result := map[string]string{}
	properties := getElement(element, "properties")
	if properties == nil {
		return result
	}

	propMap := properties.(map[string]interface{})
	for _, propName := range sortKeys(propMap) {
		mergeMaps(result, generator.generateElementDoc(propMap[propName], propName, parentPath+name+"."))
	}
	return result
}

func generateTableRow(fullName string, description string, name string) string {
	return fmt.Sprintf("| **%s** | %s |",
		fullName, normalizeDescription(description, name))
}

func getElement(obj interface{}, path ...string) interface{} {
	elem := obj
	for _, p := range path {
		elem = elem.(map[string]interface{})[p]
	}
	return elem
}

func normalizeDescription(description string, name string) any {
	description_trimmed := strings.Trim(description, " ")
	name_trimmed := strings.Trim(name, " ")
	if len(name_trimmed) == 0 {
		return description_trimmed
	}
	dParts := strings.SplitN(description_trimmed, " ", 2)
	if len(dParts) < 2 {
		return description
	}
	if !strings.EqualFold(name_trimmed, dParts[0]) {
		return description
	}
	description_trimmed = strings.Trim(dParts[1], " ")
	description_trimmed = strings.ToUpper(description_trimmed[:1]) + description_trimmed[1:]
	return description_trimmed
}

func pairsToParamsToSkip(toSkip map[string]bool, pairs [][]string, isToSkip bool) {
	for _, pair := range pairs {
		paramName := pair[1]
		toSkip[paramName] = isToSkip
	}
}

func sortKeys[T any](propMap map[string]T) []string {
	var keys []string
	for key := range propMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func mergeMaps(dest map[string]string, src map[string]string) {
	for k, v := range src {
		dest[k] = v
	}
}
