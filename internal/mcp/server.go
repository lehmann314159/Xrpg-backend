package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/yourusername/dungeon-crawler/internal/game"
	"github.com/yourusername/dungeon-crawler/internal/generator"
)

// Server implements the MCP protocol for the dungeon crawler
type Server struct {
	state *game.GameState
}

// NewServer creates a new MCP server instance
func NewServer() *Server {
	return &Server{
		state: game.NewGameState(),
	}
}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	Content   []ContentBlock          `json:"content"`
	IsError   bool                    `json:"isError,omitempty"`
	GameState *game.GameStateSnapshot `json:"gameState,omitempty"`
}

// ContentBlock represents a content block in the result
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// calculateThreat determines monster threat level relative to player
func (s *Server) calculateThreat(monster *game.Monster, player *game.Character) string {
	if player == nil || monster == nil {
		return "normal"
	}

	// Compare monster damage potential to player HP
	monsterThreat := float64(monster.Damage) / float64(player.HP)
	// Compare monster HP to player damage potential (base strength / 2 + 1d6 avg of 3.5)
	playerDamage := float64(player.Strength)/2 + 3.5
	turnsToKill := float64(monster.HP) / playerDamage

	if monsterThreat < 0.1 && turnsToKill < 2 {
		return "trivial"
	} else if monsterThreat >= 0.3 || turnsToKill >= 5 {
		if monsterThreat >= 0.5 || turnsToKill >= 8 {
			return "deadly"
		}
		return "dangerous"
	}
	return "normal"
}

// calculateAtmosphere determines room atmosphere based on threats and location
func (s *Server) calculateAtmosphere(room *game.Room, monsters []*game.Monster, player *game.Character) string {
	distance := room.X + room.Y // Manhattan distance

	// Check for dangerous/deadly monsters
	hasDangerousMonster := false
	for _, m := range monsters {
		threat := s.calculateThreat(m, player)
		if threat == "dangerous" || threat == "deadly" {
			hasDangerousMonster = true
			break
		}
	}

	if len(monsters) == 0 {
		if distance <= 1 {
			return "safe"
		}
		if distance >= 6 {
			return "mysterious"
		}
		return "tense"
	}

	if hasDangerousMonster {
		if distance >= 6 {
			return "ominous"
		}
		return "dangerous"
	}

	return "tense"
}

// calculatePhase determines game phase based on room position
func (s *Server) calculatePhase(room *game.Room) string {
	if room == nil {
		return "early_game"
	}
	distance := room.X + room.Y
	if room.IsExit {
		return "exit"
	}
	if distance <= 2 {
		return "early_game"
	}
	if distance <= 5 {
		return "mid_game"
	}
	return "late_game"
}

// calculateExplorationPct calculates percentage of dungeon explored
func (s *Server) calculateExplorationPct() float64 {
	if len(s.state.Rooms) == 0 {
		return 0
	}
	visited := 0
	for roomID := range s.state.Rooms {
		if s.state.IsRoomVisited(roomID) {
			visited++
		}
	}
	return float64(visited) / float64(len(s.state.Rooms)) * 100
}

// isItemNew checks if an item was just discovered this turn
func (s *Server) isItemNew(itemID string) bool {
	if s.state.TurnContext == nil {
		return false
	}
	for _, id := range s.state.TurnContext.NewItems {
		if id == itemID {
			return true
		}
	}
	return false
}

// isMonsterDefeated checks if a monster was defeated this turn
func (s *Server) isMonsterDefeated(monsterID string) bool {
	if s.state.TurnContext == nil {
		return false
	}
	for _, id := range s.state.TurnContext.DefeatedMonsters {
		if id == monsterID {
			return true
		}
	}
	return false
}

// Error messages as constants to avoid duplication
const (
	errNoGame       = "No game in progress. Use 'new_game' to start."
	errGameOver     = "Game is over. Use 'new_game' to play again."
	errVictory      = "You have escaped the dungeon! Victory! Use 'new_game' to play again."
	errDead         = "You are dead. Use 'new_game' to play again."
	errAlreadyWon   = "You have already won! Use 'new_game' to play again."
)

