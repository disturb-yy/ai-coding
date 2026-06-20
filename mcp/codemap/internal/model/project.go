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
