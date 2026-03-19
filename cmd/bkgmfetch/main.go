package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"bot_nardy/internal/bkgm"
)

func main() {
	out := flag.String("out", "bkgm_corpus.jsonl", "output JSONL path")
	maxPages := flag.Int("max-pages", 200, "maximum number of pages to crawl")
	delay := flag.Duration("delay", 300*time.Millisecond, "delay between requests")
	seedsFlag := flag.String("seeds", "", "comma-separated seed URLs; defaults to BKGM articles/matches/glossary")
	flag.Parse()

	seeds := bkgm.DefaultSeeds()
	if strings.TrimSpace(*seedsFlag) != "" {
		seeds = splitSeeds(*seedsFlag)
	}

	crawler := bkgm.NewCrawler(bkgm.Config{
		Seeds:    seeds,
		MaxPages: *maxPages,
		Delay:    *delay,
	})
	count, err := crawler.CrawlToJSONL(*out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bkgm fetch failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("bkgm fetch successful: wrote %d documents to %s\n", count, *out)
}

func splitSeeds(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
