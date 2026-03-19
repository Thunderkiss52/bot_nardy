package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
	"bot_nardy/internal/logging"
	"bot_nardy/internal/training"
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

type SelfLearnResult = training.TrainSummary

type suggestionCache struct {
	StateKey string
	Dice     [2]int
	Decision bot.Decision
}

type BackgroundTrainingStatus struct {
	Running   bool                   `json:"running"`
	Games     int                    `json:"games"`
	Examples  int                    `json:"examples"`
	Workers   int                    `json:"workers"`
	StartedAt time.Time              `json:"started_at"`
	LastTrain *training.TrainSummary `json:"last_train,omitempty"`
	LastError string                 `json:"last_error,omitempty"`
	GameType  string                 `json:"game_type"`
	ThinkTime int                    `json:"think_time_sec"`
}

type Service struct {
	mu               sync.Mutex
	trainingMu       sync.Mutex
	state            engine.GameState
	history          []engine.GameState
	options          GameOptions
	botW             *bot.Bot
	botB             *bot.Bot
	logger           *logging.JSONLLogger
	experienceWriter *training.JSONLWriter
	experiencePath   string
	weightsPath      string
	modelsDir        string
	gameExamples     []training.Example
	lastSuggestion   *suggestionCache
	bgTraining       BackgroundTrainingStatus
	bgStop           chan struct{}
	bgDone           chan struct{}
}

func NewService(logPath string) (*Service, error) {
	experiencePath, weightsPath, err := defaultTrainingPaths()
	if err != nil {
		return nil, err
	}
	modelsDir, err := defaultModelRegistryDir(weightsPath)
	if err != nil {
		return nil, err
	}
	s := &Service{
		experiencePath: experiencePath,
		weightsPath:    weightsPath,
		modelsDir:      modelsDir,
	}
	if os.Getenv("BOT_NARDY_WEIGHTS_PATH") == "" {
		weightsPath = bundledWeightsPath(weightsPath)
		s.weightsPath = weightsPath
	}
	if logPath != "" {
		l, err := logging.NewJSONLLogger(logPath)
		if err != nil {
			return nil, err
		}
		s.logger = l
	}
	if _, err := os.Stat(s.weightsPath); err == nil && os.Getenv("BOT_NARDY_WEIGHTS_PATH") == "" {
		_ = os.Setenv("BOT_NARDY_WEIGHTS_PATH", s.weightsPath)
	}
	experienceWriter, err := training.NewJSONLWriter(s.experiencePath)
	if err != nil {
		return nil, err
	}
	s.experienceWriter = experienceWriter
	return s, nil
}

func (s *Service) Close() error {
	s.StopBackgroundTraining()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.trainingMu.Lock()
	defer s.trainingMu.Unlock()
	if s.experienceWriter != nil {
		_ = s.experienceWriter.Close()
	}
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
		opts.ThinkTimeSec = 5
	}
	if opts.ThinkTimeSec > 5 {
		opts.ThinkTimeSec = 5
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
	s.gameExamples = nil
	s.lastSuggestion = nil

	if err := s.refreshBotsLocked(); err != nil {
		return engine.GameState{}, err
	}

	return s.state, nil
}

func (s *Service) State() engine.GameState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

func (s *Service) SwapBotSide() (engine.GameState, engine.Color, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.options.BotSide != engine.White && s.options.BotSide != engine.Black {
		return engine.GameState{}, engine.NoColor, errors.New("game is not initialized")
	}
	s.options.BotSide = s.options.BotSide.Opponent()
	s.lastSuggestion = nil
	return s.state, s.options.BotSide, nil
}

func (s *Service) LegalLines(d1, d2 int) ([]engine.TurnLine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return engine.GenerateLegalLines(s.state, d1, d2)
}

