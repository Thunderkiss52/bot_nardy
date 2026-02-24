package engine

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type GameType int

const (
	GameShort GameType = iota + 1
	GameLong
)

func (g GameType) String() string {
	switch g {
	case GameShort:
		return "short"
	case GameLong:
		return "long"
	default:
		return "unknown"
	}
}

type Color int

const (
	NoColor Color = iota
	White
	Black
)

func (c Color) String() string {
	switch c {
	case White:
		return "white"
	case Black:
		return "black"
	default:
		return "none"
	}
}

func (c Color) Opponent() Color {
	switch c {
	case White:
		return Black
	case Black:
		return White
	default:
		return NoColor
	}
}

func (c Color) Idx() int {
	if c == White {
		return 0
	}
	return 1
}

func (c Color) Direction() int {
	if c == White {
		return -1
	}
	return 1
}

type Point struct {
	Owner Color `json:"owner"`
	Count int   `json:"count"`
}

type GameState struct {
	GameType GameType  `json:"game_type"`
	Points   [25]Point `json:"points"`
	Off      [2]int    `json:"off"`
	Bar      [2]int    `json:"bar"`
	Turn     Color     `json:"turn"`
	Seed     int64     `json:"seed,omitempty"`
	Meta     StateMeta `json:"meta"`
}

type StateMeta struct {
	MoveNumber int `json:"move_number"`
}

type Move struct {
	From int `json:"from"`
	To   int `json:"to"`
	Die  int `json:"die"`
}

type TurnLine struct {
	Moves []Move `json:"moves"`
}

func (l TurnLine) DiceUsed() int {
	return len(l.Moves)
}

func (l TurnLine) Key() string {
	parts := make([]string, 0, len(l.Moves))
	for _, mv := range l.Moves {
		parts = append(parts, fmt.Sprintf("%d>%d:%d", mv.From, mv.To, mv.Die))
	}
	return strings.Join(parts, "|")
}

func (l TurnLine) String() string {
	if len(l.Moves) == 0 {
		return "pass"
	}
	parts := make([]string, 0, len(l.Moves))
	for _, mv := range l.Moves {
		parts = append(parts, fmt.Sprintf("%d/%d(%d)", mv.From, mv.To, mv.Die))
	}
	return strings.Join(parts, " ")
}

func (s GameState) Clone() GameState {
	return s
}

func (s GameState) Winner() Color {
	if s.Off[White.Idx()] >= 15 {
		return White
	}
	if s.Off[Black.Idx()] >= 15 {
		return Black
	}
	return NoColor
}

func (s GameState) IsTerminal() bool {
	return s.Winner() != NoColor
}

func (s GameState) NormalizeKey() string {
	var b strings.Builder
	b.WriteString(s.GameType.String())
	b.WriteString("|")
	b.WriteString(s.Turn.String())
	b.WriteString("|")
	for p := 1; p <= 24; p++ {
		pt := s.Points[p]
		if pt.Count == 0 {
			b.WriteString("0,")
			continue
		}
		if pt.Owner == White {
			b.WriteString("W")
		} else {
			b.WriteString("B")
		}
		b.WriteString(strconv.Itoa(pt.Count))
		b.WriteString(",")
	}
	b.WriteString("|")
	b.WriteString(strconv.Itoa(s.Bar[White.Idx()]))
	b.WriteString(",")
	b.WriteString(strconv.Itoa(s.Bar[Black.Idx()]))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(s.Off[White.Idx()]))
	b.WriteString(",")
	b.WriteString(strconv.Itoa(s.Off[Black.Idx()]))
	return b.String()
}

func (s GameState) Validate() error {
	if s.GameType != GameShort && s.GameType != GameLong {
		return errors.New("invalid game type")
	}
	if s.Turn != White && s.Turn != Black {
		return errors.New("invalid turn")
	}
	whiteCount := s.Off[White.Idx()] + s.Bar[White.Idx()]
	blackCount := s.Off[Black.Idx()] + s.Bar[Black.Idx()]
	for p := 1; p <= 24; p++ {
		pt := s.Points[p]
		if pt.Count < 0 {
			return fmt.Errorf("negative checker count at point %d", p)
		}
		if pt.Count == 0 {
			continue
		}
		if pt.Owner == White {
			whiteCount += pt.Count
		} else if pt.Owner == Black {
			blackCount += pt.Count
		} else {
			return fmt.Errorf("point %d has count but no owner", p)
		}
	}
	if whiteCount != 15 || blackCount != 15 {
		return fmt.Errorf("invariant broken: white=%d black=%d", whiteCount, blackCount)
	}
	if s.GameType == GameLong {
		if s.Bar[White.Idx()] != 0 || s.Bar[Black.Idx()] != 0 {
			return errors.New("long game cannot have bar checkers")
		}
	}
	return nil
}

func NewShortGame(seed int64) GameState {
	s := GameState{GameType: GameShort, Turn: White, Seed: seed}
	setPoint(&s, 24, White, 2)
	setPoint(&s, 13, White, 5)
	setPoint(&s, 8, White, 3)
	setPoint(&s, 6, White, 5)

	setPoint(&s, 1, Black, 2)
	setPoint(&s, 12, Black, 5)
	setPoint(&s, 17, Black, 3)
	setPoint(&s, 19, Black, 5)
	return s
}

func NewLongGame(seed int64) GameState {
	s := GameState{GameType: GameLong, Turn: White, Seed: seed}
	setPoint(&s, 24, White, 15)
	setPoint(&s, 1, Black, 15)
	return s
}

func setPoint(s *GameState, point int, owner Color, count int) {
	s.Points[point] = Point{Owner: owner, Count: count}
}

func HomeRange(c Color) (int, int) {
	if c == White {
		return 1, 6
	}
	return 19, 24
}

func HeadPoint(c Color) int {
	if c == White {
		return 24
	}
	return 1
}

func EntryPointFromBar(c Color, die int) int {
	if c == White {
		return 25 - die
	}
	return die
}

func DistanceToBearOff(c Color, point int) int {
	if c == White {
		return point
	}
	return 25 - point
}

func SortedPointsForPlayer(c Color) []int {
	pts := make([]int, 24)
	for i := range pts {
		pts[i] = i + 1
	}
	if c == White {
		sort.Slice(pts, func(i, j int) bool { return pts[i] > pts[j] })
	}
	return pts
}
