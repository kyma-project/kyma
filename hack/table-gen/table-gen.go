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
	// Regular expression pattern for reading everything between TABLE-START and TABLE-END tags
	REPattern      = `(?s)<!--\s*TABLE-START\s* -->.*<!--\s*TABLE-END\s*-->`
	SkipIdentifier = `SKIP-ELEMENT`
	// Regular expression pattern for skipping an element without its children
	RESkipPattern              = `<!--\s*` + SkipIdentifier + `\s*([^\s]+)\s*-->`
	SkipWithChildrenIdentifier = `SKIP-WITH-CHILDREN`
	// Regular expression pattern for skipping an element with its children
	RESkipWithChildrenPattern = `<!--\s*` + SkipWithChildrenIdentifier + `\s*([^\s]+)\s*-->`
)

var (
	CRDFilename string
	MDFilename  string
	APIVersion  string
	CRDKind     string
	CRDGroup    string
)

func main() {
	flag.StringVar(&CRDFilename, "crd-filename", "", "Full or relative path to the .yaml file containing crd")
	flag.StringVar(&MDFilename, "md-filename", "", "Full or relative path to the .md file containing the file where we should insert table rows")
	flag.Parse()

	if CRDFilename == "" {
		panic(fmt.Errorf("crd-filename cannot be empty. Please enter the correct filename"))
	}

	if MDFilename == "" {
		panic(fmt.Errorf("md-filename cannot be empty. Please enter the correct filename"))
	}

	elementsToSkip := getElementsToSkip()
	doc := generateDocFromCRD(elementsToSkip)
	replaceDocInMD(doc)
}

// getElementsToSkip reads MD file for SKIP tags.
// It returns a map where the key is the name of the element and the value is true if the element should be skipped with its children and it is false if the element should be skipped without its children.
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

	reSkipWithAncestors := regexp.MustCompile(RESkipWithChildrenPattern)
	pairsToParamsToSkip(elementsToSkip, reSkipWithAncestors.FindAllStringSubmatch(doc, -1), true)

	return elementsToSkip
}

// replaceDocInMD replaces the content between TABLE-START and TABLE-END tags with the newly generated content in doc.
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

// generateDocFromCRD generates table of content out of CRD.
// elementsToSkip are the elements to skip generated by getElementsToSkip function.
func generateDocFromCRD(elementsToSkip map[string]bool) string {
	input, err := os.ReadFile(CRDFilename)
	if err != nil {
		panic(err)
	}

	var obj interface{}
	if err := yaml.Unmarshal(input, &obj); err != nil {
		panic(err)
	}

	versions := getElement(obj, "spec", "versions")
	kind := getElement(obj, "spec", "names", "kind")
	group := getElement(obj, "spec", "group")
	CRDKind = kind.(string)
	CRDGroup = group.(string)

	versions = sortVersions(versions)

	table := "<div tabs name=\"CRD Specification\" group=\"crd-spec\">\n"
	open := " open"
	for _, version := range versions.([]interface{}) {
		name := getElement(version, "name")
		APIVersion = name.(string)

		table += fmt.Sprintf("<details%s>\n<summary label=\"%s\">\n%s\n</summary>\n\n", open, name.(string), name.(string))
		open = ""

		table += "**Spec:**"
		table = table + "\n" + strings.Join(generateTable(elementsToSkip, version, "spec"), "\n")
		table += "\n\n**Status:**\n"
		table = table + "\n" + strings.Join(generateTable(elementsToSkip, version, "status"), "\n")
		table += "\n\n</details>\n"
	}
	table += "</div>\n"

	return table
}

func sortVersions(versions interface{}) interface{} {
	sortedVersions := []interface{}{}
	for _, version := range versions.([]interface{}) {
		stored := getElement(version, "storage")
		if stored.(bool) {
			sortedVersions = append(sortedVersions, version)
		}
	}
	for _, version := range versions.([]interface{}) {
		stored := getElement(version, "storage")
		if !stored.(bool) {
			sortedVersions = append(sortedVersions, version)
		}
	}
	return sortedVersions

}

func generateTable(elementsToSkip map[string]bool, version interface{}, resource string) []string {
	docElements := map[string]string{}
	spec := getElement(version, "schema", "openAPIV3Schema", "properties", resource)
	mergeMaps(docElements, generateElementDoc(elementsToSkip, spec, resource, ""))

	var doc []string
	for _, propName := range sortKeys(docElements) {
		doc = append(doc, docElements[propName])
	}

	doc = append([]string{
		"<!-- " + CRDKind + " " + APIVersion + " " + CRDGroup + " -->",
		"| Parameter         | Type | Description                                   |",
		"| ---------------------------------------- | ---------|",
	}, doc...)
	return doc
}

// generateElementDoc generates table row out of some CRD element.
// It returns a map where the key is the path of an element and the value is the table row for this element.
func generateElementDoc(elementsToSkip map[string]bool, obj interface{}, name string, parentPath string) map[string]string {
	result := map[string]string{}
	element := obj.(map[string]interface{})
	elementType := element["type"].(string)
	description := ""
	if d := element["description"]; d != nil {
		description = d.(string)
	}

	fullName := fmt.Sprintf("%s%s", parentPath, name)
	skipWithChildren, shouldBeSkipped := elementsToSkip[fullName]
	if shouldBeSkipped && skipWithChildren {
		return result
	}

	if !shouldBeSkipped {
		result[fullName] = generateTableRow(fullName, elementType, description)
	}

	if elementType == "object" {
		mergeMaps(result, generateObjectDoc(elementsToSkip, element, name, parentPath))
	}

	if elementType == "array" {
		mergeMaps(result, generateArrayDoc(elementsToSkip, element, name, parentPath))
	}

	return result
}

// generateObjectDoc generates table row out of CRD object with type object.
// It returns a map where the key is the path of an element and the value is the table row for this element.
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

// generateArrayDoc generates table row out of CRD object with type array.
// It returns a map where the key is the path of an element and the value is the table row for this element.
func generateArrayDoc(elementsToSkip map[string]bool, element map[string]interface{}, name string, parentPath string) map[string]string {
	result := map[string]string{}

	skipWithChildren, shouldBeSkipped := elementsToSkip[parentPath+name]
	if skipWithChildren {
		return result
	}

	items := getElement(element, "items")
	if items == nil {
		return result
	}

	description := ""

	// in case array has a description before items element, it would take this description
	if element["description"] != nil {
		description = element["description"].(string)
	}

	itemsMap := items.(map[string]interface{})

	// in case array has a description within items element, and there is no description above, it would take this description
	if description == "" && itemsMap["description"] != nil {
		description = itemsMap["description"].(string)
	}

	result = generateObjectDoc(elementsToSkip, itemsMap, name, parentPath)

	if !shouldBeSkipped {
		result[parentPath+name] = generateTableRow(parentPath+name, description, name)
	}

	return result
}

// generateTableRow generates a row of the resulting table which we include into our MD file.
func generateTableRow(fullName, fieldType, description string) string {
	return fmt.Sprintf("| **%s** | %s | %s |",
		fullName, fieldType, description)
}

// getElement returns a specific element from obj based on the provided path.
func getElement(obj interface{}, path ...string) interface{} {
	elem := obj
	for _, p := range path {
		elem = elem.(map[string]interface{})[p]
	}
	return elem
}

func pairsToParamsToSkip(toSkip map[string]bool, pairs [][]string, isToSkipChildren bool) {
	for _, pair := range pairs {
		paramName := pair[1]
		toSkip[paramName] = isToSkipChildren
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
