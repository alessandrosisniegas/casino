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

## Game Modes

### Solo Mode
Play Blackjack against the dealer by yourself. Your stats and balance are tracked.

```bash
$ login <username> <password>
$ solo              # Enter solo mode
$ bet 100           # Start playing
$ hit
$ stand
```

Or use the quick-play shortcut:
```bash
$ login <username> <password>
$ bet 100           # Auto-enters solo mode
```

### Multiplayer Mode
Play at a shared table with 2-4 players. Everyone plays against the same dealer.

**Terminal 1:**
```bash
$ make run-client
$ login <username> <password>
$ multiplayer       # Join the table
$ ready             # Mark ready
# Wait for other players...
$ bet 100           # Place bet when round starts
$ hit               # Your turn
$ stand
```

**Terminal 2:**
```bash
$ make run-client
$ login <username2> <password2>
$ multiplayer       # Join the same table
$ ready
$ bet 50
# Wait for your turn...
```

**Multiplayer Commands:**
```
MULTIPLAYER (or MP)  # Join the multiplayer table
PLAYERS              # See who's at the table
READY                # Mark ready to start round
LEAVE                # Leave table, return to lobby
```

### Connecting to Remote Server
```bash
# Connect to a server on your local network
$ go run cmd/client/main.go --host 192.168.1.100

# Connect to a server on a different port
$ go run cmd/client/main.go --host 127.0.0.1 --port 8080
```

## Roadmap
- Week 1: Setup & Design (X)
- Week 2: Authentication & Persistence (X)
- Week 3: Core Blackjack (X)
- Week 4: UI Foundations (X)
- Week 5: Statistics Tracking (MVP) (X)
- Week 6: Complete MVP (Testing & Polishing) (X)
- Week 7: Multiplayer Foundation (X)
- Week 8: Multiplayer Game Loop (X)
- Week 9: Multiplayer Enhancements
- Week 10: Probability Analysis, Toggle, & Extended Stats
- Week 11: Complete Project (Testing & Polishing)
- Week 12: Demo

### Solo Gameplay
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