// requireActiveGame checks if a game is in progress and not over.
// Returns a ToolResult with an error message if the game is not active, or nil if OK.
func (s *Server) requireActiveGame() *ToolResult {
	if !s.state.IsInitialized() {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: errNoGame}},
		}
	}
	if s.state.GameOver {
		if s.state.Victory {
			return &ToolResult{
				Content: []ContentBlock{{Type: "text", Text: errVictory}},
			}
		}
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: errDead}},
		}
	}
	return nil
}

// requireActiveGameForAction is like requireActiveGame but uses a simpler "game over" message
// suitable for actions that don't need to distinguish between victory and death.
func (s *Server) requireActiveGameForAction() *ToolResult {
	if !s.state.IsInitialized() {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: errNoGame}},
		}
	}
	if s.state.GameOver {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: errGameOver}},
		}
	}
	return nil
}

// requireInitialized checks only if a game is initialized (for read-only operations like inventory/stats).
func (s *Server) requireInitialized() *ToolResult {
	if !s.state.IsInitialized() {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: errNoGame}},
		}
	}
	return nil
}

// beginTurn resets turn context and increments turn counters for a standard action.
func (s *Server) beginTurn() {
	s.state.ResetTurnContext()
	s.state.IncrementTurnsInRoom()
}

// beginCombatTurn resets turn context and increments both room and combat counters.
func (s *Server) beginCombatTurn() {
	s.state.ResetTurnContext()
	s.state.IncrementTurnsInRoom()
	s.state.IncrementConsecutiveCombat()
}

// beginMovementTurn resets turn context and resets room/combat counters for movement.
func (s *Server) beginMovementTurn() {
	s.state.ResetTurnContext()
	s.state.ResetTurnsInRoom()
	s.state.ResetConsecutiveCombat()
}