func (s *Service) SuggestMove(d1, d2 int) (bot.Decision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chooser := s.botForColor(s.state.Turn)
	if chooser == nil {
		return bot.Decision{}, errors.New("game is not initialized")
	}
	decision, err := chooser.ChooseMove(s.state, d1, d2)
	if err != nil {
		return bot.Decision{}, err
	}
	s.lastSuggestion = &suggestionCache{
		StateKey: s.state.NormalizeKey(),
		Dice:     [2]int{d1, d2},
		Decision: decision,
	}
	return decision, nil
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
	s.lastSuggestion = nil
	s.logMove(before, d1, d2, len(legal), nil, line, 0.0, 0, s.options.Seed)
	moveTarget, hasMoveTarget := s.moveQualityTargetLocked(before, legal, line)
	s.recordExampleLocked(before, d1, d2, len(legal), line, 0, moveTarget, hasMoveTarget)
	s.finalizeGameExamplesLocked()
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
	s.lastSuggestion = nil
	s.logMove(before, d1, d2, len(legal), nil, chosen, 0.0, 0, s.options.Seed)
	moveTarget, hasMoveTarget := s.moveQualityTargetLocked(before, legal, chosen)
	s.recordExampleLocked(before, d1, d2, len(legal), chosen, 0, moveTarget, hasMoveTarget)
	s.finalizeGameExamplesLocked()
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
	s.lastSuggestion = nil
	s.logMove(before, d1, d2, decision.LegalCount, decision.Top3, decision.ChosenLine, decision.ChosenProb, decision.ThinkElapsed, decision.Seed)
	s.recordExampleLocked(before, d1, d2, decision.LegalCount, decision.ChosenLine, decision.ChosenProb, probabilityTarget(decision.ChosenProb), true)
	s.finalizeGameExamplesLocked()

	return decision, s.state, nil
}

