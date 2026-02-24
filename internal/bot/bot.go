package bot

import (
	"errors"
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
}

type Bot struct {
	cfg Config
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
	wins  atomic.Uint64
	sims  atomic.Uint64
}

func New(cfg Config) *Bot {
	if cfg.ThinkTime <= 0 {
		cfg.ThinkTime = 8 * time.Second
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
	return &Bot{cfg: cfg}
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

	cands := b.prepareCandidates(state, legal)
	if len(cands) == 0 {
		return Decision{}, errors.New("no candidates")
	}

	deadline := start.Add(b.cfg.ThinkTime)
	target := state.Turn
	var iter atomic.Uint64
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
				idx := iter.Add(1)
				seed := b.cfg.Seed + int64(idx)*7919
				for ci := range cands {
					rng := rand.New(rand.NewSource(seed))
					winner := b.rollout(cands[ci].state, rng)
					cands[ci].sims.Add(1)
					if winner == target {
						cands[ci].wins.Add(1)
					}
				}
				if idx%32 == 0 && canEarlyStop(cands) {
					stopped.Store(true)
					return
				}
			}
		}(i)
	}
	wg.Wait()

	evals := make([]MoveEvaluation, 0, len(cands))
	for _, c := range cands {
		sims := c.sims.Load()
		if sims == 0 {
			evals = append(evals, MoveEvaluation{Line: c.line, WinProb: 0.5, Sims: 0})
			continue
		}
		wins := c.wins.Load()
		evals = append(evals, MoveEvaluation{Line: c.line, WinProb: float64(wins) / float64(sims), Sims: sims})
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
	}, nil
}

func (b *Bot) prepareCandidates(state engine.GameState, legal []engine.TurnLine) []*candidate {
	cands := make([]*candidate, 0, len(legal))
	for _, line := range legal {
		next, err := engine.ApplyTurnLine(state, line)
		if err != nil {
			continue
		}
		score := EvaluatePosition(next, state.Turn)
		cands = append(cands, &candidate{line: line, state: next, score: score})
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].score == cands[j].score {
			return cands[i].line.Key() < cands[j].line.Key()
		}
		return cands[i].score > cands[j].score
	})
	if len(cands) > b.cfg.TopK {
		cands = cands[:b.cfg.TopK]
	}
	return cands
}

func (b *Bot) rollout(state engine.GameState, rng *rand.Rand) engine.Color {
	cur := state.Clone()
	for ply := 0; ply < b.cfg.MaxPlies; ply++ {
		if cur.IsTerminal() {
			return cur.Winner()
		}
		d1 := rng.Intn(6) + 1
		d2 := rng.Intn(6) + 1
		lines, err := engine.GenerateLegalLines(cur, d1, d2)
		if err != nil {
			return cur.Turn.Opponent()
		}
		if len(lines) == 0 {
			cur.Turn = cur.Turn.Opponent()
			continue
		}
		line := selectPolicyMove(cur, lines, rng)
		next, err := engine.ApplyTurnLine(cur, line)
		if err != nil {
			return cur.Turn.Opponent()
		}
		cur = next
	}
	if EvaluatePosition(cur, cur.Turn) >= 0 {
		return cur.Turn
	}
	return cur.Turn.Opponent()
}

func selectPolicyMove(state engine.GameState, lines []engine.TurnLine, rng *rand.Rand) engine.TurnLine {
	best := lines[0]
	bestScore := math.Inf(-1)
	for _, line := range lines {
		next, err := engine.ApplyTurnLine(state, line)
		if err != nil {
			continue
		}
		s := EvaluatePosition(next, state.Turn)
		s += rng.Float64() * 0.01
		if s > bestScore {
			best = line
			bestScore = s
		}
	}
	return best
}

func canEarlyStop(cands []*candidate) bool {
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
		sims := c.sims.Load()
		if sims < 200 {
			return false
		}
		p := float64(c.wins.Load()) / float64(sims)
		margin := 1.96 * math.Sqrt((p*(1-p))/float64(sims))
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
