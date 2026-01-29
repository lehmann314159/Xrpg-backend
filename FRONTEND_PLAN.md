# Frontend Plan: Vercel AI SDK with Generative UI

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Browser                                                        │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  Next.js App (Vercel)                                     │  │
│  │  - Command input                                          │  │
│  │  - Streamed UI components from Claude                     │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ 1. User command + conversation
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Next.js Server Action                                          │
│  - Receives user command                                        │
│  - Calls Claude via Vercel AI SDK                               │
│  - Claude has access to:                                        │
│      a) MCP tools (call Go backend)                             │
│      b) UI components (render to browser)                       │
└─────────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┴───────────────────┐
          │                                       │
          ▼                                       ▼
┌──────────────────────┐              ┌──────────────────────┐
│  Go Backend (MCP)    │              │  UI Component Tools  │
│  - new_game          │              │  - showMap()         │
│  - look              │              │  - showStats()       │
│  - move              │              │  - showMonster()     │
│  - attack            │              │  - showInventory()   │
│  - take, use, equip  │              │  - showCombatLog()   │
│                      │              │  - showNotification()│
│  Returns: gameState  │              │                      │
│  JSON snapshot       │              │  Streams React       │
│                      │              │  components to UI    │
└──────────────────────┘              └──────────────────────┘
```

## Flow Example

1. **User types:** "go north"
2. **Next.js server action** sends to Claude with system prompt + tools
3. **Claude decides:** "I need to call the `move` MCP tool with direction=north"
4. **Server action** calls Go backend: `POST /mcp/call {name: "move", arguments: {direction: "north"}}`
5. **Go backend** returns `{content: [...], gameState: {...}}`
6. **Claude sees** the new game state and decides what UI to show:
   - "The player entered a room with a goblin. I'll render showMap(), showStats(), and showMonster()."
7. **UI components stream** to the browser as they're generated

## Project Structure

```
dungeon-crawler-frontend/
├── app/
│   ├── layout.tsx              # Root layout
│   ├── page.tsx                # Main game page
│   ├── actions.tsx             # Server actions (Claude + MCP calls)
│   └── globals.css             # Tailwind styles
├── components/
│   ├── ui/                     # Shadcn/ui base components
│   │   ├── card.tsx
│   │   ├── button.tsx
│   │   ├── input.tsx
│   │   └── progress.tsx
│   └── game/                   # Game-specific UI components
│       ├── DungeonMap.tsx      # 5x5 grid map
│       ├── PlayerStats.tsx     # HP bar, attributes
│       ├── MonsterCard.tsx     # Enemy display with attack option
│       ├── ItemCard.tsx        # Item display with actions
│       ├── EquipmentPanel.tsx  # Equipped gear
│       ├── CombatLog.tsx       # Recent combat events
│       ├── RoomDescription.tsx # Current room narrative
│       └── Notification.tsx    # Transient alerts
├── lib/
│   ├── mcp-client.ts           # Fetch wrapper for Go backend
│   └── tools.tsx               # Tool definitions for Claude
├── .env.local                  # ANTHROPIC_API_KEY, BACKEND_URL
├── next.config.js
├── tailwind.config.js
└── package.json
```

## Key Files

### 1. `app/actions.tsx` - Server Actions

This is the core. It:
- Receives user messages
- Maintains conversation state
- Calls Claude with both MCP tools and UI tools
- Streams UI back to client

```tsx
'use server';

import { createStreamableUI } from 'ai/rsc';
import Anthropic from '@anthropic-ai/sdk';
import { mcpTools, uiTools } from '@/lib/tools';

export async function sendMessage(userMessage: string) {
  const stream = createStreamableUI();

  // Claude sees MCP tools (to get game state) and UI tools (to render components)
  // When Claude calls a UI tool, it streams a React component to the browser

  return stream.value;
}
```

### 2. `lib/tools.tsx` - Tool Definitions

Two categories of tools:

**MCP Tools** (call Go backend, return data to Claude):
```tsx
export const mcpTools = {
  new_game: {
    description: "Start a new game",
    parameters: { character_name: string },
    execute: async (args) => {
      const response = await fetch(`${BACKEND_URL}/mcp/call`, {
        method: 'POST',
        body: JSON.stringify({ name: 'new_game', arguments: args })
      });
      return response.json(); // Returns to Claude
    }
  },
  move: { /* ... */ },
  attack: { /* ... */ },
  // ... etc
};
```

**UI Tools** (render components, stream to browser):
```tsx
export const uiTools = {
  showMap: {
    description: "Display the dungeon map showing explored areas and player position",
    parameters: { mapGrid: MapCell[][], playerPosition: {x, y} },
    render: (args) => <DungeonMap grid={args.mapGrid} player={args.playerPosition} />
  },
  showMonster: {
    description: "Display a monster card with its stats and attack button",
    parameters: { id, name, hp, maxHp, damage, description },
    render: (args) => <MonsterCard {...args} />
  },
  showStats: {
    description: "Display player stats including HP bar and attributes",
    parameters: { name, hp, maxHp, strength, dexterity, status },
    render: (args) => <PlayerStats {...args} />
  },
  // ... etc
};
```

### 3. `app/page.tsx` - Main UI

```tsx
'use client';