func (s *Service) ApplyBestLine(d1, d2 int) (bot.Decision, engine.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chooser := s.botForColor(s.state.Turn)
	if chooser == nil {
		return bot.Decision{}, engine.GameState{}, errors.New("game is not initialized")
	}
	before := s.state
	decision, ok := s.cachedSuggestionLocked(d1, d2)
	if !ok {
		var err error
		decision, err = chooser.ChooseMove(s.state, d1, d2)
		if err != nil {
			return bot.Decision{}, engine.GameState{}, err
		}
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
	s.lastSuggestion = nil
	s.logMove(before, d1, d2, decision.LegalCount, decision.Top3, decision.ChosenLine, decision.ChosenProb, decision.ThinkElapsed, decision.Seed)
	s.recordExampleLocked(before, d1, d2, decision.LegalCount, decision.ChosenLine, decision.ChosenProb, probabilityTarget(decision.ChosenProb), true)
	s.finalizeGameExamplesLocked()

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
	s.lastSuggestion = nil
	s.logMove(before, d1, d2, 0, nil, engine.TurnLine{}, 0.5, 0, s.options.Seed)
	s.recordExampleLocked(before, d1, d2, 0, engine.TurnLine{}, 0.5, 1.0, true)
	s.finalizeGameExamplesLocked()
	return s.state, nil
}

func (s *Service) SelfLearn(epochs int) (SelfLearnResult, error) {
	s.mu.Lock()
	experiencePath := s.experiencePath
	weightsPath := s.weightsPath
	modelsDir := s.modelsDir
	seed := s.options.Seed
	validationThinkTime := s.validationThinkTimeLocked()
	s.mu.Unlock()

	s.trainingMu.Lock()
	defer s.trainingMu.Unlock()

	candidatePath := weightsPath + ".candidate"
	result, err := training.TrainLinearFromJSONL(experiencePath, candidatePath, training.TrainConfig{
		Epochs:       epochs,
		LearningRate: 0.00001,
		L2:           0.000001,
		MaxExamples:  50000,
	})
	if err != nil {
		return SelfLearnResult{}, err
	}

	currentEval, err := bot.ResolveEvaluatorFromEnv()
	if err != nil {
		_ = os.Remove(candidatePath)
		return SelfLearnResult{}, err
	}
	challengerEval, err := bot.LoadLinearEvaluator(candidatePath, bot.HeuristicEvaluator{})
	if err != nil {
		_ = os.Remove(candidatePath)
		return SelfLearnResult{}, err
	}

	validationGames := 4
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	if result.ValidationExamples > 0 && result.ValidationAbsError > result.AvgAbsError*1.35 {
		result.Accepted = false
		_ = os.Remove(candidatePath)
		return result, nil
	}
	winRate, err := s.evaluateChallenger(currentEval, challengerEval, validationThinkTime, seed+91_337, validationGames)
	if err != nil {
		_ = os.Remove(candidatePath)
		return SelfLearnResult{}, err
	}
	result.ValidationGames = validationGames
	result.ValidationWinRate = winRate
	result.WeightsPath = weightsPath

	if winRate < 0.60 {
		result.Accepted = false
		_ = os.Remove(candidatePath)
		return result, nil
	}

	archivedCandidatePath, err := archiveModelFile(candidatePath, modelsDir, seed)
	if err != nil {
		_ = os.Remove(candidatePath)
		return SelfLearnResult{}, err
	}

	currentPath := ""
	if _, err := os.Stat(weightsPath); err == nil {
		currentPath = weightsPath
	}
	championPath, championName, championScore, leagueSize, err := s.selectChampionModel(currentPath, archivedCandidatePath, validationThinkTime, seed+177_013, 6)
	if err != nil {
		return SelfLearnResult{}, err
	}
	result.LeagueSize = leagueSize
	result.ChampionModel = championName
	result.ChampionScore = championScore
	result.Accepted = championPath == archivedCandidatePath

	if championPath != "" && championPath != currentPath {
		if err := copyFile(championPath, weightsPath); err != nil {
			return SelfLearnResult{}, err
		}
		_ = os.Setenv("BOT_NARDY_WEIGHTS_PATH", weightsPath)
	} else if currentPath != "" {
		_ = os.Setenv("BOT_NARDY_WEIGHTS_PATH", weightsPath)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refreshBotsLocked(); err != nil {
		return SelfLearnResult{}, err
	}
	return result, nil
}

func (s *Service) StartBackgroundTraining() (BackgroundTrainingStatus, error) {
	s.mu.Lock()
	if s.bgStop != nil {
		status := s.bgTraining
		s.mu.Unlock()
		return status, nil
	}

	thinkSec := s.backgroundThinkTimeSecLocked()
	workers := s.backgroundWorkerCountLocked()
	seed := s.options.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	status := BackgroundTrainingStatus{
		Running:   true,
		Workers:   workers,
		StartedAt: time.Now(),
		GameType:  "mixed",
		ThinkTime: thinkSec,
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	s.bgTraining = status
	s.bgStop = stop
	s.bgDone = done
	s.mu.Unlock()

	go s.runBackgroundTraining(stop, done, seed, thinkSec, workers)
	return status, nil
}

func (s *Service) StopBackgroundTraining() BackgroundTrainingStatus {
	s.mu.Lock()
	stop := s.bgStop
	done := s.bgDone
	if stop == nil {
		status := s.bgTraining
		s.mu.Unlock()
		return status
	}
	s.bgStop = nil
	s.bgDone = nil
	s.bgTraining.Running = false
	s.mu.Unlock()

	close(stop)
	<-done

	s.mu.Lock()
	status := s.bgTraining
	s.mu.Unlock()
	return status
}

func (s *Service) BackgroundTrainingStatus() BackgroundTrainingStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.bgTraining
}

func (s *Service) RunSelfPlayTrainingGame(gameType engine.GameType, thinkTime time.Duration, seed int64) (int, error) {
	examples, err := s.playSelfTrainingGame(gameType, thinkTime, seed)
	if err != nil {
		return 0, err
	}
	if len(examples) == 0 {
		return 0, nil
	}

	s.trainingMu.Lock()
	defer s.trainingMu.Unlock()
	if err := s.experienceWriter.WriteAll(examples); err != nil {
		return 0, err
	}
	return len(examples), nil
}

func (s *Service) AnalyzeLine(d1, d2 int, line engine.TurnLine) (AnalysisResult, error) {
	s.mu.Lock()
	position := s.state
	opts := s.options
	s.mu.Unlock()
	return analyzeLineForPosition(position, opts, d1, d2, line)
}

func (s *Service) AnalyzeLineForState(position engine.GameState, d1, d2 int, line engine.TurnLine) (AnalysisResult, error) {
	s.mu.Lock()
	opts := s.options
	s.mu.Unlock()
	return analyzeLineForPosition(position, opts, d1, d2, line)
}

func (s *Service) Undo() (engine.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.history) <= 1 {
		return s.state, errors.New("nothing to undo")
	}
	s.history = s.history[:len(s.history)-1]
	s.state = s.history[len(s.history)-1]
	s.lastSuggestion = nil
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
		return s.botForColor(s.state.Turn)
	}
	if s.options.BotSide == s.state.Turn {
		return s.botForColor(s.state.Turn)
	}
	return nil
}

