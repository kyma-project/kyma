package failery

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/types"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/tools/imports"
)

var invalidIdentifierChar = regexp.MustCompile("[^[:digit:][:alpha:]_]")

func getGoPathSrc() string {
	return filepath.Join(filepath.SplitList(build.Default.GOPATH)[0], "src")
}

func stripChars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, str)
}

// Generator is responsible for generating the string containing
// imports and the mock struct that will later be written out as file.
type Generator struct {
	buf bytes.Buffer

	ip               bool
	iface            *Interface
	pkg              string
	localPackageName *string

	importsWerePopulated bool
	localizationCache    map[string]string
	packagePathToName    map[string]string
	nameToPackagePath    map[string]string

	packageRoots []string
}

// NewGenerator builds a Generator.
func NewGenerator(iface *Interface, pkg string, inPackage bool) *Generator {

	var roots []string

	for _, root := range filepath.SplitList(build.Default.GOPATH) {
		roots = append(roots, filepath.Join(root, "src"))
	}

	g := &Generator{
		iface:             iface,
		pkg:               pkg,
		ip:                inPackage,
		localizationCache: make(map[string]string),
		packagePathToName: make(map[string]string),
		nameToPackagePath: make(map[string]string),
		packageRoots:      roots,
	}

	return g
}

func (g *Generator) populateImports() {
	if g.importsWerePopulated {
		return
	}
	for i := 0; i < g.iface.Type.NumMethods(); i++ {
		fn := g.iface.Type.Method(i)
		ftype := fn.Type().(*types.Signature)
		g.addImportsFromTuple(ftype.Params())
		g.addImportsFromTuple(ftype.Results())
		g.renderType(g.iface.NamedType)
	}
}

func (g *Generator) addImportsFromTuple(list *types.Tuple) {
	for i := 0; i < list.Len(); i++ {
		// We use renderType here because we need to recursively
		// resolve any types to make sure that all named types that
		// will appear in the interface file are known
		g.renderType(list.At(i).Type())
	}
}

func (g *Generator) addPackageImport(pkg *types.Package) string {
	return g.addPackageImportWithName(pkg.Path(), pkg.Name())
}

func (g *Generator) addPackageImportWithName(path, name string) string {
	path = g.getLocalizedPath(path)
	if existingName, pathExists := g.packagePathToName[path]; pathExists {
		return existingName
	}

	nonConflictingName := g.getNonConflictingName(path, name)
	g.packagePathToName[path] = nonConflictingName
	g.nameToPackagePath[nonConflictingName] = path
	return nonConflictingName
}

func (g *Generator) getNonConflictingName(path, name string) string {
	if !g.importNameExists(name) {
		return name
	}

	// The path will always contain '/' because it is enforced in getLocalizedPath
	// regardless of OS.
	directories := strings.Split(path, "/")

	cleanedDirectories := make([]string, 0, len(directories))
	for _, directory := range directories {
		cleaned := invalidIdentifierChar.ReplaceAllString(directory, "_")
		cleanedDirectories = append(cleanedDirectories, cleaned)
	}
	numDirectories := len(cleanedDirectories)
	var prospectiveName string
	for i := 1; i <= numDirectories; i++ {
		prospectiveName = strings.Join(cleanedDirectories[numDirectories-i:], "")
		if !g.importNameExists(prospectiveName) {
			return prospectiveName
		}
	}
	// Try adding numbers to the given name
	i := 2
	for {
		prospectiveName = fmt.Sprintf("%v%d", name, i)
		if !g.importNameExists(prospectiveName) {
			return prospectiveName
		}
		i++
	}
}

func (g *Generator) importNameExists(name string) bool {
	_, nameExists := g.nameToPackagePath[name]
	return nameExists
}

func (g *Generator) getLocalizedPathFromPackage(pkg *types.Package) string {
	return g.getLocalizedPath(pkg.Path())
}

