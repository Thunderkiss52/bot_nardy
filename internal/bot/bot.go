package bot

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"bot_nardy/internal/engine"
)

type Config struct {
	ThinkTime time.Duration
	TopK      int
	Workers   int
	Seed      int64
	MaxPlies  int
	Evaluator Evaluator
}

type Bot struct {
	cfg                Config
	lineCache          *boundedCache[[]engine.TurnLine]
	scoreCache         *boundedCache[float64]
	replyEquityCache   *boundedCache[float64]
	replyBestCache     *boundedCache[float64]
	counterEquityCache *boundedCache[float64]
	openingBook        *boundedCache[string]
}

type MoveEvaluation struct {
	Line    engine.TurnLine `json:"line"`
	WinProb float64         `json:"winprob"`
	Sims    uint64          `json:"sims"`
}

type Decision struct {
	LegalCount   int              `json:"legal_count"`
	ChosenLine   engine.TurnLine  `json:"chosen_line"`
	ChosenProb   float64          `json:"chosen_prob"`
	Top3         []MoveEvaluation `json:"top3"`
	ThinkElapsed time.Duration    `json:"think_elapsed"`
	Seed         int64            `json:"seed"`
}

type candidate struct {
	line  engine.TurnLine
	state engine.GameState
	score float64
	fast  float64
	prior float64
	wins  atomic.Uint64
	sims  atomic.Uint64
}

type weightedDice struct {
	d1     int
	d2     int
	weight float64
}

func New(cfg Config) *Bot {
	if cfg.ThinkTime <= 0 {
		cfg.ThinkTime = 5 * time.Second
	}
	if cfg.TopK <= 0 {
		cfg.TopK = 8
	}
	if cfg.Workers <= 0 {
		cfg.Workers = runtime.NumCPU()
	}
	if cfg.MaxPlies <= 0 {
		cfg.MaxPlies = 512
	}
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}
	if cfg.Evaluator == nil {
		cfg.Evaluator = defaultEvaluator
	}
	return &Bot{
		cfg:                cfg,
		lineCache:          newBoundedCache[[]engine.TurnLine](20_000),
		scoreCache:         newBoundedCache[float64](40_000),
		replyEquityCache:   newBoundedCache[float64](20_000),
		replyBestCache:     newBoundedCache[float64](30_000),
		counterEquityCache: newBoundedCache[float64](10_000),
		openingBook:        newBoundedCache[string](2_048),
	}
}

func (b *Bot) ChooseMove(state engine.GameState, d1, d2 int) (Decision, error) {
	start := time.Now()
	legal, err := engine.GenerateLegalLines(state, d1, d2)
	if err != nil {
		return Decision{}, err
	}
	if len(legal) == 0 {
		return Decision{
			LegalCount:   0,
			ChosenLine:   engine.TurnLine{},
			ChosenProb:   0.5,
			Top3:         nil,
			ThinkElapsed: time.Since(start),
			Seed:         b.cfg.Seed,
		}, nil
	}

	preferredLineKey, hasPreferred := b.lookupOpeningBook(state, d1, d2)
	cands := b.prepareCandidates(state, legal, preferredLineKey)
	if isOpeningPosition(state) && !hasPreferred {
		b.refineOpeningCandidates(state, cands)
	}
	if len(cands) == 0 {
		return Decision{}, errors.New("no candidates")
	}

	deadline := start.Add(b.cfg.ThinkTime)
	target := state.Turn
	priorWeight := b.simulationPriorWeight()
	var totalRollouts atomic.Uint64
	var stopped atomic.Bool
	var wg sync.WaitGroup
	for i := 0; i < b.cfg.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				if stopped.Load() || time.Now().After(deadline) {
					return
				}
				idx := totalRollouts.Add(1)
				ci := selectCandidateForSimulation(cands, priorWeight, idx)
				seed := b.cfg.Seed + int64(workerID+1)*15485863 + int64(idx)*7919
				rng := rand.New(rand.NewSource(seed))
				winner := b.rollout(cands[ci].state, target, rng)
				cands[ci].sims.Add(1)
				if winner == target {
					cands[ci].wins.Add(1)
				}
				if idx%64 == 0 && canEarlyStop(cands, priorWeight) {
					stopped.Store(true)
					return
				}
			}
		}(i)
	}
	wg.Wait()

	evals := make([]MoveEvaluation, 0, len(cands))
	for _, c := range cands {
		evals = append(evals, MoveEvaluation{
			Line:    c.line,
			WinProb: candidateWinEstimate(c, priorWeight),
			Sims:    c.sims.Load(),
		})
	}
	sort.Slice(evals, func(i, j int) bool {
		if evals[i].WinProb == evals[j].WinProb {
			return evals[i].Line.Key() < evals[j].Line.Key()
		}
		return evals[i].WinProb > evals[j].WinProb
	})
	top3N := 3
	if len(evals) < top3N {
		top3N = len(evals)
	}

	return Decision{
		LegalCount:   len(legal),
		ChosenLine:   evals[0].Line,
		ChosenProb:   evals[0].WinProb,
		Top3:         evals[:top3N],
		ThinkElapsed: time.Since(start),
		Seed:         b.cfg.Seed,
	}, b.storeOpeningBook(state, d1, d2, evals[0].Line)
}