func (s *Service) botForColor(c engine.Color) *bot.Bot {
	if c == engine.White {
		return s.botW
	}
	if c == engine.Black {
		return s.botB
	}
	return nil
}

func (s *Service) refreshBotsLocked() error {
	evaluator, err := bot.ResolveEvaluatorFromEnv()
	if err != nil {
		return err
	}

	cfg := bot.Config{
		ThinkTime: time.Duration(s.options.ThinkTimeSec) * time.Second,
		TopK:      12,
		Seed:      s.options.Seed,
		MaxPlies:  640,
		Evaluator: evaluator,
	}
	s.botW = bot.New(cfg)
	cfg.Seed += 1000003
	s.botB = bot.New(cfg)
	return nil
}

func (s *Service) recordExampleLocked(before engine.GameState, d1, d2 int, legalCount int, chosen engine.TurnLine, chosenProb float64, moveTarget float64, hasMoveTarget bool) {
	if s.options.Opponent != OpponentHuman {
		return
	}
	s.gameExamples = append(s.gameExamples, training.Example{
		StateBefore:   before,
		GameType:      before.GameType.String(),
		StateKey:      before.NormalizeKey(),
		Perspective:   before.Turn.String(),
		Turn:          before.Turn.String(),
		MoveNumber:    before.Meta.MoveNumber,
		Dice:          [2]int{d1, d2},
		LegalCount:    legalCount,
		ChosenLine:    chosen,
		ChosenProb:    chosenProb,
		HasMoveTarget: hasMoveTarget,
		MoveTarget:    moveTarget,
		Features:      bot.ExtractFeatures(before, before.Turn),
	})
}

func (s *Service) moveQualityTargetLocked(before engine.GameState, legal []engine.TurnLine, chosen engine.TurnLine) (float64, bool) {
	chooser := s.botForColor(before.Turn)
	if chooser == nil {
		return 0, false
	}
	if len(legal) <= 1 {
		return 1.0, true
	}

	best := -1e18
	chosenScore := -1e18
	chosenKey := chosen.Key()
	for _, line := range legal {
		next, err := engine.ApplyTurnLine(before, line)
		if err != nil {
			continue
		}
		score := chooser.Evaluate(next, before.Turn)
		if score > best {
			best = score
		}
		if line.Key() == chosenKey {
			chosenScore = score
		}
	}
	if best <= -1e17 || chosenScore <= -1e17 {
		return 0, false
	}
	gap := best - chosenScore
	if gap <= 0 {
		return 1.0, true
	}
	return 1.0 - 2.0*clampFloat(gap/80.0, 0, 1), true
}

func (s *Service) cachedSuggestionLocked(d1, d2 int) (bot.Decision, bool) {
	if s.lastSuggestion == nil {
		return bot.Decision{}, false
	}
	if s.lastSuggestion.StateKey != s.state.NormalizeKey() {
		return bot.Decision{}, false
	}
	if s.lastSuggestion.Dice != [2]int{d1, d2} {
		return bot.Decision{}, false
	}
	return s.lastSuggestion.Decision, true
}

