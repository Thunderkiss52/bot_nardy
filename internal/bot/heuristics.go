package bot

import "bot_nardy/internal/engine"

func shortScore(state engine.GameState, p engine.Color) float64 {
	pip := float64(engine.PipCount(state, p))
	bar := float64(state.Bar[p.Idx()])
	blots := float64(engine.CountBlots(state, p))
	exposed := float64(engine.CountExposedBlotsShort(state, p))
	made := float64(engine.CountMadePoints(state, p))
	prime := float64(engine.LongestPrime(state, p))
	home := float64(engine.HomeMadePoints(state, p))
	outer := float64(engine.CountPointsInOuterBoard(state, p))
	anchors := float64(engine.CountAnchorsShort(state, p))
	entryBlocks := float64(engine.EntryBlockPoints(state, p))
	shots := float64(engine.CountDirectShotsShort(state, p))
	escaped := float64(engine.CountEscapedCheckersShort(state, p))
	raceLead := float64(engine.RaceLead(state, p))
	ready := engine.BearoffReady(state, p)
	contact := engine.ContactExists(state)

	if !contact {
		return -1.18*pip + 0.72*raceLead - 1.1*blots + 1.4*made + 1.8*home + 0.8*outer + 11.0*ready
	}

	return -0.94*pip - 28.0*bar - 3.0*blots - 7.0*exposed + 2.8*made + 4.5*prime + 3.2*home +
		1.2*outer + 2.6*anchors + 2.1*entryBlocks + 1.5*shots + 0.8*escaped + 10.0*ready
}

func longScore(state engine.GameState, p engine.Color) float64 {
	pip := float64(engine.PipCount(state, p))
	blocks := float64(engine.CountBlocksLong(state, p))
	run := float64(engine.MaxBlockRunLong(state, p))
	spread := engine.Spread(state, p)
	singletons := float64(engine.CountSingletons(state, p))
	homeMass := float64(engine.HomeMass(state, p))
	headCount := float64(engine.CheckersOnHead(state, p))
	backPoint := float64(engine.BackmostPoint(state, p))
	raceLead := float64(engine.RaceLead(state, p))

	if engine.BearoffReady(state, p) > 0 {
		return -1.22*pip + 0.55*raceLead + 1.2*homeMass - 0.5*singletons
	}

	return -1.06*pip + 0.45*raceLead + 2.2*blocks + 4.0*run - 1.7*spread - 1.1*singletons +
		0.9*homeMass - 1.2*headCount - 0.18*backPoint
}
