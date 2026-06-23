package cognitive

import (
	"fmt"
	"sort"
	"strings"

	"github.com/disturb-yy/codemap/internal/model"
)

var riskModuleMap = map[string]string{
	"auth":       "security",
	"payment":    "payment",
	"order":      "business_state",
	"inventory":  "consistency",
	"user":       "user_data",
	"permission": "access_control",
	"token":      "security",
	"session":    "security",
}

var sensitiveRouteParts = []string{"/login", "/pay", "/refund", "/delete", "/admin"}

func inferRiskAreas(modules []moduleScore, routes []*model.Route, dependencyCounts, dependentCounts, callerCounts map[string]int) []RiskArea {
	risks := make(map[string]RiskArea)
	add := func(name, level, reason string) {
		if name == "" {
			return
		}
		existing, ok := risks[name]
		if !ok || riskRank(level) > riskRank(existing.Level) {
			risks[name] = RiskArea{Name: name, Level: level, Reason: reason}
		}
	}

	for _, module := range modules {
		lower := strings.ToLower(module.module + " " + module.path)
		for key, riskName := range riskModuleMap {
			if strings.Contains(lower, key) {
				add(riskName, "high", fmt.Sprintf("Critical module %q is related to %s", module.module, riskName))
			}
		}
		if dependencyCounts[module.module] > 5 {
			add(module.module+"_dependencies", "medium", "Module has more than 5 outgoing dependencies")
		}
		if dependentCounts[module.module] > 5 {
			add(module.module+"_impact", "high", "Module has more than 5 dependent modules")
		}
		if callerCounts[module.module] > 3 {
			add(module.module+"_callers", "high", "Module has more than 3 callers in call graph")
		}
	}

	for _, route := range routes {
		path := strings.ToLower(route.Path)
		for _, part := range sensitiveRouteParts {
			if strings.Contains(path, part) {
				add("sensitive_route", "high", fmt.Sprintf("Route %s %s touches sensitive endpoint %s", route.Method, route.Path, part))
			}
		}
	}

	result := make([]RiskArea, 0, len(risks))
	for _, risk := range risks {
		result = append(result, risk)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if riskRank(result[i].Level) == riskRank(result[j].Level) {
			return result[i].Name < result[j].Name
		}
		return riskRank(result[i].Level) > riskRank(result[j].Level)
	})
	return result
}

func riskRank(level string) int {
	switch level {
	case "high":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}
