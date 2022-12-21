package main

import (
	"fmt"
	"os"
	"regexp"
	"sigs.k8s.io/yaml"
	"sort"
	"strings"
	"time"
)

type IndentType int

const (
	IndentSpace IndentType = iota
	IndentNbsp
	IndentParentPath
)
const indentType = IndentParentPath

// TODO: use relative path or from param
const InputCRDFilename = `/Users/I567085/src/2022jul/kyma/components/function-controller/config/crd/bases/serverless.kyma-project.io_functions.yaml`
const APIVersion = "v1alpha2"

// TODO: use relative path or from param
const OutputMDFilename = `/Users/I567085/src/2022jul/kyma/docs/05-technical-reference/00-custom-resources/svls-01-function.md`

// TODO: rename - this script create only SPEC
const ReplacePartIdentifier = `FUNCTION-CRD-PARAMETERS-TABLE`
const REPatternToReplace = `(?s)<!--\s*` + ReplacePartIdentifier + `\(START\).*<!--\s*` + ReplacePartIdentifier + `\(END\)[^\n]*`

const KeepIdentifier = `KEEP-THIS`
const REPatternKeep = `<!--\s*` + KeepIdentifier + `\s*-->\s*[|]\s*\*{2}([^|]*)\*{2}.*`

func main() {
	elementsToKeep := getElementsToKeep()
	newDoc := generateDocFromCRD(elementsToKeep)
	replaceDocInMD(newDoc)
}

func getElementsToKeep() map[string]string {
	inDoc, err := os.ReadFile(OutputMDFilename)
	if err != nil {
		panic(err)
	}

	re, err := regexp.Compile(REPatternToReplace)
	if err != nil {
		panic(err)
	}

	tablePart := re.FindString(string(inDoc))

	reKeep, err := regexp.Compile(REPatternKeep)
	if err != nil {
		panic(err)
	}

	rows := reKeep.FindAllStringSubmatch(tablePart, -1)

	toKeep := map[string]string{}
	for _, pair := range rows {
		paramName := pair[1]
		rowContent := pair[0]
		toKeep[paramName] = rowContent
	}
	return toKeep
}

func replaceDocInMD(doc string) {
	inDoc, err := os.ReadFile(OutputMDFilename)
	if err != nil {
		panic(err)
	}

	re, err := regexp.Compile(REPatternToReplace)
	if err != nil {
		panic(err)
	}

	newContent := strings.Join([]string{
		"<!-- " + ReplacePartIdentifier + "(START) -->",
		//TODO: for test only - remove row with date (unnecessary changes after regeneration without changes) or detect changes
		"<!-- generated: " + time.Now().String() + " -->",
		doc,
		"<!-- " + ReplacePartIdentifier + "(END) -->",
	}, "\n")

	outDoc := re.ReplaceAll(inDoc, []byte(newContent))

	outFile, err := os.OpenFile(OutputMDFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()
	outFile.Write(outDoc)
}

func generateDocFromCRD(elementsToKeep map[string]string) string {
	input, err := os.ReadFile(InputCRDFilename)
	if err != nil {
		panic(err)
	}

	//TODO: check why unmarshalling to CustomResource don't work
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
		for k, v := range generateElementDoc(functionSpec, "spec", true, 0, "", elementsToKeep) {
			docElements[k] = v
		}
	}

	for k, v := range elementsToKeep {
		docElements[k] = v
	}

	var doc []string
	for _, propName := range sortedKeys(docElements) {
		doc = append(doc, docElements[propName])
	}

	return strings.Join(doc, "\n")
}

func getElement(obj interface{}, path ...string) interface{} {
	elem := obj
	for _, p := range path {
		elem = elem.(map[string]interface{})[p]
	}
	return elem
}

func generateElementDoc(obj interface{}, name string, required bool, indent int, parentPath string, elementsToKeep map[string]string) map[string]string {
	result := map[string]string{}
	element := obj.(map[string]interface{})
	elementType := element["type"].(string)
	description := ""
	if d := element["description"]; d != nil {
		description = d.(string)
	}
	fullName := fmt.Sprintf("%s%s", parentPath, name)
	_, isRowToKeep := elementsToKeep[fullName]
	var generatedRow string
	if !isRowToKeep {
		//	generatedRow = rowToKeep
		//} else {
		generatedRow =
			fmt.Sprintf("| **%s** | %s | %s |",
				fullName, yesNo(required), normalizeDescription(description, name))
	}
	result[fullName] = generatedRow

	if elementType == "object" {
		for k, v := range generateObjectDoc(element, name, indent, parentPath, elementsToKeep) {
			result[k] = v
		}
	}
	return result
}

func normalizeDescription(description string, name string) any {
	d := strings.Trim(description, " ")
	n := strings.Trim(name, " ")
	if len(n) == 0 {
		return d
	}
	dParts := strings.SplitN(d, " ", 2)
	if len(dParts) < 2 {
		return description
	}
	if !strings.EqualFold(n, dParts[0]) {
		return description
	}
	d = strings.Trim(dParts[1], " ")
	d = strings.ToUpper(d[:1]) + d[1:]
	return d
}

func generateObjectDoc(element map[string]interface{}, name string, indent int, parentPath string, elementsToKeep map[string]string) map[string]string {
	result := map[string]string{}
	properties := getElement(element, "properties")
	if properties == nil {
		return result
	}
	var requiredChildren []interface{}
	if rc := getElement(element, "required"); rc != nil {
		requiredChildren = rc.([]interface{})
	}
	propMap := properties.(map[string]interface{})
	for _, propName := range sortedKeys(propMap) {
		propRequired := contains(requiredChildren, name)
		for k, v := range generateElementDoc(propMap[propName], propName, propRequired,
			indent+1, parentPath+name+".", elementsToKeep) {
			result[k] = v
		}
	}
	return result
}

func sortedKeys[T any](propMap map[string]T) []string {
	var keys []string
	for key := range propMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func contains(s []interface{}, e string) bool {
	for _, a := range s {
		if a.(string) == e {
			return true
		}
	}
	return false
}
