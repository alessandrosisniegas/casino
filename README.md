# Casino
Terminal-based Casino starting with Blackjack in Go.
MVP is single-player with SQLite backed stats.
Completion will include TCP multiplayer and probability guidance.

## Usage
```bash
make build
make test       # run all tests
make run-server # start server in one terminal
make run-client # start client in another terminal
```

### Commands

**Account Management:**
```
SIGNUP <user> <pass>  # Create a new account
LOGIN <user> <pass>   # Login to your account
LOGOUT                # Logout from your account
WHOAMI                # Show current login status
```

**Playing Blackjack:**
```
BET <amount>          # Start a game (e.g., BET 10 for $10)
HIT                   # Draw another card
STAND                 # End your turn
DOUBLEDOWN            # Double bet, draw one card, end turn
```

**Account Info:**
```
BALANCE               # Check your current balance
STATS                 # View your game statistics
```

**Other:**
```
HELP                  # Show all available commands
QUIT                  # Disconnect from server
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
- Week 3: Core Blackjack (X)
- Week 4: UI Foundations
- Week 5: Statistics Tracking (MVP)
- Week 6: Complete MVP (Testing & Polishing)
- Week 7: Multiplayer Foundation
- Week 8: Multiplayer Game Loop
- Week 9: Multiplayer Enhancements
- Week 10: Probability Analysis, Toggle, & Extended Stats
- Week 11: Complete Project (Testing & Polishing)
- Week 12: Demo

### Gameplay
Server:
```
➜  casino git:(main) ✗ make build
go build ./...
➜  casino git:(main) ✗ make run-server
cd cmd/server && go run .
Casino Server listening on 127.0.0.1:9090
Database initialized at ../../data/casino.db
Type 'help' for server commands, 'quit' to shutdown
server> 2025/10/11 17:33:01 Client connection error: read tcp 127.0.0.1:9090->127.0.0.1:55606: use of closed network connection
quit
Shutting down server...
Server stopped.
```
Client:
```
➜  casino git:(main) ✗ make run-client
cd cmd/client && go run .
Connected to Casino server at 127.0.0.1:9090
Type 'help' for available commands or 'quit' to exit.

OK Welcome to Casino! Use SIGNUP <username> <password> or LOGIN <username> <password>
> SIGNUP frankie fs_blank
OK Account created for frankie with balance $1000.00
> LOGIN frankie fs_blank
OK Welcome back, frankie! Balance: $1000.00
> help
OK Available commands:
Account Management:
  SIGNUP <username> <password> - Create a new account
  LOGIN <username> <password>  - Login to your account
  LOGOUT                       - Logout from your account
  BALANCE                      - Check your current balance
  STATS                        - View your game statistics
  WHOAMI                       - Show current login status

Blackjack Game:
  BET <amount>                 - Start a game and place bet (in dollars)
  HIT                          - Draw another card
  STAND                        - End your turn
  DOUBLEDOWN                   - Double bet, draw one card, end turn

Other:
  HELP                         - Show this help message
  QUIT                         - Disconnect from server

Username & Password requirements:
  - 2-30 characters long
  - Letters, numbers, and underscores only
  - No whitespace allowed
  - Password cannot be the same as username

> balance
OK Balance: $1000.00
> stats
OK Stats for frankie:
  Games Played: 0
  Games Won: 0
  Games Lost: 0
  Win Rate: 0.0%
  Total Bet: $0.00
  Total Won: $0.00
  Net: $0.00
  Biggest Win: $0.00
  Biggest Loss: $0.00

> whoami
OK Logged in as: frankie (ID: 7, Balance: $1000.00)
> BET 100
OK Game started!
Bet: $100.00
Player Hand: 5♦ K♠ (Value: 15)
Dealer Hand: 10♠ [Hidden]

Actions: HIT, STAND, DOUBLEDOWN
   
> hit
OK
Bet: $100.00
Player Hand: 5♦ K♠ A♠ (Value: 16)
Dealer Hand: 10♠ [Hidden]

Actions: HIT, STAND, DOUBLEDOWN

> hit
OK
Bet: $100.00
Player Hand: 5♦ K♠ A♠ A♦ (Value: 17)
Dealer Hand: 10♠ [Hidden]

Actions: HIT, STAND, DOUBLEDOWN
> hit
OK
Bet: $100.00
Player Hand: 5♦ K♠ A♠ A♦ J♣ (Value: 27)
Dealer Hand: 10♠ 7♥ (Value: 17)

Result: Bust! Dealer wins.
Payout: $0.00
 
> balance
OK Balance: $900.00
> bet 200
OK Game started!
Bet: $200.00
Player Hand: 9♥ 7♥ (Value: 16)
Dealer Hand: 9♠ [Hidden]

Actions: HIT, STAND, DOUBLEDOWN
> hit
OK
Bet: $200.00
Player Hand: 9♥ 7♥ K♥ (Value: 26)
Dealer Hand: 9♠ K♣ (Value: 19)

Result: Bust! Dealer wins.
Payout: $0.00

> bet 400
OK Game started!
Bet: $400.00
Player Hand: 2♠ 8♦ (Value: 10)
Dealer Hand: 7♥ [Hidden]

Actions: HIT, STAND, DOUBLEDOWN
> hit  
OK
Bet: $400.00
Player Hand: 2♠ 8♦ 7♠ (Value: 17)
Dealer Hand: 7♥ [Hidden]

Actions: HIT, STAND, DOUBLEDOWN
> stand
OK
Bet: $400.00
Player Hand: 2♠ 8♦ 7♠ (Value: 17)
Dealer Hand: 7♥ 8♥ 10♠ (Value: 25)

Result: You win!
Payout: $800.00

> balance
OK Balance: $1100.00
> bet 500
OK Game started!
Bet: $500.00
Player Hand: 3♠ J♥ (Value: 13)
Dealer Hand: 6♣ [Hidden]

Actions: HIT, STAND, DOUBLEDOWN
> doubledown
OK Doubled down!
> Bet: $1000.00
Player Hand: 3♠ J♥ A♥ (Value: 14)
Dealer Hand: 6♣ 3♥ 8♠ (Value: 17)

Result: Dealer wins.
Payout: $0.00

> balance
OK Balance: $100.00
> stats
OK Stats for frankie:
  Games Played: 4
  Games Won: 1
  Games Lost: 3
  Win Rate: 25.0%
  Total Bet: $1700.00
  Total Won: $800.00
  Net: $-900.00
  Biggest Win: $400.00
  Biggest Loss: $1000.00

> logout
OK Logged out successfully
> whoami
ERROR Not logged in
> quit
Connection to server lost
```