// buildGameStateSnapshot creates a snapshot of the current game state for the frontend
func (s *Server) buildGameStateSnapshot() *game.GameStateSnapshot {
	if !s.state.IsInitialized() {
		return nil
	}

	snapshot := &game.GameStateSnapshot{
		GameOver: s.state.GameOver,
		Victory:  s.state.Victory,
	}

	// Character view
	if s.state.Character != nil {
		char := s.state.Character
		status := "Healthy"
		if !char.IsAlive {
			status = "Dead"
		} else if char.HP <= char.MaxHP/4 {
			status = "Critical"
		} else if char.HP <= char.MaxHP/2 {
			status = "Wounded"
		}

		snapshot.Character = &game.CharacterView{
			ID:        char.ID,
			Name:      char.Name,
			HP:        char.HP,
			MaxHP:     char.MaxHP,
			Strength:  char.Strength,
			Dexterity: char.Dexterity,
			IsAlive:   char.IsAlive,
			Status:    status,
		}
	}

	// Current room view
	room := s.state.GetCurrentRoom()
	monsters := make([]*game.Monster, 0)
	if room != nil {
		exits := s.state.GetRoomExits(room.ID)
		exitDirs := make([]string, 0, len(exits))
		for dir := range exits {
			exitDirs = append(exitDirs, dir)
		}

		monsters = s.state.GetRoomMonsters(room.ID)

		// Check if this is the first visit (room was just marked visited this turn)
		isFirstVisit := false
		if s.state.TurnContext != nil && s.state.TurnContext.LastEvent != nil {
			if s.state.TurnContext.LastEvent.Type == "movement" {
				isFirstVisit = true
			}
		}

		snapshot.CurrentRoom = &game.RoomView{
			ID:           room.ID,
			Name:         room.Name,
			Description:  room.Description,
			IsEntrance:   room.IsEntrance,
			IsExit:       room.IsExit,
			X:            room.X,
			Y:            room.Y,
			Exits:        exitDirs,
			Atmosphere:   s.calculateAtmosphere(room, monsters, s.state.Character),
			IsFirstVisit: isFirstVisit,
		}

		// Monsters in current room
		snapshot.Monsters = make([]*game.MonsterView, 0, len(monsters))
		for _, m := range monsters {
			snapshot.Monsters = append(snapshot.Monsters, &game.MonsterView{
				ID:          m.ID,
				Name:        m.Name,
				Description: m.Description,
				HP:          m.HP,
				MaxHP:       m.MaxHP,
				Damage:      m.Damage,
				Threat:      s.calculateThreat(m, s.state.Character),
				IsDefeated:  s.isMonsterDefeated(m.ID),
			})
		}

		// Items in current room
		roomItems := s.state.GetRoomItems(room.ID)
		snapshot.RoomItems = make([]*game.ItemView, 0, len(roomItems))
		for _, item := range roomItems {
			snapshot.RoomItems = append(snapshot.RoomItems, &game.ItemView{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				Type:        item.Type,
				Damage:      item.Damage,
				Armor:       item.Armor,
				Healing:     item.Healing,
				Rarity:      item.Rarity,
				IsEquipped:  item.IsEquipped,
				IsNew:       s.isItemNew(item.ID),
			})
		}
	}

	// Inventory
	inventory := s.state.GetInventory()
	snapshot.Inventory = make([]*game.ItemView, 0, len(inventory))
	for _, item := range inventory {
		snapshot.Inventory = append(snapshot.Inventory, &game.ItemView{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Type:        item.Type,
			Damage:      item.Damage,
			Armor:       item.Armor,
			Healing:     item.Healing,
			Rarity:      item.Rarity,
			IsEquipped:  item.IsEquipped,
			IsNew:       s.isItemNew(item.ID),
		})
	}

	// Equipment
	snapshot.Equipment = &game.EquipmentView{}
	if s.state.Character != nil {
		if s.state.Character.EquippedWeaponID != nil {
			weapon := s.state.Items[*s.state.Character.EquippedWeaponID]
			if weapon != nil {
				snapshot.Equipment.Weapon = &game.ItemView{
					ID:          weapon.ID,
					Name:        weapon.Name,
					Description: weapon.Description,
					Type:        weapon.Type,
					Damage:      weapon.Damage,
					Rarity:      weapon.Rarity,
					IsEquipped:  true,
				}
			}
		}
		if s.state.Character.EquippedArmorID != nil {
			armor := s.state.Items[*s.state.Character.EquippedArmorID]
			if armor != nil {
				snapshot.Equipment.Armor = &game.ItemView{
					ID:          armor.ID,
					Name:        armor.Name,
					Description: armor.Description,
					Type:        armor.Type,
					Armor:       armor.Armor,
					Rarity:      armor.Rarity,
					IsEquipped:  true,
				}
			}
		}
	}

	// Turn context data
	if s.state.TurnContext != nil {
		snapshot.Event = s.state.TurnContext.LastEvent
		snapshot.CombatResult = s.state.TurnContext.LastCombatResult
		snapshot.InventoryDelta = s.state.TurnContext.InventoryDelta

		// Game context
		snapshot.Context = &game.GameContext{
			Phase:             s.calculatePhase(room),
			TurnsInRoom:       s.state.TurnContext.TurnsInRoom,
			ConsecutiveCombat: s.state.TurnContext.ConsecutiveCombat,
			ExplorationPct:    s.calculateExplorationPct(),
		}
	}

	// Map grid (5x5)
	gridSize := 5
	snapshot.MapGrid = make([][]game.MapCell, gridSize)
	for y := 0; y < gridSize; y++ {
		snapshot.MapGrid[y] = make([]game.MapCell, gridSize)
		for x := 0; x < gridSize; x++ {
			cell := game.MapCell{
				X:      x,
				Y:      y,
				Status: "unknown",
			}

			mapRoom := s.state.GetRoomAt(x, y)
			if mapRoom != nil {
				cell.RoomID = mapRoom.ID

				// Get exits for this room
				exits := s.state.GetRoomExits(mapRoom.ID)
				cell.Exits = make([]string, 0, len(exits))
				for dir := range exits {
					cell.Exits = append(cell.Exits, dir)
				}

				if room != nil && mapRoom.ID == room.ID {
					cell.Status = "current"
					cell.HasPlayer = true
				} else if mapRoom.IsExit && s.state.IsRoomVisited(mapRoom.ID) {
					cell.Status = "exit"
				} else if s.state.IsRoomVisited(mapRoom.ID) {
					cell.Status = "visited"
				} else if s.state.IsRoomAdjacent(mapRoom.ID) {
					cell.Status = "adjacent"
				}
			}

			snapshot.MapGrid[y][x] = cell
		}
	}

	return snapshot
}