func (b *Bot) prepareCandidates(state engine.GameState, legal []engine.TurnLine, preferredLineKey string) []*candidate {
	cands := make([]*candidate, 0, len(legal))
	for _, line := range legal {
		next, err := engine.ApplyTurnLine(state, line)
		if err != nil {
			continue
		}
		fast := b.scorePosition(next, state.Turn) + tacticalLineScore(state, next, line) + openingBias(state, next, line)
		if preferredLineKey != "" && line.Key() == preferredLineKey {
			fast += 24.0
		}
		cands = append(cands, &candidate{line: line, state: next, score: fast, fast: fast})
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].score == cands[j].score {
			return cands[i].line.Key() < cands[j].line.Key()
		}
		return cands[i].score > cands[j].score
	})

	deepK := b.deepCandidateCount()
	if deepK > len(cands) {
		deepK = len(cands)
	}
	candidateWorkers, replyWorkers := b.deepEvalParallelism(deepK)
	if candidateWorkers <= 1 {
		for i := 0; i < deepK; i++ {
			reply := b.expectedReplyEquityWithWorkers(cands[i].state, state.Turn, replyWorkers)
			cands[i].score = 0.45*cands[i].fast + 0.55*reply
		}
	} else {
		jobs := make(chan int, deepK)
		var wg sync.WaitGroup
		for wi := 0; wi < candidateWorkers; wi++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for idx := range jobs {
					reply := b.expectedReplyEquityWithWorkers(cands[idx].state, state.Turn, replyWorkers)
					cands[idx].score = 0.45*cands[idx].fast + 0.55*reply
				}
			}()
		}
		for i := 0; i < deepK; i++ {
			jobs <- i
		}
		close(jobs)
		wg.Wait()
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].score == cands[j].score {
			if cands[i].fast == cands[j].fast {
				return cands[i].line.Key() < cands[j].line.Key()
			}
			return cands[i].fast > cands[j].fast
		}
		return cands[i].score > cands[j].score
	})
	if len(cands) > b.cfg.TopK {
		cands = cands[:b.cfg.TopK]
	}
	for _, cand := range cands {
		cand.prior = scoreToWinProb(cand.score)
	}
	return cands
}

func (b *Bot) refineOpeningCandidates(state engine.GameState, cands []*candidate) {
	if !isOpeningPosition(state) || len(cands) == 0 {
		return
	}

	counterK := len(cands)
	if counterK > 4 {
		counterK = 4
	}
	replyWorkers := b.replyWorkers(21)
	for i := 0; i < counterK; i++ {
		counter := b.expectedCounterPlayEquity(cands[i].state, state.Turn, replyWorkers)
		cands[i].score = 0.18*cands[i].fast + 0.42*cands[i].score + 0.40*counter
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].score == cands[j].score {
			if cands[i].fast == cands[j].fast {
				return cands[i].line.Key() < cands[j].line.Key()
			}
			return cands[i].fast > cands[j].fast
		}
		return cands[i].score > cands[j].score
	})
	if len(cands) > b.cfg.TopK {
		cands = cands[:b.cfg.TopK]
	}
	for _, cand := range cands {
		cand.prior = scoreToWinProb(cand.score)
	}
}