func (s *Service) finalizeGameExamplesLocked() {
	if !s.state.IsTerminal() || len(s.gameExamples) == 0 || s.experienceWriter == nil {
		return
	}
	winner := s.state.Winner().String()
	for idx := range s.gameExamples {
		s.gameExamples[idx].Winner = winner
		if s.gameExamples[idx].Perspective == winner {
			s.gameExamples[idx].OutcomeValue = 1
		} else {
			s.gameExamples[idx].OutcomeValue = -1
		}
	}
	s.trainingMu.Lock()
	_ = s.experienceWriter.WriteAll(s.gameExamples)
	s.trainingMu.Unlock()
	s.gameExamples = nil
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

func analyzeLineForPosition(position engine.GameState, opts GameOptions, d1, d2 int, line engine.TurnLine) (AnalysisResult, error) {
	evaluator, err := bot.ResolveEvaluatorFromEnv()
	if err != nil {
		return AnalysisResult{}, err
	}
	currentBot := bot.New(bot.Config{
		ThinkTime: time.Duration(opts.ThinkTimeSec) * time.Second,
		TopK:      12,
		Seed:      opts.Seed + 777,
		MaxPlies:  640,
		Evaluator: evaluator,
	})
	decision, err := currentBot.ChooseMove(position, d1, d2)
	if err != nil {
		return AnalysisResult{}, err
	}
	if decision.LegalCount == 0 {
		return AnalysisResult{Category: "exact", Delta: 0, BestLine: engine.TurnLine{}, BestWinProb: 0.5}, nil
	}
	if !engine.IsLegalLine(position, d1, d2, line) {
		return AnalysisResult{}, errors.New("illegal line for analysis")
	}

	bestNext, _ := engine.ApplyTurnLine(position, decision.ChosenLine)
	bestScore := bot.EvaluatePosition(bestNext, position.Turn)
	lineNext, _ := engine.ApplyTurnLine(position, line)
	lineScore := bot.EvaluatePosition(lineNext, position.Turn)
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

func defaultTrainingPaths() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	dir := filepath.Join(home, ".codex", "memories")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	return filepath.Join(dir, "bot_nardy_training_examples.jsonl"), filepath.Join(dir, "bot_nardy_weights.json"), nil
}

func bundledWeightsPath(defaultPath string) string {
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}
	exePath, err := os.Executable()
	if err != nil {
		return defaultPath
	}
	bundledPath := filepath.Join(filepath.Dir(exePath), filepath.Base(defaultPath))
	if bundledPath == defaultPath {
		return defaultPath
	}
	if _, err := os.Stat(bundledPath); err != nil {
		return defaultPath
	}
	if err := copyFile(bundledPath, defaultPath); err == nil {
		return defaultPath
	}
	return bundledPath
}

func probabilityTarget(prob float64) float64 {
	return clampFloat(2*prob-1, -1, 1)
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

type selfPlayResult struct {
	examples []training.Example
	err      error
}

func (s *Service) runBackgroundTraining(stop <-chan struct{}, done chan<- struct{}, baseSeed int64, thinkSec int, workers int) {
	defer close(done)

	if workers < 1 {
		workers = 1
	}

	results := make(chan selfPlayResult, workers*2)
	var gameCounter atomic.Uint64
	var wg sync.WaitGroup

	for workerID := 0; workerID < workers; workerID++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}

				gameIdx := int(gameCounter.Add(1) - 1)
				gameType := engine.GameShort
				if gameIdx%2 == 1 {
					gameType = engine.GameLong
				}
				seed := baseSeed + int64(workerID+1)*1_000_003 + int64(gameIdx)*1009
				examples, err := s.playSelfTrainingGame(gameType, time.Duration(thinkSec)*time.Second, seed)

				select {
				case <-stop:
					return
				case results <- selfPlayResult{examples: examples, err: err}:
				}
			}
		}(workerID)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	gamesSinceTrain := 0
	retrainEvery := workers * 4
	if retrainEvery < 4 {
		retrainEvery = 4
	}

	for {
		var result selfPlayResult
		var ok bool
		select {
		case <-stop:
			return
		case result, ok = <-results:
			if !ok {
				return
			}
		}

		examples := result.examples
		err := result.err
		if err != nil {
			s.mu.Lock()
			s.bgTraining.LastError = err.Error()
			s.mu.Unlock()
			continue
		}
		if len(examples) > 0 {
			s.trainingMu.Lock()
			err = s.experienceWriter.WriteAll(examples)
			s.trainingMu.Unlock()
			if err != nil {
				s.mu.Lock()
				s.bgTraining.LastError = err.Error()
				s.mu.Unlock()
				continue
			}
			s.mu.Lock()
			s.bgTraining.Games++
			s.bgTraining.Examples += len(examples)
			s.bgTraining.LastError = ""
			s.mu.Unlock()
			gamesSinceTrain++
		}
		if gamesSinceTrain >= retrainEvery {
			result, err := s.SelfLearn(8)
			s.mu.Lock()
			if err != nil {
				s.bgTraining.LastError = err.Error()
			} else {
				summary := training.TrainSummary(result)
				s.bgTraining.LastTrain = &summary
				s.bgTraining.LastError = ""
			}
			s.mu.Unlock()
			gamesSinceTrain = 0
		}
	}
}

