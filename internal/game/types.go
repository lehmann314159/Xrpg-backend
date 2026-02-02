package game

import "time"

// Character represents a player character
type Character struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	HP              int        `json:"hp"`
	MaxHP           int        `json:"max_hp"`
	Strength        int        `json:"strength"`
	Dexterity       int        `json:"dexterity"`
	CurrentRoomID   string     `json:"current_room_id"`
	IsAlive         bool       `json:"is_alive"`
	CreatedAt       time.Time  `json:"created_at"`
	DiedAt          *time.Time `json:"died_at,omitempty"`
	EquippedWeaponID *string   `json:"equipped_weapon_id,omitempty"`
	EquippedArmorID  *string   `json:"equipped_armor_id,omitempty"`
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
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        string  `json:"type"` // weapon, armor, consumable, key, treasure
	Damage      int     `json:"damage"`
	Armor       int     `json:"armor"`
	Healing     int     `json:"healing"`
	Rarity      string  `json:"rarity"` // common, uncommon, rare, legendary
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

// EventInfo provides structured metadata about the last game event
type EventInfo struct {
	Type     string   `json:"type"`     // combat, discovery, movement, interaction, death, victory
	Subtype  string   `json:"subtype"`  // attack_hit, attack_miss, enemy_defeated, item_found, etc.
	Entities []string `json:"entities"` // IDs of involved monsters/items
}

// AttackResult represents the detailed outcome of a single attack
type AttackResult struct {
	AttackerName string `json:"attackerName"`
	TargetName   string `json:"targetName"`
	Damage       int    `json:"damage"`
	WasHit       bool   `json:"wasHit"`
	WasCritical  bool   `json:"wasCritical"`
	RemainingHP  int    `json:"remainingHp"`
}

// EnhancedCombatResult provides detailed combat information for the frontend
type EnhancedCombatResult struct {
	PlayerAttack  *AttackResult `json:"playerAttack,omitempty"`
	EnemyAttack   *AttackResult `json:"enemyAttack,omitempty"`
	EnemyDefeated bool          `json:"enemyDefeated"`
	PlayerDied    bool          `json:"playerDied"`
}

// InventoryDelta tracks changes to inventory this turn
type InventoryDelta struct {
	Added   []string `json:"added,omitempty"`
	Removed []string `json:"removed,omitempty"`
	Used    []string `json:"used,omitempty"`
}

// GameContext provides contextual information about game progression
type GameContext struct {
	Phase             string  `json:"phase"`             // early_game, mid_game, late_game, exit
	TurnsInRoom       int     `json:"turnsInRoom"`
	ConsecutiveCombat int     `json:"consecutiveCombat"`
	ExplorationPct    float64 `json:"explorationPct"`
}

// UIPanel represents a suggested UI component
type UIPanel struct {
	Type     string                 `json:"type"`
	Priority string                 `json:"priority"` // high, medium, low
	Data     map[string]interface{} `json:"data"`
}

// === View Types for Frontend ===

// CharacterView is a frontend-friendly view of character state
type CharacterView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	HP        int    `json:"hp"`
	MaxHP     int    `json:"maxHp"`
	Strength  int    `json:"strength"`
	Dexterity int    `json:"dexterity"`
	IsAlive   bool   `json:"isAlive"`
	Status    string `json:"status"` // "Healthy", "Wounded", "Critical", "Dead"
}

// RoomView is a frontend-friendly view of a room
type RoomView struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	IsEntrance   bool     `json:"isEntrance"`
	IsExit       bool     `json:"isExit"`
	X            int      `json:"x"`
	Y            int      `json:"y"`
	Exits        []string `json:"exits"` // Available exit directions
	Atmosphere   string   `json:"atmosphere"`   // safe, tense, dangerous, mysterious, ominous
	IsFirstVisit bool     `json:"isFirstVisit"`
}

// MonsterView is a frontend-friendly view of a monster
type MonsterView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	HP          int    `json:"hp"`
	MaxHP       int    `json:"maxHp"`
	Damage      int    `json:"damage"`
	Threat      string `json:"threat"`      // trivial, normal, dangerous, deadly
	IsDefeated  bool   `json:"isDefeated"`
}

// ItemView is a frontend-friendly view of an item
type ItemView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // weapon, armor, consumable, key, treasure
	Damage      int    `json:"damage,omitempty"`
	Armor       int    `json:"armor,omitempty"`
	Healing     int    `json:"healing,omitempty"`
	Rarity      string `json:"rarity"` // common, uncommon, rare, legendary
	IsEquipped  bool   `json:"isEquipped"`
	IsNew       bool   `json:"isNew,omitempty"`
}

// EquipmentView shows currently equipped items
type EquipmentView struct {
	Weapon *ItemView `json:"weapon,omitempty"`
	Armor  *ItemView `json:"armor,omitempty"`
}

// MapCell represents a single cell in the map grid
type MapCell struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	RoomID    string `json:"roomId,omitempty"`
	Status    string `json:"status"` // "unknown", "visited", "current", "adjacent", "exit"
	HasPlayer bool   `json:"hasPlayer"`
	Exits     []string `json:"exits,omitempty"` // Available directions
}

// GameStateSnapshot is the complete game state for the frontend
type GameStateSnapshot struct {
	Character      *CharacterView        `json:"character,omitempty"`
	CurrentRoom    *RoomView             `json:"currentRoom,omitempty"`
	Monsters       []*MonsterView        `json:"monsters,omitempty"`
	RoomItems      []*ItemView           `json:"roomItems,omitempty"`
	Inventory      []*ItemView           `json:"inventory,omitempty"`
	Equipment      *EquipmentView        `json:"equipment,omitempty"`
	MapGrid        [][]MapCell           `json:"mapGrid,omitempty"`
	GameOver       bool                  `json:"gameOver"`
	Victory        bool                  `json:"victory"`
	TurnNumber     int                   `json:"turnNumber"`
	Message        string                `json:"message,omitempty"` // Event message for transient notifications
	Event          *EventInfo            `json:"event,omitempty"`
	CombatResult   *EnhancedCombatResult `json:"combatResult,omitempty"`
	InventoryDelta *InventoryDelta       `json:"inventoryDelta,omitempty"`
	Context        *GameContext          `json:"context,omitempty"`
}
