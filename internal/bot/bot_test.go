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