func (s *Service) playSelfTrainingGame(gameType engine.GameType, thinkTime time.Duration, seed int64) ([]training.Example, error) {
	evaluator, err := bot.ResolveEvaluatorFromEnv()
	if err != nil {
		return nil, err
	}

	whiteBot := bot.New(bot.Config{
		ThinkTime: thinkTime,
		TopK:      12,
		Workers:   1,
		MaxPlies:  640,
		Seed:      seed + 1,
		Evaluator: evaluator,
	})
	blackBot := bot.New(bot.Config{
		ThinkTime: thinkTime,
		TopK:      12,
		Workers:   1,
		MaxPlies:  640,
		Seed:      seed + 2,
		Evaluator: evaluator,
	})

	state := engine.NewShortGame(seed)
	if gameType == engine.GameLong {
		state = engine.NewLongGame(seed)
	}
	rng := rand.New(rand.NewSource(seed + 17))
	examples := make([]training.Example, 0, 256)

	for ply := 0; ply < 2048; ply++ {
		if state.IsTerminal() {
			break
		}
		d1 := rng.Intn(6) + 1
		d2 := rng.Intn(6) + 1
		thinker := whiteBot
		if state.Turn == engine.Black {
			thinker = blackBot
		}
		decision, err := thinker.ChooseMove(state, d1, d2)
		if err != nil {
			return nil, err
		}
		examples = append(examples, training.Example{
			StateBefore:   state,
			GameType:      state.GameType.String(),
			StateKey:      state.NormalizeKey(),
			Perspective:   state.Turn.String(),
			Turn:          state.Turn.String(),
			MoveNumber:    state.Meta.MoveNumber,
			Dice:          [2]int{d1, d2},
			LegalCount:    decision.LegalCount,
			ChosenLine:    decision.ChosenLine,
			ChosenProb:    decision.ChosenProb,
			HasMoveTarget: true,
			MoveTarget:    probabilityTarget(decision.ChosenProb),
			Features:      bot.ExtractFeatures(state, state.Turn),
		})
		if decision.LegalCount == 0 {
			state.Turn = state.Turn.Opponent()
			state.Meta.MoveNumber++
			continue
		}
		next, err := engine.ApplyTurnLine(state, decision.ChosenLine)
		if err != nil {
			return nil, err
		}
		state = next
	}

	winner := engine.White
	if state.IsTerminal() {
		winner = state.Winner()
	} else if bot.EvaluatePosition(state, engine.White) < 0 {
		winner = engine.Black
	}
	for idx := range examples {
		examples[idx].Winner = winner.String()
		if examples[idx].Perspective == winner.String() {
			examples[idx].OutcomeValue = 1
		} else {
			examples[idx].OutcomeValue = -1
		}
	}
	return examples, nil
}

