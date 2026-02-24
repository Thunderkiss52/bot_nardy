package bot

import "bot_nardy/internal/engine"

func EvaluatePosition(state engine.GameState, perspective engine.Color) float64 {
	if state.IsTerminal() {
		if state.Winner() == perspective {
			return 1_000_000
		}
		return -1_000_000
	}
	if state.GameType == engine.GameShort {
		s := shortScore(state, perspective)
		o := shortScore(state, perspective.Opponent())
		return s - o
	}
	s := longScore(state, perspective)
	o := longScore(state, perspective.Opponent())
	return s - o
}

func shortScore(state engine.GameState, p engine.Color) float64 {
	contactWeight := 0.35
	if engine.ContactExists(state) {
		contactWeight = 1.0
	}

	pip := float64(engine.PipCount(state, p))
	bar := float64(state.Bar[p.Idx()])
	blots := float64(engine.CountBlots(state, p))
	made := float64(engine.CountMadePoints(state, p))
	prime := float64(engine.LongestPrime(state, p))
	home := float64(engine.HomeMadePoints(state, p))
	ready := engine.BearoffReady(state, p)

	return -1.00*pip - 30.0*bar - 6.0*blots*contactWeight + 3.0*made + 4.0*prime + 2.5*home + 12.0*ready
}

func longScore(state engine.GameState, p engine.Color) float64 {
	pip := float64(engine.PipCount(state, p))
	blocks := float64(engine.CountBlocksLong(state, p))
	run := float64(engine.MaxBlockRunLong(state, p))
	spread := engine.Spread(state, p)
	singletons := float64(engine.CountSingletons(state, p))
	homeMass := float64(engine.HomeMass(state, p))

	return -1.10*pip + 2.0*blocks + 3.5*run - 1.5*spread - 1.0*singletons + 0.8*homeMass
}
