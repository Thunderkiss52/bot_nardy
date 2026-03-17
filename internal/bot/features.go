package bot

import "bot_nardy/internal/engine"

type FeatureVector struct {
	GameType         string  `json:"game_type"`
	Perspective      string  `json:"perspective"`
	Turn             string  `json:"turn"`
	PipSelf          int     `json:"pip_self"`
	PipOpp           int     `json:"pip_opp"`
	RaceLead         int     `json:"race_lead"`
	BarSelf          int     `json:"bar_self"`
	BarOpp           int     `json:"bar_opp"`
	OffSelf          int     `json:"off_self"`
	OffOpp           int     `json:"off_opp"`
	BlotsSelf        int     `json:"blots_self"`
	BlotsOpp         int     `json:"blots_opp"`
	ExposedSelf      int     `json:"exposed_self"`
	ExposedOpp       int     `json:"exposed_opp"`
	MadeSelf         int     `json:"made_self"`
	MadeOpp          int     `json:"made_opp"`
	PrimeSelf        int     `json:"prime_self"`
	PrimeOpp         int     `json:"prime_opp"`
	HomeMadeSelf     int     `json:"home_made_self"`
	HomeMadeOpp      int     `json:"home_made_opp"`
	EntryBlockSelf   int     `json:"entry_block_self"`
	EntryBlockOpp    int     `json:"entry_block_opp"`
	AnchorsSelf      int     `json:"anchors_self"`
	AnchorsOpp       int     `json:"anchors_opp"`
	ShotsSelf        int     `json:"shots_self"`
	ShotsOpp         int     `json:"shots_opp"`
	EscapedSelf      int     `json:"escaped_self"`
	EscapedOpp       int     `json:"escaped_opp"`
	BlocksSelf       int     `json:"blocks_self"`
	BlocksOpp        int     `json:"blocks_opp"`
	BlockRunSelf     int     `json:"block_run_self"`
	BlockRunOpp      int     `json:"block_run_opp"`
	SpreadSelf       float64 `json:"spread_self"`
	SpreadOpp        float64 `json:"spread_opp"`
	SingletonsSelf   int     `json:"singletons_self"`
	SingletonsOpp    int     `json:"singletons_opp"`
	HomeMassSelf     int     `json:"home_mass_self"`
	HomeMassOpp      int     `json:"home_mass_opp"`
	HeadSelf         int     `json:"head_self"`
	HeadOpp          int     `json:"head_opp"`
	BackmostSelf     int     `json:"backmost_self"`
	BackmostOpp      int     `json:"backmost_opp"`
	OuterSelf        int     `json:"outer_self"`
	OuterOpp         int     `json:"outer_opp"`
	BearoffReadySelf float64 `json:"bearoff_ready_self"`
	BearoffReadyOpp  float64 `json:"bearoff_ready_opp"`
	Contact          bool    `json:"contact"`
}

func ExtractFeatures(state engine.GameState, perspective engine.Color) FeatureVector {
	opp := perspective.Opponent()
	return FeatureVector{
		GameType:         state.GameType.String(),
		Perspective:      perspective.String(),
		Turn:             state.Turn.String(),
		PipSelf:          engine.PipCount(state, perspective),
		PipOpp:           engine.PipCount(state, opp),
		RaceLead:         engine.RaceLead(state, perspective),
		BarSelf:          state.Bar[perspective.Idx()],
		BarOpp:           state.Bar[opp.Idx()],
		OffSelf:          state.Off[perspective.Idx()],
		OffOpp:           state.Off[opp.Idx()],
		BlotsSelf:        engine.CountBlots(state, perspective),
		BlotsOpp:         engine.CountBlots(state, opp),
		ExposedSelf:      engine.CountExposedBlotsShort(state, perspective),
		ExposedOpp:       engine.CountExposedBlotsShort(state, opp),
		MadeSelf:         engine.CountMadePoints(state, perspective),
		MadeOpp:          engine.CountMadePoints(state, opp),
		PrimeSelf:        engine.LongestPrime(state, perspective),
		PrimeOpp:         engine.LongestPrime(state, opp),
		HomeMadeSelf:     engine.HomeMadePoints(state, perspective),
		HomeMadeOpp:      engine.HomeMadePoints(state, opp),
		EntryBlockSelf:   engine.EntryBlockPoints(state, perspective),
		EntryBlockOpp:    engine.EntryBlockPoints(state, opp),
		AnchorsSelf:      engine.CountAnchorsShort(state, perspective),
		AnchorsOpp:       engine.CountAnchorsShort(state, opp),
		ShotsSelf:        engine.CountDirectShotsShort(state, perspective),
		ShotsOpp:         engine.CountDirectShotsShort(state, opp),
		EscapedSelf:      engine.CountEscapedCheckersShort(state, perspective),
		EscapedOpp:       engine.CountEscapedCheckersShort(state, opp),
		BlocksSelf:       engine.CountBlocksLong(state, perspective),
		BlocksOpp:        engine.CountBlocksLong(state, opp),
		BlockRunSelf:     engine.MaxBlockRunLong(state, perspective),
		BlockRunOpp:      engine.MaxBlockRunLong(state, opp),
		SpreadSelf:       engine.Spread(state, perspective),
		SpreadOpp:        engine.Spread(state, opp),
		SingletonsSelf:   engine.CountSingletons(state, perspective),
		SingletonsOpp:    engine.CountSingletons(state, opp),
		HomeMassSelf:     engine.HomeMass(state, perspective),
		HomeMassOpp:      engine.HomeMass(state, opp),
		HeadSelf:         engine.CheckersOnHead(state, perspective),
		HeadOpp:          engine.CheckersOnHead(state, opp),
		BackmostSelf:     engine.BackmostPoint(state, perspective),
		BackmostOpp:      engine.BackmostPoint(state, opp),
		OuterSelf:        engine.CountPointsInOuterBoard(state, perspective),
		OuterOpp:         engine.CountPointsInOuterBoard(state, opp),
		BearoffReadySelf: engine.BearoffReady(state, perspective),
		BearoffReadyOpp:  engine.BearoffReady(state, opp),
		Contact:          engine.ContactExists(state),
	}
}

