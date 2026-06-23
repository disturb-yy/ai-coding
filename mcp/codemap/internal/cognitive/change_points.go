package cognitive

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/disturb-yy/codemap/internal/model"
)

type CognitiveService struct {
	repo Repository
}

func NewService(repo Repository) *CognitiveService {
	return &CognitiveService{repo: repo}
}

func (s *CognitiveService) FindChangePoints(ctx context.Context, req FindChangePointsRequest) (*FindChangePointsResponse, error) {
	if strings.TrimSpace(req.Requirement) == "" {
		return nil, fmt.Errorf("requirement is required")
	}
	topK := req.TopK
	if topK <= 0 {
		topK = 5
	}

	tokens := TokenizeRequirement(req.Requirement)
	features, err := s.repo.GetFeatureMap()
	if err != nil {
		return nil, fmt.Errorf("get feature map: %w", err)
	}
	hints, err := s.repo.GetNavigationHints()
	if err != nil {
		return nil, fmt.Errorf("get navigation hints: %w", err)
	}
	modules, err := s.repo.SearchModule("")
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}

	matched := matchFeatures(tokens, features)
	routes, flows, err := s.findRelatedFacts(tokens, matched)
	if err != nil {
		return nil, err
	}

	callerCounts, calleeCounts, dependencyCounts, dependentCounts := s.collectGraphCounts(modules, tokens, matched)
	moduleScores := scoreModules(tokens, matched, hints, modules, routes, flows, callerCounts, calleeCounts)
	fileScores := scoreFiles(tokens, ClassifyRequirement(tokens), moduleScores, hints, routes, flows)
	risks := inferRiskAreas(moduleScores, routes, dependencyCounts, dependentCounts, callerCounts)

	return &FindChangePointsResponse{
		Requirement:      req.Requirement,
		MatchedFeatures:  toMatchedFeatures(matched, topK),
		CandidateModules: toCandidateModules(moduleScores, topK),
		CandidateFiles:   toCandidateFiles(fileScores, topK),
		RelatedRoutes:    limitStrings(relatedRoutes(matched, routes), topK*2),
		RelatedFlows:     limitStrings(relatedFlows(matched, flows), topK*2),
		RiskAreas:        limitRisks(risks, topK),
		NextActions:      buildNextActions(moduleScores, fileScores, topK),
	}, nil
}

func (s *CognitiveService) findRelatedFacts(tokens []string, matches []featureMatch) ([]*model.Route, []*model.Flow, error) {
	seenRoutes := make(map[string]*model.Route)
	seenFlows := make(map[string]*model.Flow)
	addRoutes := func(query string) error {
		routes, err := s.repo.FindRoutes(query)
		if err != nil {
			return fmt.Errorf("find routes %q: %w", query, err)
		}
		for _, route := range routes {
			seenRoutes[route.Method+" "+route.Path+" "+route.Handler] = route
		}
		return nil
	}
	addFlows := func(query string) error {
		flows, err := s.repo.SearchFlow(query)
		if err != nil {
			return fmt.Errorf("search flow %q: %w", query, err)
		}
		for _, flow := range flows {
			seenFlows[flow.ID+" "+flow.Name] = flow
		}
		return nil
	}

	for _, token := range tokens {
		if err := addRoutes(token); err != nil {
			return nil, nil, err
		}
		if err := addFlows(token); err != nil {
			return nil, nil, err
		}
	}
	for _, match := range matches {
		for _, route := range match.feature.Routes {
			parts := strings.Fields(route)
			if len(parts) >= 2 {
				_ = addRoutes(parts[1])
			}
		}
		for _, flow := range match.feature.Flows {
			_ = addFlows(flow)
		}
	}

	routes := make([]*model.Route, 0, len(seenRoutes))
	for _, route := range seenRoutes {
		routes = append(routes, route)
	}
	sort.SliceStable(routes, func(i, j int) bool {
		return routes[i].Method+" "+routes[i].Path < routes[j].Method+" "+routes[j].Path
	})
	flows := make([]*model.Flow, 0, len(seenFlows))
	for _, flow := range seenFlows {
		flows = append(flows, flow)
	}
	sort.SliceStable(flows, func(i, j int) bool {
		return flows[i].Name < flows[j].Name
	})
	return routes, flows, nil
}

