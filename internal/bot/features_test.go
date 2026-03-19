package bot

import (
	"testing"

	"bot_nardy/internal/engine"
)

func TestExtractFeaturesHasStableCoreMetrics(t *testing.T) {
	s := engine.NewShortGame(1)
	f := ExtractFeatures(s, engine.White)
	if f.GameType != "short" {
		t.Fatalf("expected short game type, got %s", f.GameType)
	}
	if f.Perspective != "white" {
		t.Fatalf("expected white perspective, got %s", f.Perspective)
	}
	if f.PipSelf <= 0 || f.PipOpp <= 0 {
		t.Fatalf("expected positive pip counts, got self=%d opp=%d", f.PipSelf, f.PipOpp)
	}
	values := f.Values()
	if values["pip_self"] != float64(f.PipSelf) {
		t.Fatalf("expected pip_self value to match feature")
	}
	if _, ok := values["phase_contact"]; !ok {
		t.Fatalf("expected phase_contact derived feature")
	}
	if _, ok := values["pip_adv"]; !ok {
		t.Fatalf("expected pip_adv derived feature")
	}
	if _, ok := values["contact_prime_adv"]; !ok {
		t.Fatalf("expected contact interaction feature")
	}
}

func TestLinearEvaluatorUsesFeatureWeights(t *testing.T) {
	s := engine.NewShortGame(1)
	ev := NewLinearEvaluator("test", 3, map[string]float64{
		"pip_self": -1,
		"pip_opp":  1,
	}, nil, nil)
	score := ev.Evaluate(s, engine.White)
	if score == 0 {
		t.Fatalf("expected non-zero score")
	}
}
