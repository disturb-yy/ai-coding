package java

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/disturb-yy/codemap/internal/model"
)

type Analyzer struct{}

func New() *Analyzer { return &Analyzer{} }

type javaFile struct {
	path       string
	srcRoot    string
	modulePath string
	pkg        string
	imports    []string
	fields     map[string]string
	types      []javaType
	methods    []javaMethod
}

type javaType struct {
	name        string
	isInterface bool
	isPublic    bool
	baseRoutes  []routeMapping
}

type javaMethod struct {
	name        string
	owner       string
	isPublic    bool
	annotations []string
	body        []string
	params      map[string]string
	locals      map[string]string
}

type routeMapping struct {
	method string
	path   string
}

var (
	packageRE = regexp.MustCompile(`^\s*package\s+([A-Za-z_][\w.]*)\s*;`)
	importRE  = regexp.MustCompile(`^\s*import\s+(?:static\s+)?([A-Za-z_][\w.]*(?:\.\*)?)\s*;`)
	typeRE    = regexp.MustCompile(`\b(public\s+)?(?:abstract\s+|final\s+)?(class|interface|enum)\s+([A-Za-z_$][\w$]*)`)
	methodRE  = regexp.MustCompile(`\b(public|protected|private)?\s*(?:static\s+|final\s+|abstract\s+|synchronized\s+|default\s+|native\s+)*[\w<>\[\], ?.$]+\s+([A-Za-z_$][\w$]*)\s*\(([^)]*)\)`)
	fieldRE   = regexp.MustCompile(`\b(?:private|protected|public)?\s*(?:final\s+|static\s+)*([A-Z][\w$]*(?:<[^;=(){}]+>)?)\s+([a-zA-Z_$][\w$]*)\s*(?:[;=,])`)
	callRE    = regexp.MustCompile(`\b([A-Za-z_$][\w$]*)\s*\.\s*([A-Za-z_$][\w$]*)\s*\(`)
)

var springRouteAnnotations = map[string]string{
	"GetMapping":     "GET",
	"PostMapping":    "POST",
	"PutMapping":     "PUT",
	"DeleteMapping":  "DELETE",
	"PatchMapping":   "PATCH",
	"RequestMapping": "ANY",
}

func (a *Analyzer) Analyze(ctx context.Context, root string) (*model.Project, error) {
	project := &model.Project{Name: filepath.Base(root), Root: root}
	srcRoots := inferSourceRoots(root)
	if len(srcRoots) == 0 {
		return project, fmt.Errorf("no Java source directory found (expected src/main/java, src, or .java files)")
	}

	var files []*javaFile
	moduleIndex := make(map[string]*model.Module)
	packageToModule := make(map[string]string)
	classToModule := make(map[string]string)

	for _, srcRoot := range srcRoots {
		err := filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if info.IsDir() || filepath.Ext(path) != ".java" {
				return nil
			}
			jf, err := parseJavaFile(path, srcRoot)
			if err != nil || jf == nil {
				return nil
			}
			files = append(files, jf)
			mod := ensureModule(moduleIndex, jf.modulePath, jf.pkg)
			for _, typ := range jf.types {
				if typ.isPublic {
					mod.ExportedTypes = appendStrIfNew(mod.ExportedTypes, typ.name)
					if typ.isInterface {
						mod.KeyInterfaces = appendStrIfNew(mod.KeyInterfaces, typ.name)
					}
				}
				classToModule[typ.name] = jf.modulePath
				if jf.pkg != "" {
					classToModule[jf.pkg+"."+typ.name] = jf.modulePath
				}
			}
			if jf.pkg != "" {
				packageToModule[jf.pkg] = jf.modulePath
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	for _, jf := range files {
		mod := ensureModule(moduleIndex, jf.modulePath, jf.pkg)
		importAliases := buildImportAliases(jf.imports, classToModule, packageToModule)

		for _, method := range jf.methods {
			if method.isPublic && method.owner != "" {
				mod.ExportedMethods = appendStrIfNew(mod.ExportedMethods, method.owner+"."+method.name)
			}
		}

		for _, dep := range fileDependencies(jf, importAliases, packageToModule) {
			if dep != "" && dep != jf.modulePath {
				addDepIfNew(mod, dep)
			}
		}

		project.Routes = append(project.Routes, extractRoutes(jf)...)
		project.CallEdges = append(project.CallEdges, extractCallEdges(jf, importAliases, classToModule, packageToModule)...)
	}

	project.Flows = flowsFromEdges(project.CallEdges)
	for _, module := range moduleIndex {
		project.Modules = append(project.Modules, module)
	}
	sort.Slice(project.Modules, func(i, j int) bool {
		return project.Modules[i].Path < project.Modules[j].Path
	})

	return project, nil
}

func inferSourceRoots(root string) []string {
	for _, c := range []string{
		filepath.Join(root, "src", "main", "java"),
		filepath.Join(root, "src"),
		root,
	} {
		if info, err := os.Stat(c); err == nil && info.IsDir() && containsJava(c) {
			return []string{c}
		}
	}
	return nil
}

func containsJava(root string) bool {
	found := false
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".java" {
			found = true
		}
		return nil
	})
	return found
}

