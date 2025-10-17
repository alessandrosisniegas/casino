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
- Week 4: UI Foundations (X)
- Week 5: Statistics Tracking (MVP) (X)
- Week 6: Complete MVP (Testing & Polishing) (X)
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
➜  casino git:(main) make run-client                                            
cd cmd/client && go run .
Connected to Casino server at 127.0.0.1:9090
Type 'help' for available commands or 'quit' to exit.

OK Welcome to Casino! Use SIGNUP <username> <password> or LOGIN <username> <password>

$ signup charles oliveira 
OK Account created for charles with balance $10000.00

$ login charles oliveira
OK Welcome back, charles! Balance: $10000.00

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
  SURRENDER                    - Forfeit hand, get half bet back

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
Player Hand: [7♠] [9♣] (Value: 16)
Dealer Hand: [9♦] [Hidden]

Actions: HIT, STAND, DOUBLEDOWN, SURRENDER

$ doubledown
OK Doubled down!
Bet: $1000.00
Player Hand: [7♠] [9♣] [6♥] (Value: 22)
Dealer Hand: [9♦] [5♦] (Value: 14)

Result: Bust! Dealer wins.
Payout: $0.00


$ bet 1000
OK Game started!
Bet: $1000.00
Player Hand: [9♥] [4♣] (Value: 13)
Dealer Hand: [9♣] [Hidden]

Actions: HIT, STAND, DOUBLEDOWN, SURRENDER

$ hit
OK
Bet: $1000.00
Player Hand: [9♥] [4♣] [6♥] (Value: 19)
Dealer Hand: [9♣] [Hidden]

Actions: HIT, STAND

$ stand
OK
Bet: $1000.00
Player Hand: [9♥] [4♣] [6♥] (Value: 19)
Dealer Hand: [9♣] [2♥] [5♥] [Q♣] (Value: 26)

Result: You win!
Payout: $2000.00


$ bet 1500
OK Game started!
Bet: $1500.00
Player Hand: [K♥] [J♠] (Value: 20)
Dealer Hand: [2♣] [Hidden]

Actions: HIT, STAND, DOUBLEDOWN, SURRENDER

$ stand
OK
Bet: $1500.00
Player Hand: [K♥] [J♠] (Value: 20)
Dealer Hand: [2♣] [K♠] [J♦] (Value: 22)

Result: You win!
Payout: $3000.00


$ bet 2500
OK Game started!
Bet: $2500.00
Player Hand: [9♠] [6♦] (Value: 15)
Dealer Hand: [6♥] [Hidden]

Actions: HIT, STAND, DOUBLEDOWN, SURRENDER

$ surrender
OK Surrendered!
Bet: $2500.00
Player Hand: [9♠] [6♦] (Value: 15)
Dealer Hand: [6♥] [2♦] (Value: 8)

Result: Surrendered - half bet returned.
Payout: $1250.00


$ balance
OK Balance: $10250.00

$ stats
OK Stats for charles:
  Games Played: 4
  Games Won: 2
  Games Lost: 2
  Win Rate: 50.0%
  Total Bet: $6000.00
  Total Won: $6250.00
  Net: $250.00
  Avg Bet: $1500.00
  Biggest Win: $1500.00
  Biggest Loss: $1250.00

$ logout
OK Logged out successfully

$ quit
Connection to server lost
```