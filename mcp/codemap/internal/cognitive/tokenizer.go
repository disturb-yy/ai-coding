package cognitive

import (
	"regexp"
	"strings"
)

var tokenPattern = regexp.MustCompile(`[a-zA-Z0-9]+`)

var synonymMap = map[string][]string{
	"cancel": {"cancel", "cancellation", "cancelled", "close", "terminate"},
	"create": {"create", "add", "new", "insert"},
	"delete": {"delete", "remove"},
	"update": {"update", "modify", "edit", "change"},
	"login":  {"login", "signin", "authenticate", "auth"},
	"refund": {"refund", "rollback", "return"},
}

var stopTokens = map[string]bool{
	"a": true, "an": true, "and": true, "for": true, "of": true, "the": true, "to": true,
	"with": true, "functionality": true, "feature": true, "support": true,
}

func TokenizeRequirement(requirement string) []string {
	seen := make(map[string]bool)
	var tokens []string
	for _, raw := range tokenPattern.FindAllString(strings.ToLower(requirement), -1) {
		token := normalizeToken(raw)
		if token == "" || stopTokens[token] || seen[token] {
			continue
		}
		seen[token] = true
		tokens = append(tokens, token)
	}
	return tokens
}

func expandTokens(tokens []string) map[string]bool {
	expanded := make(map[string]bool)
	for _, token := range tokens {
		expanded[token] = true
		for root, words := range synonymMap {
			if token == root || containsString(words, token) {
				expanded[root] = true
				for _, word := range words {
					expanded[normalizeToken(word)] = true
				}
			}
		}
	}
	return expanded
}

func normalizeToken(token string) string {
	token = strings.TrimSpace(strings.ToLower(token))
	if token == "" {
		return ""
	}
	for root, words := range synonymMap {
		if token == root || containsString(words, token) {
			return root
		}
	}
	switch {
	case strings.HasSuffix(token, "ies") && len(token) > 3:
		return normalizeStem(token[:len(token)-3] + "y")
	case strings.HasSuffix(token, "ing") && len(token) > 4:
		return normalizeStem(token[:len(token)-3])
	case strings.HasSuffix(token, "ed") && len(token) > 3:
		return normalizeStem(token[:len(token)-2])
	case strings.HasSuffix(token, "s") && len(token) > 3:
		return normalizeStem(token[:len(token)-1])
	}
	return token
}

func normalizeStem(stem string) string {
	for root, words := range synonymMap {
		if stem == root || stem+"e" == root || containsString(words, stem) || containsString(words, stem+"e") {
			return root
		}
	}
	return stem
}

func ClassifyRequirement(tokens []string) RequirementType {
	tokenSet := make(map[string]bool)
	for _, token := range tokens {
		tokenSet[token] = true
	}
	if tokenSet["auth"] || tokenSet["login"] || tokenSet["token"] || tokenSet["permission"] || tokenSet["session"] {
		return ReqAddAuth
	}
	if tokenSet["field"] || tokenSet["column"] || tokenSet["request"] || tokenSet["response"] || tokenSet["param"] {
		return ReqAddField
	}
	if tokenSet["fix"] || tokenSet["bug"] || tokenSet["error"] || tokenSet["failed"] || tokenSet["wrong"] {
		return ReqFixBug
	}
	if tokenSet["api"] || tokenSet["endpoint"] || tokenSet["route"] {
		return ReqAddAPI
	}
	if tokenSet["change"] || tokenSet["modify"] || tokenSet["update"] || tokenSet["logic"] || tokenSet["rule"] {
		return ReqModifyLogic
	}
	return ReqUnknown
}