func calculateImport(set []string, path string) string {
	for _, root := range set {
		if strings.HasPrefix(path, root) {
			packagePath, err := filepath.Rel(root, path)
			if err == nil {
				return packagePath
			} else {
				log.Printf("Unable to localize path %v, %v", path, err)
			}
		}
	}
	return path
}

// TODO(@IvanMalison): Is there not a better way to get the actual
// import path of a package?
func (g *Generator) getLocalizedPath(path string) string {
	if strings.HasSuffix(path, ".go") {
		path, _ = filepath.Split(path)
	}
	if localized, ok := g.localizationCache[path]; ok {
		return localized
	}
	directories := strings.Split(path, string(filepath.Separator))
	numDirectories := len(directories)
	vendorIndex := -1
	for i := 1; i <= numDirectories; i++ {
		dir := directories[numDirectories-i]
		if dir == "vendor" {
			vendorIndex = numDirectories - i
			break
		}
	}

	toReturn := path
	if vendorIndex >= 0 {
		toReturn = filepath.Join(directories[vendorIndex+1:]...)
	} else if filepath.IsAbs(path) {
		toReturn = calculateImport(g.packageRoots, path)
	}

	// Enforce '/' slashes for import paths in every OS.
	toReturn = filepath.ToSlash(toReturn)

	g.localizationCache[path] = toReturn
	return toReturn
}

func (g *Generator) mockName() string {
	if g.ip {
		if ast.IsExported(g.iface.Name) {
			return "Failing" + g.iface.Name
		}
		first := true
		return "failing" + strings.Map(func(r rune) rune {
			if first {
				first = false
				return unicode.ToUpper(r)
			}
			return r
		}, g.iface.Name)
	}

	return g.iface.Name
}

func (g *Generator) unescapedImportPath(imp *ast.ImportSpec) string {
	return strings.Replace(imp.Path.Value, "\"", "", -1)
}

func (g *Generator) getImportStringFromSpec(imp *ast.ImportSpec) string {
	if name, ok := g.packagePathToName[g.unescapedImportPath(imp)]; ok {
		return fmt.Sprintf("import %s %s\n", name, imp.Path.Value)
	}
	return fmt.Sprintf("import %s\n", imp.Path.Value)
}

func (g *Generator) sortedImportNames() (importNames []string) {
	for name := range g.nameToPackagePath {
		importNames = append(importNames, name)
	}
	sort.Strings(importNames)
	return
}

func (g *Generator) generateImports() {
	pkgPath := g.nameToPackagePath[g.iface.Pkg.Name()]
	// Sort by import name so that we get a deterministic order
	for _, name := range g.sortedImportNames() {
		path := g.nameToPackagePath[name]
		if g.ip && path == pkgPath {
			continue
		}
		g.printf("import %s \"%s\"\n", name, path)
	}
}

// GeneratePrologue generates the prologue of the mock.
func (g *Generator) GeneratePrologue(pkg string) {
	g.populateImports()
	if g.ip {
		g.printf("package %s\n\n", g.iface.Pkg.Name())
	} else {
		g.printf("package %v\n\n", pkg)
	}

	g.generateImports()
	g.printf("\n")
}

// GeneratePrologueNote adds a note after the prologue to the output
// string.
func (g *Generator) GeneratePrologueNote(note string) {
	g.printf("// Code generated by failery v%s. DO NOT EDIT.\n", SemVer)
	if note != "" {
		g.printf("\n")
		for _, n := range strings.Split(note, "\\n") {
			g.printf("// %s\n", n)
		}
	}
	g.printf("\n")
}

// ErrNotInterface is returned when the given type is not an interface
// type.
var ErrNotInterface = errors.New("expression not an interface")

func (g *Generator) printf(s string, vals ...interface{}) {
	fmt.Fprintf(&g.buf, s, vals...)
}