func (b *Bot) deepCandidateCount() int {
	switch {
	case b.cfg.ThinkTime <= 500*time.Millisecond:
		return 4
	case b.cfg.ThinkTime <= 1500*time.Millisecond:
		return 5
	case b.cfg.ThinkTime <= 3*time.Second:
		return 6
	case b.cfg.ThinkTime <= 5*time.Second:
		return 8
	default:
		return 10
	}
}

func (b *Bot) expectedReplyEquity(state engine.GameState, perspective engine.Color) float64 {
	return b.expectedReplyEquityWithWorkers(state, perspective, b.replyWorkers(len(weightedDiceOutcomes())))
}

func (b *Bot) expectedReplyEquityWithWorkers(state engine.GameState, perspective engine.Color, workers int) float64 {
	cacheKey := replyEquityCacheKey(state, perspective)
	if cached, ok := b.replyEquityCache.Get(cacheKey); ok {
		return cached
	}

	outcomes := weightedDiceOutcomes()
	workers = b.replyWorkersForBudget(len(outcomes), workers)
	totalWeight := 0.0
	totalScore := 0.0

	if workers <= 1 {
		for _, dice := range outcomes {
			score, ok := b.replyOutcomeScore(state, perspective, dice)
			if !ok {
				continue
			}
			totalScore += dice.weight * score
			totalWeight += dice.weight
		}
	} else {
		type replyResult struct {
			weight float64
			score  float64
			ok     bool
		}
		jobs := make(chan weightedDice, len(outcomes))
		results := make(chan replyResult, len(outcomes))
		var wg sync.WaitGroup

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for dice := range jobs {
					score, ok := b.replyOutcomeScore(state, perspective, dice)
					results <- replyResult{weight: dice.weight, score: score, ok: ok}
				}
			}()
		}

		for _, dice := range outcomes {
			jobs <- dice
		}
		close(jobs)
		wg.Wait()
		close(results)

		for result := range results {
			if !result.ok {
				continue
			}
			totalScore += result.weight * result.score
			totalWeight += result.weight
		}
	}

	if totalWeight == 0 {
		score := b.scorePosition(state, perspective)
		b.replyEquityCache.Set(cacheKey, score)
		return score
	}
	score := totalScore / totalWeight
	b.replyEquityCache.Set(cacheKey, score)
	return score
}

func (b *Bot) expectedCounterPlayEquity(state engine.GameState, perspective engine.Color, workers int) float64 {
	cacheKey := replyEquityCacheKey(state, perspective) + "|counter"
	if cached, ok := b.counterEquityCache.Get(cacheKey); ok {
		return cached
	}

	outcomes := weightedDiceOutcomes()
	totalWeight := 0.0
	totalScore := 0.0
	for _, dice := range outcomes {
		lines, err := b.generateLines(state, dice.d1, dice.d2)
		if err != nil {
			continue
		}
		if len(lines) == 0 {
			passed := state.Clone()
			passed.Turn = passed.Turn.Opponent()
			passed.Meta.MoveNumber++
			totalScore += dice.weight * b.expectedReplyEquityWithWorkers(passed, perspective, workers)
			totalWeight += dice.weight
			continue
		}

		bestReply := state
		bestReplySet := false
		best := math.Inf(-1)
		replyPlayer := state.Turn
		for _, line := range lines {
			next, err := engine.ApplyTurnLine(state, line)
			if err != nil {
				continue
			}
			replyScore := b.scorePosition(next, replyPlayer) + tacticalLineScore(state, next, line) + openingBias(state, next, line)
			if replyScore > best {
				best = replyScore
				bestReply = next
				bestReplySet = true
			}
		}
		if !bestReplySet {
			continue
		}
		totalScore += dice.weight * b.expectedReplyEquityWithWorkers(bestReply, perspective, workers)
		totalWeight += dice.weight
	}
	if totalWeight == 0 {
		score := b.scorePosition(state, perspective)
		b.counterEquityCache.Set(cacheKey, score)
		return score
	}
	score := totalScore / totalWeight
	b.counterEquityCache.Set(cacheKey, score)
	return score
}

