package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
	"bot_nardy/internal/logging"
)

type OpponentMode string

const (
	OpponentHuman OpponentMode = "human"
	OpponentBot   OpponentMode = "bot"
)

type GameOptions struct {
	GameType     engine.GameType `json:"game_type"`
	BotSide      engine.Color    `json:"bot_side"`
	Opponent     OpponentMode    `json:"opponent"`
	ThinkTimeSec int             `json:"think_time_sec"`
	ShowTop3     bool            `json:"show_top3"`
	Seed         int64           `json:"seed"`
}

type AnalysisResult struct {
	Category    string          `json:"category"`
	Delta       float64         `json:"delta"`
	BestLine    engine.TurnLine `json:"best_line"`
	BestWinProb float64         `json:"best_winprob"`
}

type Service struct {
	mu      sync.Mutex
	state   engine.GameState
	history []engine.GameState
	options GameOptions
	botW    *bot.Bot
	botB    *bot.Bot
	logger  *logging.JSONLLogger
}

func NewService(logPath string) (*Service, error) {
	s := &Service{}
	if logPath != "" {
		l, err := logging.NewJSONLLogger(logPath)
		if err != nil {
			return nil, err
		}
		s.logger = l
	}
	return s, nil
}

func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.logger != nil {
		return s.logger.Close()
	}
	return nil
}

func (s *Service) Start(opts GameOptions) (engine.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if opts.GameType != engine.GameShort && opts.GameType != engine.GameLong {
		return engine.GameState{}, errors.New("unsupported game type")
	}
	if opts.BotSide != engine.White && opts.BotSide != engine.Black {
		return engine.GameState{}, errors.New("invalid bot side")
	}
	if opts.ThinkTimeSec < 1 {
		opts.ThinkTimeSec = 8
	}
	if opts.ThinkTimeSec > 20 {
		opts.ThinkTimeSec = 20
	}
	if opts.Opponent == "" {
		opts.Opponent = OpponentHuman
	}
	if opts.Seed == 0 {
		opts.Seed = time.Now().UnixNano()
	}

	if opts.GameType == engine.GameShort {
		s.state = engine.NewShortGame(opts.Seed)
	} else {
		s.state = engine.NewLongGame(opts.Seed)
	}
	s.history = []engine.GameState{s.state}
	s.options = opts

	cfg := bot.Config{
		ThinkTime: time.Duration(opts.ThinkTimeSec) * time.Second,
		TopK:      10,
		Seed:      opts.Seed,
	}
	s.botW = bot.New(cfg)
	cfg.Seed += 1000003
	s.botB = bot.New(cfg)

	return s.state, nil
}

func (s *Service) State() engine.GameState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

func (s *Service) LegalLines(d1, d2 int) ([]engine.TurnLine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return engine.GenerateLegalLines(s.state, d1, d2)
}

func (s *Service) ApplyHumanLine(d1, d2 int, line engine.TurnLine) (engine.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentBotForTurn() != nil {
		return engine.GameState{}, errors.New("current turn belongs to bot")
	}
	legal, err := engine.GenerateLegalLines(s.state, d1, d2)
	if err != nil {
		return engine.GameState{}, err
	}
	if len(legal) == 0 {
		return engine.GameState{}, errors.New("no legal lines available")
	}
	if !containsLine(legal, line) {
		return engine.GameState{}, errors.New("illegal line")
	}
	before := s.state
	next, err := engine.ApplyTurnLine(before, line)
	if err != nil {
		return engine.GameState{}, err
	}
	s.state = next
	s.history = append(s.history, next)
	s.logMove(before, d1, d2, len(legal), nil, line, 0.0, 0, s.options.Seed)
	return next, nil
}

func (s *Service) ApplyHumanLineByIndex(d1, d2 int, index int) (engine.GameState, engine.TurnLine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentBotForTurn() != nil {
		return engine.GameState{}, engine.TurnLine{}, errors.New("current turn belongs to bot")
	}
	legal, err := engine.GenerateLegalLines(s.state, d1, d2)
	if err != nil {
		return engine.GameState{}, engine.TurnLine{}, err
	}
	if len(legal) == 0 {
		return engine.GameState{}, engine.TurnLine{}, errors.New("no legal lines available")
	}
	if index < 0 || index >= len(legal) {
		return engine.GameState{}, engine.TurnLine{}, fmt.Errorf("line index out of range: %d", index)
	}
	chosen := legal[index]
	before := s.state
	next, err := engine.ApplyTurnLine(before, chosen)
	if err != nil {
		return engine.GameState{}, engine.TurnLine{}, err
	}
	s.state = next
	s.history = append(s.history, next)
	s.logMove(before, d1, d2, len(legal), nil, chosen, 0.0, 0, s.options.Seed)
	return next, chosen, nil
}

