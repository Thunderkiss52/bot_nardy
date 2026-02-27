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
