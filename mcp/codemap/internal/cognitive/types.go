package cognitive

import "github.com/disturb-yy/codemap/internal/model"

type Repository interface {
	SearchModule(query string) ([]*model.Module, error)
	FindRoutes(query string) ([]*model.Route, error)
	SearchFlow(query string) ([]*model.Flow, error)
	FindCallers(funcName string) ([]*model.CallEdge, error)
	FindCallees(module string) ([]*model.CallEdge, error)
	GetFeatureMap() ([]model.FeatureEntry, error)
	GetNavigationHints() ([]model.NavHintEntry, error)
}

type FindChangePointsRequest struct {
	Requirement string `json:"requirement"`
	TopK        int    `json:"top_k,omitempty"`
}

type FindChangePointsResponse struct {
	Requirement      string            `json:"requirement"`
	MatchedFeatures  []MatchedFeature  `json:"matched_features"`
	CandidateModules []CandidateModule `json:"candidate_modules"`
	CandidateFiles   []CandidateFile   `json:"candidate_files"`
	RelatedRoutes    []string          `json:"related_routes"`
	RelatedFlows     []string          `json:"related_flows"`
	RiskAreas        []RiskArea        `json:"risk_areas"`
	NextActions      []string          `json:"next_actions"`
}

type MatchedFeature struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type CandidateModule struct {
	Module string  `json:"module"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type CandidateFile struct {
	File   string  `json:"file"`
	Module string  `json:"module"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type RiskArea struct {
	Name   string `json:"name"`
	Level  string `json:"level"`
	Reason string `json:"reason"`
}

type RequirementType string

const (
	ReqAddAPI      RequirementType = "add_api"
	ReqModifyLogic RequirementType = "modify_logic"
	ReqFixBug      RequirementType = "fix_bug"
	ReqAddField    RequirementType = "add_field"
	ReqAddAuth     RequirementType = "add_auth"
	ReqUnknown     RequirementType = "unknown"
)
