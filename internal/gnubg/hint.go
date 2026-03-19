package gnubg

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
)

type Teacher struct {
	Binary  string
	Timeout time.Duration
}

type HintMove struct {
	Line     engine.TurnLine
	Notation string
	WinProb  float64
	Equity   float64
}

var hintLineRE = regexp.MustCompile(`^\s*(?:\d+\.\s+)?([0-9.]+)\s+([0-9.]+)\s+([0-9.]+)\s+([0-9.]+)\s+([0-9.]+)\s+\(([+-]?[0-9.]+)\)\s+(.+?)\s*$`)

func (t Teacher) Analyze(state engine.GameState, d1, d2, topN int) ([]HintMove, error) {
	if topN <= 0 {
		topN = 3
	}
	if strings.TrimSpace(t.Binary) == "" {
		t.Binary = "gnubg"
	}
	if t.Timeout <= 0 {
		t.Timeout = 20 * time.Second
	}
	if state.GameType != engine.GameShort {
		return nil, fmt.Errorf("gnubg teacher currently supports short only")
	}
	if err := state.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()

	script := buildHintScript(state, d1, d2)
	cmd := exec.CommandContext(ctx, t.Binary, "-t")
	cmd.Stdin = strings.NewReader(script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gnubg failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	hints, err := ParseHintOutput(string(out))
	if err != nil {
		return nil, err
	}
	legal, err := engine.GenerateLegalLines(state, d1, d2)
	if err != nil {
		return nil, err
	}

	matched := make([]HintMove, 0, topN)
	for _, parsed := range hints {
		line, ok := MatchHintToLegal(state, legal, parsed.Notation)
		if !ok {
			continue
		}
		matched = append(matched, HintMove{
			Line:     line,
			Notation: parsed.Notation,
			WinProb:  parsed.WinProb,
			Equity:   parsed.Equity,
		})
		if len(matched) >= topN {
			break
		}
	}
	if len(matched) == 0 {
		return nil, fmt.Errorf("no gnubg hints matched legal lines")
	}
	return matched, nil
}

func buildHintScript(state engine.GameState, d1, d2 int) string {
	var b strings.Builder
	b.WriteString("new session\n")
	b.WriteString("set player both human\n")
	b.WriteString("set turn X\n")
	b.WriteString("set board ")
	b.WriteString(EncodePositionID(state))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("set dice %d %d\n", d1, d2))
	b.WriteString("hint\n")
	b.WriteString("quit\n")
	return b.String()
}

type parsedHint struct {
	Notation string
	WinProb  float64
	Equity   float64
}

func ParseHintOutput(output string) ([]parsedHint, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	hints := make([]parsedHint, 0, 8)
	for scanner.Scan() {
		line := scanner.Text()
		m := hintLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		winProb, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return nil, err
		}
		equity, err := strconv.ParseFloat(m[6], 64)
		if err != nil {
			return nil, err
		}
		hints = append(hints, parsedHint{
			Notation: cleanNotation(m[7]),
			WinProb:  winProb,
			Equity:   equity,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(hints) == 0 {
		return nil, fmt.Errorf("no hint lines found in gnubg output")
	}
	return hints, nil
}

func MatchHintToLegal(state engine.GameState, legal []engine.TurnLine, notation string) (engine.TurnLine, bool) {
	want := normalizeNotationTokens(notation)
	for _, line := range legal {
		got := normalizeNotationTokens(FormatNotation(state, line))
		if len(got) != len(want) {
			continue
		}
		match := true
		for i := range want {
			if got[i] != want[i] {
				match = false
				break
			}
		}
		if match {
			return line, true
		}
	}
	return engine.TurnLine{}, false
}

func FormatNotation(state engine.GameState, line engine.TurnLine) string {
	if len(line.Moves) == 0 {
		return "pass"
	}
	tokens := make([]string, 0, len(line.Moves))
	tmp := state.Clone()
	for _, mv := range line.Moves {
		token := fromLabel(tmp.Turn, mv.From) + "/" + toLabel(tmp.Turn, mv.To)
		if isHit(tmp, mv) {
			token += "*"
		}
		tokens = append(tokens, token)
		_ = applyForNotation(&tmp, mv)
	}
	return strings.Join(tokens, " ")
}

func applyForNotation(state *engine.GameState, mv engine.Move) error {
	next, err := engine.ApplyTurnLine(*state, engine.TurnLine{Moves: []engine.Move{mv}})
	if err != nil {
		return err
	}
	next.Turn = state.Turn
	next.Meta = state.Meta
	*state = next
	return nil
}

func isHit(state engine.GameState, mv engine.Move) bool {
	if state.GameType != engine.GameShort || mv.To == 0 {
		return false
	}
	pt := state.Points[mv.To]
	return pt.Owner == state.Turn.Opponent() && pt.Count == 1
}

func fromLabel(player engine.Color, point int) string {
	if point == 0 {
		return "bar"
	}
	return relativeLabel(player, point)
}

func toLabel(player engine.Color, point int) string {
	if point == 0 {
		return "off"
	}
	return relativeLabel(player, point)
}

func relativeLabel(player engine.Color, point int) string {
	if point < 1 || point > 24 {
		return strconv.Itoa(point)
	}
	rel := point
	if player == engine.Black {
		rel = 25 - point
	}
	return strconv.Itoa(rel)
}

func normalizeNotationTokens(notation string) []string {
	notation = strings.ToLower(cleanNotation(notation))
	if notation == "" || notation == "pass" {
		return nil
	}
	notation = strings.NewReplacer(",", " ", "\t", " ", "*", "").Replace(notation)
	fields := strings.Fields(notation)
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		count := 1
		if idx := strings.Index(field, "("); idx > 0 && strings.HasSuffix(field, ")") {
			n, err := strconv.Atoi(field[idx+1 : len(field)-1])
			if err == nil && n > 0 {
				count = n
				field = field[:idx]
			}
		}
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		for i := 0; i < count; i++ {
			out = append(out, field)
		}
	}
	sort.Strings(out)
	return out
}

func cleanNotation(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func HintMovesToEvaluations(hints []HintMove) []bot.MoveEvaluation {
	evals := make([]bot.MoveEvaluation, 0, len(hints))
	for _, hint := range hints {
		evals = append(evals, bot.MoveEvaluation{
			Line:    hint.Line,
			WinProb: hint.WinProb,
			Sims:    0,
		})
	}
	return evals
}