// ListTools returns all available MCP tools
func (s *Server) ListTools() []Tool {
	return []Tool{
		{
			Name:        "new_game",
			Description: "Start a new game with a character",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"character_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of your character",
					},
				},
				"required": []string{"character_name"},
			},
		},
		{
			Name:        "look",
			Description: "Look around the current room to see exits, monsters, items, and traps",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "move",
			Description: "Move in a direction (north, south, east, west)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"direction": map[string]interface{}{
						"type":        "string",
						"description": "Direction to move (north, south, east, west)",
						"enum":        []string{"north", "south", "east", "west"},
					},
				},
				"required": []string{"direction"},
			},
		},
		{
			Name:        "attack",
			Description: "Attack a monster in the current room",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"target_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the monster to attack",
					},
				},
				"required": []string{"target_id"},
			},
		},
		{
			Name:        "take",
			Description: "Pick up an item from the current room",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"item_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the item to pick up",
					},
				},
				"required": []string{"item_id"},
			},
		},
		{
			Name:        "use",
			Description: "Use an item from inventory",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"item_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the item to use",
					},
				},
				"required": []string{"item_id"},
			},
		},
		{
			Name:        "inventory",
			Description: "View current inventory",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "stats",
			Description: "View character stats",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "map",
			Description: "View the dungeon map showing explored areas",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "equip",
			Description: "Equip a weapon or armor from your inventory",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"item_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the item to equip",
					},
				},
				"required": []string{"item_id"},
			},
		},
	}
}

// CallTool executes an MCP tool
func (s *Server) CallTool(name string, arguments map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "new_game":
		charName, ok := arguments["character_name"].(string)
		if !ok || charName == "" {
			charName = "Hero"
		}
		return s.handleNewGame(charName)
	case "look":
		return s.handleLook()
	case "move":
		direction, ok := arguments["direction"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid direction")
		}
		return s.handleMove(direction)
	case "attack":
		targetID, ok := arguments["target_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid target_id")
		}
		return s.handleAttack(targetID)
	case "take":
		itemID, ok := arguments["item_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid item_id")
		}
		return s.handleTake(itemID)
	case "use":
		itemID, ok := arguments["item_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid item_id")
		}
		return s.handleUse(itemID)
	case "inventory":
		return s.handleInventory()
	case "stats":
		return s.handleStats()
	case "map":
		return s.handleMap()
	case "equip":
		itemID, ok := arguments["item_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid item_id")
		}
		return s.handleEquip(itemID)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// handleNewGame starts a new game
func (s *Server) handleNewGame(characterName string) (*ToolResult, error) {
	// Reset game state
	s.state = game.NewGameState()

	// Create character
	character := game.NewCharacter(characterName)
	s.state.Character = character

	// Generate dungeon
	seed := time.Now().UnixNano()
	gen := generator.NewDungeonGenerator(seed)
	dungeon, rooms, connections, err := gen.GenerateDungeon(1) // Depth 1 for MVP
	if err != nil {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Failed to generate dungeon: %v", err)}},
			IsError: true,
		}, nil
	}

	s.state.Dungeon = dungeon

	// Add rooms and connections
	for _, room := range rooms {
		s.state.AddRoom(room)
	}
	for _, conn := range connections {
		s.state.AddConnection(conn)
	}

	// Find entrance and set character's starting position
	for _, room := range rooms {
		if room.IsEntrance {
			character.CurrentRoomID = room.ID
			s.state.MarkRoomVisited(room.ID)
			break
		}
	}

	// Populate rooms with monsters and items
	for _, room := range rooms {
		difficulty := room.X + room.Y // Manhattan distance from entrance
		monsters, items, traps := gen.PopulateRoom(room, difficulty)
		for _, m := range monsters {
			s.state.AddMonster(m)
		}
		for _, item := range items {
			s.state.AddItem(item)
		}
		for _, trap := range traps {
			s.state.AddTrap(trap)
		}
	}

	// Initialize turn context with game start event
	s.state.ResetTurnContext()
	s.state.SetLastEvent(&game.EventInfo{
		Type:    "interaction",
		Subtype: "game_start",
	})

	// Build response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== NEW GAME STARTED ===\n\n"))
	sb.WriteString(fmt.Sprintf("Welcome, %s!\n\n", character.Name))
	sb.WriteString(fmt.Sprintf("You find yourself at the entrance of a dark dungeon.\n"))
	sb.WriteString(fmt.Sprintf("Your goal: reach the exit on the other side.\n"))
	sb.WriteString(fmt.Sprintf("Beware of the monsters that lurk within!\n\n"))
	sb.WriteString(fmt.Sprintf("Stats: HP %d/%d | STR %d | DEX %d\n\n",
		character.HP, character.MaxHP, character.Strength, character.Dexterity))
	sb.WriteString("Use 'look' to see your surroundings.")

	return &ToolResult{
		Content:   []ContentBlock{{Type: "text", Text: sb.String()}},
		GameState: s.buildGameStateSnapshot(),
	}, nil
}