func parseJavaFile(filePath, srcRoot string) (*javaFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	lines := stripBlockComments(strings.Split(string(data), "\n"))
	jf := &javaFile{
		path:       filePath,
		srcRoot:    srcRoot,
		modulePath: modulePath(filePath, srcRoot),
		fields:     make(map[string]string),
	}
	var pendingAnnotations []string
	var currentType string
	var inMethod bool
	var current javaMethod
	braceDepth := 0

	for _, raw := range lines {
		line := stripLineComment(strings.TrimSpace(raw))
		if line == "" {
			continue
		}
		if matches := packageRE.FindStringSubmatch(line); len(matches) == 2 {
			jf.pkg = matches[1]
			continue
		}
		if matches := importRE.FindStringSubmatch(line); len(matches) == 2 {
			jf.imports = append(jf.imports, matches[1])
			continue
		}
		if strings.HasPrefix(line, "@") {
			pendingAnnotations = append(pendingAnnotations, line)
			if !strings.Contains(line, ")") && !strings.Contains(line, "(") {
				continue
			}
			continue
		}
		if inMethod {
			current.body = append(current.body, line)
			for name, typ := range extractVars(line) {
				current.locals[name] = typ
			}
			braceDepth += braceDelta(line)
			if braceDepth <= 0 {
				jf.methods = append(jf.methods, current)
				inMethod = false
			}
			continue
		}
		if typ, ok := parseType(line, pendingAnnotations); ok {
			currentType = typ.name
			jf.types = append(jf.types, typ)
			pendingAnnotations = nil
			continue
		}
		if method, ok := parseMethod(line, currentType, pendingAnnotations); ok {
			method.body = append(method.body, line)
			method.locals = copyStringMap(jf.fields)
			for name, typ := range extractVars(line) {
				method.locals[name] = typ
			}
			method.params = extractParams(methodSignatureParams(line))
			for name, typ := range method.params {
				method.locals[name] = typ
			}
			pendingAnnotations = nil
			braceDepth = braceDelta(line)
			if braceDepth <= 0 {
				jf.methods = append(jf.methods, method)
			} else {
				current = method
				inMethod = true
			}
			continue
		}
		for name, typ := range extractVars(line) {
			jf.fields[name] = typ
		}
		pendingAnnotations = nil
	}
	return jf, nil
}

func stripBlockComments(lines []string) []string {
	out := make([]string, 0, len(lines))
	inBlock := false
	for _, line := range lines {
		for {
			if inBlock {
				end := strings.Index(line, "*/")
				if end < 0 {
					line = ""
					break
				}
				line = line[end+2:]
				inBlock = false
				continue
			}
			start := strings.Index(line, "/*")
			if start < 0 {
				break
			}
			end := strings.Index(line[start+2:], "*/")
			if end < 0 {
				line = line[:start]
				inBlock = true
				break
			}
			line = line[:start] + line[start+2+end+2:]
		}
		out = append(out, line)
	}
	return out
}

func stripLineComment(line string) string {
	if idx := strings.Index(line, "//"); idx >= 0 {
		return strings.TrimSpace(line[:idx])
	}
	return line
}

func parseType(line string, annotations []string) (javaType, bool) {
	matches := typeRE.FindStringSubmatch(line)
	if len(matches) != 4 {
		return javaType{}, false
	}
	var base []routeMapping
	for _, ann := range annotations {
		base = append(base, routeMappingsFromAnnotation(ann)...)
	}
	return javaType{
		name:        matches[3],
		isInterface: matches[2] == "interface",
		isPublic:    strings.TrimSpace(matches[1]) == "public",
		baseRoutes:  base,
	}, true
}

