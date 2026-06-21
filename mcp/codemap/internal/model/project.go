package model

// Project 表示整个项目的知识模型。
type Project struct {
	Name      string      `json:"name"`
	Root      string      `json:"root"`
	Modules   []*Module   `json:"modules"`
	Routes    []*Route    `json:"routes"`
	Flows     []*Flow     `json:"flows"`
	CallEdges []*CallEdge `json:"call_edges"`
}

// Module 表示一个逻辑模块。
type Module struct {
	Name              string   `json:"name"`
	Path              string   `json:"path"`
	Dependencies      []string `json:"dependencies"`
	ExportedTypes     []string `json:"exported_types"`
	ExportedFunctions []string `json:"exported_functions"`
	ExportedMethods   []string `json:"exported_methods"`
	KeyInterfaces     []string `json:"key_interfaces"`
}

// Route 表示一个 HTTP 路由。
type Route struct {
	Path    string `json:"path"`
	Method  string `json:"method"`
	Handler string `json:"handler"`
	Module  string `json:"module"`
}

// Flow 表示模块间的调用/数据流。
type Flow struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Trigger string   `json:"trigger"`
	Steps   []string `json:"steps"`
}

// CallEdge 表示一次函数调用（call graph 边）。
type CallEdge struct {
	CallerModule string `json:"caller_module"`
	CallerFunc   string `json:"caller_func"`
	CalleeModule string `json:"callee_module"`
	CalleeFunc   string `json:"callee_func"`
}

// FeatureEntry represents a business feature for the get_feature_map MCP tool.
type FeatureEntry struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Modules     []string `json:"modules"`
	Routes      []string `json:"routes"`
	Flows       []string `json:"flows"`
}

// NavHintEntry represents a navigation hint for the get_navigation_hints MCP tool.
type NavHintEntry struct {
	Feature        string   `json:"feature"`
	StartFiles     []string `json:"start_files"`
	Routes         []string `json:"routes"`
	RelatedModules []string `json:"related_modules"`
	RelatedFlows   []string `json:"related_flows"`
	Risk           []string `json:"risk"`
}