// handleLook shows the current room
func (s *Server) handleLook() (*ToolResult, error) {
	if errResult := s.requireActiveGame(); errResult != nil {
		return errResult, nil
	}

	room := s.state.GetCurrentRoom()
	if room == nil {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: "Error: Current room not found."}},
			IsError: true,
		}, nil
	}

	s.beginTurn()
	s.state.SetLastEvent(&game.EventInfo{
		Type:    "interaction",
		Subtype: "look",
	})

	var sb strings.Builder

	// Room header
	sb.WriteString(fmt.Sprintf("=== %s ===\n\n", room.Name))
	sb.WriteString(fmt.Sprintf("%s\n\n", room.Description))

	// Special room markers
	if room.IsEntrance {
		sb.WriteString("[This is the dungeon entrance]\n\n")
	}
	if room.IsExit {
		sb.WriteString("[This is the dungeon exit - reach here to win!]\n\n")
	}

	// Exits
	exits := s.state.GetRoomExits(room.ID)
	if len(exits) > 0 {
		exitDirs := make([]string, 0, len(exits))
		for dir := range exits {
			exitDirs = append(exitDirs, dir)
		}
		sb.WriteString(fmt.Sprintf("Exits: %s\n\n", strings.Join(exitDirs, ", ")))
	} else {
		sb.WriteString("Exits: none\n\n")
	}

	// Monsters
	monsters := s.state.GetRoomMonsters(room.ID)
	if len(monsters) > 0 {
		sb.WriteString("Monsters:\n")
		for _, m := range monsters {
			sb.WriteString(fmt.Sprintf("  - %s (HP: %d/%d) [ID: %s]\n", m.Name, m.HP, m.MaxHP, m.ID))
			sb.WriteString(fmt.Sprintf("    %s\n", m.Description))
		}
		sb.WriteString("\n")
	}

	// Items on the floor
	items := s.state.GetRoomItems(room.ID)
	if len(items) > 0 {
		sb.WriteString("Items:\n")
		for _, item := range items {
			sb.WriteString(fmt.Sprintf("  - %s [ID: %s]\n", item.Name, item.ID))
			sb.WriteString(fmt.Sprintf("    %s\n", item.Description))
		}
		sb.WriteString("\n")
	}

	// Warning if monsters block exit
	if len(monsters) > 0 {
		sb.WriteString("⚔️  Monsters block your path! Defeat them to proceed.\n")
	}

	return &ToolResult{
		Content:   []ContentBlock{{Type: "text", Text: sb.String()}},
		GameState: s.buildGameStateSnapshot(),
	}, nil
}