func (b *Bot) replyOutcomeScore(state engine.GameState, perspective engine.Color, dice weightedDice) (float64, bool) {
	lines, err := b.generateLines(state, dice.d1, dice.d2)
	if err != nil {
		return 0, false
	}
	if len(lines) == 0 {
		passed := state.Clone()
		passed.Turn = passed.Turn.Opponent()
		passed.Meta.MoveNumber++
		return b.scorePosition(passed, perspective), true
	}

	cacheKey := replyBestCacheKey(state, perspective, dice.d1, dice.d2)
	if cached, ok := b.replyBestCache.Get(cacheKey); ok {
		return cached, true
	}
	score := b.bestReplyScore(state, lines, perspective)
	b.replyBestCache.Set(cacheKey, score)
	return score, true
}

func (b *Bot) bestReplyScore(state engine.GameState, lines []engine.TurnLine, perspective engine.Color) float64 {
	replyPlayer := state.Turn
	best := math.Inf(-1)
	bestScoreForPerspective := b.scorePosition(state, perspective)

	for _, line := range lines {
		next, err := engine.ApplyTurnLine(state, line)
		if err != nil {
			continue
		}
		replyScore := b.scorePosition(next, replyPlayer) + tacticalLineScore(state, next, line) + openingBias(state, next, line)
		if replyScore > best {
			best = replyScore
			bestScoreForPerspective = b.scorePosition(next, perspective)
		}
	}

	return bestScoreForPerspective
}

func tacticalLineScore(before, after engine.GameState, line engine.TurnLine) float64 {
	player := before.Turn
	opp := player.Opponent()
	score := 0.0

	score += float64(after.Off[player.Idx()]-before.Off[player.Idx()]) * 18.0
	score += float64(after.Bar[opp.Idx()]-before.Bar[opp.Idx()]) * 16.0
	score += float64(engine.CountMadePoints(after, player)-engine.CountMadePoints(before, player)) * 2.8
	score += float64(engine.LongestPrime(after, player)-engine.LongestPrime(before, player)) * 3.2
	score += float64(engine.HomeMadePoints(after, player)-engine.HomeMadePoints(before, player)) * 2.2
	score += float64(engine.CountDirectShotsShort(after, player)-engine.CountDirectShotsShort(before, player)) * 1.0
	score -= float64(engine.CountExposedBlotsShort(after, player)-engine.CountExposedBlotsShort(before, player)) * 5.0
	score -= float64(engine.CountBlots(after, player)-engine.CountBlots(before, player)) * 1.8

	if before.GameType == engine.GameLong {
		score += float64(engine.MaxBlockRunLong(after, player)-engine.MaxBlockRunLong(before, player)) * 2.4
		score -= float64(engine.CheckersOnHead(after, player)-engine.CheckersOnHead(before, player)) * 0.8
	}
	if len(line.Moves) == 0 {
		score -= 2.0
	}
	return score
}

func (b *Bot) rollout(state engine.GameState, target engine.Color, rng *rand.Rand) engine.Color {
	cur := state.Clone()
	for ply := 0; ply < b.cfg.MaxPlies; ply++ {
		if cur.IsTerminal() {
			return cur.Winner()
		}
		d1 := rng.Intn(6) + 1
		d2 := rng.Intn(6) + 1
		lines, err := b.generateLines(cur, d1, d2)
		if err != nil {
			return cur.Turn.Opponent()
		}
		if len(lines) == 0 {
			cur.Turn = cur.Turn.Opponent()
			continue
		}
		line := b.selectPolicyMove(cur, lines, rng)
		next, err := engine.ApplyTurnLine(cur, line)
		if err != nil {
			return cur.Turn.Opponent()
		}
		cur = next
	}
	if b.scorePosition(cur, target) >= 0 {
		return target
	}
	return target.Opponent()
}

func (b *Bot) selectPolicyMove(state engine.GameState, lines []engine.TurnLine, rng *rand.Rand) engine.TurnLine {
	best := lines[0]
	bestScore := math.Inf(-1)
	for _, line := range lines {
		next, err := engine.ApplyTurnLine(state, line)
		if err != nil {
			continue
		}
		s := b.scorePosition(next, state.Turn) + tacticalLineScore(state, next, line) + openingBias(state, next, line)
		s += rng.Float64() * 0.01
		if s > bestScore {
			best = line
			bestScore = s
		}
	}
	return best
}

