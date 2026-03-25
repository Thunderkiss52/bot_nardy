package desktop

import "testing"

func TestApplyDiceHumanReturnsAnalysis(t *testing.T) {
	api, err := NewAPI("")
	if err != nil {
		t.Fatalf("new api: %v", err)
	}
	defer api.Close()

	_, err = api.StartGame(StartRequest{
		GameType:  "long",
		BotSide:   "black",
		Opponent:  "human",
		ThinkTime: 1,
		ShowTop3:  true,
		Seed:      1,
	})
	if err != nil {
		t.Fatalf("start game: %v", err)
	}

	resp, err := api.ApplyDice(1, 2, 0)
	if err != nil {
		t.Fatalf("apply dice: %v", err)
	}
	if resp.Applied == nil {
		t.Fatalf("expected applied line")
	}
	if resp.Analysis == nil {
		t.Fatalf("expected analysis result")
	}
}

func TestApplyBestMoveUsesStrongestLineForHumanTurn(t *testing.T) {
	api, err := NewAPI("")
	if err != nil {
		t.Fatalf("new api: %v", err)
	}
	defer api.Close()

	_, err = api.StartGame(StartRequest{
		GameType:  "short",
		BotSide:   "black",
		Opponent:  "human",
		ThinkTime: 1,
		ShowTop3:  true,
		Seed:      7,
	})
	if err != nil {
		t.Fatalf("start game: %v", err)
	}

	decision, err := api.SuggestMove(3, 1)
	if err != nil {
		t.Fatalf("suggest move: %v", err)
	}

	resp, err := api.ApplyBestMove(3, 1)
	if err != nil {
		t.Fatalf("apply best move: %v", err)
	}
	if resp.Applied == nil {
		t.Fatalf("expected applied line")
	}
	if resp.Applied.Key() != decision.ChosenLine.Key() {
		t.Fatalf("expected best line %s, got %s", decision.ChosenLine, resp.Applied)
	}
	if resp.Analysis == nil || resp.Analysis.Category != "exact" {
		t.Fatalf("expected exact analysis, got %+v", resp.Analysis)
	}
}

func TestBackgroundTrainingLifecycle(t *testing.T) {
	api, err := NewAPI("")
	if err != nil {
		t.Fatalf("new api: %v", err)
	}
	defer api.Close()

	started, err := api.StartBackgroundTraining()
	if err != nil {
		t.Fatalf("start background training: %v", err)
	}
	if !started.Status.Running {
		t.Fatalf("expected running background training")
	}

	status := api.BackgroundTrainingStatus()
	if !status.Status.Running {
		t.Fatalf("expected running status from background training status")
	}

	stopped, err := api.StopBackgroundTraining()
	if err != nil {
		t.Fatalf("stop background training: %v", err)
	}
	if stopped.Status.Running {
		t.Fatalf("expected stopped background training")
	}
}

func TestSwapBotSideUpdatesBotTurn(t *testing.T) {
	api, err := NewAPI("")
	if err != nil {
		t.Fatalf("new api: %v", err)
	}
	defer api.Close()

	_, err = api.StartGame(StartRequest{
		GameType:  "short",
		BotSide:   "black",
		Opponent:  "human",
		ThinkTime: 1,
		Seed:      11,
	})
	if err != nil {
		t.Fatalf("start game: %v", err)
	}

	resp, err := api.SwapBotSide()
	if err != nil {
		t.Fatalf("swap bot side: %v", err)
	}
	if !resp.IsBotTurn {
		t.Fatalf("expected bot turn after swapping bot to white")
	}
}

func TestEditCheckerUpdatesState(t *testing.T) {
	api, err := NewAPI("")
	if err != nil {
		t.Fatalf("new api: %v", err)
	}
	defer api.Close()

	_, err = api.StartGame(StartRequest{
		GameType:  "short",
		BotSide:   "black",
		Opponent:  "human",
		ThinkTime: 1,
		Seed:      5,
	})
	if err != nil {
		t.Fatalf("start game: %v", err)
	}

	resp, err := api.EditChecker(EditCheckerRequest{From: 24, To: 23})
	if err != nil {
		t.Fatalf("edit checker: %v", err)
	}
	if resp.State.Points[24].Count != 1 || resp.State.Points[23].Count != 1 {
		t.Fatalf("unexpected edited state: p24=%+v p23=%+v", resp.State.Points[24], resp.State.Points[23])
	}
}
