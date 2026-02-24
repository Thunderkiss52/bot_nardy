package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"bot_nardy/internal/app"
	"bot_nardy/internal/engine"
)

func main() {
	game := flag.String("game", "short", "short|long")
	botSide := flag.String("bot", "black", "white|black")
	opponent := flag.String("opponent", "human", "human|bot")
	think := flag.Int("think", 8, "think time seconds (1..20)")
	seed := flag.Int64("seed", 0, "optional random seed")
	logPath := flag.String("log", "moves.jsonl", "path for JSONL move logs")
	flag.Parse()

	svc, err := app.NewService(*logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "service init failed: %v\n", err)
		os.Exit(1)
	}
	defer svc.Close()

	opts := app.GameOptions{
		GameType:     parseGameType(*game),
		BotSide:      parseColor(*botSide),
		Opponent:     app.OpponentMode(*opponent),
		ThinkTimeSec: *think,
		ShowTop3:     true,
		Seed:         *seed,
	}
	state, err := svc.Start(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("started %s game. turn=%s\n", state.GameType, state.Turn)
	scan := bufio.NewScanner(os.Stdin)
	for {
		if state.IsTerminal() {
			fmt.Printf("winner: %s\n", state.Winner())
			return
		}

		fmt.Print("dice (d1 d2), undo, export <path>, quit> ")
		if !scan.Scan() {
			return
		}
		line := strings.TrimSpace(scan.Text())
		if line == "quit" {
			return
		}
		if line == "undo" {
			state, err = svc.Undo()
			if err != nil {
				fmt.Printf("undo error: %v\n", err)
			}
			continue
		}
		if strings.HasPrefix(line, "export ") {
			path := strings.TrimSpace(strings.TrimPrefix(line, "export "))
			if err := svc.ExportState(path); err != nil {
				fmt.Printf("export error: %v\n", err)
			} else {
				fmt.Printf("exported to %s\n", path)
			}
			continue
		}

		d1, d2, err := parseDice(line)
		if err != nil {
			fmt.Printf("input error: %v\n", err)
			continue
		}

		if isBotTurn(opts, state.Turn) {
			decision, next, err := svc.BotMove(d1, d2)
			if err != nil {
				fmt.Printf("bot error: %v\n", err)
				continue
			}
			sims := uint64(0)
			if len(decision.Top3) > 0 {
				sims = decision.Top3[0].Sims
			}
			fmt.Printf("bot: %s (p=%.3f sims=%d)\n", decision.ChosenLine, decision.ChosenProb, sims)
			for i, ev := range decision.Top3 {
				fmt.Printf("  #%d %s -> %.3f (%d)\n", i+1, ev.Line, ev.WinProb, ev.Sims)
			}
			state = next
			continue
		}

		legal, err := svc.LegalLines(d1, d2)
		if err != nil {
			fmt.Printf("legal gen error: %v\n", err)
			continue
		}
		if len(legal) == 0 {
			fmt.Println("no legal moves, pass")
			next, err := svc.PassTurnIfNoLegal(d1, d2)
			if err != nil {
				fmt.Printf("pass error: %v\n", err)
				continue
			}
			state = next
			continue
		}
		for i, l := range legal {
			fmt.Printf("%d: %s\n", i, l)
		}
		fmt.Print("choose line index> ")
		if !scan.Scan() {
			return
		}
		idx, err := strconv.Atoi(strings.TrimSpace(scan.Text()))
		if err != nil || idx < 0 || idx >= len(legal) {
			fmt.Println("invalid index")
			continue
		}
		next, err := svc.ApplyHumanLine(d1, d2, legal[idx])
		if err != nil {
			fmt.Printf("apply error: %v\n", err)
			continue
		}
		analysis, err := svc.AnalyzeLine(d1, d2, legal[idx])
		if err == nil {
			fmt.Printf("analysis: %s delta=%.4f best=%s\n", analysis.Category, analysis.Delta, analysis.BestLine)
		}
		state = next
	}
}

func isBotTurn(opts app.GameOptions, turn engine.Color) bool {
	if opts.Opponent == app.OpponentBot {
		return true
	}
	return opts.BotSide == turn
}

func parseDice(line string) (int, int, error) {
	parts := strings.Fields(line)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected two dice")
	}
	d1, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	d2, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	if d1 < 1 || d1 > 6 || d2 < 1 || d2 > 6 {
		return 0, 0, fmt.Errorf("dice must be 1..6")
	}
	return d1, d2, nil
}

func parseGameType(s string) engine.GameType {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "long" {
		return engine.GameLong
	}
	return engine.GameShort
}

func parseColor(s string) engine.Color {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "white" {
		return engine.White
	}
	return engine.Black
}