func canEarlyStop(cands []*candidate, priorWeight float64) bool {
	if len(cands) < 2 {
		return false
	}
	type bound struct {
		idx   int
		lower float64
		upper float64
	}
	bounds := make([]bound, 0, len(cands))
	for i, c := range cands {
		effectiveSims := float64(c.sims.Load()) + priorWeight
		if effectiveSims < 200 {
			return false
		}
		p := candidateWinEstimate(c, priorWeight)
		margin := 1.96 * math.Sqrt(math.Max(0.000001, p*(1-p))/effectiveSims)
		bounds = append(bounds, bound{idx: i, lower: p - margin, upper: p + margin})
	}
	sort.Slice(bounds, func(i, j int) bool {
		return bounds[i].lower > bounds[j].lower
	})
	best := bounds[0]
	for _, b := range bounds[1:] {
		if best.lower <= b.upper {
			return false
		}
	}
	return true
}

func selectCandidateForSimulation(cands []*candidate, priorWeight float64, totalRollouts uint64) int {
	bestIdx := 0
	bestScore := math.Inf(-1)
	logTotal := math.Log(float64(totalRollouts) + 2)

	for i, cand := range cands {
		sims := cand.sims.Load()
		if sims == 0 {
			ucb := cand.prior + 0.35
			if ucb > bestScore {
				bestScore = ucb
				bestIdx = i
			}
			continue
		}
		estimate := candidateWinEstimate(cand, priorWeight)
		bonus := 0.9 * math.Sqrt(logTotal/float64(sims))
		ucb := estimate + bonus
		if ucb > bestScore {
			bestScore = ucb
			bestIdx = i
		}
	}
	return bestIdx
}

func candidateWinEstimate(c *candidate, priorWeight float64) float64 {
	wins := float64(c.wins.Load())
	sims := float64(c.sims.Load())
	return (wins + priorWeight*c.prior) / (sims + priorWeight)
}

func scoreToWinProb(score float64) float64 {
	return 1.0 / (1.0 + math.Exp(-score/36.0))
}

func (b *Bot) simulationPriorWeight() float64 {
	switch {
	case b.cfg.ThinkTime <= time.Second:
		return 8
	case b.cfg.ThinkTime <= 3*time.Second:
		return 6
	case b.cfg.ThinkTime <= 5*time.Second:
		return 5
	default:
		return 4
	}
}

func (b *Bot) replyWorkers(outcomes int) int {
	return b.replyWorkersForBudget(outcomes, b.cfg.Workers)
}

func (b *Bot) replyWorkersForBudget(outcomes, budget int) int {
	workers := budget
	if workers <= 1 {
		return 1
	}
	if workers > 4 {
		workers = 4
	}
	if workers > outcomes {
		workers = outcomes
	}
	if workers < 1 {
		return 1
	}
	return workers
}

func (b *Bot) deepEvalParallelism(deepK int) (int, int) {
	if deepK <= 1 {
		return 1, b.replyWorkers(21)
	}

	totalBudget := b.cfg.Workers
	if totalBudget <= 1 {
		return 1, 1
	}

	candidateWorkers := totalBudget / 2
	if candidateWorkers < 1 {
		candidateWorkers = 1
	}
	if candidateWorkers > 3 {
		candidateWorkers = 3
	}
	if candidateWorkers > deepK {
		candidateWorkers = deepK
	}

	replyWorkers := totalBudget / candidateWorkers
	if replyWorkers < 1 {
		replyWorkers = 1
	}
	replyWorkers = b.replyWorkersForBudget(21, replyWorkers)

	return candidateWorkers, replyWorkers
}

func (b *Bot) Evaluate(state engine.GameState, perspective engine.Color) float64 {
	if b == nil {
		return EvaluatePosition(state, perspective)
	}
	return b.scorePosition(state, perspective)
}

