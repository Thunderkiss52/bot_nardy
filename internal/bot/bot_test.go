package bot

import (
	"testing"
	"time"

	"bot_nardy/internal/engine"
)

func TestChooseMoveReturnsLegal(t *testing.T) {
	s := engine.NewShortGame(1)
	b := New(Config{ThinkTime: 120 * time.Millisecond, TopK: 4, Workers: 2, Seed: 42})
	dec, err := b.ChooseMove(s, 3, 1)
	if err != nil {
		t.Fatal(err)
	}
	if dec.LegalCount == 0 {
		t.Fatal("expected legal move")
	}
	if !engine.IsLegalLine(s, 3, 1, dec.ChosenLine) {
		t.Fatalf("bot chose illegal line: %s", dec.ChosenLine)
	}
}

func TestWeightedDiceOutcomesCover21Rolls(t *testing.T) {
	outcomes := weightedDiceOutcomes()
	if len(outcomes) != 21 {
		t.Fatalf("expected 21 unique weighted rolls, got %d", len(outcomes))
	}

	totalWeight := 0.0
	for _, outcome := range outcomes {
		totalWeight += outcome.weight
	}
	if totalWeight != 36 {
		t.Fatalf("expected weight 36, got %.0f", totalWeight)
	}
}

func TestExpectedReplyEquityStableAcrossWorkerCounts(t *testing.T) {
	state := engine.NewShortGame(11)

	single := New(Config{ThinkTime: 150 * time.Millisecond, TopK: 4, Workers: 1, Seed: 101})
	parallel := New(Config{ThinkTime: 150 * time.Millisecond, TopK: 4, Workers: 4, Seed: 101})

	gotSingle := single.expectedReplyEquity(state, engine.White)
	gotParallel := parallel.expectedReplyEquity(state, engine.White)

	if gotSingle != gotParallel {
		t.Fatalf("expected same equity, single=%.6f parallel=%.6f", gotSingle, gotParallel)
	}
}

func TestPrepareCandidatesStableAcrossWorkerCounts(t *testing.T) {
	state := engine.NewShortGame(21)
	legal, err := engine.GenerateLegalLines(state, 3, 1)
	if err != nil {
		t.Fatalf("generate legal lines: %v", err)
	}
	if len(legal) < 2 {
		t.Fatalf("expected multiple legal lines, got %d", len(legal))
	}

	single := New(Config{ThinkTime: 150 * time.Millisecond, TopK: 8, Workers: 1, Seed: 201})
	parallel := New(Config{ThinkTime: 150 * time.Millisecond, TopK: 8, Workers: 4, Seed: 201})

	singleCands := single.prepareCandidates(state, legal)
	parallelCands := parallel.prepareCandidates(state, legal)
	if len(singleCands) == 0 || len(parallelCands) == 0 {
		t.Fatalf("expected non-empty candidates")
	}

	if singleCands[0].line.Key() != parallelCands[0].line.Key() {
		t.Fatalf("expected same top candidate, single=%s parallel=%s", singleCands[0].line, parallelCands[0].line)
	}
}