func (s *Service) evaluateChallenger(current, challenger bot.Evaluator, thinkTime time.Duration, seed int64, games int) (float64, error) {
	if games <= 0 {
		games = 6
	}
	if thinkTime <= 0 {
		thinkTime = 250 * time.Millisecond
	}

	challengerWins := 0
	for i := 0; i < games; i++ {
		gameType := engine.GameShort
		if i%2 == 1 {
			gameType = engine.GameLong
		}

		challengerColor := engine.White
		whiteEval := challenger
		blackEval := current
		if i%2 == 1 {
			challengerColor = engine.Black
			whiteEval = current
			blackEval = challenger
		}

		winner, err := s.playEvaluationGame(gameType, thinkTime, seed+int64(i)*4099, whiteEval, blackEval)
		if err != nil {
			return 0, err
		}
		if winner == challengerColor {
			challengerWins++
		}
	}

	return float64(challengerWins) / float64(games), nil
}

func (s *Service) playEvaluationGame(gameType engine.GameType, thinkTime time.Duration, seed int64, whiteEval, blackEval bot.Evaluator) (engine.Color, error) {
	if thinkTime <= 0 {
		thinkTime = 10 * time.Millisecond
	}
	if thinkTime > 20*time.Millisecond {
		thinkTime = 20 * time.Millisecond
	}

	whiteBot := bot.New(bot.Config{
		ThinkTime: thinkTime,
		TopK:      4,
		Workers:   1,
		MaxPlies:  48,
		Seed:      seed + 1,
		Evaluator: whiteEval,
	})
	blackBot := bot.New(bot.Config{
		ThinkTime: thinkTime,
		TopK:      4,
		Workers:   1,
		MaxPlies:  48,
		Seed:      seed + 2,
		Evaluator: blackEval,
	})

	state := engine.NewShortGame(seed)
	if gameType == engine.GameLong {
		state = engine.NewLongGame(seed)
	}
	rng := rand.New(rand.NewSource(seed + 17))

	for ply := 0; ply < 128; ply++ {
		if state.IsTerminal() {
			return state.Winner(), nil
		}
		d1 := rng.Intn(6) + 1
		d2 := rng.Intn(6) + 1
		thinker := whiteBot
		if state.Turn == engine.Black {
			thinker = blackBot
		}
		decision, err := thinker.ChooseMove(state, d1, d2)
		if err != nil {
			return state.Turn.Opponent(), err
		}
		if decision.LegalCount == 0 {
			state.Turn = state.Turn.Opponent()
			state.Meta.MoveNumber++
			continue
		}
		next, err := engine.ApplyTurnLine(state, decision.ChosenLine)
		if err != nil {
			return state.Turn.Opponent(), err
		}
		state = next
	}

	if whiteEval.Evaluate(state, engine.White) >= 0 {
		return engine.White, nil
	}
	return engine.Black, nil
}

type modelEntry struct {
	path string
	name string
	eval bot.Evaluator
}

func (s *Service) selectChampionModel(currentPath, challengerPath string, thinkTime time.Duration, seed int64, recentLimit int) (string, string, float64, int, error) {
	paths, err := s.modelLeaguePaths(currentPath, challengerPath, recentLimit)
	if err != nil {
		return "", "", 0, 0, err
	}
	if len(paths) == 0 {
		return "", "", 0, 0, nil
	}

	models := make([]modelEntry, 0, len(paths))
	for _, path := range paths {
		eval, err := bot.LoadLinearEvaluator(path, bot.HeuristicEvaluator{})
		if err != nil {
			return "", "", 0, 0, err
		}
		models = append(models, modelEntry{
			path: path,
			name: filepath.Base(path),
			eval: eval,
		})
	}

	scores := make(map[string]float64, len(models))
	for _, model := range models {
		scores[model.path] = 0
	}

	for i := 0; i < len(models); i++ {
		for j := i + 1; j < len(models); j++ {
			scoreI, scoreJ, err := s.playModelPair(models[i], models[j], thinkTime, seed+int64(i*97+j*389))
			if err != nil {
				return "", "", 0, 0, err
			}
			scores[models[i].path] += scoreI
			scores[models[j].path] += scoreJ
		}
	}

	best := models[0]
	bestScore := scores[best.path]
	for _, model := range models[1:] {
		score := scores[model.path]
		if score > bestScore {
			best = model
			bestScore = score
			continue
		}
		if score == bestScore && model.path == currentPath {
			best = model
			bestScore = score
		}
	}

	return best.path, best.name, bestScore, len(models), nil
}