func (b *Bot) generateLines(state engine.GameState, d1, d2 int) ([]engine.TurnLine, error) {
	key := fmt.Sprintf("%s|%d:%d", state.NormalizeKey(), d1, d2)
	if cached, ok := b.lineCache.Get(key); ok {
		return cached, nil
	}
	lines, err := engine.GenerateLegalLines(state, d1, d2)
	if err != nil {
		return nil, err
	}
	b.lineCache.Set(key, lines)
	return lines, nil
}

func weightedDiceOutcomes() []weightedDice {
	outcomes := make([]weightedDice, 0, 21)
	for d1 := 1; d1 <= 6; d1++ {
		for d2 := d1; d2 <= 6; d2++ {
			weight := 2.0
			if d1 == d2 {
				weight = 1.0
			}
			outcomes = append(outcomes, weightedDice{d1: d1, d2: d2, weight: weight})
		}
	}
	return outcomes
}

func isOpeningPosition(state engine.GameState) bool {
	return state.Meta.MoveNumber <= 1
}

func (b *Bot) lookupOpeningBook(state engine.GameState, d1, d2 int) (string, bool) {
	if !isOpeningPosition(state) {
		return "", false
	}
	key := fmt.Sprintf("%s|opening|%d:%d", state.NormalizeKey(), d1, d2)
	if cached, ok := b.openingBook.Get(key); ok {
		return cached, true
	}
	return "", false
}

func (b *Bot) storeOpeningBook(state engine.GameState, d1, d2 int, line engine.TurnLine) error {
	if !isOpeningPosition(state) || len(line.Moves) == 0 {
		return nil
	}
	key := fmt.Sprintf("%s|opening|%d:%d", state.NormalizeKey(), d1, d2)
	b.openingBook.Set(key, line.Key())
	return nil
}

func (b *Bot) scorePosition(state engine.GameState, perspective engine.Color) float64 {
	cacheKey := scoreCacheKey(state, perspective)
	if cached, ok := b.scoreCache.Get(cacheKey); ok {
		return cached
	}
	score := b.cfg.Evaluator.Evaluate(state, perspective)
	b.scoreCache.Set(cacheKey, score)
	return score
}

func scoreCacheKey(state engine.GameState, perspective engine.Color) string {
	return state.NormalizeKey() + "|" + perspective.String()
}

func replyEquityCacheKey(state engine.GameState, perspective engine.Color) string {
	return state.NormalizeKey() + "|reply|" + perspective.String()
}

func replyBestCacheKey(state engine.GameState, perspective engine.Color, d1, d2 int) string {
	return fmt.Sprintf("%s|bestreply|%s|%d:%d", state.NormalizeKey(), perspective.String(), d1, d2)
}

func openingBias(before, after engine.GameState, line engine.TurnLine) float64 {
	if before.Meta.MoveNumber > 1 {
		return 0
	}

	player := before.Turn
	score := 0.0

	switch before.GameType {
	case engine.GameShort:
		score += float64(engine.HomeMadePoints(after, player)-engine.HomeMadePoints(before, player)) * 6.0
		score += float64(engine.CountMadePoints(after, player)-engine.CountMadePoints(before, player)) * 3.0
		score += float64(engine.CountAnchorsShort(after, player)-engine.CountAnchorsShort(before, player)) * 2.2
		score += float64(engine.CountEscapedCheckersShort(after, player)-engine.CountEscapedCheckersShort(before, player)) * 1.6
		score -= float64(engine.CountExposedBlotsShort(after, player)-engine.CountExposedBlotsShort(before, player)) * 3.0
		score -= float64(engine.CountBlots(after, player)-engine.CountBlots(before, player)) * 1.2
	case engine.GameLong:
		score += float64(engine.CountMadePoints(after, player)-engine.CountMadePoints(before, player)) * 2.5
		score += float64(engine.MaxBlockRunLong(after, player)-engine.MaxBlockRunLong(before, player)) * 2.4
		score += float64(engine.Spread(before, player)-engine.Spread(after, player)) * 0.8
		score -= float64(engine.CheckersOnHead(after, player)-engine.CheckersOnHead(before, player)) * 1.0
		score -= float64(engine.CountSingletons(after, player)-engine.CountSingletons(before, player)) * 0.8
	}

	if len(line.Moves) == 0 {
		score -= 2.0
	}
	return score
}
