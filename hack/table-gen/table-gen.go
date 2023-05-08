package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"sigs.k8s.io/yaml"
)

const (
	// Regular expression pattern for reading everything between TABLE-START and TABLE-END tags
	REPattern = `(?s)<!--\s*TABLE-START\s* -->.*<!--\s*TABLE-END\s*-->`

	// template to be used for rendering the crd documentation. Has to iterate over all versions and spec and status.
	// The versions will be sorted:
	// 1. stored version
	// 2. served version
	// within those version alphanumeric ordering applies

	documentationTemplate = `
{{- range $version := . -}}
### {{ $version.GKV }}

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
{{- range $prop := $version.Spec }}
| **{{range $i, $v := $prop.Path}}{{if $i}}.{{end}}{{$v}}{{end}}** | {{ $prop.ElemType }} | {{ $prop.Description }} |
{{- end }}

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
{{- range $prop := $version.Status }}
| **{{range $i, $v := $prop.Path}}{{if $i}}.{{end}}{{$v}}{{end}}** | {{ $prop.ElemType }} | {{ $prop.Description }} |
{{- end }}

{{ end -}}`
)

var (
	CRDFilename string
	MDFilename  string
	APIVersion  string
	CRDKind     string
	CRDGroup    string
)

// element contains one tree element. can be a simple type (string,
type element struct {
	name        string
	description string
	elemtype    string
	required    bool
	items       *element
	properties  []*element
}

type flatElement struct {
	Path        []string
	Description string
	ElemType    string
}

type crdVersion struct {
	GKV            string // API-GroupKindVersion
	Spec, Status   []flatElement
	Stored, Served bool
}

func (e *element) String() string {
	s := fmt.Sprintf("-----\nname:%v\ndesc:%v\ntype:%v\nreq:%v", e.name, e.description, e.elemtype, e.required)
	s = fmt.Sprintf("%v\nitems: %v", s, e.items)
	for _, p := range e.properties {
		s = fmt.Sprintf("%v \n - %v", s, p)
	}
	return s
}

