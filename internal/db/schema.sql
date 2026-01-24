-- Dungeon Crawler Database Schema

-- Characters (players)
CREATE TABLE IF NOT EXISTS characters (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    hp INTEGER NOT NULL,
    max_hp INTEGER NOT NULL,
    strength INTEGER NOT NULL,
    dexterity INTEGER NOT NULL,
    current_room_id TEXT,
    is_alive BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    died_at TIMESTAMP
);

-- Rooms in the dungeon
CREATE TABLE IF NOT EXISTS rooms (
    id TEXT PRIMARY KEY,
    dungeon_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    is_entrance BOOLEAN DEFAULT 0,
    is_exit BOOLEAN DEFAULT 0,
    x INTEGER,
    y INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Connections between rooms
CREATE TABLE IF NOT EXISTS room_connections (
    id TEXT PRIMARY KEY,
    room_id TEXT NOT NULL,
    direction TEXT NOT NULL, -- north, south, east, west
    connected_room_id TEXT NOT NULL,
    FOREIGN KEY (room_id) REFERENCES rooms(id),
    FOREIGN KEY (connected_room_id) REFERENCES rooms(id)
);

-- Monsters in rooms
CREATE TABLE IF NOT EXISTS monsters (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    hp INTEGER NOT NULL,
    max_hp INTEGER NOT NULL,
    damage INTEGER NOT NULL,
    room_id TEXT NOT NULL,
    is_alive BOOLEAN DEFAULT 1,
    loot_table TEXT, -- JSON array of possible item drops
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (room_id) REFERENCES rooms(id)
);

-- Items (in rooms or inventory)
CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL, -- weapon, armor, consumable, key, treasure
    damage INTEGER DEFAULT 0,
    armor INTEGER DEFAULT 0,
    healing INTEGER DEFAULT 0,
    room_id TEXT,
    character_id TEXT,
    is_equipped BOOLEAN DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (room_id) REFERENCES rooms(id),
    FOREIGN KEY (character_id) REFERENCES characters(id)
);

-- Traps in rooms
CREATE TABLE IF NOT EXISTS traps (
    id TEXT PRIMARY KEY,
    room_id TEXT NOT NULL,
    description TEXT,
    damage INTEGER NOT NULL,
    is_triggered BOOLEAN DEFAULT 0,
    is_discovered BOOLEAN DEFAULT 0,
    difficulty INTEGER DEFAULT 10, -- DC for detection/disarm
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (room_id) REFERENCES rooms(id)
);

-- Game events log (for UI generation and history)
CREATE TABLE IF NOT EXISTS game_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id TEXT NOT NULL,
    event_type TEXT NOT NULL, -- room_entered, combat_started, item_taken, etc.
    event_data TEXT, -- JSON data
    suggested_ui TEXT, -- JSON array of UI suggestions
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (character_id) REFERENCES characters(id)
);

-- Dungeons (to support multiple runs)
CREATE TABLE IF NOT EXISTS dungeons (
    id TEXT PRIMARY KEY,
    seed INTEGER,
    depth INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User UI preferences (which panels they keep/discard)
CREATE TABLE IF NOT EXISTS ui_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id TEXT NOT NULL,
    panel_type TEXT NOT NULL,
    is_enabled BOOLEAN DEFAULT 1,
    position INTEGER,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (character_id) REFERENCES characters(id)
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_characters_alive ON characters(is_alive);
CREATE INDEX IF NOT EXISTS idx_monsters_room ON monsters(room_id, is_alive);
CREATE INDEX IF NOT EXISTS idx_items_room ON items(room_id);
CREATE INDEX IF NOT EXISTS idx_items_character ON items(character_id);
CREATE INDEX IF NOT EXISTS idx_room_connections ON room_connections(room_id);
CREATE INDEX IF NOT EXISTS idx_events_character ON game_events(character_id);