func (s *CognitiveService) collectGraphCounts(modules []*model.Module, tokens []string, matches []featureMatch) (map[string]int, map[string]int, map[string]int, map[string]int) {
	callers := make(map[string]int)
	callees := make(map[string]int)
	deps := make(map[string]int)
	dependents := make(map[string]int)
	moduleByPath := make(map[string]string)
	for _, module := range modules {
		moduleByPath[module.Path] = module.Name
		moduleByPath[module.Name] = module.Name
	}
	for _, module := range modules {
		deps[module.Name] = len(module.Dependencies)
		for _, dep := range module.Dependencies {
			name := moduleByPath[dep]
			if name == "" {
				name = dep
			}
			dependents[name]++
		}
	}

	queries := make(map[string]bool)
	for _, token := range tokens {
		queries[token] = true
	}
	for _, match := range matches {
		for _, module := range match.feature.Modules {
			queries[module] = true
		}
	}
	for query := range queries {
		if edges, err := s.repo.FindCallers(query); err == nil {
			for _, edge := range edges {
				callers[edge.CallerModule]++
			}
		}
		if edges, err := s.repo.FindCallees(query); err == nil {
			for _, edge := range edges {
				callees[edge.CalleeModule]++
			}
		}
	}
	return callers, callees, deps, dependents
}

func buildNextActions(modules []moduleScore, files []fileScore, topK int) []string {
	var actions []string
	for _, module := range limitModuleScores(modules, min(topK, 3)) {
		actions = append(actions, fmt.Sprintf("search_module(%s)", module.module))
	}
	if len(modules) > 0 {
		actions = append(actions, fmt.Sprintf("call_graph(%s)", modules[0].module))
		actions = append(actions, fmt.Sprintf("impact_analysis(%s)", modules[0].module))
	}
	for _, file := range limitFileScores(files, min(topK, 2)) {
		actions = append(actions, fmt.Sprintf("open_file(%s)", file.file))
	}
	return uniqueStrings(actions)
}

func relatedRoutes(matches []featureMatch, routes []*model.Route) []string {
	var result []string
	for _, match := range matches {
		result = append(result, match.feature.Routes...)
	}
	for _, route := range routes {
		result = append(result, route.Method+" "+route.Path)
	}
	return uniqueStrings(result)
}

func relatedFlows(matches []featureMatch, flows []*model.Flow) []string {
	var result []string
	for _, match := range matches {
		result = append(result, match.feature.Flows...)
	}
	for _, flow := range flows {
		result = append(result, flow.Name)
	}
	return uniqueStrings(result)
}

func normalizeScore(score, maxScore float64) float64 {
	if maxScore <= 0 {
		return 0
	}
	return math.Round((score/maxScore)*100) / 100
}

func buildMatchReason(reasons []string) string {
	if len(reasons) == 0 {
		return "Related to requirement terms"
	}
	return "Matched " + strings.Join(uniqueStrings(reasons), ", ")
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func limitStrings(values []string, limit int) []string {
	if limit <= 0 || len(values) <= limit {
		return values
	}
	return values[:limit]
}

func limitRisks(values []RiskArea, limit int) []RiskArea {
	if limit <= 0 || len(values) <= limit {
		return values
	}
	return values[:limit]
}

func limitFeatureMatches(values []featureMatch, limit int) []featureMatch {
	if limit <= 0 || len(values) <= limit {
		return values
	}
	return values[:limit]
}

func limitModuleScores(values []moduleScore, limit int) []moduleScore {
	if limit <= 0 || len(values) <= limit {
		return values
	}
	return values[:limit]
}

func limitFileScores(values []fileScore, limit int) []fileScore {
	if limit <= 0 || len(values) <= limit {
		return values
	}
	return values[:limit]
}

func maxFeatureScore(values []featureMatch) float64 {
	var maxScore float64
	for _, value := range values {
		if value.score > maxScore {
			maxScore = value.score
		}
	}
	return maxScore
}

func maxModuleScore(values []moduleScore) float64 {
	var maxScore float64
	for _, value := range values {
		if value.score > maxScore {
			maxScore = value.score
		}
	}
	return maxScore
}

func maxFileScore(values []fileScore) float64 {
	var maxScore float64
	for _, value := range values {
		if value.score > maxScore {
			maxScore = value.score
		}
	}
	return maxScore
}

func isSynonymOnly(token string, textTokens map[string]bool) bool {
	if textTokens[token] {
		return false
	}
	for root, words := range synonymMap {
		if token == root && containsString(words, token) {
			return false
		}
	}
	return false
}