func main() {
	flag.StringVar(&CRDFilename, "crd-filename", "", "Full or relative Path to the .yaml file containing crd")
	flag.StringVar(&MDFilename, "md-filename", "", "Full or relative Path to the .md file containing the file where we should insert table rows")
	flag.Parse()

	if CRDFilename == "" {
		panic(fmt.Errorf("crd-filename cannot be empty. Please enter the correct filename"))
	}

	if MDFilename == "" {
		panic(fmt.Errorf("md-filename cannot be empty. Please enter the correct filename"))
	}

	doc := generateDocFromCRD()
	replaceDocInMD(doc)
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
func generateDocFromCRD() string {
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

	var crdVersions []crdVersion
	for _, version := range versions.([]interface{}) {
		if v, ok := version.(map[string]interface{}); ok {
			crd := crdVersion{}
			crd.Stored = v["storage"].(bool)
			crd.Served = v["served"].(bool)
			name := getElement(version, "name")
			APIVersion = name.(string)
			crd.GKV = fmt.Sprintf("%v.%v/%v", CRDKind, CRDGroup, APIVersion)
			resource := "spec"
			pathList(version, resource)
			crd.Spec = pathList(version, "spec")
			crd.Status = pathList(version, "status")
			crdVersions = append(crdVersions, crd)
		}
	}

	// sort in reverse order
	sort.Slice(crdVersions, func(i, j int) bool {
		// both are stored or not stored. Falling back to GKV comparison
		if crdVersions[i].Stored == crdVersions[j].Stored {
			return crdVersions[i].GKV > crdVersions[j].GKV
		}
		if crdVersions[i].Stored && !crdVersions[j].Stored {
			return true // stored is more than not stored
		}
		if crdVersions[i].Served && !crdVersions[j].Served {
			return true // served is more than not served
		}
		return false
	})
	return generateSnippet(crdVersions)
}

func generateSnippet(versions []crdVersion) string {
	tmpl, err := template.New("").Parse(documentationTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var b strings.Builder
	err = tmpl.Execute(&b, versions)
	if err != nil {
		log.Fatal(err)
	}
	return b.String()

}

func pathList(version interface{}, resource string) []flatElement {
	elem := getElement(version, "schema", "openAPIV3Schema", "properties", resource)
	e := convertUnstructuredToElementTree(elem, resource, true)
	fe := flatten(e)
	fe = filter(fe, resource)
	return fe
}

func filter(elements []flatElement, pathElement string) []flatElement {
	var elems []flatElement
	for _, elem := range elements {
		if len(elem.Path) > 0 {
			if elem.Path[0] == pathElement {
				elem.Path = elem.Path[1:]
			}
			if len(elem.Path) > 0 {
				elems = append(elems, elem)
			}
		}
	}
	return elems
}

// flatten converts the recursive datastructure of the element into a list of flatElement.
// The names of the elements and their position gets converted into a flat list of path segments
func flatten(e *element) []flatElement {
	if e == nil {
		return nil
	}
	var elems []flatElement
	elem := flatElement{
		Path:        []string{e.name},
		Description: e.description,
		ElemType:    e.elemtype,
	}
	if e.required {
		elem.ElemType += " **required**"
	}

	// recurse into child properties
	for _, p := range e.properties {
		fes := flatten(p)
		for _, fe := range fes {
			fe.Path = append([]string{e.name}, fe.Path...)
			elems = append(elems, fe)
		}
	}
	if e.elemtype == "array" {
		elems = flattenArray(e, &elem, &elems)
	}

	// sort the list by path
	elems = append(elems, elem)
	sort.Slice(elems, func(i, j int) bool {
		return strings.Join(elems[i].Path, "") < strings.Join(elems[j].Path, "")
	})
	return elems
}

func flattenArray(e *element, flatElem *flatElement, flatElems *[]flatElement) []flatElement {
	items := flatten(e.items)
	fes := *flatElems
	// handle an array of objects
	if e.items != nil && e.items.elemtype == "object" {
		// if it is an object we can use the description of the anonymous object to fill gaps in the description of the list
		if flatElem.Description == "" {
			flatElem.Description = items[0].Description
		}
		// the child object is stored in "items" we need to clean this as it would otherwise show up in the path list
		items = filter(items, "items")
		for _, item := range items {
			item.Path = append([]string{e.name}, item.Path...)
			fes = append(fes, item)
		}
	} else { // handle array of simple type
		for _, item := range items {
			flatElem.ElemType = fmt.Sprintf("[]%v", item.ElemType)
		}
	}
	return fes
}

// getElement returns a specific element from obj based on the provided Path.
func getElement(obj interface{}, path ...string) interface{} {
	elem := obj
	for _, p := range path {
		elem = elem.(map[string]interface{})[p]
	}
	return elem
}

// convertUnstructuredToElementTree is a rather simple converter from interface to a tree structure of elements
func convertUnstructuredToElementTree(obj interface{}, name string, required bool) *element {
	e := element{}
	m, ok := obj.(map[string]interface{})
	if !ok {
		return &e
	}

	e.name = name
	e.required = required
	if d, ok := m["description"].(string); ok {
		e.description = d
	}
	e.elemtype = m["type"].(string)

	if e.elemtype == "object" {
		handleObjectType(&e, m)
	}

	if e.elemtype == "array" {
		// store the allowed child type of the list in "items"
		if p, ok := m["items"].(map[string]interface{}); ok {
			e.items = convertUnstructuredToElementTree(p, "items", false)
		}
	}
	return &e
}

func handleObjectType(e *element, m map[string]interface{}) {
	e.properties = []*element{}

	// find required properties
	req := []interface{}{}
	if r, ok := m["required"].([]interface{}); ok {
		req = r
	}

	// recurse into child properties
	if p, ok := m["properties"].(map[string]interface{}); ok {
		for n, ce := range p {
			e.properties = append(e.properties, convertUnstructuredToElementTree(ce, n, contains(req, n)))
		}
	}

	// additionalProperties is an unstructed map of string to type
	if p, ok := m["additionalProperties"].(map[string]interface{}); ok {
		e.elemtype = fmt.Sprintf("%v%v", "map[string]", p["type"].(string))
	}
}

func contains(list []interface{}, value string) bool {
	for _, i := range list {
		if i.(string) == value {
			return true
		}
	}
	return false
}
