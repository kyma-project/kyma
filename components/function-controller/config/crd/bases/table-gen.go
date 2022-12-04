package main

import (
	"fmt"
	"os"
	"regexp"
	"sigs.k8s.io/yaml"
	"strings"
)

type IndentType int

const (
	IndentSpace IndentType = iota
	IndentNbsp
	IndentParentPath
)
const indentType = IndentSpace

const InputCRDFilename = `/Users/I567085/src/2022jul/kyma/components/function-controller/config/crd/bases/serverless.kyma-project.io_functions.yaml`
const APIVersion = "v1alpha2"

const OutputMDFilename = `/Users/I567085/src/2022jul/kyma/docs/05-technical-reference/00-custom-resources/svls-01-function.md`
const REPatternToReplace = `<!--\s*FUNCTION-CRD-PARAMETERS-TABLE\(START\).*<<!--\s*FUNCTION-CRD-PARAMETERS-TABLE\(START\)[^\n]*`

func main() {
	newDoc := generateDocFromCRD()
	replaceDocInMD(newDoc)
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

	outDoc := re.ReplaceAll(inDoc, []byte(doc))

	outFile, err := os.OpenFile(OutputMDFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()
	outFile.Write(outDoc)
}

func generateDocFromCRD() string {
	input, err := os.ReadFile(InputCRDFilename)
	if err != nil {
		panic(err)
	}

	var obj interface{}
	if err := yaml.Unmarshal(input, &obj); err != nil {
		panic(err)
	}

	var doc []string

	versions := getElement(obj, "spec", "versions")
	for _, version := range versions.([]interface{}) {
		name := getElement(version, "name")
		if name.(string) != APIVersion {
			continue
		}
		functionSpec := getElement(version, "schema", "openAPIV3Schema", "properties", "spec")
		doc = append(doc, generateElementDoc(functionSpec, "spec", true, 0, "")...)
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

func generateElementDoc(obj interface{}, name string, required bool, indent int, parentPath string) []string {
	var result []string
	element := obj.(map[string]interface{})
	elementType := element["type"].(string)
	description := ""
	if d := element["description"]; d != nil {
		description = d.(string)
	}
	result = append(result,
		fmt.Sprintf("| %s%s | %s | %s |",
			getIndent(indent, parentPath), name,
			yesNo(required), description))

	if elementType == "object" {
		result = append(result,
			generateObjectDoc(element, name, indent, parentPath)...)
	}
	return result
}

func generateObjectDoc(element map[string]interface{}, name string, indent int, parentPath string) []string {
	var result []string
	properties := getElement(element, "properties")
	if properties == nil {
		return result
	}
	var requiredChildren []interface{}
	if rc := getElement(element, "required"); rc != nil {
		requiredChildren = rc.([]interface{})
	}
	for propName, propVal := range properties.(map[string]interface{}) {
		propRequired := contains(requiredChildren, name)
		result = append(result,
			generateElementDoc(propVal, propName, propRequired,
				indent+1, parentPath+name+".")...)
	}
	return result
}

func getIndent(indent int, path string) string {
	switch indentType {
	case IndentSpace:
		return strings.Repeat(" ", indent*2)
	case IndentNbsp:
		return strings.Repeat("&nbsp", indent*2)
	case IndentParentPath:
		return path
	default:
		panic("Unexpected indent type!")
	}
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