// handleMove moves the character
func (s *Server) handleMove(direction string) (*ToolResult, error) {
	if errResult := s.requireActiveGame(); errResult != nil {
		return errResult, nil
	}

	s.beginMovementTurn()

	err := s.state.MoveCharacter(direction)
	if err != nil {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: err.Error()}},
		}, nil
	}

	s.state.SetLastEvent(&game.EventInfo{
		Type:    "movement",
		Subtype: "room_enter",
	})

	// Check for victory
	if s.state.Victory {
		s.state.SetLastEvent(&game.EventInfo{
			Type:    "victory",
			Subtype: "dungeon_escaped",
		})
		return &ToolResult{
			Content:   []ContentBlock{{Type: "text", Text: "You step through the exit and escape the dungeon!\n\n🏆 VICTORY! 🏆\n\nCongratulations, brave adventurer! Use 'new_game' to play again."}},
			GameState: s.buildGameStateSnapshot(),
		}, nil
	}

	// Show the new room
	newRoom := s.state.GetCurrentRoom()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You move %s...\n\n", direction))
	sb.WriteString(fmt.Sprintf("=== %s ===\n\n", newRoom.Name))
	sb.WriteString(fmt.Sprintf("%s\n", newRoom.Description))

	// Check for monsters
	monsters := s.state.GetRoomMonsters(newRoom.ID)
	if len(monsters) > 0 {
		sb.WriteString("\n⚔️  Danger! Monsters ahead!\n")
		for _, m := range monsters {
			sb.WriteString(fmt.Sprintf("  - %s (HP: %d/%d) [ID: %s]\n", m.Name, m.HP, m.MaxHP, m.ID))
		}
	}

	return &ToolResult{
		Content:   []ContentBlock{{Type: "text", Text: sb.String()}},
		GameState: s.buildGameStateSnapshot(),
	}, nil
}

// handleAttack attacks a monster
func (s *Server) handleAttack(targetID string) (*ToolResult, error) {
	if errResult := s.requireActiveGameForAction(); errResult != nil {
		return errResult, nil
	}

	// Find the monster
	monster, ok := s.state.Monsters[targetID]
	if !ok {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: "Monster not found. Use 'look' to see available targets."}},
		}, nil
	}

	// Check monster is in current room
	if monster.RoomID != s.state.Character.CurrentRoomID {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: "That monster is not in this room."}},
		}, nil
	}

	// Check monster is alive
	if !monster.IsAlive {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: "That monster is already dead."}},
		}, nil
	}

	s.beginCombatTurn()

	// Calculate equipment bonuses
	weaponBonus := 0
	armorBonus := 0
	if s.state.Character.EquippedWeaponID != nil {
		weapon := s.state.Items[*s.state.Character.EquippedWeaponID]
		if weapon != nil {
			weaponBonus = weapon.Damage
		}
	}
	if s.state.Character.EquippedArmorID != nil {
		armor := s.state.Items[*s.state.Character.EquippedArmorID]
		if armor != nil {
			armorBonus = armor.Armor
		}
	}

	// Execute combat turn
	result, enhanced, _ := game.ExecuteCombatTurn(s.state.Character, monster, "attack", weaponBonus, armorBonus)

	// Store enhanced combat result
	s.state.SetLastCombatResult(enhanced)

	// Determine event subtype based on outcome
	eventSubtype := "attack_hit"
	if enhanced.PlayerAttack != nil && !enhanced.PlayerAttack.WasHit {
		eventSubtype = "attack_miss"
	}

	var sb strings.Builder
	sb.WriteString("=== COMBAT ===\n\n")
	sb.WriteString(result.Message)
	sb.WriteString("\n")

	// Check for player death
	if result.AttackerDied {
		s.state.KillCharacter()
		s.state.SetLastEvent(&game.EventInfo{
			Type:     "death",
			Subtype:  "player_died",
			Entities: []string{targetID},
		})
		sb.WriteString("\n💀 YOU HAVE DIED 💀\n\nUse 'new_game' to try again.")
	} else if result.DefenderDied {
		s.state.KillMonster(targetID)
		s.state.RecordMonsterDefeated(targetID)
		s.state.SetLastEvent(&game.EventInfo{
			Type:     "combat",
			Subtype:  "enemy_defeated",
			Entities: []string{targetID},
		})
		sb.WriteString(fmt.Sprintf("\n✨ The %s has been defeated!\n", monster.Name))

		// Check if room is clear
		if !s.state.HasMonstersInRoom(s.state.Character.CurrentRoomID) {
			sb.WriteString("\nThe room is now clear. You may proceed.")
		}
	} else {
		s.state.SetLastEvent(&game.EventInfo{
			Type:     "combat",
			Subtype:  eventSubtype,
			Entities: []string{targetID},
		})
	}

	return &ToolResult{
		Content:   []ContentBlock{{Type: "text", Text: sb.String()}},
		GameState: s.buildGameStateSnapshot(),
	}, nil
}