func (f FeatureVector) Values() map[string]float64 {
	phaseContact := boolFloat(f.Contact)
	phaseBearoff := boolFloat(!f.Contact && (f.BearoffReadySelf > 0 || f.BearoffReadyOpp > 0 || f.OffSelf > 0 || f.OffOpp > 0))
	phaseRace := boolFloat(!f.Contact && phaseBearoff == 0)
	gameShort := boolFloat(f.GameType == "short")
	gameLong := boolFloat(f.GameType == "long")
	onTurn := boolFloat(f.Perspective == f.Turn)

	pipAdv := float64(f.PipOpp - f.PipSelf)
	barAdv := float64(f.BarOpp - f.BarSelf)
	offAdv := float64(f.OffSelf - f.OffOpp)
	blotAdv := float64(f.BlotsOpp - f.BlotsSelf)
	exposedAdv := float64(f.ExposedOpp - f.ExposedSelf)
	madeAdv := float64(f.MadeSelf - f.MadeOpp)
	primeAdv := float64(f.PrimeSelf - f.PrimeOpp)
	homeMadeAdv := float64(f.HomeMadeSelf - f.HomeMadeOpp)
	entryBlockAdv := float64(f.EntryBlockSelf - f.EntryBlockOpp)
	anchorAdv := float64(f.AnchorsSelf - f.AnchorsOpp)
	shotAdv := float64(f.ShotsSelf - f.ShotsOpp)
	escapedAdv := float64(f.EscapedSelf - f.EscapedOpp)
	blocksAdv := float64(f.BlocksSelf - f.BlocksOpp)
	blockRunAdv := float64(f.BlockRunSelf - f.BlockRunOpp)
	spreadAdv := f.SpreadOpp - f.SpreadSelf
	singletonAdv := float64(f.SingletonsOpp - f.SingletonsSelf)
	homeMassAdv := float64(f.HomeMassSelf - f.HomeMassOpp)
	headAdv := float64(f.HeadOpp - f.HeadSelf)
	backmostAdv := float64(f.BackmostOpp - f.BackmostSelf)
	outerAdv := float64(f.OuterSelf - f.OuterOpp)
	bearoffReadyAdv := f.BearoffReadySelf - f.BearoffReadyOpp

	return map[string]float64{
		"pip_self":           float64(f.PipSelf),
		"pip_opp":            float64(f.PipOpp),
		"race_lead":          float64(f.RaceLead),
		"bar_self":           float64(f.BarSelf),
		"bar_opp":            float64(f.BarOpp),
		"off_self":           float64(f.OffSelf),
		"off_opp":            float64(f.OffOpp),
		"blots_self":         float64(f.BlotsSelf),
		"blots_opp":          float64(f.BlotsOpp),
		"exposed_self":       float64(f.ExposedSelf),
		"exposed_opp":        float64(f.ExposedOpp),
		"made_self":          float64(f.MadeSelf),
		"made_opp":           float64(f.MadeOpp),
		"prime_self":         float64(f.PrimeSelf),
		"prime_opp":          float64(f.PrimeOpp),
		"home_made_self":     float64(f.HomeMadeSelf),
		"home_made_opp":      float64(f.HomeMadeOpp),
		"entry_block_self":   float64(f.EntryBlockSelf),
		"entry_block_opp":    float64(f.EntryBlockOpp),
		"anchors_self":       float64(f.AnchorsSelf),
		"anchors_opp":        float64(f.AnchorsOpp),
		"shots_self":         float64(f.ShotsSelf),
		"shots_opp":          float64(f.ShotsOpp),
		"escaped_self":       float64(f.EscapedSelf),
		"escaped_opp":        float64(f.EscapedOpp),
		"blocks_self":        float64(f.BlocksSelf),
		"blocks_opp":         float64(f.BlocksOpp),
		"block_run_self":     float64(f.BlockRunSelf),
		"block_run_opp":      float64(f.BlockRunOpp),
		"spread_self":        f.SpreadSelf,
		"spread_opp":         f.SpreadOpp,
		"singletons_self":    float64(f.SingletonsSelf),
		"singletons_opp":     float64(f.SingletonsOpp),
		"home_mass_self":     float64(f.HomeMassSelf),
		"home_mass_opp":      float64(f.HomeMassOpp),
		"head_self":          float64(f.HeadSelf),
		"head_opp":           float64(f.HeadOpp),
		"backmost_self":      float64(f.BackmostSelf),
		"backmost_opp":       float64(f.BackmostOpp),
		"outer_self":         float64(f.OuterSelf),
		"outer_opp":          float64(f.OuterOpp),
		"bearoff_ready_self": f.BearoffReadySelf,
		"bearoff_ready_opp":  f.BearoffReadyOpp,
		"contact":            boolFloat(f.Contact),
		"game_short":         gameShort,
		"game_long":          gameLong,
		"on_turn":            onTurn,
		"phase_contact":      phaseContact,
		"phase_race":         phaseRace,
		"phase_bearoff":      phaseBearoff,
		"pip_adv":            pipAdv,
		"bar_adv":            barAdv,
		"off_adv":            offAdv,
		"blot_adv":           blotAdv,
		"exposed_adv":        exposedAdv,
		"made_adv":           madeAdv,
		"prime_adv":          primeAdv,
		"home_made_adv":      homeMadeAdv,
		"entry_block_adv":    entryBlockAdv,
		"anchor_adv":         anchorAdv,
		"shot_adv":           shotAdv,
		"escaped_adv":        escapedAdv,
		"blocks_adv":         blocksAdv,
		"block_run_adv":      blockRunAdv,
		"spread_adv":         spreadAdv,
		"singleton_adv":      singletonAdv,
		"home_mass_adv":      homeMassAdv,
		"head_adv":           headAdv,
		"backmost_adv":       backmostAdv,
		"outer_adv":          outerAdv,
		"bearoff_ready_adv":  bearoffReadyAdv,
		"contact_bar_adv":    phaseContact * barAdv,
		"contact_blot_adv":   phaseContact * blotAdv,
		"contact_prime_adv":  phaseContact * primeAdv,
		"contact_anchor_adv": phaseContact * anchorAdv,
		"contact_shot_adv":   phaseContact * shotAdv,
		"contact_escape_adv": phaseContact * escapedAdv,
		"race_pip_adv":       phaseRace * pipAdv,
		"race_off_adv":       phaseRace * offAdv,
		"race_spread_adv":    phaseRace * spreadAdv,
		"race_home_mass_adv": phaseRace * homeMassAdv,
		"bearoff_off_adv":    phaseBearoff * offAdv,
		"bearoff_pip_adv":    phaseBearoff * pipAdv,
		"bearoff_home_adv":   phaseBearoff * homeMassAdv,
		"bearoff_ready_x":    phaseBearoff * bearoffReadyAdv,
		"short_anchor_adv":   gameShort * anchorAdv,
		"short_entry_adv":    gameShort * entryBlockAdv,
		"short_bar_adv":      gameShort * barAdv,
		"short_shot_adv":     gameShort * shotAdv,
		"long_blocks_adv":    gameLong * blocksAdv,
		"long_run_adv":       gameLong * blockRunAdv,
		"long_head_adv":      gameLong * headAdv,
		"long_back_adv":      gameLong * backmostAdv,
	}
}

func boolFloat(v bool) float64 {
	if v {
		return 1
	}
	return 0
}
