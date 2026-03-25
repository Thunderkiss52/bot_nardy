package app

import (
	"strings"
	"testing"

	"bot_nardy/internal/engine"
)

func TestIsBotTurnSwitchesAfterHumanMove(t *testing.T) {
	svc, err := NewService("")
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Close()

	_, err = svc.Start(GameOptions{
		GameType:     engine.GameLong,
		BotSide:      engine.Black,
		Opponent:     OpponentHuman,
		ThinkTimeSec: 1,
		Seed:         1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if svc.IsBotTurn() {
		t.Fatal("white turn should be human")
	}

	_, _, err = svc.ApplyHumanLineByIndex(1, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !svc.IsBotTurn() {
		t.Fatal("after white move, black bot turn expected")
	}
}

func TestPassTurnIfNoLegal(t *testing.T) {
	svc, err := NewService("")
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Close()

	svc.options = GameOptions{BotSide: engine.Black, Opponent: OpponentHuman}
	svc.state = engine.GameState{GameType: engine.GameShort, Turn: engine.White}
	svc.state.Bar[engine.White.Idx()] = 1
	svc.state.Off[engine.White.Idx()] = 14

	svc.state.Points[24] = engine.Point{Owner: engine.Black, Count: 2}
	svc.state.Points[23] = engine.Point{Owner: engine.Black, Count: 2}
	svc.state.Points[22] = engine.Point{Owner: engine.Black, Count: 2}
	svc.state.Points[21] = engine.Point{Owner: engine.Black, Count: 2}
	svc.state.Points[20] = engine.Point{Owner: engine.Black, Count: 2}
	svc.state.Points[19] = engine.Point{Owner: engine.Black, Count: 2}
	svc.state.Points[1] = engine.Point{Owner: engine.Black, Count: 3}
	svc.history = []engine.GameState{svc.state}

	next, err := svc.PassTurnIfNoLegal(1, 6)
	if err != nil {
		t.Fatal(err)
	}
	if next.Turn != engine.Black {
		t.Fatalf("expected turn to pass to black")
	}
	if next.Meta.MoveNumber != 1 {
		t.Fatalf("expected move number increment")
	}
}

func TestSwapBotSideChangesTurnOwnership(t *testing.T) {
	svc, err := NewService("")
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Close()

	_, err = svc.Start(GameOptions{
		GameType:     engine.GameShort,
		BotSide:      engine.Black,
		Opponent:     OpponentHuman,
		ThinkTimeSec: 1,
		Seed:         1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if svc.IsBotTurn() {
		t.Fatal("white turn should start as human")
	}

	_, side, err := svc.SwapBotSide()
	if err != nil {
		t.Fatal(err)
	}
	if side != engine.White {
		t.Fatalf("expected bot side to become white, got %v", side)
	}
	if !svc.IsBotTurn() {
		t.Fatal("white turn should belong to bot after swap")
	}
}

func TestEditCheckerMovesCheckerAndUndo(t *testing.T) {
	svc, err := NewService("")
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Close()

	_, err = svc.Start(GameOptions{
		GameType:     engine.GameShort,
		BotSide:      engine.Black,
		Opponent:     OpponentHuman,
		ThinkTimeSec: 1,
		Seed:         1,
	})
	if err != nil {
		t.Fatal(err)
	}

	next, err := svc.EditChecker(24, 23, engine.NoColor)
	if err != nil {
		t.Fatal(err)
	}
	if next.Points[24].Count != 1 || next.Points[24].Owner != engine.White {
		t.Fatalf("expected one white checker left on 24, got %+v", next.Points[24])
	}
	if next.Points[23].Count != 1 || next.Points[23].Owner != engine.White {
		t.Fatalf("expected one white checker on 23, got %+v", next.Points[23])
	}

	undone, err := svc.Undo()
	if err != nil {
		t.Fatal(err)
	}
	if undone.Points[24].Count != 2 || undone.Points[24].Owner != engine.White {
		t.Fatalf("expected undo to restore point 24, got %+v", undone.Points[24])
	}
	if undone.Points[23].Count != 0 || undone.Points[23].Owner != engine.NoColor {
		t.Fatalf("expected undo to clear point 23, got %+v", undone.Points[23])
	}
}

func TestEditCheckerRejectsEmptySource(t *testing.T) {
	svc, err := NewService("")
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Close()

	_, err = svc.Start(GameOptions{
		GameType:     engine.GameShort,
		BotSide:      engine.Black,
		Opponent:     OpponentHuman,
		ThinkTimeSec: 1,
		Seed:         1,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.EditChecker(23, 22, engine.NoColor)
	if err == nil || !strings.Contains(err.Error(), "source point is empty") {
		t.Fatalf("expected empty source error, got %v", err)
	}
}

func TestEditCheckerRejectsOpponentDestination(t *testing.T) {
	svc, err := NewService("")
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Close()

	_, err = svc.Start(GameOptions{
		GameType:     engine.GameShort,
		BotSide:      engine.Black,
		Opponent:     OpponentHuman,
		ThinkTimeSec: 1,
		Seed:         1,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.EditChecker(24, 1, engine.NoColor)
	if err == nil || !strings.Contains(err.Error(), "destination point belongs to the opponent") {
		t.Fatalf("expected opponent destination error, got %v", err)
	}
}
