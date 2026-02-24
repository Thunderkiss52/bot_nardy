package engine

import "testing"

func TestShortBarMandatory(t *testing.T) {
	s := GameState{GameType: GameShort, Turn: White}
	s.Bar[White.Idx()] = 1
	setPoint(&s, 24, White, 14)
	setPoint(&s, 19, Black, 2)
	setPoint(&s, 1, Black, 13)

	lines, err := GenerateLegalLines(s, 6, 1)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	for _, line := range lines {
		if len(line.Moves) == 0 {
			t.Fatalf("expected at least one move from bar")
		}
		if line.Moves[0].From != 0 {
			t.Fatalf("first move must enter from bar, got %+v", line.Moves[0])
		}
	}
}

func TestLongHeadRule(t *testing.T) {
	s := NewLongGame(1)
	lines, err := GenerateLegalLines(s, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range lines {
		fromHead := 0
		for _, mv := range line.Moves {
			if mv.From == 24 {
				fromHead++
			}
		}
		if fromHead > 1 {
			t.Fatalf("head rule violated: %s", line)
		}
	}
}

func TestMaxDiceUsageHigherDieWhenOnlyOne(t *testing.T) {
	s := GameState{GameType: GameLong, Turn: White}
	setPoint(&s, 2, White, 1)
	s.Off[White.Idx()] = 14
	setPoint(&s, 1, Black, 15)

	lines, err := GenerateLegalLines(s, 6, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 {
		t.Fatal("expected legal lines")
	}
	for _, line := range lines {
		if line.DiceUsed() != 1 {
			t.Fatalf("expected one die, got %d", line.DiceUsed())
		}
		if line.Moves[0].Die != 6 {
			t.Fatalf("expected higher die 6, got %d", line.Moves[0].Die)
		}
	}
}
