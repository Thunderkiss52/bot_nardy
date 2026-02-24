# Desktop Nardy Engine (Short + Long)

MVP backend implementation in Go with:

- Deterministic game engine for `short` (backgammon) and `long` nardy.
- Legal move generation with dice ordering, max dice usage, bar entry (short), head rule (long), and long blockade validation.
- Monte-Carlo rollout bot with configurable think-time, parallel workers, top-K candidate filtering, CRN-style seeding, and early-stop by confidence intervals.
- Analysis mode primitives (`exact/inaccuracy/mistake/blunder` + delta).
- JSONL move logging.
- CLI (`cmd/nardy`) for manual dice input and validation.
- Self-play benchmark CLI (`cmd/selfplay`) to estimate strong-vs-baseline win rate.
- Desktop shell on Wails (`cmd/desktop`) with board, manual dice, Top-3, analysis, undo, and export.
- Quality runner (`cmd/quality`) with winrate + AEL + blunder metrics.

## Build

```bash
go test ./...
go run ./cmd/nardy -game short -bot black -opponent human -think 8
go run ./cmd/selfplay -game short -n 200 -strong-think 4 -base-think 1
go run ./cmd/quality -game both -n 300 -strong-think 800ms -baseline-think 120ms

# desktop UI (requires Wails SDK + deps)
go run -tags wails ./cmd/desktop

# macOS build (run on macOS host)
GOCACHE=/tmp/go-cache go run github.com/wailsapp/wails/v2/cmd/wails@v2.11.0 build -tags wails -platform darwin/universal -clean
```

## Project structure

- `internal/engine`: state model, rules, move generation, invariants.
- `internal/bot`: heuristics and rollout search.
- `internal/app`: high-level game service APIs (ready for Wails bindings).
- `internal/desktop`: Wails-facing API bridge.
- `internal/logging`: JSONL move logger.
- `cmd/nardy`: manual play loop for MVP verification.
- `cmd/selfplay`: benchmark runner.
- `cmd/quality`: quality metrics runner.
- `cmd/desktop`: desktop app entrypoint (`main_stub.go` by default, `main_wails.go` with `-tags wails`).
- `cmd/desktop/frontend`: static desktop UI assets.

## Notes

Desktop mode uses local static frontend assets with Wails bindings (`window.go.desktop.API`).
`cmd/quality` is intended for threshold tracking from the technical spec.
Cross-compiling Wails desktop to macOS from Linux is not supported by Wails. Use a macOS host or `.github/workflows/build-macos-desktop.yml`.
