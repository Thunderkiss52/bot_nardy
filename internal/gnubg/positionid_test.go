package gnubg

import (
	"strings"
	"testing"

	"bot_nardy/internal/engine"
)

func TestEncodePositionIDStartingPosition(t *testing.T) {
	state := engine.NewShortGame(1)
	got := EncodePositionID(state)
	want := "4HPwATDgc/ABMA"
	if got != want {
		t.Fatalf("unexpected position id: got %q want %q", got, want)
	}
}

func TestParseHintOutput(t *testing.T) {
	output := `
Win    W(g)   W(bg)  L(g)   L(bg)  Equity       Move
0.542  0.142  0.008  0.113  0.008  (+0.114)     6/5 8/5
0.505  0.120  0.008  0.122  0.007  (+0.009)     24/23 23/20
`
	hints, err := ParseHintOutput(output)
	if err != nil {
		t.Fatalf("ParseHintOutput failed: %v", err)
	}
	if len(hints) != 2 {
		t.Fatalf("expected 2 hints, got %d", len(hints))
	}
	if hints[0].Notation != "6/5 8/5" {
		t.Fatalf("unexpected notation: %q", hints[0].Notation)
	}
	if hints[0].WinProb != 0.542 {
		t.Fatalf("unexpected winprob: %v", hints[0].WinProb)
	}
}

func TestMatchHintToLegalWhiteAndBlackPerspective(t *testing.T) {
	state := engine.NewShortGame(1)
	legal, err := engine.GenerateLegalLines(state, 3, 1)
	if err != nil {
		t.Fatalf("GenerateLegalLines failed: %v", err)
	}
	line, ok := MatchHintToLegal(state, legal, "8/5 6/5")
	if !ok {
		t.Fatal("expected to match white notation")
	}
	if !strings.Contains(FormatNotation(state, line), "8/5") {
		t.Fatalf("unexpected formatted notation: %q", FormatNotation(state, line))
	}

	black := engine.NewShortGame(1)
	black.Turn = engine.Black
	legal, err = engine.GenerateLegalLines(black, 6, 2)
	if err != nil {
		t.Fatalf("GenerateLegalLines failed: %v", err)
	}
	line, ok = MatchHintToLegal(black, legal, "24/18 13/11")
	if !ok {
		t.Fatal("expected to match black notation")
	}
	if FormatNotation(black, line) != "24/18 13/11" && FormatNotation(black, line) != "13/11 24/18" {
		t.Fatalf("unexpected black notation: %q", FormatNotation(black, line))
	}
}