var builtinTypes = map[string]bool{
	"ComplexType": true,
	"FloatType":   true,
	"IntegerType": true,
	"Type":        true,
	"Type1":       true,
	"bool":        true,
	"byte":        true,
	"complex128":  true,
	"complex64":   true,
	"error":       true,
	"float32":     true,
	"float64":     true,
	"int":         true,
	"int16":       true,
	"int32":       true,
	"int64":       true,
	"int8":        true,
	"rune":        true,
	"string":      true,
	"uint":        true,
	"uint16":      true,
	"uint32":      true,
	"uint64":      true,
	"uint8":       true,
	"uintptr":     true,
}

type namer interface {
	Name() string
}

func (g *Generator) renderType(typ types.Type) string {
	switch t := typ.(type) {
	case *types.Named:
		o := t.Obj()
		if o.Pkg() == nil || o.Pkg().Name() == "main" || (g.ip && o.Pkg() == g.iface.Pkg) {
			return o.Name()
		}
		return g.addPackageImport(o.Pkg()) + "." + o.Name()
	case *types.Basic:
		return t.Name()
	case *types.Pointer:
		return "*" + g.renderType(t.Elem())
	case *types.Slice:
		return "[]" + g.renderType(t.Elem())
	case *types.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), g.renderType(t.Elem()))
	case *types.Signature:
		switch t.Results().Len() {
		case 0:
			return fmt.Sprintf(
				"func(%s)",
				g.renderTypeTuple(t.Params()),
			)
		case 1:
			return fmt.Sprintf(
				"func(%s) %s",
				g.renderTypeTuple(t.Params()),
				g.renderType(t.Results().At(0).Type()),
			)
		default:
			return fmt.Sprintf(
				"func(%s)(%s)",
				g.renderTypeTuple(t.Params()),
				g.renderTypeTuple(t.Results()),
			)
		}
	case *types.Map:
		kt := g.renderType(t.Key())
		vt := g.renderType(t.Elem())

		return fmt.Sprintf("map[%s]%s", kt, vt)
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			return "chan " + g.renderType(t.Elem())
		case types.RecvOnly:
			return "<-chan " + g.renderType(t.Elem())
		default:
			return "chan<- " + g.renderType(t.Elem())
		}
	case *types.Struct:
		var fields []string

		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)

			if f.Anonymous() {
				fields = append(fields, g.renderType(f.Type()))
			} else {
				fields = append(fields, fmt.Sprintf("%s %s", f.Name(), g.renderType(f.Type())))
			}
		}

		return fmt.Sprintf("struct{%s}", strings.Join(fields, ";"))
	case *types.Interface:
		if t.NumMethods() != 0 {
			panic("Unable to mock inline interfaces with methods")
		}

		return "interface{}"
	case namer:
		return t.Name()
	default:
		panic(fmt.Sprintf("un-namable type: %#v (%T)", t, t))
	}
}

func (g *Generator) renderTypeTuple(tup *types.Tuple) string {
	var parts []string

	for i := 0; i < tup.Len(); i++ {
		v := tup.At(i)

		parts = append(parts, g.renderType(v.Type()))
	}

	return strings.Join(parts, " , ")
}

func isNillable(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Pointer, *types.Array, *types.Map, *types.Interface, *types.Signature, *types.Chan, *types.Slice:
		return true
	case *types.Named:
		return isNillable(t.Underlying())
	}
	return false
}

type paramList struct {
	Names    []string
	Types    []string
	Params   []string
	Nilable  []bool
	Variadic bool
}

