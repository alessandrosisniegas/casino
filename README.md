# Casino
Terminal-based Casino starting with Blackjack in Go.
MVP is single-player with SQLite backed stats.
Completion will include TCP multiplayer and probability guidance.

## Usage
```bash
make build
make test       # run all tests
make run-server
make run-client # in another terminal
```

## Structure
- `cmd/server` — Server
- `cmd/client` — Client
- `core/game` — Blackjack Game
- `core/vault` — SQLite Database
- `core/security` — Authentication & Security
- `data/` — Runtime Data

## Roadmap
- Week 1: Setup & Design (X)
- Week 2: Authentication & Persistence (X)
- Week 3: Core Blackjack
- Week 4: UI Foundations
- Week 5: Statistics Tracking (MVP)
- Week 6: Complete MVP (Testing & Polishing)
- Week 7: Multiplayer Foundation
- Week 8: Multiplayer Game Loop
- Week 9: Multiplayer Enhancements
- Week 10: Probability Analysis, Toggle, & Extended Stats
- Week 11: Complete Project (Testing & Polishing)
- Week 12: Demo
