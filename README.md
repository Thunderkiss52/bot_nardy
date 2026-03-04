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

# local macOS helper script
./scripts/build-macos.sh
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

For GitLab CI use `.gitlab-ci.yml` with a macOS runner tagged `macos`. Build artifact is uploaded as `build/bin/desktop-nardy-engine-macos.zip`.
For GitHub Actions (macOS signing/notarization), configure repository secrets:

- `MACOS_CERT_P12_BASE64` (base64 of Developer ID Application `.p12`)
- `MACOS_CERT_PASSWORD` (password for `.p12`)
- `MACOS_SIGN_IDENTITY` (example: `Developer ID Application: Your Name (TEAMID)`)
- `APPLE_ID`
- `APPLE_APP_SPECIFIC_PASSWORD`
- `APPLE_TEAM_ID`

### GitLab Artifact Download (macOS)

1. Push branch to GitLab and wait for job `macos:build` to finish.
2. Open `CI/CD -> Pipelines -> <pipeline> -> macos:build`.
3. Download artifact `desktop-nardy-engine-macos-<sha>.zip`.

You can also download artifact by API:

```bash
curl --location \
  --header "PRIVATE-TOKEN: <YOUR_GITLAB_TOKEN>" \
  "https://<YOUR_GITLAB_HOST>/api/v4/projects/<PROJECT_ID>/jobs/artifacts/<BRANCH>/download?job=macos:build" \
  --output desktop-nardy-engine-macos.zip
```

Then on macOS:

```bash
unzip desktop-nardy-engine-macos.zip
open desktop-nardy-engine.app
```

If `spctl` shows `source=no signature`, build was not Developer ID signed/notarized.
For production-ready build in CI set:

- `MACOS_SIGN_IDENTITY` (example: `Developer ID Application: Your Name (TEAMID)`)
- `MACOS_NOTARY_PROFILE` (keychain profile name for `xcrun notarytool`)