func (s *Service) playModelPair(a, b modelEntry, thinkTime time.Duration, seed int64) (float64, float64, error) {
	whiteWinner, err := s.playEvaluationGame(engine.GameShort, thinkTime, seed+1, a.eval, b.eval)
	if err != nil {
		return 0, 0, err
	}
	blackWinner, err := s.playEvaluationGame(engine.GameLong, thinkTime, seed+2, b.eval, a.eval)
	if err != nil {
		return 0, 0, err
	}

	scoreA := 0.0
	scoreB := 0.0
	if whiteWinner == engine.White {
		scoreA++
	} else {
		scoreB++
	}
	if blackWinner == engine.White {
		scoreB++
	} else {
		scoreA++
	}
	return scoreA, scoreB, nil
}

func (s *Service) modelLeaguePaths(currentPath, challengerPath string, recentLimit int) ([]string, error) {
	seen := map[string]struct{}{}
	paths := make([]string, 0, recentLimit+2)
	appendPath := func(path string) {
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}

	appendPath(currentPath)
	appendPath(challengerPath)

	recent, err := recentModelFiles(s.modelsDir, recentLimit)
	if err != nil {
		return nil, err
	}
	for _, path := range recent {
		appendPath(path)
	}
	return paths, nil
}

func (s *Service) backgroundThinkTimeSecLocked() int {
	switch {
	case s.options.ThinkTimeSec >= 5:
		return 2
	case s.options.ThinkTimeSec >= 2:
		return s.options.ThinkTimeSec - 1
	default:
		return 1
	}
}

func (s *Service) validationThinkTimeLocked() time.Duration {
	thinkSec := s.backgroundThinkTimeSecLocked()
	switch {
	case thinkSec >= 5:
		return 20 * time.Millisecond
	case thinkSec >= 3:
		return 15 * time.Millisecond
	default:
		return 10 * time.Millisecond
	}
}

func defaultModelRegistryDir(weightsPath string) (string, error) {
	dir := filepath.Join(filepath.Dir(weightsPath), "bot_nardy_models")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func archiveModelFile(sourcePath, modelsDir string, seed int64) (string, error) {
	name := fmt.Sprintf("model_%s_%d.json", time.Now().UTC().Format("20060102_150405"), seed)
	dest := filepath.Join(modelsDir, name)
	if err := os.Rename(sourcePath, dest); err == nil {
		return dest, nil
	}
	if err := copyFile(sourcePath, dest); err != nil {
		return "", err
	}
	_ = os.Remove(sourcePath)
	return dest, nil
}

func recentModelFiles(modelsDir string, limit int) ([]string, error) {
	if limit <= 0 {
		return nil, nil
	}
	entries, err := os.ReadDir(modelsDir)
	if err != nil {
		return nil, err
	}
	type item struct {
		path    string
		modTime time.Time
	}
	items := make([]item, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		items = append(items, item{
			path:    filepath.Join(modelsDir, entry.Name()),
			modTime: info.ModTime(),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].modTime.After(items[j].modTime)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	paths := make([]string, 0, len(items))
	for _, item := range items {
		paths = append(paths, item.path)
	}
	return paths, nil
}

func copyFile(sourcePath, destPath string) error {
	raw, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, raw, 0o644)
}

func (s *Service) backgroundWorkerCountLocked() int {
	cpus := runtime.NumCPU()
	switch {
	case cpus <= 2:
		return 1
	case cpus <= 4:
		return 2
	case cpus <= 8:
		return 3
	default:
		return 4
	}
}
