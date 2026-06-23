package cognitive

import (
	"fmt"
	"math"
	"path"
	"sort"
	"strings"

	"github.com/disturb-yy/codemap/internal/model"
)

type featureMatch struct {
	feature model.FeatureEntry
	score   float64
	reasons []string
}

type moduleScore struct {
	module string
	path   string
	score  float64
	reason reasonSet
}

type fileScore struct {
	file   string
	module string
	score  float64
	reason reasonSet
}

type reasonSet map[string]bool

func (r reasonSet) add(reason string) {
	if reason != "" {
		r[reason] = true
	}
}

func (r reasonSet) string() string {
	if len(r) == 0 {
		return "Related to requirement terms"
	}
	var reasons []string
	for reason := range r {
		reasons = append(reasons, reason)
	}
	sort.Strings(reasons)
	return strings.Join(reasons, "; ")
}

func matchFeatures(tokens []string, features []model.FeatureEntry) []featureMatch {
	expanded := expandTokens(tokens)
	var matches []featureMatch
	for _, feat := range features {
		var score float64
		var reasons []string
		score += scoreText(expanded, feat.Name, 5, 2, "feature name", &reasons)
		score += scoreText(expanded, feat.Description, 3, 2, "feature description", &reasons)
		for _, module := range feat.Modules {
			score += scoreText(expanded, module, 4, 2, "module", &reasons)
		}
		for _, route := range feat.Routes {
			score += scoreText(expanded, route, 3, 2, "route", &reasons)
		}
		for _, flow := range feat.Flows {
			score += scoreText(expanded, flow, 4, 2, "flow", &reasons)
		}
		if score > 0 {
			matches = append(matches, featureMatch{feature: feat, score: score, reasons: uniqueStrings(reasons)})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score == matches[j].score {
			return matches[i].feature.Name < matches[j].feature.Name
		}
		return matches[i].score > matches[j].score
	})
	return matches
}

func toMatchedFeatures(matches []featureMatch, limit int) []MatchedFeature {
	limited := limitFeatureMatches(matches, limit)
	maxScore := maxFeatureScore(limited)
	result := make([]MatchedFeature, 0, len(limited))
	for _, match := range limited {
		result = append(result, MatchedFeature{
			Name:   match.feature.Name,
			Score:  normalizeScore(match.score, maxScore),
			Reason: buildMatchReason(match.reasons),
		})
	}
	return result
}

func scoreModules(tokens []string, matches []featureMatch, hints []model.NavHintEntry, modules []*model.Module, routes []*model.Route, flows []*model.Flow, callers map[string]int, callees map[string]int) []moduleScore {
	expanded := expandTokens(tokens)
	moduleByKey := make(map[string]*model.Module)
	for _, module := range modules {
		moduleByKey[module.Name] = module
		moduleByKey[module.Path] = module
	}

	scores := make(map[string]*moduleScore)
	ensure := func(name string) *moduleScore {
		if name == "" {
			name = "unknown"
		}
		if mod, ok := moduleByKey[name]; ok {
			name = mod.Name
		}
		if scores[name] == nil {
			pathValue := name
			if mod, ok := moduleByKey[name]; ok {
				pathValue = mod.Path
			}
			scores[name] = &moduleScore{module: name, path: pathValue, reason: make(reasonSet)}
		}
		return scores[name]
	}

	for _, module := range modules {
		if score := scoreText(expanded, module.Name+" "+module.Path, 10, 2, "", nil); score > 0 {
			ms := ensure(module.Name)
			ms.score += score
			ms.reason.add("Direct requirement match")
		}
		for _, dep := range module.Dependencies {
			if scoreText(expanded, dep, 10, 2, "", nil) > 0 {
				ms := ensure(module.Name)
				ms.score += 5
				ms.reason.add("Dependency neighbor matches requirement")
			}
		}
	}

	for _, match := range matches {
		for _, module := range match.feature.Modules {
			ms := ensure(module)
			ms.score += 8 + math.Min(match.score/10, 4)
			ms.reason.add("Matched feature")
		}
	}

	hintsByFeature := map[string]model.NavHintEntry{}
	for _, hint := range hints {
		hintsByFeature[hint.Feature] = hint
	}
	for _, match := range matches {
		hint, ok := hintsByFeature[match.feature.Name]
		if !ok {
			continue
		}
		for _, module := range hint.RelatedModules {
			ms := ensure(module)
			ms.score += 5
			ms.reason.add("Navigation hint related module")
		}
	}

	for _, route := range routes {
		ms := ensure(route.Module)
		ms.score += 7
		ms.reason.add("Route owner")
	}
	for _, flow := range flows {
		ms := ensure(flow.Trigger)
		ms.score += 7
		ms.reason.add("Flow owner")
	}
	for module, count := range callers {
		ms := ensure(module)
		ms.score += math.Min(float64(count), 4)
		ms.reason.add("Impact analysis caller")
	}
	for module, count := range callees {
		ms := ensure(module)
		ms.score += math.Min(float64(count), 4)
		ms.reason.add("Call graph neighbor")
	}

	result := make([]moduleScore, 0, len(scores))
	for _, score := range scores {
		if score.score > 0 {
			result = append(result, *score)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].score == result[j].score {
			return result[i].module < result[j].module
		}
		return result[i].score > result[j].score
	})
	return result
}

