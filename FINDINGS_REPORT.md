# Code Review Findings — Bottesmo

**Date:** 2026-07-09
**Scope:** Comprehensive quality/sanity review of all Go source code, unit tests, frontend (JS/CSS/templates), specs, and deployment artifacts.

---

## Test Results Summary

| Test Suite | Result |
|---|---|
| `go test ./...` (52 tests) | ✅ All passed |
| Playwright E2E (33 tests) | ⚠️ 29 passed, 4 failed |

The 4 Playwright failures are pre-existing test bugs (see below). No application logic failures were detected.

---

## Critical / High Severity Findings

### H-1: No graceful shutdown in `cmd/server/main.go` (Severity: High, Category: Bug)

**Location:** `cmd/server/main.go:76`

**Description:** `srv.ListenAndServe()` is called without any signal handling (`SIGINT`/`SIGTERM`). The server terminates immediately on `SIGKILL`, dropping active WebSocket/SSE connections and in-memory game state.

**Recommendation:** Add `signal.Notify` + `srv.Shutdown(ctx)` in a goroutine to enable graceful termination with a configurable drain timeout.

---

### H-2: SSE write deadline disabled, enabling indefinite goroutine leaks (Severity: High, Category: Bug)

**Location:** `internal/handlers/multiplayer.go:547`

**Description:** `SetWriteDeadline(time.Time{})` disables the write deadline on SSE responses. If an SSE client pauses reading, the server goroutine blocks indefinitely writing to the TCP buffer. A slow-client DoS attack could exhaust server goroutines.

**Recommendation:** Set a reasonable write deadline (e.g. 30s) and handle timeout errors to clean up the SSE goroutine. Alternatively, implement a heartbeat mechanism.

---

### H-3: `crypto/rand.Read` error silently ignored in ID generation (Severity: High, Category: Security)

**Location:** `internal/handlers/game.go:63-65`

**Description:** `crypto/rand.Read(b)` returns `(n int, err error)` but both return values are discarded. If `crypto/rand.Read` fails (e.g. entropy pool exhausted on early boot / container startup), the session/player ID will be all zeros, making multiple sessions share the same ID.

**Recommendation:** At minimum `log.Fatal(err)` or `panic(err)` — a game server cannot operate with predictable IDs.

---

### H-4: SSE `onerror` reconnects unconditionally after `leaveRoom()` (Severity: High, Category: Bug)

**Location:** `web/static/multiplayer.js:300-302`

**Description:** When `leaveRoom()` calls `mp.eventSource.close()`, this triggers the `onerror` handler, which calls `setupSSE(roomCode, playerID)` again after 3 seconds — reconnecting to a room the player has already left. This creates a zombie SSE connection that leaks server-side goroutines.

**Recommendation:** Check a `leftRoom` flag in `onerror` before reconnecting.

---

### H-5: Auth token validation bypassed on most multiplayer endpoints (Severity: High, Category: Security)

**Location:** `internal/handlers/multiplayer.go` — `MultiGuessHandler`, `LeaveRoomHandler`, etc.

**Description:** Only `RestartGameHandler` validates the player token. `MultiGuessHandler`, `LeaveRoomHandler`, and others accept any `playerID` without verifying the caller owns that ID. An attacker who knows a valid `roomCode` and `playerID` can submit guesses or leave on behalf of other players.

**Recommendation:** Validate the player token on all mutating endpoints. The token should be sent as a header (not body param) to avoid leaking in server logs.

---

### H-6: Spec mismatch — API routes documented with `/room/` prefix don't match implementation (Severity: High, Category: Documentation)

**Location:** `specs/api.md:26-31`

**Description:** The spec documents routes like `/api/multiplayer/room/create`, `/api/multiplayer/room/join`, `/api/multiplayer/room/start`, `/api/multiplayer/room/leave`, `/api/multiplayer/room/restart`. The actual implementation uses `/api/multiplayer/create`, `/api/multiplayer/join`, `/api/multiplayer/start`, `/api/multiplayer/leave`, `/api/multiplayer/restart` (without `/room/`).

**Recommendation:** Update `specs/api.md` to match the actual routes.

---

## Medium Severity Findings

### M-1: `words.txt` only contains 11 words vs 50 documented in specs (Severity: Medium, Category: Documentation)

**Location:** `specs/dictionary.md:9`, `internal/dictionary/words.txt`

