package main

import (
	"fmt"
	"os"
	"regexp"
	"sigs.k8s.io/yaml"
	"sort"
	"strings"
)

// TODO: consider move to param?
const APIVersion = "v1alpha2"

// TODO: use relative path or from param
const CRDFilename = `/Users/I567085/src/2022jul/kyma/components/function-controller/config/crd/bases/serverless.kyma-project.io_functions.yaml`

// TODO: use relative path or from param
const MDFilename = `/Users/I567085/src/2022jul/kyma/docs/05-technical-reference/00-custom-resources/svls-01-function.md`

const FunctionSpecIdentifier = `FUNCTION-SPEC`
const REFunctionSpecPattern = `(?s)<!--\s*` + FunctionSpecIdentifier + `-START\s* -->.*<!--\s*` + FunctionSpecIdentifier + `-END\s*-->`

const KeepThisIdentifier = `KEEP-THIS`
const REKeepThisPattern = `\s*[|]\s*\*{2}([^*]*)\*{2}.*<!--\s*` + KeepThisIdentifier + `\s*-->`

type FunctionSpecGenerator struct {
	elementsToKeep map[string]string
}

func main() {
	elementsToKeep := getElementsToKeep()
	gen := CreateFunctionSpecGenerator(elementsToKeep)
	doc := gen.generateDocFromCRD()
	replaceDocInMD(doc)
}

func getElementsToKeep() map[string]string {
	inDoc, err := os.ReadFile(MDFilename)
	if err != nil {
		panic(err)
	}

	reFunSpec := regexp.MustCompile(REFunctionSpecPattern)
	funSpecPart := reFunSpec.FindString(string(inDoc))
	reKeep := regexp.MustCompile(REKeepThisPattern)
	rowsToKeep := reKeep.FindAllStringSubmatch(funSpecPart, -1)

	toKeep := map[string]string{}
	for _, pair := range rowsToKeep {
		rowContent := pair[0]
		paramName := pair[1]
		toKeep[paramName] = rowContent
	}
	return toKeep
}

func replaceDocInMD(doc string) {
	inDoc, err := os.ReadFile(MDFilename)
	if err != nil {
		panic(err)
	}

	newContent := strings.Join([]string{
		"<!-- " + FunctionSpecIdentifier + "-START -->",
		doc,
		"<!-- " + FunctionSpecIdentifier + "-END -->",
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

func CreateFunctionSpecGenerator(toKeep map[string]string) FunctionSpecGenerator {
	return FunctionSpecGenerator{
		elementsToKeep: toKeep,
	}
}

func (g *FunctionSpecGenerator) generateDocFromCRD() string {
	input, err := os.ReadFile(CRDFilename)
	if err != nil {
		panic(err)
	}

	// why unmarshalling to CustomResource don't work?
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
		for k, v := range g.generateElementDoc(functionSpec, "spec", true, "") {
			docElements[k] = v
		}
	}

	for k, v := range g.elementsToKeep {
		docElements[k] = v
	}

	var doc []string
	for _, propName := range sortedKeys(docElements) {
		doc = append(doc, docElements[propName])
	}
	return strings.Join(doc, "\n")
}

func (g *FunctionSpecGenerator) generateElementDoc(obj interface{}, name string, required bool, parentPath string) map[string]string {
	result := map[string]string{}
	element := obj.(map[string]interface{})
	elementType := element["type"].(string)
	description := ""
	if d := element["description"]; d != nil {
		description = d.(string)
	}

	fullName := fmt.Sprintf("%s%s", parentPath, name)
	_, isRowToKeep := g.elementsToKeep[fullName]
	if !isRowToKeep {
		result[fullName] =
			fmt.Sprintf("| **%s** | %s | %s |",
				fullName, yesNo(required), normalizeDescription(description, name))
	}

	if elementType == "object" {
		for k, v := range g.generateObjectDoc(element, name, parentPath) {
			result[k] = v
		}
	}
	return result
}

func (g *FunctionSpecGenerator) generateObjectDoc(element map[string]interface{}, name string, parentPath string) map[string]string {
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
		for k, v := range g.generateElementDoc(propMap[propName], propName, propRequired, parentPath+name+".") {
			result[k] = v
		}
	}
	return result
}

func getElement(obj interface{}, path ...string) interface{} {
	elem := obj
	for _, p := range path {
		elem = elem.(map[string]interface{})[p]
	}
	return elem
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