func toCandidateModules(scores []moduleScore, limit int) []CandidateModule {
	limited := limitModuleScores(scores, limit)
	maxScore := maxModuleScore(limited)
	result := make([]CandidateModule, 0, len(limited))
	for _, score := range limited {
		result = append(result, CandidateModule{
			Module: score.module,
			Score:  normalizeScore(score.score, maxScore),
			Reason: score.reason.string(),
		})
	}
	return result
}

func scoreFiles(tokens []string, reqType RequirementType, modules []moduleScore, hints []model.NavHintEntry, routes []*model.Route, flows []*model.Flow) []fileScore {
	files := make(map[string]*fileScore)
	ensure := func(file, module string) *fileScore {
		file = strings.TrimSpace(file)
		if file == "" {
			return nil
		}
		if files[file] == nil {
			files[file] = &fileScore{file: file, module: module, reason: make(reasonSet)}
		}
		if files[file].module == "" {
			files[file].module = module
		}
		return files[file]
	}

	moduleScoreByName := make(map[string]float64)
	for _, module := range modules {
		moduleScoreByName[module.module] = module.score
	}

	for _, hint := range hints {
		hintScore := relatedHintScore(hint, moduleScoreByName)
		if hintScore == 0 {
			continue
		}
		module := firstString(hint.RelatedModules)
		for _, file := range hint.StartFiles {
			if fs := ensure(file, module); fs != nil {
				fs.score += hintScore + 9
				fs.reason.add("Navigation start file")
			}
		}
	}

	for _, route := range routes {
		file := routeToFile(route)
		if fs := ensure(file, route.Module); fs != nil {
			fs.score += 12
			fs.reason.add("Route handler entry point")
		}
	}

	for _, flow := range flows {
		file := flowToFile(flow)
		if fs := ensure(file, flow.Trigger); fs != nil {
			fs.score += 10
			fs.reason.add("Flow entry point")
		}
	}

	for _, module := range modules {
		base := math.Min(module.score/2, 10)
		for _, file := range suggestedFilesForModule(module.path, reqType) {
			if fs := ensure(file, module.module); fs != nil {
				fs.score += base + filePriority(file, reqType)
				fs.reason.add(fmt.Sprintf("%s layer candidate", reqType))
			}
		}
	}

	expanded := expandTokens(tokens)
	for _, fs := range files {
		if score := scoreText(expanded, fs.file, 5, 2, "", nil); score > 0 {
			fs.score += score
			fs.reason.add("File path matches requirement")
		}
	}

	result := make([]fileScore, 0, len(files))
	for _, score := range files {
		if score.score > 0 {
			result = append(result, *score)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].score == result[j].score {
			return result[i].file < result[j].file
		}
		return result[i].score > result[j].score
	})
	return result
}

