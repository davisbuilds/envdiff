package analyzers

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

var tokenExpansions = map[string][]string{
	"DB": {"DATABASE"},
	"PG": {"POSTGRES"},
}

var postgresSuffixes = map[string]struct{}{
	"HOST":     {},
	"PORT":     {},
	"USER":     {},
	"PASSWORD": {},
	"DATABASE": {},
	"DBNAME":   {},
}

type AliasCandidate struct {
	CandidateName string
	Score         float64
	Reason        string
}

func FindAliasCandidates(missingName string, definedNames map[string]struct{}) []AliasCandidate {
	threshold := 0.8
	candidates := []AliasCandidate{}
	missingCanonical := CanonicalName(missingName)
	missingTokens := canonicalTokens(missingName)

	names := make([]string, 0, len(definedNames))
	for name := range definedNames {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, definedName := range names {
		if definedName == missingName {
			continue
		}

		definedCanonical := CanonicalName(definedName)
		definedTokens := canonicalTokens(definedName)
		tokenOverlap := jaccard(missingTokens, definedTokens)
		sequenceRatio := sequenceRatio(missingCanonical, definedCanonical)

		if missingCanonical == definedCanonical {
			candidates = append(candidates, AliasCandidate{
				CandidateName: definedName,
				Score:         0.99,
				Reason: fmt.Sprintf(
					"Canonical token expansion matches: %s == %s.",
					missingCanonical,
					definedCanonical,
				),
			})
			continue
		}

		if tokenOverlap >= 2.0/3.0 && sequenceRatio >= threshold {
			candidates = append(candidates, AliasCandidate{
				CandidateName: definedName,
				Score:         math.Round(sequenceRatio*100) / 100,
				Reason: fmt.Sprintf(
					"Token overlap %.2f and name similarity %.2f suggest drift.",
					tokenOverlap,
					sequenceRatio,
				),
			})
		}
	}

	sort.SliceStable(candidates, func(left int, right int) bool {
		if candidates[left].Score != candidates[right].Score {
			return candidates[left].Score > candidates[right].Score
		}
		return candidates[left].CandidateName < candidates[right].CandidateName
	})
	return candidates
}

func CanonicalName(name string) string {
	return strings.Join(canonicalTokens(name), "_")
}

func canonicalTokens(name string) []string {
	upperName := strings.ToUpper(name)
	if strings.HasPrefix(upperName, "PG") {
		suffix := upperName[2:]
		if _, ok := postgresSuffixes[suffix]; ok {
			if suffix == "DBNAME" {
				return []string{"POSTGRES", "DATABASE"}
			}
			return []string{"POSTGRES", suffix}
		}
	}

	rawTokens := strings.Split(upperName, "_")
	tokens := []string{}
	for _, token := range rawTokens {
		if expansion, ok := tokenExpansions[token]; ok {
			tokens = append(tokens, expansion...)
		} else {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func jaccard(left []string, right []string) float64 {
	leftSet := map[string]struct{}{}
	rightSet := map[string]struct{}{}
	union := map[string]struct{}{}
	for _, value := range left {
		leftSet[value] = struct{}{}
		union[value] = struct{}{}
	}
	for _, value := range right {
		rightSet[value] = struct{}{}
		union[value] = struct{}{}
	}
	if len(union) == 0 {
		return 0
	}
	intersection := 0
	for value := range leftSet {
		if _, ok := rightSet[value]; ok {
			intersection++
		}
	}
	return float64(intersection) / float64(len(union))
}

func sequenceRatio(left string, right string) float64 {
	if len(left)+len(right) == 0 {
		return 1
	}
	lcs := longestCommonSubsequence(left, right)
	return 2 * float64(lcs) / float64(len(left)+len(right))
}

func longestCommonSubsequence(left string, right string) int {
	leftRunes := []rune(left)
	rightRunes := []rune(right)
	previous := make([]int, len(rightRunes)+1)
	for _, leftRune := range leftRunes {
		current := make([]int, len(rightRunes)+1)
		for rightIndex, rightRune := range rightRunes {
			if leftRune == rightRune {
				current[rightIndex+1] = previous[rightIndex] + 1
			} else if previous[rightIndex+1] > current[rightIndex] {
				current[rightIndex+1] = previous[rightIndex+1]
			} else {
				current[rightIndex+1] = current[rightIndex]
			}
		}
		previous = current
	}
	return previous[len(rightRunes)]
}