// handleTake picks up an item
func (s *Server) handleTake(itemID string) (*ToolResult, error) {
	if errResult := s.requireActiveGameForAction(); errResult != nil {
		return errResult, nil
	}

	item, ok := s.state.Items[itemID]
	if !ok {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: "Item not found. Use 'look' to see available items."}},
		}, nil
	}

	s.beginTurn()

	err := s.state.TakeItem(itemID)
	if err != nil {
		return &ToolResult{
			Content:   []ContentBlock{{Type: "text", Text: err.Error()}},
			GameState: s.buildGameStateSnapshot(),
		}, nil
	}

	// Track item taken
	s.state.RecordItemTaken(itemID)
	s.state.SetLastEvent(&game.EventInfo{
		Type:     "discovery",
		Subtype:  "item_found",
		Entities: []string{itemID},
	})

	return &ToolResult{
		Content:   []ContentBlock{{Type: "text", Text: fmt.Sprintf("You pick up the %s.", item.Name)}},
		GameState: s.buildGameStateSnapshot(),
	}, nil
}

// handleUse uses an item from inventory
func (s *Server) handleUse(itemID string) (*ToolResult, error) {
	if errResult := s.requireActiveGameForAction(); errResult != nil {
		return errResult, nil
	}

	s.beginTurn()

	// Track item used before it's removed
	s.state.RecordItemUsed(itemID)

	message, err := s.state.UseItem(itemID)
	if err != nil {
		return &ToolResult{
			Content:   []ContentBlock{{Type: "text", Text: err.Error()}},
			GameState: s.buildGameStateSnapshot(),
		}, nil
	}

	s.state.SetLastEvent(&game.EventInfo{
		Type:     "interaction",
		Subtype:  "item_used",
		Entities: []string{itemID},
	})

	return &ToolResult{
		Content:   []ContentBlock{{Type: "text", Text: message}},
		GameState: s.buildGameStateSnapshot(),
	}, nil
}

// handleInventory shows the character's inventory
func (s *Server) handleInventory() (*ToolResult, error) {
	if errResult := s.requireInitialized(); errResult != nil {
		return errResult, nil
	}

	items := s.state.GetInventory()

	var sb strings.Builder
	sb.WriteString("=== INVENTORY ===\n\n")

	if len(items) == 0 {
		sb.WriteString("Your inventory is empty.\n")
	} else {
		for _, item := range items {
			equippedMarker := ""
			if item.IsEquipped {
				equippedMarker = " [EQUIPPED]"
			}
			sb.WriteString(fmt.Sprintf("- %s%s [ID: %s]\n", item.Name, equippedMarker, item.ID))
			sb.WriteString(fmt.Sprintf("  %s\n", item.Description))
			if item.Type == "consumable" && item.Healing > 0 {
				sb.WriteString(fmt.Sprintf("  (Heals %d HP)\n", item.Healing))
			}
			if item.Type == "weapon" && item.Damage > 0 {
				sb.WriteString(fmt.Sprintf("  (Damage +%d)\n", item.Damage))
			}
			if item.Type == "armor" && item.Armor > 0 {
				sb.WriteString(fmt.Sprintf("  (Armor +%d)\n", item.Armor))
			}
		}
	}

	return &ToolResult{
		Content:   []ContentBlock{{Type: "text", Text: sb.String()}},
		GameState: s.buildGameStateSnapshot(),
	}, nil
}