func toCandidateFiles(scores []fileScore, limit int) []CandidateFile {
	limited := limitFileScores(scores, limit)
	maxScore := maxFileScore(limited)
	result := make([]CandidateFile, 0, len(limited))
	for _, score := range limited {
		result = append(result, CandidateFile{
			File:   score.file,
			Module: score.module,
			Score:  normalizeScore(score.score, maxScore),
			Reason: score.reason.string(),
		})
	}
	return result
}

func scoreText(tokens map[string]bool, text string, directWeight, synonymWeight float64, label string, reasons *[]string) float64 {
	var score float64
	textTokens := tokenSet(text)
	for token := range tokens {
		if textTokens[token] {
			if isSynonymOnly(token, textTokens) {
				score += synonymWeight
			} else {
				score += directWeight
			}
			if reasons != nil && label != "" {
				*reasons = append(*reasons, label)
			}
		}
	}
	return score
}

func tokenSet(text string) map[string]bool {
	set := make(map[string]bool)
	for _, token := range TokenizeRequirement(text) {
		set[token] = true
	}
	for _, part := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return r == '_' || r == '-' || r == '/' || r == '.' || r == ':'
	}) {
		if token := normalizeToken(part); token != "" {
			set[token] = true
		}
	}
	return set
}

func routeToFile(route *model.Route) string {
	if route.Handler != "" && strings.Contains(route.Handler, "/") {
		return route.Handler
	}
	if route.Module != "" {
		return path.Join(route.Module, "handler.go")
	}
	return strings.TrimPrefix(strings.ReplaceAll(route.Path, "/", "_"), "_")
}

func flowToFile(flow *model.Flow) string {
	if flow.Trigger == "" {
		return flow.Name
	}
	return path.Join(flow.Trigger, "service.go")
}

func suggestedFilesForModule(modulePath string, reqType RequirementType) []string {
	if modulePath == "" {
		return nil
	}
	var names []string
	switch reqType {
	case ReqAddAPI:
		names = []string{"router.go", "controller.go", "handler.go", "service.go", "dto.go"}
	case ReqModifyLogic:
		names = []string{"service.go", "usecase.go", "domain.go", "biz.go"}
	case ReqFixBug:
		names = []string{"handler.go", "service.go"}
	case ReqAddField:
		names = []string{"model.go", "entity.go", "dto.go", "repository.go", "migration.go"}
	case ReqAddAuth:
		names = []string{"middleware.go", "auth.go", "service.go", "controller.go"}
	default:
		names = []string{"service.go", "handler.go", "repository.go"}
	}
	files := make([]string, 0, len(names))
	for _, name := range names {
		files = append(files, path.Join(modulePath, name))
	}
	return files
}

func filePriority(file string, reqType RequirementType) float64 {
	file = strings.ToLower(file)
	priority := map[RequirementType][]string{
		ReqAddAPI:      {"router", "controller", "handler", "service", "dto"},
		ReqModifyLogic: {"service", "usecase", "domain", "biz"},
		ReqFixBug:      {"handler", "service"},
		ReqAddField:    {"model", "entity", "dto", "repository", "migration"},
		ReqAddAuth:     {"middleware", "auth", "service", "controller"},
	}
	for i, part := range priority[reqType] {
		if strings.Contains(file, part) {
			return float64(len(priority[reqType])-i) * 2
		}
	}
	return 1
}

func relatedHintScore(hint model.NavHintEntry, modules map[string]float64) float64 {
	var score float64
	for _, module := range hint.RelatedModules {
		if modules[module] > score {
			score = modules[module]
		}
	}
	return math.Min(score/2, 10)
}
