package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"bot_nardy/internal/app"
	"bot_nardy/internal/engine"
)

func main() {
	rounds := flag.Int("rounds", 8, "number of train/validate rounds")
	selfPlayGames := flag.Int("selfplay-games", 48, "self-play games per round")
	epochs := flag.Int("epochs", 12, "training epochs per round")
	think := flag.Int("think", 2, "self-play think time seconds")
	seed := flag.Int64("seed", time.Now().UnixNano(), "seed")
	logDir := flag.String("log-dir", "", "optional dir for round logs")
	flag.Parse()

	svc, err := app.NewService("")
	if err != nil {
		panic(err)
	}
	defer svc.Close()

	if _, err := svc.Start(app.GameOptions{
		GameType:     engine.GameShort,
		BotSide:      engine.Black,
		Opponent:     app.OpponentHuman,
		ThinkTimeSec: min(max(*think, 1), 5),
		Seed:         *seed,
	}); err != nil {
		panic(err)
	}

	rng := rand.New(rand.NewSource(*seed))
	for round := 1; round <= *rounds; round++ {
		fmt.Printf("round=%d selfplay=%d epochs=%d\n", round, *selfPlayGames, *epochs)

		for i := 0; i < *selfPlayGames; i++ {
			gameType := engine.GameShort
			if i%2 == 1 {
				gameType = engine.GameLong
			}
			if _, err := svc.RunSelfPlayTrainingGame(gameType, time.Duration(*think)*time.Second, *seed+int64(round)*100000+int64(i)*101+int64(rng.Intn(97))); err != nil {
				panic(err)
			}
		}

		summary, err := svc.SelfLearn(*epochs)
		if err != nil {
			panic(err)
		}
		fmt.Printf(
			"round=%d accepted=%v champion=%s champion_score=%.2f validation=%.3f league=%d examples=%d err=%.4f\n",
			round,
			summary.Accepted,
			summary.ChampionModel,
			summary.ChampionScore,
			summary.ValidationWinRate,
			summary.LeagueSize,
			summary.Examples,
			summary.AvgAbsError,
		)

		if *logDir != "" {
			if err := os.MkdirAll(*logDir, 0o755); err != nil {
				panic(err)
			}
			path := filepath.Join(*logDir, fmt.Sprintf("round_%02d.txt", round))
			content := fmt.Sprintf(
				"round=%d accepted=%v champion=%s champion_score=%.2f validation=%.3f league=%d examples=%d err=%.4f\n",
				round,
				summary.Accepted,
				summary.ChampionModel,
				summary.ChampionScore,
				summary.ValidationWinRate,
				summary.LeagueSize,
				summary.Examples,
				summary.AvgAbsError,
			)
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				panic(err)
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