// handleStats shows character stats
func (s *Server) handleStats() (*ToolResult, error) {
	if errResult := s.requireInitialized(); errResult != nil {
		return errResult, nil
	}

	char := s.state.Character

	var sb strings.Builder
	sb.WriteString("=== CHARACTER STATS ===\n\n")
	sb.WriteString(fmt.Sprintf("Name: %s\n", char.Name))
	sb.WriteString(fmt.Sprintf("HP: %d/%d\n", char.HP, char.MaxHP))
	sb.WriteString(fmt.Sprintf("Strength: %d\n", char.Strength))
	sb.WriteString(fmt.Sprintf("Dexterity: %d\n", char.Dexterity))
	sb.WriteString(fmt.Sprintf("Status: %s\n", func() string {
		if !char.IsAlive {
			return "Dead"
		}
		if char.HP <= char.MaxHP/4 {
			return "Critical"
		}
		if char.HP <= char.MaxHP/2 {
			return "Wounded"
		}
		return "Healthy"
	}()))

	// Show equipped items
	sb.WriteString("\n--- Equipment ---\n")
	if char.EquippedWeaponID != nil {
		weapon := s.state.Items[*char.EquippedWeaponID]
		if weapon != nil {
			sb.WriteString(fmt.Sprintf("Weapon: %s (+%d damage)\n", weapon.Name, weapon.Damage))
		}
	} else {
		sb.WriteString("Weapon: None (bare hands)\n")
	}
	if char.EquippedArmorID != nil {
		armor := s.state.Items[*char.EquippedArmorID]
		if armor != nil {
			sb.WriteString(fmt.Sprintf("Armor: %s (+%d defense)\n", armor.Name, armor.Armor))
		}
	} else {
		sb.WriteString("Armor: None\n")
	}

	if s.state.Victory {
		sb.WriteString("\n🏆 VICTORIOUS 🏆\n")
	}

	// Also show current room summary
	room := s.state.GetCurrentRoom()
	if room != nil {
		sb.WriteString(fmt.Sprintf("\nLocation: %s\n", room.Name))
	}

	// Show inventory count
	inv := s.state.GetInventory()
	sb.WriteString(fmt.Sprintf("Inventory: %d items\n", len(inv)))

	return &ToolResult{
		Content:   []ContentBlock{{Type: "text", Text: sb.String()}},
		GameState: s.buildGameStateSnapshot(),
	}, nil
}

// handleMap shows the dungeon map
func (s *Server) handleMap() (*ToolResult, error) {
	if errResult := s.requireInitialized(); errResult != nil {
		return errResult, nil
	}

	mapStr := s.state.RenderMap(5) // 5x5 grid

	return &ToolResult{
		Content:   []ContentBlock{{Type: "text", Text: mapStr}},
		GameState: s.buildGameStateSnapshot(),
	}, nil
}

// handleEquip equips a weapon or armor
func (s *Server) handleEquip(itemID string) (*ToolResult, error) {
	if errResult := s.requireActiveGameForAction(); errResult != nil {
		return errResult, nil
	}

	item, ok := s.state.Items[itemID]
	if !ok {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: "Item not found. Use 'inventory' to see your items."}},
		}, nil
	}

	// Check item is in inventory
	if item.CharacterID == nil || *item.CharacterID != s.state.Character.ID {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: "That item is not in your inventory. Pick it up first with 'take'."}},
		}, nil
	}

	char := s.state.Character
	var sb strings.Builder

	switch item.Type {
	case "weapon":
		// Unequip old weapon if any
		if char.EquippedWeaponID != nil {
			oldWeapon := s.state.Items[*char.EquippedWeaponID]
			if oldWeapon != nil {
				oldWeapon.IsEquipped = false
				sb.WriteString(fmt.Sprintf("You unequip the %s.\n", oldWeapon.Name))
			}
		}
		// Equip new weapon
		char.EquippedWeaponID = &item.ID
		item.IsEquipped = true
		sb.WriteString(fmt.Sprintf("You equip the %s. (Damage +%d)", item.Name, item.Damage))

	case "armor":
		// Unequip old armor if any
		if char.EquippedArmorID != nil {
			oldArmor := s.state.Items[*char.EquippedArmorID]
			if oldArmor != nil {
				oldArmor.IsEquipped = false
				sb.WriteString(fmt.Sprintf("You unequip the %s.\n", oldArmor.Name))
			}
		}
		// Equip new armor
		char.EquippedArmorID = &item.ID
		item.IsEquipped = true
		sb.WriteString(fmt.Sprintf("You equip the %s. (Armor +%d)", item.Name, item.Armor))

	default:
		return &ToolResult{
			Content:   []ContentBlock{{Type: "text", Text: fmt.Sprintf("Cannot equip %s - it's not a weapon or armor.", item.Name)}},
			GameState: s.buildGameStateSnapshot(),
		}, nil
	}

	return &ToolResult{
		Content:   []ContentBlock{{Type: "text", Text: sb.String()}},
		GameState: s.buildGameStateSnapshot(),
	}, nil
}

// Helper to marshal to JSON for debugging
func toJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
