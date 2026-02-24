package desktop

import (
	"errors"
	"fmt"
	"strings"

	"bot_nardy/internal/app"
	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
)

type API struct {
	svc *app.Service
}

type StartRequest struct {
	GameType  string `json:"gameType"`
	BotSide   string `json:"botSide"`
	Opponent  string `json:"opponent"`
	ThinkTime int    `json:"thinkTime"`
	ShowTop3  bool   `json:"showTop3"`
	Seed      int64  `json:"seed"`
	LogPath   string `json:"logPath"`
}

type TurnResponse struct {
	State     engine.GameState    `json:"state"`
	IsBotTurn bool                `json:"isBotTurn"`
	Decision  *bot.Decision       `json:"decision,omitempty"`
	Analysis  *app.AnalysisResult `json:"analysis,omitempty"`
	Applied   *engine.TurnLine    `json:"applied,omitempty"`
}

func NewAPI(logPath string) (*API, error) {
	svc, err := app.NewService(logPath)
	if err != nil {
		return nil, err
	}
	return &API{svc: svc}, nil
}

func (a *API) Close() error {
	if a == nil || a.svc == nil {
		return nil
	}
	return a.svc.Close()
}

func (a *API) StartGame(req StartRequest) (TurnResponse, error) {
	state, err := a.svc.Start(app.GameOptions{
		GameType:     parseGameType(req.GameType),
		BotSide:      parseColor(req.BotSide),
		Opponent:     app.OpponentMode(normalizeLower(req.Opponent)),
		ThinkTimeSec: req.ThinkTime,
		ShowTop3:     req.ShowTop3,
		Seed:         req.Seed,
	})
	if err != nil {
		return TurnResponse{}, err
	}
	return TurnResponse{State: state, IsBotTurn: a.svc.IsBotTurn()}, nil
}

func (a *API) State() TurnResponse {
	return TurnResponse{State: a.svc.State(), IsBotTurn: a.svc.IsBotTurn()}
}

func (a *API) LegalLines(d1, d2 int) ([]engine.TurnLine, error) {
	return a.svc.LegalLines(d1, d2)
}

func (a *API) ApplyDice(d1, d2 int, lineIndex int) (TurnResponse, error) {
	if d1 < 1 || d1 > 6 || d2 < 1 || d2 > 6 {
		return TurnResponse{}, fmt.Errorf("dice must be in 1..6")
	}

	if a.svc.IsBotTurn() {
		decision, state, err := a.svc.BotMove(d1, d2)
		if err != nil {
			return TurnResponse{}, err
		}
		return TurnResponse{State: state, IsBotTurn: a.svc.IsBotTurn(), Decision: &decision}, nil
	}

	legal, err := a.svc.LegalLines(d1, d2)
	if err != nil {
		return TurnResponse{}, err
	}
	if len(legal) == 0 {
		state, err := a.svc.PassTurnIfNoLegal(d1, d2)
		if err != nil {
			return TurnResponse{}, err
		}
		return TurnResponse{State: state, IsBotTurn: a.svc.IsBotTurn()}, nil
	}
	if lineIndex < 0 || lineIndex >= len(legal) {
		return TurnResponse{}, errors.New("line index required for human turn")
	}

	state, chosen, err := a.svc.ApplyHumanLineByIndex(d1, d2, lineIndex)
	if err != nil {
		return TurnResponse{}, err
	}
	analysis, err := a.svc.AnalyzeLine(d1, d2, chosen)
	if err != nil {
		return TurnResponse{}, err
	}
	return TurnResponse{
		State:     state,
		IsBotTurn: a.svc.IsBotTurn(),
		Applied:   &chosen,
		Analysis:  &analysis,
	}, nil
}

func (a *API) Undo() (TurnResponse, error) {
	state, err := a.svc.Undo()
	if err != nil {
		return TurnResponse{}, err
	}
	return TurnResponse{State: state, IsBotTurn: a.svc.IsBotTurn()}, nil
}

func (a *API) Export(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("export path is empty")
	}
	return a.svc.ExportState(path)
}

func parseGameType(v string) engine.GameType {
	if normalizeLower(v) == "long" {
		return engine.GameLong
	}
	return engine.GameShort
}

func parseColor(v string) engine.Color {
	if normalizeLower(v) == "white" {
		return engine.White
	}
	return engine.Black
}

func normalizeLower(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}
