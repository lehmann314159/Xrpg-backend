package game

import "time"

// Character represents a player character
type Character struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	HP            int       `json:"hp"`
	MaxHP         int       `json:"max_hp"`
	Strength      int       `json:"strength"`
	Dexterity     int       `json:"dexterity"`
	CurrentRoomID string    `json:"current_room_id"`
	IsAlive       bool      `json:"is_alive"`
	CreatedAt     time.Time `json:"created_at"`
	DiedAt        *time.Time `json:"died_at,omitempty"`
}

// Room represents a location in the dungeon
type Room struct {
	ID          string   `json:"id"`
	DungeonID   string   `json:"dungeon_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsEntrance  bool     `json:"is_entrance"`
	IsExit      bool     `json:"is_exit"`
	X           int      `json:"x"`
	Y           int      `json:"y"`
	Exits       []string `json:"exits"` // Populated from connections
}

// RoomConnection represents a connection between rooms
type RoomConnection struct {
	ID               string `json:"id"`
	RoomID           string `json:"room_id"`
	Direction        string `json:"direction"`
	ConnectedRoomID  string `json:"connected_room_id"`
}

// Monster represents an enemy
type Monster struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	HP          int       `json:"hp"`
	MaxHP       int       `json:"max_hp"`
	Damage      int       `json:"damage"`
	RoomID      string    `json:"room_id"`
	IsAlive     bool      `json:"is_alive"`
	LootTable   []string  `json:"loot_table"` // Item IDs that can drop
}

// Item represents an object that can be picked up
type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // weapon, armor, consumable, key, treasure
	Damage      int    `json:"damage"`
	Armor       int    `json:"armor"`
	Healing     int    `json:"healing"`
	RoomID      *string `json:"room_id,omitempty"`
	CharacterID *string `json:"character_id,omitempty"`
	IsEquipped  bool    `json:"is_equipped"`
}

// Trap represents a hazard in a room
type Trap struct {
	ID           string `json:"id"`
	RoomID       string `json:"room_id"`
	Description  string `json:"description"`
	Damage       int    `json:"damage"`
	IsTriggered  bool   `json:"is_triggered"`
	IsDiscovered bool   `json:"is_discovered"`
	Difficulty   int    `json:"difficulty"`
}

// GameEvent represents an event in the game for UI generation
type GameEvent struct {
	ID          int       `json:"id"`
	CharacterID string    `json:"character_id"`
	EventType   string    `json:"event_type"`
	EventData   string    `json:"event_data"` // JSON
	SuggestedUI string    `json:"suggested_ui"` // JSON array of UI panels
	Timestamp   time.Time `json:"timestamp"`
}

// Dungeon represents a generated dungeon
type Dungeon struct {
	ID        string    `json:"id"`
	Seed      int64     `json:"seed"`
	Depth     int       `json:"depth"`
	CreatedAt time.Time `json:"created_at"`
}

// CombatResult represents the outcome of a combat action
type CombatResult struct {
	AttackerDamage int    `json:"attacker_damage"`
	DefenderDamage int    `json:"defender_damage"`
	AttackerHP     int    `json:"attacker_hp"`
	DefenderHP     int    `json:"defender_hp"`
	AttackerDied   bool   `json:"attacker_died"`
	DefenderDied   bool   `json:"defender_died"`
	Message        string `json:"message"`
}

// UIPanel represents a suggested UI component
type UIPanel struct {
	Type     string                 `json:"type"`
	Priority string                 `json:"priority"` // high, medium, low
	Data     map[string]interface{} `json:"data"`
}