**Description:** The spec says words.txt has "50 mots" but the actual file has 11 words (April 2025 weekly batch). This severely limits game variety in solo/daily modes.

**Recommendation:** Either update the spec to reflect reality, or expand the dictionary.

---

### M-2: `GenerateWordSequence` silently fills with `"AAAAAA"` when dictionary is empty (Severity: Medium, Category: Bug)

**Location:** `internal/game/multiplayer.go:86-93`

**Description:** If `RandomWord` fails 10 times consecutively, the function inserts `"AAAAAA"` as a placeholder word. `"AAAAAA"` is not in the dictionary, so players can never guess it (first-letter validation would pass, but `IsValid("AAAAAA")` returns false). The function never returns an error, so the caller has no way to detect this failure.

**Recommendation:** If the dictionary is empty/misconfigured, return an error early rather than producing unplayable words.

---

### M-3: Spec says client-side validation via Typo.js, but server actually validates (Severity: Medium, Category: Documentation)

**Location:** `specs/dictionary.md:39`

**Description:** The spec states "La validation orthographique des mots saisis par le joueur est effectuée côté client via Typo.js. Le serveur ne valide que la longueur, la première lettre et la correspondance." In reality, both `GuessHandler` and `MultiGuessHandler` call `dictionary.IsValid()` for server-side validation.

**Recommendation:** Update the spec to reflect actual server-side validation behavior.

---

### M-4: Playwright tests use `page.request()` which is not a valid API (Severity: Medium, Category: Test Gap)

**Location:** `bottesmo.spec.js:168,276`

**Description:** Tests 7 and the SSE content-type test use `page.request().post(...)`. In Playwright, the fixture is `request` (available as a test parameter), not `page.request()`. This causes `TypeError: page.request is not a function`.

Additionally, Test 1 assumes the `firstLetter` from `beforeAll` is valid for the game created on page navigation, which is not guaranteed — each page load creates a new random game with a different target word.

**Recommendation:** Use the `request` fixture directly and fix the first-letter assumption. Tests 1 and 7 are unreliable as written.

---

### M-5: Hardcoded BASE URL in spec files ignores Playwright config (Severity: Medium, Category: Style)

**Location:** `bottesmo.spec.js:3`, `multiplayer_restart.spec.js:3`

**Description:** Both spec files hardcode `const BASE = 'http://localhost:3118'` rather than using the Playwright config's `baseURL`. If the config changes, tests silently break. The `baseURL` from the config is already set to 3130.

**Recommendation:** Remove the hardcoded `BASE` and use the baseURL from config (prepend `/api/...` to paths).

---

### M-6: `page.request()` vs `request` fixture (same as M-4, additional instances) (Severity: Medium, Category: Test Gap)

**Location:** `bottesmo.spec.js:168,276`

**Description:** The same `page.request()` error appears in two tests. Additionally, these tests directly call API endpoints without going through UI pages, so they should use the `request` fixture or `apiRequestContext` instead.

**Recommendation:** Fix all occurrences of `page.request()` to use the `request` fixture.

---

### M-7: `AGENTS.md` shows stale version `0.1.0` (Severity: Medium, Category: Documentation)

**Location:** `AGENTS.md:10`

**Description:** The AGENTS.md example references `Version = "0.1.0"`, but `internal/version/version.go` actually contains `0.2.0`. This is only an example in documentation, but it may confuse automated tools.

**Recommendation:** Update the example to match the actual version.

---

### M-8: `README.md` doesn't mention the multiplayer mode (Severity: Medium, Category: Documentation)

**Location:** `README.md:5-8`

**Description:** The README only documents "Mot du Jour" and "Solo" modes. The multiplayer mode is not mentioned at all, despite being a major feature of the application.

**Recommendation:** Add a multiplayer section to `README.md` with basic usage instructions.

---

### M-9: `.bottega/project.md` missing multiplayer files from tree (Severity: Medium, Category: Documentation)

**Location:** `.bottega/project.md:6-31`

**Description:** The architecture tree doesn't list `internal/handlers/multiplayer.go`, `internal/game/multiplayer.go`, `web/static/multiplayer.js`, or `web/templates/multiplayer.html`.

**Recommendation:** Update the file tree to include all multiplayer components.

---

### M-10: `TestAddPlayerMaxPlayers` uses `rune('0'+i)` for player IDs with bug for i≥10 (Severity: Medium, Category: Style)

**Location:** `internal/game/multiplayer_test.go:81`

