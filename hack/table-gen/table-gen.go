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

const (
	REPattern                   = `(?s)<!--\s*TABLE-START\s* -->.*<!--\s*TABLE-END\s*-->`
	SkipIdentifier              = `SKIP-ELEMENT`
	RESkipPattern               = `<!--\s*` + SkipIdentifier + `\s*([^\s]+)\s*-->`
	SkipWithAncestorsIdentifier = `SKIP-WITH-ANCESTORS`
	RESkipWithAncestorsPattern  = `<!--\s*` + SkipWithAncestorsIdentifier + `\s*([^\s-]+)\s*-->`
)

var (
	CRDFilename string
	MDFilename  string
	APIVersion  string
	CRDKind     string
)

func main() {
	flag.StringVar(&CRDFilename, "crd-filename", "../../installation/resources/crds/telemetry/tracepipelines.crd.yaml", "Full or relative path to the .yaml file containing crd")
	flag.StringVar(&MDFilename, "md-filename", "../../docs/05-technical-reference/00-custom-resources/telemetry-03-tracepipeline.md", "Full or relative path to the .md file containing the file where we should insert table rows")
	flag.Parse()

	elementsToSkip := getElementsToSkip()
	doc := generateDocFromCRD(elementsToSkip)
	replaceDocInMD(doc)
}

func getElementsToSkip() map[string]bool {
	inDoc, err := os.ReadFile(MDFilename)
	if err != nil {
		panic(err)
	}

	doc := string(inDoc)
	reSkip := regexp.MustCompile(RESkipPattern)
	elementsToSkip := map[string]bool{
		"spec":   false,
		"status": false,
	}

	pairsToParamsToSkip(elementsToSkip, reSkip.FindAllStringSubmatch(doc, -1), false)

	reSkipWithAncestors := regexp.MustCompile(RESkipWithAncestorsPattern)
	pairsToParamsToSkip(elementsToSkip, reSkipWithAncestors.FindAllStringSubmatch(doc, -1), true)

	return elementsToSkip
}

func replaceDocInMD(doc string) {
	inDoc, err := os.ReadFile(MDFilename)
	if err != nil {
		panic(err)
	}

	newContent := strings.Join([]string{
		"<!-- TABLE-START -->",
		doc + "<!-- TABLE-END -->",
	}, "\n")
	re := regexp.MustCompile(REPattern)
	outDoc := re.ReplaceAll(inDoc, []byte(newContent))

	outFile, err := os.OpenFile(MDFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()
	outFile.Write(outDoc)
}

func generateDocFromCRD(elementsToSkip map[string]bool) string {
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
	kind := getElement(obj, "spec", "names", "kind")
	CRDKind = kind.(string)

	for _, version := range versions.([]interface{}) {
		name := getElement(version, "name")
		APIVersion = name.(string)

		spec := getElement(version, "schema", "openAPIV3Schema", "properties", "spec")
		mergeMaps(docElements, generateElementDoc(elementsToSkip, spec, "spec", ""))

		status := getElement(version, "schema", "openAPIV3Schema", "properties", "status")
		mergeMaps(docElements, generateElementDoc(elementsToSkip, status, "status", ""))
	}

	var doc []string
	for _, propName := range sortKeys(docElements) {
		doc = append(doc, docElements[propName])
	}

	doc = append([]string{
		"<!-- " + CRDKind + " -->",
		"| Parameter         | Description                                   |",
		"| ---------------------------------------- | ---------|",
	}, doc...)

	return strings.Join(doc, "\n")
}

func generateElementDoc(elementsToSkip map[string]bool, obj interface{}, name string, parentPath string) map[string]string {
	result := map[string]string{}
	element := obj.(map[string]interface{})
	elementType := element["type"].(string)
	description := ""
	if d := element["description"]; d != nil {
		description = d.(string)
	}

	fullName := fmt.Sprintf("%s%s", parentPath, name)
	skipWithAncestors, shouldBeSkipped := elementsToSkip[fullName]
	if shouldBeSkipped && skipWithAncestors {
		return result
	}

	if !shouldBeSkipped {
		result[fullName] = generateTableRow(fullName, description, name)
	}

	if elementType == "object" {
		mergeMaps(result, generateObjectDoc(elementsToSkip, element, name, parentPath))
	}

	if elementType == "array" {
		mergeMaps(result, generateArrayDoc(elementsToSkip, element, name, parentPath))
	}

	return result
}

func generateObjectDoc(elementsToSkip map[string]bool, element map[string]interface{}, name string, parentPath string) map[string]string {
	result := map[string]string{}
	properties := getElement(element, "properties")
	if properties == nil {
		return result
	}

	propMap := properties.(map[string]interface{})
	for _, propName := range sortKeys(propMap) {
		mergeMaps(result, generateElementDoc(elementsToSkip, propMap[propName], propName, parentPath+name+"."))
	}
	return result
}

func generateArrayDoc(elementsToSkip map[string]bool, element map[string]interface{}, name string, parentPath string) map[string]string {
	result := map[string]string{}
	properties := getElement(element, "items")
	if properties == nil {
		return result
	}

	description := ""

	if element["description"] != nil {
		description = element["description"].(string)
	}

	propMap := properties.(map[string]interface{})

	if description == "" && propMap["description"] != nil {
		description = propMap["description"].(string)
	}

	result = generateObjectDoc(elementsToSkip, propMap, name, parentPath)

	result[parentPath+name] = generateTableRow(parentPath+name, description, name)

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
