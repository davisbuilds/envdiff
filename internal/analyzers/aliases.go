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

// AliasIndex precomputes the canonical token decomposition of a set of defined
// names plus an inverted token index. Built once, it answers Candidates for
// many missing names without re-deriving every defined name and without
// comparing against names that share no tokens — those can clear neither the
// canonical-equality nor the token-overlap thresholds, so skipping them yields
// identical results while avoiding the O(missing x defined) comparison blowup.
type aliasEntry struct {
	name      string
	canonical string
	tokenSet  map[string]struct{}
}

type AliasIndex struct {
	entries []aliasEntry
	byToken map[string][]int
}

func BuildAliasIndex(definedNames map[string]struct{}) *AliasIndex {
	names := make([]string, 0, len(definedNames))
	for name := range definedNames {
		names = append(names, name)
	}
	sort.Strings(names)

	index := &AliasIndex{byToken: map[string][]int{}}
	for _, name := range names {
		tokens := canonicalTokens(name)
		set := tokenSet(tokens)
		position := len(index.entries)
		index.entries = append(index.entries, aliasEntry{
			name:      name,
			canonical: strings.Join(tokens, "_"),
			tokenSet:  set,
		})
		for token := range set {
			index.byToken[token] = append(index.byToken[token], position)
		}
	}
	return index
}

func (index *AliasIndex) Candidates(missingName string) []AliasCandidate {
	threshold := 0.8
	missingTokens := canonicalTokens(missingName)
	missingSet := tokenSet(missingTokens)
	missingCanonical := strings.Join(missingTokens, "_")

	// Only defined names sharing at least one token can match; gather them once.
	considered := map[int]struct{}{}
	candidates := []AliasCandidate{}
	for token := range missingSet {
		for _, position := range index.byToken[token] {
			if _, ok := considered[position]; ok {
				continue
			}
			considered[position] = struct{}{}

			entry := index.entries[position]
			if entry.name == missingName {
				continue
			}

			if missingCanonical == entry.canonical {
				candidates = append(candidates, AliasCandidate{
					CandidateName: entry.name,
					Score:         0.99,
					Reason: fmt.Sprintf(
						"Canonical token expansion matches: %s == %s.",
						missingCanonical,
						entry.canonical,
					),
				})
				continue
			}

			tokenOverlap := jaccard(missingSet, entry.tokenSet)
			ratio := sequenceRatio(missingCanonical, entry.canonical)
			if tokenOverlap >= 2.0/3.0 && ratio >= threshold {
				candidates = append(candidates, AliasCandidate{
					CandidateName: entry.name,
					Score:         math.Round(ratio*100) / 100,
					Reason: fmt.Sprintf(
						"Token overlap %.2f and name similarity %.2f suggest drift.",
						tokenOverlap,
						ratio,
					),
				})
			}
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

// FindAliasCandidates builds a one-shot index for definedNames and queries it.
// Prefer reusing a BuildAliasIndex across many missing names.
func FindAliasCandidates(missingName string, definedNames map[string]struct{}) []AliasCandidate {
	return BuildAliasIndex(definedNames).Candidates(missingName)
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

func tokenSet(tokens []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		set[token] = struct{}{}
	}
	return set
}

func jaccard(left map[string]struct{}, right map[string]struct{}) float64 {
	if len(left) == 0 && len(right) == 0 {
		return 0
	}
	intersection := 0
	for value := range left {
		if _, ok := right[value]; ok {
			intersection++
		}
	}
	union := len(left) + len(right) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
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
