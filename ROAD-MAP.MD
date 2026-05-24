
---

### Phase 1 — Infrastructure *(foundation, everything else depends on this)*

1. **Env config struct** — load port, DB DSN, Redis URL from env in main.go; fail fast if required vars are missing
2. **PostgreSQL** — `users`, `sessions`, `matches`, `scores` tables
3. **Redis** — session tokens, active room state, WebSocket pub/sub fan-out
4. **Migrations** — `golang-migrate`, versioned SQL under `db/migrations/`
5. **Auth** — cookie-based sessions: register / login / logout; guest play gets a generated nickname

---

### Phase 2 — Per-Game Configuration *(parallel with Phase 1 UI work)*

6. Add a `Config` interface to `models.Game`; implement per-game config structs:

| Game | Config fields |
|---|---|
| **Dual N-Back** | N level (1–9), mode (position / audio / dual), trial count, inter-stimulus interval |
| **Reversi** | Difficulty (Easy / Medium / Hard) |

7. Render a game-specific config form on `/{game}/machine` and `/{game}/match` **before** the session starts — replaces the current placeholder buttons

---

### Phase 3 — WebSocket & Real-Time Match *(depends on Phase 1 Redis)*

8. `internal/ws/hub.go` — central `Hub`: create/join/broadcast/leave rooms
9. Routes: `POST /{game}/match/room` → room code; `GET /{game}/match/room/{code}/ws` → upgrade
10. JSON message envelope: `{ "type": "move|state|end|chat|ping", "payload": … }`
11. Redis pub/sub — fan-out moves across server instances
12. 30-second reconnect grace window; board state restored from Redis on rejoin

---

### Phase 4 — Game Logic *(depends on Phase 2 config, Phase 3 for multiplayer)*

**Dual N-Back**
- Stimulus sequence generator (random grid position + optional audio cue)
- Response evaluator: compare against N steps back → hit / miss / false-alarm
- Session scorer: accuracy %, auto-advance N after two consecutive good sessions

**Reversi**
- Board: `[8][8]int8`, valid-move generator, flip logic
- Machine AI: minimax + alpha-beta pruning; search depth 4 / 6 / 8 per difficulty
- Multiplayer: delegate moves through WebSocket hub

---

### Phase 5 — Leaderboards *(depends on Phase 1 DB + Phase 4 scoring)*

13. `scores` table: `user_id`, `game_slug`, `score INT`, `metadata JSONB`, `created_at`
14. Per-game metrics: N-Back → best accuracy + highest N reached; Reversi → ELO + W/L record
15. `GET /{game}/leader-boards` — top 50 paginated; personal best row when authenticated
16. Score submission is **server-side only** at session end (client sends no score payload)

---

### Phase 6 — Optimization *(parallel, after Phase 4 is stable)*

17. Reversi AI: iterative deepening + transposition table
18. Redis cache for leaderboard top-N (TTL 60 s), invalidated on new score write
19. Fingerprinted static asset filenames + far-future `Cache-Control`
20. `pprof` endpoint behind an internal-only middleware flag

---

**Relevant files to create/modify**

- game.go — add `Config` interface + per-game config structs
- `internal/models/user.go` — new: `User` struct
- `internal/models/score.go` — new: `Score` struct, per-game metric helpers
- `internal/store/` — new: `postgres` + `redis` store interfaces and implementations
- `internal/handlers/auth.go` — new: register/login/logout
- `internal/handlers/ws.go` — new: WebSocket hub + room HTTP handlers
- `db/migrations/` — new: versioned SQL files
- main.go — wire config, stores, auth middleware

---

**Decisions / open questions**

1. **Auth depth** — full registration vs. OAuth (GitHub/Google)? Recommendation: start with email+password, add OAuth later.
2. **N-Back multiplayer** — does it need a real-time versus mode, or is it always solo? Recommendation: solo only for now; leaderboard is score-based.
3. **Database for Redis** — use Redis just for rooms/sessions, or also cache heavy leaderboard queries there too? Recommendation: both (Phase 6 adds the cache layer on top of PG queries).