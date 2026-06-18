package analyzers

import "testing"

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

func TestCanonicalTokensExpandPostgresShorthand(t *testing.T) {
	tokens := canonicalTokens("PGHOST")
	if len(tokens) != 2 || tokens[0] != "POSTGRES" || tokens[1] != "HOST" {
		t.Fatalf("tokens = %v, want POSTGRES HOST", tokens)
	}
}