func (g *Generator) genList(list *types.Tuple, variadic bool) *paramList {
	var params paramList

	if list == nil {
		return &params
	}

	for i := 0; i < list.Len(); i++ {
		v := list.At(i)

		ts := g.renderType(v.Type())

		if variadic && i == list.Len()-1 {
			t := v.Type()
			switch t := t.(type) {
			case *types.Slice:
				params.Variadic = true
				ts = "..." + g.renderType(t.Elem())
			default:
				panic("bad variadic type!")
			}
		}

		pname := v.Name()

		if g.nameCollides(pname) || pname == "" {
			pname = fmt.Sprintf("_a%d", i)
		}

		params.Names = append(params.Names, pname)
		params.Types = append(params.Types, ts)

		params.Params = append(params.Params, fmt.Sprintf("%s %s", pname, ts))
		params.Nilable = append(params.Nilable, isNillable(v.Type()))
	}

	return &params
}

func (g *Generator) nameCollides(pname string) bool {
	if pname == g.pkg {
		return true
	}
	return g.importNameExists(pname)
}

// ErrNotSetup is returned when the generator is not configured.
var ErrNotSetup = errors.New("not setup")

// Generate builds a string that constitutes a valid go source file
// containing the mock of the relevant interface.
func (g *Generator) Generate() error {
	g.populateImports()
	if g.iface == nil {
		return ErrNotSetup
	}

	g.printf(
		"// %s is an autogenerated failing mock type for the %s type\n", g.mockName(),
		g.iface.Name,
	)

	g.printf(
		"type %s struct {\n\terr error\n}\n\n", g.mockName(),
	)

	capitalizedMockName := strings.Title(g.mockName())
	g.printf(
		"// New%s creates a new %s type instance\nfunc New%s(err error) *%s {\n\treturn &%s{err: err}\n}\n\n",
		capitalizedMockName, g.mockName(), capitalizedMockName, g.mockName(), g.mockName(),
	)

	for i := 0; i < g.iface.Type.NumMethods(); i++ {
		fn := g.iface.Type.Method(i)

		ftype := fn.Type().(*types.Signature)
		fname := fn.Name()

		params := g.genList(ftype.Params(), ftype.Variadic())
		returns := g.genList(ftype.Results(), false)

		if len(params.Names) == 0 {
			g.printf("// %s provides a failing mock function with given fields:\n", fname)
		} else {
			g.printf(
				"// %s provides a failing mock function with given fields: %s\n", fname,
				strings.Join(params.Names, ", "),
			)
		}
		g.printf(
			"func (_m *%s) %s(%s) ", g.mockName(), fname,
			strings.Join(params.Params, ", "),
		)

		switch len(returns.Types) {
		case 0:
			g.printf("{\n")
		case 1:
			g.printf("%s {\n", returns.Types[0])
		default:
			g.printf("(%s) {\n", strings.Join(returns.Types, ", "))
		}

		var formattedParamNames string
		for i, name := range params.Names {
			if i > 0 {
				formattedParamNames += ", "
			}

			paramType := params.Types[i]
			// for variable args, move the ... to the end.
			if strings.Index(paramType, "...") == 0 {
				name += "..."
			}
			formattedParamNames += name
		}

		if len(returns.Types) > 0 {

			var ret []string

			for idx, typ := range returns.Types {
				g.printf("\tvar r%d %s\n", idx, typ)
				if typ == "error" {
					g.printf("\tr%d = _m.err\n", idx)
				} else if typ == "*error" {
					g.printf("\tr%d = &_m.err\n", idx)
				}

				ret = append(ret, fmt.Sprintf("r%d", idx))
			}

			g.printf("\n\treturn %s\n", strings.Join(ret, ", "))
		}

		g.printf("}\n")
	}

	return nil
}

func (g *Generator) Write(w io.Writer) error {
	opt := &imports.Options{Comments: true}
	theBytes := g.buf.Bytes()

	res, err := imports.Process("mock.go", theBytes, opt)
	if err != nil {
		line := "--------------------------------------------------------------------------------------------"
		fmt.Fprintf(os.Stderr, "Between the lines is the file (mock.go) failery generated in-memory but detected as invalid:\n%s\n%s\n%s\n", line, g.buf.String(), line)
		return err
	}

	w.Write(res)
	return nil
}
