package analyzers

import (
	"strings"
	"testing"
)

func TestFindAliasCandidatesDetectsConservativeTokenDrift(t *testing.T) {
	candidates := FindAliasCandidates("OPENAI_API_KEY", map[string]struct{}{"OPENAI_KEY": {}})

	if len(candidates) != 1 {
		t.Fatalf("candidate count = %d, want 1", len(candidates))
	}
	if candidates[0].CandidateName != "OPENAI_KEY" {
		t.Fatalf("candidate = %s, want OPENAI_KEY", candidates[0].CandidateName)
	}
	if candidates[0].Reason != "Token overlap 0.67 and name similarity 0.83 suggest drift." {
		t.Fatalf("reason = %q", candidates[0].Reason)
	}
}

func TestFindAliasCandidatesSkipsExactNamesAndReportsCanonicalMatches(t *testing.T) {
	candidates := FindAliasCandidates(
		"DATABASE_URL",
		map[string]struct{}{"DATABASE_URL": {}, "DB_URL": {}},
	)

	if len(candidates) != 1 {
		t.Fatalf("candidate count = %d, want 1 (exact name skipped)", len(candidates))
	}
	if candidates[0].CandidateName != "DB_URL" {
		t.Fatalf("candidate = %s, want DB_URL", candidates[0].CandidateName)
	}
	if candidates[0].Score != 0.99 {
		t.Fatalf("score = %v, want 0.99 for a canonical match", candidates[0].Score)
	}
	if !strings.Contains(candidates[0].Reason, "Canonical token expansion matches") {
		t.Fatalf("reason = %q, want canonical-match phrasing", candidates[0].Reason)
	}
}

func TestCanonicalTokensExpandPostgresShorthand(t *testing.T) {
	tokens := canonicalTokens("PGHOST")
	if len(tokens) != 2 || tokens[0] != "POSTGRES" || tokens[1] != "HOST" {
		t.Fatalf("tokens = %v, want POSTGRES HOST", tokens)
	}
}