import { useState } from 'react';
import { useActions } from 'ai/rsc';

export default function GamePage() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const { sendMessage } = useActions();

  const handleSubmit = async () => {
    const response = await sendMessage(input);
    setMessages(prev => [...prev, { role: 'user', content: input }, response]);
    setInput('');
  };

  return (
    <div className="game-container">
      {/* Rendered UI components appear here */}
      <div className="game-panels">
        {messages.map(m => m.display)}
      </div>

      {/* Command input */}
      <input
        value={input}
        onChange={e => setInput(e.target.value)}
        onKeyDown={e => e.key === 'Enter' && handleSubmit()}
        placeholder="What do you want to do?"
      />
    </div>
  );
}
```

### 4. `lib/mcp-client.ts` - Backend Communication

```tsx
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';

export async function callMcpTool(name: string, args: Record<string, unknown>) {
  const response = await fetch(`${BACKEND_URL}/mcp/call`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, arguments: args }),
  });

  if (!response.ok) {
    throw new Error(`MCP call failed: ${response.statusText}`);
  }

  return response.json();
}
```

## System Prompt for Claude

```
You are the game master for a dungeon crawler RPG. You control the game interface.

When the player gives a command:
1. Use the appropriate MCP tool (move, attack, take, etc.) to execute the action
2. Based on the game state returned, decide which UI components to display
3. Always show relevant information - map, stats, any monsters or items present

UI Guidelines:
- Always show the map and player stats
- Show monster cards when enemies are present
- Show item cards for items in the room or inventory when relevant
- Use notifications for important events (damage taken, items found, etc.)
- Describe what happens narratively before showing the UI components

Available MCP tools: new_game, look, move, attack, take, use, equip, inventory, stats, map

Available UI components: showMap, showStats, showMonster, showItem, showEquipment, showCombatLog, showNotification, showRoomDescription
```

## Environment Variables

```bash
# .env.local
ANTHROPIC_API_KEY=sk-ant-...
BACKEND_URL=http://localhost:8080

# For production (Vercel dashboard)
ANTHROPIC_API_KEY=sk-ant-...
BACKEND_URL=https://your-lightsail-server.com
```

## Dependencies

```json
{
  "dependencies": {
    "next": "14.x",
    "react": "18.x",
    "react-dom": "18.x",
    "ai": "^3.0.0",
    "@anthropic-ai/sdk": "^0.24.0",
    "tailwindcss": "^3.4.0",
    "class-variance-authority": "^0.7.0",
    "clsx": "^2.1.0"
  }
}
```

## Setup Steps

1. **Create Next.js project:**
   ```bash
   npx create-next-app@latest dungeon-crawler-frontend --typescript --tailwind --app
   cd dungeon-crawler-frontend
   ```

2. **Install AI SDK:**
   ```bash
   npm install ai @anthropic-ai/sdk
   ```

3. **Add shadcn/ui (optional but recommended):**
   ```bash
   npx shadcn-ui@latest init
   npx shadcn-ui@latest add card button input progress
   ```

4. **Create environment file:**
   ```bash
   echo "ANTHROPIC_API_KEY=your-key-here" > .env.local
   echo "BACKEND_URL=http://localhost:8080" >> .env.local
   ```

5. **Implement files** in order:
   - `lib/mcp-client.ts`
   - `components/game/*.tsx` (UI components)
   - `lib/tools.tsx`
   - `app/actions.tsx`
   - `app/page.tsx`

## Backend Requirements (Already Done)

The Go backend changes I made are still needed:
- ✅ CORS middleware in `cmd/server/main.go`
- ✅ `GameStateSnapshot` types in `internal/game/types.go`
- ✅ `gameState` field in `ToolResult` responses
- ✅ All handlers return game state

## Questions to Consider

1. **Conversation memory:** Should Claude remember the full game history, or just recent turns? (Token cost vs. context)

2. **UI persistence:** When Claude streams new UI, should it replace all panels or append? Probably replace with fresh state each turn.

3. **Error handling:** What if the Go backend is down? Show error UI component?

4. **Loading states:** Stream a "thinking..." indicator while Claude processes?