func parseMethod(line, owner string, annotations []string) (javaMethod, bool) {
	if strings.Contains(line, " class ") || strings.Contains(line, " interface ") || strings.Contains(line, " enum ") {
		return javaMethod{}, false
	}
	if isControlLine(line) {
		return javaMethod{}, false
	}
	matches := methodRE.FindStringSubmatch(line)
	if len(matches) != 4 {
		return javaMethod{}, false
	}
	return javaMethod{
		name:        matches[2],
		owner:       owner,
		isPublic:    matches[1] == "public",
		annotations: append([]string(nil), annotations...),
		locals:      make(map[string]string),
		params:      extractParams(matches[3]),
	}, true
}

func isControlLine(line string) bool {
	for _, prefix := range []string{"if ", "for ", "while ", "switch ", "catch ", "return ", "new "} {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func methodSignatureParams(line string) string {
	open := strings.Index(line, "(")
	close := strings.Index(line, ")")
	if open < 0 || close < open {
		return ""
	}
	return line[open+1 : close]
}

func extractParams(params string) map[string]string {
	result := make(map[string]string)
	for _, param := range strings.Split(params, ",") {
		fields := strings.Fields(strings.TrimSpace(param))
		if len(fields) < 2 {
			continue
		}
		name := fields[len(fields)-1]
		name = strings.TrimPrefix(name, "...")
		typ := sanitizeType(fields[len(fields)-2])
		if name != "" && typ != "" {
			result[name] = typ
		}
	}
	return result
}

func extractVars(line string) map[string]string {
	result := make(map[string]string)
	for _, match := range fieldRE.FindAllStringSubmatch(line, -1) {
		if len(match) == 3 {
			result[match[2]] = sanitizeType(match[1])
		}
	}
	return result
}

func sanitizeType(typ string) string {
	typ = strings.TrimSpace(typ)
	if idx := strings.Index(typ, "<"); idx >= 0 {
		typ = typ[:idx]
	}
	typ = strings.TrimPrefix(typ, "final ")
	typ = strings.TrimSuffix(typ, "[]")
	return typ
}

func copyStringMap(values map[string]string) map[string]string {
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}

func braceDelta(line string) int {
	return strings.Count(line, "{") - strings.Count(line, "}")
}

func ensureModule(moduleIndex map[string]*model.Module, path, pkg string) *model.Module {
	if path == "" {
		path = "."
	}
	mod, ok := moduleIndex[path]
	if ok {
		return mod
	}
	name := pkg
	if name == "" {
		name = filepath.Base(path)
	}
	mod = &model.Module{Name: name, Path: path}
	moduleIndex[path] = mod
	return mod
}

func modulePath(filePath, srcRoot string) string {
	rel, err := filepath.Rel(srcRoot, filepath.Dir(filePath))
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

func buildImportAliases(imports []string, classToModule, packageToModule map[string]string) map[string]string {
	aliases := make(map[string]string)
	for _, imp := range imports {
		if isStdLib(imp) {
			continue
		}
		if strings.HasSuffix(imp, ".*") {
			pkg := strings.TrimSuffix(imp, ".*")
			if mod := packageToModule[pkg]; mod != "" {
				aliases[pkg] = mod
			}
			continue
		}
		parts := strings.Split(imp, ".")
		if len(parts) == 0 {
			continue
		}
		simple := parts[len(parts)-1]
		if mod := classToModule[imp]; mod != "" {
			aliases[simple] = mod
		}
	}
	return aliases
}

func fileDependencies(jf *javaFile, aliases, packageToModule map[string]string) []string {
	seen := make(map[string]bool)
	for _, imp := range jf.imports {
		if isStdLib(imp) {
			continue
		}
		if strings.HasSuffix(imp, ".*") {
			if mod := packageToModule[strings.TrimSuffix(imp, ".*")]; mod != "" {
				seen[mod] = true
			}
			continue
		}
		parts := strings.Split(imp, ".")
		if len(parts) > 0 {
			if mod := aliases[parts[len(parts)-1]]; mod != "" {
				seen[mod] = true
			}
		}
	}
	deps := make([]string, 0, len(seen))
	for dep := range seen {
		deps = append(deps, dep)
	}
	sort.Strings(deps)
	return deps
}

func extractRoutes(jf *javaFile) []*model.Route {
	var routes []*model.Route
	baseByType := make(map[string][]routeMapping)
	for _, typ := range jf.types {
		baseByType[typ.name] = typ.baseRoutes
	}
	for _, method := range jf.methods {
		methodMappings := routeMappingsFromAnnotations(method.annotations)
		if len(methodMappings) == 0 {
			continue
		}
		baseMappings := baseByType[method.owner]
		if len(baseMappings) == 0 {
			baseMappings = []routeMapping{{method: "ANY", path: ""}}
		}
		for _, base := range baseMappings {
			for _, mapping := range methodMappings {
				httpMethod := mapping.method
				if httpMethod == "ANY" && base.method != "ANY" {
					httpMethod = base.method
				}
				routes = append(routes, &model.Route{
					Path:    joinRoutePath(base.path, mapping.path),
					Method:  httpMethod,
					Handler: fmt.Sprintf("%s/%s.%s", jf.modulePath, method.owner, method.name),
					Module:  jf.modulePath,
				})
			}
		}
	}
	return routes
}

func routeMappingsFromAnnotations(annotations []string) []routeMapping {
	var mappings []routeMapping
	for _, ann := range annotations {
		mappings = append(mappings, routeMappingsFromAnnotation(ann)...)
	}
	return mappings
}

func routeMappingsFromAnnotation(annotation string) []routeMapping {
	name := annotationName(annotation)
	method, ok := springRouteAnnotations[name]
	if !ok {
		return nil
	}
	if name == "RequestMapping" {
		if detected := requestMappingMethod(annotation); detected != "" {
			method = detected
		}
	}
	paths := annotationPaths(annotation)
	if len(paths) == 0 {
		paths = []string{""}
	}
	mappings := make([]routeMapping, 0, len(paths))
	for _, path := range paths {
		mappings = append(mappings, routeMapping{method: method, path: path})
	}
	return mappings
}

func annotationName(annotation string) string {
	annotation = strings.TrimPrefix(strings.TrimSpace(annotation), "@")
	for i, r := range annotation {
		if !(r == '_' || r == '$' || r == '.' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			parts := strings.Split(annotation[:i], ".")
			return parts[len(parts)-1]
		}
	}
	parts := strings.Split(annotation, ".")
	return parts[len(parts)-1]
}

func annotationPaths(annotation string) []string {
	start := strings.Index(annotation, "(")
	end := strings.LastIndex(annotation, ")")
	if start < 0 || end <= start {
		return nil
	}
	args := annotation[start+1 : end]
	var paths []string
	for _, key := range []string{"value", "path"} {
		re := regexp.MustCompile(key + `\s*=\s*(?:\{\s*)?"([^"]+)"`)
		for _, match := range re.FindAllStringSubmatch(args, -1) {
			paths = append(paths, match[1])
		}
	}
	if len(paths) == 0 {
		re := regexp.MustCompile(`"([^"]+)"`)
		for _, match := range re.FindAllStringSubmatch(args, -1) {
			if strings.HasPrefix(match[1], "/") {
				paths = append(paths, match[1])
			}
		}
	}
	return uniqueStrings(paths)
}

func requestMappingMethod(annotation string) string {
	for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"} {
		if strings.Contains(annotation, "RequestMethod."+method) {
			return method
		}
	}
	return ""
}

func joinRoutePath(base, child string) string {
	if base == "" && child == "" {
		return "/"
	}
	joined := "/" + strings.Trim(strings.TrimSpace(base)+"/"+strings.TrimSpace(child), "/")
	return strings.ReplaceAll(joined, "//", "/")
}

func extractCallEdges(jf *javaFile, aliases, classToModule, packageToModule map[string]string) []*model.CallEdge {
	var edges []*model.CallEdge
	localMethods := make(map[string]bool)
	for _, method := range jf.methods {
		localMethods[method.name] = true
	}
	for _, method := range jf.methods {
		caller := method.owner + "." + method.name
		for _, line := range method.body {
			for _, match := range callRE.FindAllStringSubmatch(line, -1) {
				if len(match) != 3 {
					continue
				}
				target, callee := match[1], match[2]
				calleeModule := resolveTargetModule(target, method.locals, aliases, classToModule, packageToModule, jf.modulePath)
				if calleeModule == "" {
					continue
				}
				edge := &model.CallEdge{
					CallerModule: jf.modulePath,
					CallerFunc:   caller,
					CalleeModule: calleeModule,
					CalleeFunc:   target + "." + callee,
				}
				if !callEdgeExists(edges, edge) {
					edges = append(edges, edge)
				}
			}
			if !methodRE.MatchString(line) {
				for _, local := range localMethodCalls(line, localMethods) {
					edge := &model.CallEdge{
						CallerModule: jf.modulePath,
						CallerFunc:   caller,
						CalleeModule: jf.modulePath,
						CalleeFunc:   method.owner + "." + local,
					}
					if !callEdgeExists(edges, edge) {
						edges = append(edges, edge)
					}
				}
			}
		}
	}
	return edges
}

func resolveTargetModule(target string, locals, aliases, classToModule, packageToModule map[string]string, currentModule string) string {
	if typ := locals[target]; typ != "" {
		if mod := aliases[typ]; mod != "" {
			return mod
		}
		if mod := classToModule[typ]; mod != "" {
			return mod
		}
	}
	if mod := aliases[target]; mod != "" {
		return mod
	}
	if mod := classToModule[target]; mod != "" {
		return mod
	}
	if mod := packageToModule[target]; mod != "" {
		return mod
	}
	if target == "this" || target == "super" {
		return currentModule
	}
	return ""
}

func localMethodCalls(line string, localMethods map[string]bool) []string {
	var calls []string
	for name := range localMethods {
		if strings.Contains(line, name+"(") && !strings.Contains(line, "."+name+"(") {
			calls = append(calls, name)
		}
	}
	sort.Strings(calls)
	return calls
}

func flowsFromEdges(edges []*model.CallEdge) []*model.Flow {
	grouped := make(map[string][]string)
	for _, edge := range edges {
		if edge.CallerModule == edge.CalleeModule {
			continue
		}
		key := edge.CallerModule + "\x00" + edge.CalleeModule
		grouped[key] = appendStrIfNew(grouped[key], edge.CalleeFunc)
	}
	var flows []*model.Flow
	for key, funcs := range grouped {
		parts := strings.Split(key, "\x00")
		sort.Strings(funcs)
		from, to := parts[0], parts[1]
		steps := []string{fmt.Sprintf("%s -> %s", from, to)}
		for _, fn := range funcs {
			steps = append(steps, "  call: "+fn)
		}
		flows = append(flows, &model.Flow{
			ID:      safeID(from + "_to_" + to),
			Name:    fmt.Sprintf("%s calls %s", from, to),
			Trigger: from,
			Steps:   steps,
		})
	}
	sort.Slice(flows, func(i, j int) bool {
		return flows[i].ID < flows[j].ID
	})
	return flows
}

func safeID(id string) string {
	id = strings.ReplaceAll(id, "/", "_")
	id = strings.ReplaceAll(id, ".", "_")
	return id
}

func isStdLib(imp string) bool {
	prefixes := []string{
		"java.", "javax.", "jakarta.", "org.w3c.", "org.xml.",
		"org.springframework.", "org.hibernate.", "org.apache.",
		"org.junit.", "org.mockito.", "org.slf4j.", "org.testng.",
		"com.google.common.", "com.google.gson.", "com.fasterxml.",
		"lombok.", "io.swagger.",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(imp, p) {
			return true
		}
	}
	return false
}

func appendStrIfNew(slice []string, s string) []string {
	for _, existing := range slice {
		if existing == s {
			return slice
		}
	}
	return append(slice, s)
}

func uniqueStrings(values []string) []string {
	var out []string
	for _, value := range values {
		out = appendStrIfNew(out, value)
	}
	return out
}

func addDepIfNew(mod *model.Module, dep string) {
	for _, existing := range mod.Dependencies {
		if existing == dep {
			return
		}
	}
	mod.Dependencies = append(mod.Dependencies, dep)
	sort.Strings(mod.Dependencies)
}

func callEdgeExists(edges []*model.CallEdge, edge *model.CallEdge) bool {
	for _, existing := range edges {
		if existing.CallerModule == edge.CallerModule &&
			existing.CallerFunc == edge.CallerFunc &&
			existing.CalleeModule == edge.CalleeModule &&
			existing.CalleeFunc == edge.CalleeFunc {
			return true
		}
	}
	return false
}
