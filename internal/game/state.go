package game

import (
	"fmt"
	"strings"
	"time"
)

// GameState holds all in-memory game state
type GameState struct {
	Character    *Character
	Dungeon      *Dungeon
	Rooms        map[string]*Room             // keyed by room ID
	RoomsByCoord map[string]*Room             // keyed by "x,y"
	Connections  map[string][]*RoomConnection // keyed by room ID
	Monsters     map[string]*Monster          // keyed by monster ID
	Items        map[string]*Item             // keyed by item ID
	Traps        map[string]*Trap             // keyed by trap ID
	VisitedRooms map[string]bool              // keyed by room ID
	GameOver     bool
	Victory      bool
}

// NewGameState creates an empty game state
func NewGameState() *GameState {
	return &GameState{
		Rooms:        make(map[string]*Room),
		RoomsByCoord: make(map[string]*Room),
		Connections:  make(map[string][]*RoomConnection),
		Monsters:     make(map[string]*Monster),
		Items:        make(map[string]*Item),
		Traps:        make(map[string]*Trap),
		VisitedRooms: make(map[string]bool),
	}
}

// IsInitialized returns true if a game has been started
func (gs *GameState) IsInitialized() bool {
	return gs.Character != nil && gs.Dungeon != nil
}

// GetCurrentRoom returns the room the character is in
func (gs *GameState) GetCurrentRoom() *Room {
	if gs.Character == nil {
		return nil
	}
	return gs.Rooms[gs.Character.CurrentRoomID]
}

// GetRoomExits returns the exits available from a room
func (gs *GameState) GetRoomExits(roomID string) map[string]string {
	exits := make(map[string]string) // direction -> room ID
	for _, conn := range gs.Connections[roomID] {
		exits[conn.Direction] = conn.ConnectedRoomID
	}
	return exits
}

// GetRoomMonsters returns all alive monsters in a room
func (gs *GameState) GetRoomMonsters(roomID string) []*Monster {
	monsters := make([]*Monster, 0)
	for _, m := range gs.Monsters {
		if m.RoomID == roomID && m.IsAlive {
			monsters = append(monsters, m)
		}
	}
	return monsters
}

// GetRoomItems returns all items in a room (not in inventory)
func (gs *GameState) GetRoomItems(roomID string) []*Item {
	items := make([]*Item, 0)
	for _, item := range gs.Items {
		if item.RoomID != nil && *item.RoomID == roomID {
			items = append(items, item)
		}
	}
	return items
}

// GetRoomTraps returns all traps in a room
func (gs *GameState) GetRoomTraps(roomID string) []*Trap {
	traps := make([]*Trap, 0)
	for _, trap := range gs.Traps {
		if trap.RoomID == roomID {
			traps = append(traps, trap)
		}
	}
	return traps
}

// GetInventory returns all items carried by the character
func (gs *GameState) GetInventory() []*Item {
	items := make([]*Item, 0)
	if gs.Character == nil {
		return items
	}
	for _, item := range gs.Items {
		if item.CharacterID != nil && *item.CharacterID == gs.Character.ID {
			items = append(items, item)
		}
	}
	return items
}

// HasMonstersInRoom returns true if there are alive monsters in the room
func (gs *GameState) HasMonstersInRoom(roomID string) bool {
	return len(gs.GetRoomMonsters(roomID)) > 0
}

// MoveCharacter moves the character in a direction
func (gs *GameState) MoveCharacter(direction string) error {
	if gs.Character == nil {
		return fmt.Errorf("no character")
	}
	if !gs.Character.IsAlive {
		return fmt.Errorf("character is dead")
	}

	currentRoomID := gs.Character.CurrentRoomID

	// Check for monsters blocking movement
	if gs.HasMonstersInRoom(currentRoomID) {
		return fmt.Errorf("cannot leave while monsters are present - defeat them first")
	}

	// Find the exit in the given direction
	exits := gs.GetRoomExits(currentRoomID)
	newRoomID, ok := exits[direction]
	if !ok {
		return fmt.Errorf("cannot move %s - no exit in that direction", direction)
	}

	// Move the character
	gs.Character.CurrentRoomID = newRoomID
	gs.VisitedRooms[newRoomID] = true

	// Check for victory
	newRoom := gs.Rooms[newRoomID]
	if newRoom != nil && newRoom.IsExit {
		gs.Victory = true
		gs.GameOver = true
	}

	return nil
}

// TakeItem moves an item from the room to the character's inventory
func (gs *GameState) TakeItem(itemID string) error {
	if gs.Character == nil {
		return fmt.Errorf("no character")
	}

	item, ok := gs.Items[itemID]
	if !ok {
		return fmt.Errorf("item not found")
	}

	// Check item is in current room
	if item.RoomID == nil || *item.RoomID != gs.Character.CurrentRoomID {
		return fmt.Errorf("item is not in this room")
	}

	// Check item is not already in someone's inventory
	if item.CharacterID != nil {
		return fmt.Errorf("item is already being carried")
	}

	// Move to inventory
	item.RoomID = nil
	item.CharacterID = &gs.Character.ID

	return nil
}

