package app

import (
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
