# Xrpg Backend

Go MCP server for a turn-based dungeon crawler with procedural generation.

## Related Repositories

- **Frontend:** [Xrpg-frontend](https://github.com/lehmann314159/Xrpg-frontend) - Next.js + Vercel AI SDK

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Browser                                                    │
│  Next.js frontend with streamed UI components               │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Next.js Server Actions (Vercel)                            │
│  Claude via Vercel AI SDK                                   │
│  - Interprets user commands                                 │
│  - Calls MCP tools on Go backend                            │
│  - Decides which UI components to render                    │
└──────────────┬─────────────────────────┬────────────────────┘
               │                         │
               ▼                         ▼
┌──────────────────────────┐  ┌──────────────────────────────┐
│  Go Backend (MCP Server) │  │  UI Component Tools          │
│  ← THIS REPO             │  │  - showMap()                 │
│  - Game logic            │  │  - showStats()               │
│  - Procedural generation │  │  - showMonster()             │
│  - State management      │  │  - showInventory()           │
│  Returns: gameState JSON │  │  Streams React → browser     │
└──────────────────────────┘  └──────────────────────────────┘
```

## Project Structure

```
Xrpg-backend/
├── cmd/server/main.go        # Entry point, CORS middleware
├── internal/
│   ├── game/                 # Game logic
│   │   ├── types.go          # Data structures + view types
│   │   ├── state.go          # Game state management
│   │   ├── combat.go         # Combat mechanics
│   │   └── character.go      # Character management
│   ├── mcp/server.go         # MCP protocol + handlers
│   ├── generator/procgen.go  # Dungeon generation
│   └── db/                   # SQLite layer
├── Dockerfile
└── docker-compose.yml
```

## Getting Started

### Prerequisites

- Go 1.21+

### Setup

```bash
go mod download

# Run the server
go run cmd/server/main.go
```

The server runs on `http://localhost:8080` by default.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `DB_PATH` | `./dungeon-crawler.db` | SQLite database path |
| `CORS_ORIGINS` | `localhost:3000,localhost:5173` | Comma-separated allowed origins |

### Testing

```bash
# List MCP tools
curl http://localhost:8080/mcp/tools

# Start a new game
curl -X POST http://localhost:8080/mcp/call \
  -H "Content-Type: application/json" \
  -d '{"name": "new_game", "arguments": {"character_name": "Hero"}}'

# Look around
curl -X POST http://localhost:8080/mcp/call \
  -H "Content-Type: application/json" \
  -d '{"name": "look", "arguments": {}}'
```

## MCP Tools

The server exposes these tools via `/mcp/call`:

| Tool | Description | Arguments |
|------|-------------|-----------|
| `new_game` | Start a new game | `character_name` |
| `look` | Examine current room | - |
| `move` | Move in a direction | `direction` (north/south/east/west) |
| `attack` | Attack a monster | `target_id` |
| `take` | Pick up an item | `item_id` |
| `use` | Use an item | `item_id` |
| `equip` | Equip weapon/armor | `item_id` |
| `inventory` | View inventory | - |
| `stats` | View character stats | - |
| `map` | View dungeon map | - |

All responses include a `gameState` field with the full game state snapshot for UI rendering.

## Game Mechanics

### Combat
- Turn-based with d20 attack rolls
- Damage uses d6 + weapon/strength modifiers
- Armor reduces incoming damage
- Monsters block movement until defeated

### Progression
- No character leveling
- Find better weapons and armor
- Consumables restore HP

### Permadeath
- Character dies = game over
- Start fresh with a new dungeon

### Dungeon
- 5x5 procedurally generated grid
- Entrance at (0,0), exit at (4,4)
- Monsters, items, and traps scale with distance from entrance

## Deployment

### Docker

```bash
docker-compose up
```

### Manual (Lightsail/VPS)

```bash
go build -o dungeon-crawler cmd/server/main.go
CORS_ORIGINS=https://your-vercel-app.vercel.app ./dungeon-crawler
```

## License

MIT
