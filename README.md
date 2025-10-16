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
make stop       # stop the server
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
SURRENDER             # Forfeit hand, get half bet back
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
server> quit
Shutting down server...
Server stopped.
➜  casino git:(main) ✗ make stop
No process running on port 9090
```
Client:
```
➜  casino git:(main) ✗ make run-client
cd cmd/client && go run .
Connected to Casino server at 127.0.0.1:9090
Type 'help' for available commands or 'quit' to exit.

OK Welcome to Casino! Use SIGNUP <username> <password> or LOGIN <username> <password>

$ signup mark zuckerburg
OK Account created for mark with balance $10000.00

$ login mark zuckerburg
OK Welcome back, mark! Balance: $10000.00

$ help
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

$ bet 500
OK Game started!
Bet: $500.00
Player Hand: [6♥] [7♦] (Value: 13)
Dealer Hand: [4♥] [Hidden]

Actions: HIT, STAND, DOUBLEDOWN

$ doubledown
OK Doubled down!
Bet: $1000.00
Player Hand: [6♥] [7♦] [4♦] (Value: 17)
Dealer Hand: [4♥] [J♠] [2♥] [6♦] (Value: 22)

Result: You win!
Payout: $2000.00


$ balance
OK Balance: $11000.00

$ bet 1000
OK Game started!
Bet: $1000.00
Player Hand: [3♣] [8♥] (Value: 11)
Dealer Hand: [3♦] [Hidden]

Actions: HIT, STAND, DOUBLEDOWN

$ hit
OK
Bet: $1000.00
Player Hand: [3♣] [8♥] [Q♦] (Value: 21)
Dealer Hand: [3♦] [Hidden]

Actions: HIT, STAND

$ stand
OK
Bet: $1000.00
Player Hand: [3♣] [8♥] [Q♦] (Value: 21)
Dealer Hand: [3♦] [A♠] [A♦] [A♥] [6♣] [6♠] (Value: 18)

Result: You win!
Payout: $2000.00


$ bet 2000
OK Game started!
Bet: $2000.00
Player Hand: [3♠] [5♦] (Value: 8)
Dealer Hand: [4♠] [Hidden]

Actions: HIT, STAND, DOUBLEDOWN

$ hit
OK
Bet: $2000.00
Player Hand: [3♠] [5♦] [7♣] (Value: 15)
Dealer Hand: [4♠] [Hidden]

Actions: HIT, STAND

$ stand
OK
Bet: $2000.00
Player Hand: [3♠] [5♦] [7♣] (Value: 15)
Dealer Hand: [4♠] [5♠] [A♥] (Value: 20)

Result: Dealer wins.
Payout: $0.00


$ balance
OK Balance: $10000.00

$ stats
OK Stats for mark:
  Games Played: 3
  Games Won: 2
  Games Lost: 1
  Win Rate: 66.7%
  Total Bet: $4000.00
  Total Won: $4000.00
  Net: $0.00
  Avg Bet: $1333.33
  Biggest Win: $1000.00
  Biggest Loss: $2000.00

$ logout
OK Logged out successfully

$ quit
Connection to server lost
```