func (s *Service) BotMove(d1, d2 int) (bot.Decision, engine.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentBot := s.currentBotForTurn()
	if currentBot == nil {
		return bot.Decision{}, engine.GameState{}, errors.New("current turn belongs to human")
	}
	before := s.state
	decision, err := currentBot.ChooseMove(s.state, d1, d2)
	if err != nil {
		return bot.Decision{}, engine.GameState{}, err
	}

	if decision.LegalCount > 0 {
		next, err := engine.ApplyTurnLine(s.state, decision.ChosenLine)
		if err != nil {
			return bot.Decision{}, engine.GameState{}, err
		}
		s.state = next
	} else {
		s.state.Turn = s.state.Turn.Opponent()
		s.state.Meta.MoveNumber++
	}
	s.history = append(s.history, s.state)
	s.logMove(before, d1, d2, decision.LegalCount, decision.Top3, decision.ChosenLine, decision.ChosenProb, decision.ThinkElapsed, decision.Seed)

	return decision, s.state, nil
}

func (s *Service) PassTurnIfNoLegal(d1, d2 int) (engine.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentBotForTurn() != nil {
		return engine.GameState{}, errors.New("current turn belongs to bot")
	}
	legal, err := engine.GenerateLegalLines(s.state, d1, d2)
	if err != nil {
		return engine.GameState{}, err
	}
	if len(legal) > 0 {
		return engine.GameState{}, errors.New("pass is only allowed when no legal lines exist")
	}
	before := s.state
	s.state.Turn = s.state.Turn.Opponent()
	s.state.Meta.MoveNumber++
	s.history = append(s.history, s.state)
	s.logMove(before, d1, d2, 0, nil, engine.TurnLine{}, 0.5, 0, s.options.Seed)
	return s.state, nil
}

func (s *Service) AnalyzeLine(d1, d2 int, line engine.TurnLine) (AnalysisResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentBot := bot.New(bot.Config{
		ThinkTime: time.Duration(s.options.ThinkTimeSec) * time.Second,
		TopK:      10,
		Seed:      s.options.Seed + 777,
	})
	decision, err := currentBot.ChooseMove(s.state, d1, d2)
	if err != nil {
		return AnalysisResult{}, err
	}
	if decision.LegalCount == 0 {
		return AnalysisResult{Category: "exact", Delta: 0, BestLine: engine.TurnLine{}, BestWinProb: 0.5}, nil
	}
	if !engine.IsLegalLine(s.state, d1, d2, line) {
		return AnalysisResult{}, errors.New("illegal line for analysis")
	}

	bestNext, _ := engine.ApplyTurnLine(s.state, decision.ChosenLine)
	bestScore := bot.EvaluatePosition(bestNext, s.state.Turn)
	lineNext, _ := engine.ApplyTurnLine(s.state, line)
	lineScore := bot.EvaluatePosition(lineNext, s.state.Turn)
	delta := normalizeDelta(bestScore - lineScore)

	category := classifyDelta(delta)
	if line.Key() == decision.ChosenLine.Key() {
		category = "exact"
		delta = 0
	}
	return AnalysisResult{
		Category:    category,
		Delta:       delta,
		BestLine:    decision.ChosenLine,
		BestWinProb: decision.ChosenProb,
	}, nil
}

func (s *Service) Undo() (engine.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.history) <= 1 {
		return s.state, errors.New("nothing to undo")
	}
	s.history = s.history[:len(s.history)-1]
	s.state = s.history[len(s.history)-1]
	return s.state, nil
}

func (s *Service) ExportState(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(s.state)
}

func (s *Service) IsBotTurn() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentBotForTurn() != nil
}

func (s *Service) currentBotForTurn() *bot.Bot {
	if s.options.Opponent == OpponentBot {
		if s.state.Turn == engine.White {
			return s.botW
		}
		return s.botB
	}
	if s.options.BotSide == s.state.Turn {
		if s.state.Turn == engine.White {
			return s.botW
		}
		return s.botB
	}
	return nil
}

func normalizeDelta(scoreDiff float64) float64 {
	if scoreDiff <= 0 {
		return 0
	}
	delta := scoreDiff / 100.0
	if delta > 1 {
		return 1
	}
	return delta
}

func containsLine(lines []engine.TurnLine, line engine.TurnLine) bool {
	k := line.Key()
	for _, l := range lines {
		if l.Key() == k {
			return true
		}
	}
	return false
}

func (s *Service) logMove(before engine.GameState, d1, d2 int, legalCount int, top3 []bot.MoveEvaluation, chosen engine.TurnLine, winProb float64, think time.Duration, seed int64) {
	if s.logger == nil {
		return
	}
	_ = s.logger.Write(logging.MoveLog{
		Timestamp:   time.Now(),
		StateBefore: before,
		Dice:        [2]int{d1, d2},
		LegalCount:  legalCount,
		Top3:        top3,
		ChosenLine:  chosen,
		WinProb:     winProb,
		ThinkMillis: think.Milliseconds(),
		Seed:        seed,
	})
}

func classifyDelta(delta float64) string {
	switch {
	case delta < 0.005:
		return "exact"
	case delta < 0.02:
		return "inaccuracy"
	case delta < 0.03:
		return "mistake"
	default:
		return "blunder"
	}
}

func (s *Service) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fmt.Sprintf("game=%s turn=%s move=%d", s.state.GameType, s.state.Turn, s.state.Meta.MoveNumber)
}