**Description:** The loop runs `for i := 1; i < 20; i++` and creates player IDs with `"p" + string(rune('0'+i))`. For i=10, `'0' + 10` = rune(58) = `':'`, producing IDs like `"p:"`, `"p;"`, etc. The test still passes because these are valid unique strings, but it's unintended and confusing.

**Recommendation:** Use `fmt.Sprintf("p%d", i)` or `strconv.Itoa`.

---

### M-11: `CleanupSessions` goroutine runs forever with no cancellation (Severity: Medium, Category: Bug)

**Location:** `internal/handlers/game.go:292-303`

**Description:** The `CleanupSessions` goroutine has no context or stop channel. It runs forever with `time.Sleep(5*time.Minute)`. While acceptable for the production binary, it prevents clean shutdown in tests (goroutine leak).

**Recommendation:** Add a `ctx` parameter to `CleanupSessions` and select on `ctx.Done()`.

---

## Low Severity / Informational Findings

### L-1: `computeResults` uses byte indexing for first-letter check but rune slices for the rest

**Location:** `internal/game/game.go:22-27`

**Detail:** `word[0] != target[0]` compares bytes (not runes), while the rest of the function uses `[]rune()`. For pure ASCII this is safe, but it's an inconsistency.

### L-2: `Game.Attempts` is nil (not empty slice) when created

**Location:** `internal/game/game.go:8-15`

**Detail:** `NewGame` doesn't initialize `Attempts`. JSON serialization produces `null` instead of `[]`. Functionally equivalent but a minor JSON inconsistency.

### L-3: `GetRankings` silently ignores `ComputeAttemptResults` errors

**Location:** `internal/game/multiplayer.go:239`

**Detail:** `results, _ := ComputeAttemptResults(...)`. If `computeResults` fails for any reason (e.g. empty target), the error is silently dropped and results will be nil.

### L-4: No `:focus-visible` styles for keyboard accessibility

**Location:** `web/static/style.css`

**Detail:** The CSS has no `:focus-visible` outlines on `.kb-key`, `.multi-btn`, or `.mode-btn`. Keyboard users navigating with Tab will see no focus indicator (browser default overridden by `border: none` on keys).

### L-5: `execCommand('copy')` is deprecated

**Location:** `web/static/multiplayer.js:397`

**Detail:** `document.execCommand('copy')` is deprecated in favor of `navigator.clipboard.writeText()`. Works in all current browsers but will eventually be removed.

### L-6: CSS `#submit-btn:disabled` uses `opacity: 0.5` without `pointer-events: none`

**Location:** `web/static/style.css:237-240`

**Detail:** The disabled submit button has `opacity: 0.5` and `cursor: not-allowed` but no `pointer-events: none`. Click events are still captured by the button — the JS handler checks `gameState.gameOver` before proceeding, so this is functionally safe, but disabled buttons should ideally use `pointer-events: none`.

### L-7: No CI/CD workflow found

**Location:** `.github/workflows/`

**Detail:** No CI workflow exists. The project has no automated CI pipeline for running tests on push/PR. Adding a basic Go + Playwright CI workflow would prevent regression.

### L-8: `words.txt` symlink in root doesn't exist

**Location:** `README.md:31`

**Detail:** The README mentions `words.txt` at the project root as a symlink to `internal/dictionary/words.txt`, but no such symlink exists in the repository. A dangling README reference.

---

## Summary

| Severity | Count |
|----------|-------|
| Critical | 0 |
| High | 6 |
| Medium | 11 |
| Low / Info | 8 |
| **Total** | **25** |

### Key takeaways

1. **Security:** Token validation is missing on most multiplayer endpoints (H-5). SSE write deadline is disabled (H-2). ID generation ignores crypto errors (H-3).
2. **Correctness:** SSE reconnection creates zombie connections after leaving a room (H-4). Empty dictionary produces unplayable `"AAAAAA"` filler words (M-2).
3. **Documentation:** API routes in specs don't match implementation (H-6). Dictionary word count is wrong (M-1). Client vs server validation is misdescribed (M-3). README doesn't mention multiplayer (M-8). `AGENTS.md` version is stale (M-7).
4. **Tests:** 4 Playwright tests fail due to test bugs (M-4/M-5/M-6) — not application bugs. All 52 Go unit tests pass.
5. **Operations:** No graceful shutdown (H-1). No CI pipeline (L-7).
