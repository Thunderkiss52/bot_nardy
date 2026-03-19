package desktop

import (
	"errors"
	"fmt"
	"strings"

	"bot_nardy/internal/app"
	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
	"bot_nardy/internal/training"
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

type SelfLearnResponse struct {
	Result training.TrainSummary `json:"result"`
}

type BackgroundTrainingResponse struct {
	Status app.BackgroundTrainingStatus `json:"status"`
}

type ImportLogResponse struct {
	Result app.ImportLogResult `json:"result"`
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

func (a *API) SuggestMove(d1, d2 int) (bot.Decision, error) {
	if d1 < 1 || d1 > 6 || d2 < 1 || d2 > 6 {
		return bot.Decision{}, fmt.Errorf("dice must be in 1..6")
	}
	return a.svc.SuggestMove(d1, d2)
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

	before := a.svc.State()
	state, chosen, err := a.svc.ApplyHumanLineByIndex(d1, d2, lineIndex)
	if err != nil {
		return TurnResponse{}, err
	}
	analysis, err := a.svc.AnalyzeLineForState(before, d1, d2, chosen)
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

func (a *API) ApplyBestMove(d1, d2 int) (TurnResponse, error) {
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

	decision, state, err := a.svc.ApplyBestLine(d1, d2)
	if err != nil {
		return TurnResponse{}, err
	}
	analysis := &app.AnalysisResult{
		Category:    "exact",
		Delta:       0,
		BestLine:    decision.ChosenLine,
		BestWinProb: decision.ChosenProb,
	}
	if decision.LegalCount == 0 {
		analysis.BestWinProb = 0.5
	}
	applied := decision.ChosenLine
	if decision.LegalCount == 0 {
		applied = engine.TurnLine{}
	}

	return TurnResponse{
		State:     state,
		IsBotTurn: a.svc.IsBotTurn(),
		Decision:  &decision,
		Analysis:  analysis,
		Applied:   &applied,
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

func (a *API) SelfLearn(epochs int) (SelfLearnResponse, error) {
	result, err := a.svc.SelfLearn(epochs)
	if err != nil {
		return SelfLearnResponse{}, err
	}
	return SelfLearnResponse{Result: result}, nil
}

func (a *API) StartBackgroundTraining() (BackgroundTrainingResponse, error) {
	status, err := a.svc.StartBackgroundTraining()
	if err != nil {
		return BackgroundTrainingResponse{}, err
	}
	return BackgroundTrainingResponse{Status: status}, nil
}

func (a *API) StopBackgroundTraining() (BackgroundTrainingResponse, error) {
	return BackgroundTrainingResponse{Status: a.svc.StopBackgroundTraining()}, nil
}

func (a *API) BackgroundTrainingStatus() BackgroundTrainingResponse {
	return BackgroundTrainingResponse{Status: a.svc.BackgroundTrainingStatus()}
}

func (a *API) ImportMoveLog(path string) (ImportLogResponse, error) {
	result, err := a.svc.ImportMoveLog(strings.TrimSpace(path))
	if err != nil {
		return ImportLogResponse{}, err
	}
	return ImportLogResponse{Result: result}, nil
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