// UseItem uses a consumable item from inventory
func (gs *GameState) UseItem(itemID string) (string, error) {
	if gs.Character == nil {
		return "", fmt.Errorf("no character")
	}

	item, ok := gs.Items[itemID]
	if !ok {
		return "", fmt.Errorf("item not found")
	}

	// Check item is in inventory
	if item.CharacterID == nil || *item.CharacterID != gs.Character.ID {
		return "", fmt.Errorf("item is not in your inventory")
	}

	if item.Type != "consumable" {
		return "", fmt.Errorf("cannot use this item - it's not consumable")
	}

	// Apply effects
	var message string
	if item.Healing > 0 {
		oldHP := gs.Character.HP
		gs.Character.Heal(item.Healing)
		healed := gs.Character.HP - oldHP
		message = fmt.Sprintf("You drink the %s and recover %d HP! (HP: %d/%d)",
			item.Name, healed, gs.Character.HP, gs.Character.MaxHP)
	} else {
		message = fmt.Sprintf("You use the %s.", item.Name)
	}

	// Remove item from inventory
	delete(gs.Items, itemID)

	return message, nil
}

// KillMonster marks a monster as dead and potentially drops loot
func (gs *GameState) KillMonster(monsterID string) []*Item {
	monster, ok := gs.Monsters[monsterID]
	if !ok {
		return nil
	}

	monster.IsAlive = false
	monster.HP = 0

	// Drop loot (for now, no loot drops - items are pre-placed)
	return nil
}

// KillCharacter marks the character as dead
func (gs *GameState) KillCharacter() {
	if gs.Character == nil {
		return
	}
	gs.Character.IsAlive = false
	gs.Character.HP = 0
	now := time.Now()
	gs.Character.DiedAt = &now
	gs.GameOver = true
	gs.Victory = false
}

// AddRoom adds a room to the game state
func (gs *GameState) AddRoom(room *Room) {
	gs.Rooms[room.ID] = room
	gs.RoomsByCoord[fmt.Sprintf("%d,%d", room.X, room.Y)] = room
}

// AddConnection adds a room connection to the game state
func (gs *GameState) AddConnection(conn *RoomConnection) {
	gs.Connections[conn.RoomID] = append(gs.Connections[conn.RoomID], conn)
}

// AddMonster adds a monster to the game state
func (gs *GameState) AddMonster(monster *Monster) {
	gs.Monsters[monster.ID] = monster
}

// AddItem adds an item to the game state
func (gs *GameState) AddItem(item *Item) {
	gs.Items[item.ID] = item
}

// AddTrap adds a trap to the game state
func (gs *GameState) AddTrap(trap *Trap) {
	gs.Traps[trap.ID] = trap
}

// MarkRoomVisited marks a room as visited
func (gs *GameState) MarkRoomVisited(roomID string) {
	gs.VisitedRooms[roomID] = true
}

// IsRoomVisited returns true if a room has been visited
func (gs *GameState) IsRoomVisited(roomID string) bool {
	return gs.VisitedRooms[roomID]
}

// GetRoomAt returns the room at the given coordinates
func (gs *GameState) GetRoomAt(x, y int) *Room {
	return gs.RoomsByCoord[fmt.Sprintf("%d,%d", x, y)]
}

// IsRoomAdjacent returns true if a room is adjacent to a visited room
func (gs *GameState) IsRoomAdjacent(roomID string) bool {
	room := gs.Rooms[roomID]
	if room == nil {
		return false
	}

	// Check all 4 adjacent positions
	adjacentCoords := []struct{ x, y int }{
		{room.X - 1, room.Y},
		{room.X + 1, room.Y},
		{room.X, room.Y - 1},
		{room.X, room.Y + 1},
	}

	for _, coord := range adjacentCoords {
		adjRoom := gs.GetRoomAt(coord.x, coord.y)
		if adjRoom != nil && gs.IsRoomVisited(adjRoom.ID) {
			// Check if there's actually a connection between them
			exits := gs.GetRoomExits(adjRoom.ID)
			for _, connectedID := range exits {
				if connectedID == roomID {
					return true
				}
			}
		}
	}
	return false
}

// RenderMap generates a box-drawing map of the dungeon
func (gs *GameState) RenderMap(gridSize int) string {
	currentRoom := gs.GetCurrentRoom()
	if currentRoom == nil {
		return "No map available"
	}

	var sb strings.Builder

	// Top border
	sb.WriteString("┌")
	for x := 0; x < gridSize; x++ {
		sb.WriteString("───")
		if x < gridSize-1 {
			sb.WriteString("┬")
		}
	}
	sb.WriteString("┐\n")

	// Rows (from top to bottom, so y goes from gridSize-1 to 0)
	for y := gridSize - 1; y >= 0; y-- {
		// Room row
		sb.WriteString("│")
		for x := 0; x < gridSize; x++ {
			room := gs.GetRoomAt(x, y)
			cell := "   " // Unknown

			if room != nil {
				if room.ID == currentRoom.ID {
					cell = " @ " // Current location
				} else if room.IsExit && gs.IsRoomVisited(room.ID) {
					cell = " E " // Exit (discovered)
				} else if gs.IsRoomVisited(room.ID) {
					cell = " # " // Explored
				} else if gs.IsRoomAdjacent(room.ID) {
					cell = " ? " // Adjacent/accessible
				}
			}

			sb.WriteString(cell)
			if x < gridSize-1 {
				sb.WriteString("│")
			}
		}
		sb.WriteString("│\n")

		// Row separator or bottom border
		if y > 0 {
			sb.WriteString("├")
			for x := 0; x < gridSize; x++ {
				sb.WriteString("───")
				if x < gridSize-1 {
					sb.WriteString("┼")
				}
			}
			sb.WriteString("┤\n")
		}
	}

	// Bottom border
	sb.WriteString("└")
	for x := 0; x < gridSize; x++ {
		sb.WriteString("───")
		if x < gridSize-1 {
			sb.WriteString("┴")
		}
	}
	sb.WriteString("┘\n")

	// Legend
	sb.WriteString("\n@ = You  # = Explored  ? = Adjacent  E = Exit")

	return sb.String()
}