### Multiplayer Gameplay
Server:
```
➜  casino git:(main) ✗ make run-server
cd cmd/server && go run .
Casino Server listening on 127.0.0.1:9090
Database initialized at ../../data/casino.db
Type 'help' for server commands, 'quit' to shutdown
server> 2025/12/05 10:13:43 Client connection error: read tcp 127.0.0.1:9090->127.0.0.1:56703: use of closed network connection
2025/12/05 10:13:46 Client connection error: read tcp 127.0.0.1:9090->127.0.0.1:56704: use of closed network connection
server> quit
Shutting down server...
Server stopped.
```
Client #1:
```
➜  casino git:(main) ✗ make run-client
cd cmd/client && go run .
Connected to Casino server at 127.0.0.1:9090
Type 'help' for available commands or 'quit' to exit.

OK Welcome to Casino! Use SIGNUP <username> <password> or LOGIN <username> <password>

$ login real madrid
OK Welcome back, real! Balance: $6900.00
Available modes: SOLO, MULTIPLAYER
Use SOLO or MULTIPLAYER to choose a mode, or BET <amount> for quick solo play.

$ multiplayer
OK Joined multiplayer table (1/4 players)
Players at table (1/4):
  Seat 1: real
Type READY when you're ready to play.

$ 
inter joined the table (2/4 players)

$ ready
OK Marked ready. Waiting for other players...

$ 
inter is ready!

$ 
--- NEW ROUND ---
Place your bets! (30 seconds)
Use: BET <amount>

$ bet 500
OK Bet placed: $500.00

$ 
inter bets $300.00

$ 
--- DEALING ---
Seat 1 - real: [3♥] [Q♠] (Value: 13) | Bet: $500.00
Seat 2 - inter: [8♠] [4♦] (Value: 12) | Bet: $300.00
Dealer: [3♠] [Hidden]

YOUR TURN (30 seconds)
Actions: HIT, STAND, DOUBLEDOWN, SURRENDER

$ hit 
OK You hit and got [7♥]
Your hand: [3♥] [Q♠] [7♥] (Value: 20)

YOUR TURN (30 seconds)
Seat 1 - real: [3♥] [Q♠] [7♥] (Value: 20) | Bet: $500.00
Seat 2 - inter: [8♠] [4♦] (Value: 12) | Bet: $300.00
Dealer: [3♠] [Hidden]
Actions: HIT, STAND

$ stand
OK You stand with 20

$ 
inter's turn...

$ 
inter doubles down and gets [2♦] (now at 14)

$ 
--- DEALER TURN ---
Dealer: [3♠] [4♣] [K♠] (Value: 17)
--- RESULTS ---
real: PLAYER_WIN → WIN (+$500.00)
inter: DEALER_WIN → LOSS (-$600.00)
Round complete! Type READY to play again.

$ ready
OK Marked ready. Waiting for other players...

$ 
inter is ready!

$ 
--- NEW ROUND ---
Place your bets! (30 seconds)
Use: BET <amount>

$ bet 600
OK Bet placed: $600.00

$ 
inter bets $400.00

$ 
--- DEALING ---
Seat 1 - real: [8♦] [10♠] (Value: 18) | Bet: $600.00
Seat 2 - inter: [J♥] [5♦] (Value: 15) | Bet: $400.00
Dealer: [A♠] [K♣] (Value: 21)


--- DEALER TURN ---
Dealer: [A♠] [K♣] (Value: 21)
--- RESULTS ---
real: DEALER_WIN → LOSS (-$600.00)
inter: DEALER_WIN → LOSS (-$400.00)
Round complete! Type READY to play again.

$ ready
OK Marked ready. Waiting for other players...

$ 
inter is ready!

$ 
--- NEW ROUND ---
Place your bets! (30 seconds)
Use: BET <amount>

$ bet 700
OK Bet placed: $700.00

$ 
inter bets $500.00

$ 
--- DEALING ---
Seat 1 - real: [J♣] [A♦] (Value: 21) | Bet: $700.00
Seat 2 - inter: [A♣] [5♠] (Value: 16) | Bet: $500.00
Dealer: [8♥] [Hidden]

inter's turn...

$ 
inter stands with 16

$ 
--- DEALER TURN ---
Dealer: [8♥] [10♥] (Value: 18)
--- RESULTS ---
real: PLAYER_BLACKJACK → WIN (+$1050.00)
inter: DEALER_WIN → LOSS (-$500.00)
Round complete! Type READY to play again.

$ leave
OK Left multiplayer mode. Back in lobby.
Available modes: SOLO, MULTIPLAYER

$ quit
```
Client #2:
```
➜  casino git:(main) ✗ make run-client
cd cmd/client && go run .
Connected to Casino server at 127.0.0.1:9090
Type 'help' for available commands or 'quit' to exit.

OK Welcome to Casino! Use SIGNUP <username> <password> or LOGIN <username> <password>

$ login inter milan
OK Welcome back, inter! Balance: $10300.00
Available modes: SOLO, MULTIPLAYER
Use SOLO or MULTIPLAYER to choose a mode, or BET <amount> for quick solo play.

$ multiplayer
OK Joined multiplayer table (2/4 players)
Players at table (2/4):
  Seat 1: real
  Seat 2: inter
Type READY when you're ready to play.

$ 
real is ready!

$ ready
OK Marked ready. Waiting for other players...

$ 
--- NEW ROUND ---
Place your bets! (30 seconds)
Use: BET <amount>

$ 
real bets $500.00

$ bet 300
OK Bet placed: $300.00

$ 
--- DEALING ---
Seat 1 - real: [3♥] [Q♠] (Value: 13) | Bet: $500.00
Seat 2 - inter: [8♠] [4♦] (Value: 12) | Bet: $300.00
Dealer: [3♠] [Hidden]

real's turn...

$ 
real hits and gets [7♥] (now at 20)

$ 
real stands with 20

$ 
YOUR TURN (30 seconds)
Seat 1 - real: [3♥] [Q♠] [7♥] (Value: 20) | Bet: $500.00
Seat 2 - inter: [8♠] [4♦] (Value: 12) | Bet: $300.00
Dealer: [3♠] [Hidden]
Actions: HIT, STAND, DOUBLEDOWN, SURRENDER

$ doubledown
OK Doubled down! Got [2♦]
Your hand: [8♠] [4♦] [2♦] (Value: 14)

$ 
--- DEALER TURN ---
Dealer: [3♠] [4♣] [K♠] (Value: 17)
--- RESULTS ---
real: PLAYER_WIN → WIN (+$500.00)
inter: DEALER_WIN → LOSS (-$600.00)
Round complete! Type READY to play again.

$ 
real is ready!

$ ready
OK Marked ready. Waiting for other players...

$ 
--- NEW ROUND ---
Place your bets! (30 seconds)
Use: BET <amount>

$ 
real bets $600.00

$ bet 400
OK Bet placed: $400.00

$ 
--- DEALING ---
Seat 1 - real: [8♦] [10♠] (Value: 18) | Bet: $600.00
Seat 2 - inter: [J♥] [5♦] (Value: 15) | Bet: $400.00
Dealer: [A♠] [K♣] (Value: 21)


--- DEALER TURN ---
Dealer: [A♠] [K♣] (Value: 21)
--- RESULTS ---
real: DEALER_WIN → LOSS (-$600.00)
inter: DEALER_WIN → LOSS (-$400.00)
Round complete! Type READY to play again.

$ 
real is ready!

$ ready
OK Marked ready. Waiting for other players...

$ 
--- NEW ROUND ---
Place your bets! (30 seconds)
Use: BET <amount>

$ 
real bets $700.00

$ bet 500
OK Bet placed: $500.00

$ 
--- DEALING ---
Seat 1 - real: [J♣] [A♦] (Value: 21) | Bet: $700.00
Seat 2 - inter: [A♣] [5♠] (Value: 16) | Bet: $500.00
Dealer: [8♥] [Hidden]

YOUR TURN (30 seconds)
Actions: HIT, STAND, DOUBLEDOWN, SURRENDER

$ stand
OK You stand with 16

$ 
--- DEALER TURN ---
Dealer: [8♥] [10♥] (Value: 18)
--- RESULTS ---
real: PLAYER_BLACKJACK → WIN (+$1050.00)
inter: DEALER_WIN → LOSS (-$500.00)
Round complete! Type READY to play again.

$ 
real left the table

$ leave
OK Left multiplayer mode. Back in lobby.
Available modes: SOLO, MULTIPLAYER

$ quit
```