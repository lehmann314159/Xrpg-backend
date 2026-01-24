package generator

import (
	"crypto/rand"
	"fmt"
	mrand "math/rand"

	"github.com/yourusername/dungeon-crawler/internal/game"
)

const GridSize = 5

// DungeonGenerator handles procedural dungeon generation
type DungeonGenerator struct {
	seed   int64
	random *mrand.Rand
}

// generateID creates a simple random ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// NewDungeonGenerator creates a new dungeon generator
func NewDungeonGenerator(seed int64) *DungeonGenerator {
	return &DungeonGenerator{
		seed:   seed,
		random: mrand.New(mrand.NewSource(seed)),
	}
}

// coord represents a position in the grid
type coord struct {
	x, y int
}

// edge represents a connection between two rooms
type edge struct {
	from, to  coord
	direction string
}

// GenerateDungeon creates a new procedural dungeon
func (dg *DungeonGenerator) GenerateDungeon(depth int) (*game.Dungeon, []*game.Room, []*game.RoomConnection, error) {
	dungeon := &game.Dungeon{
		ID:    generateID(),
		Seed:  dg.seed,
		Depth: depth,
	}

	// Generate 5x5 grid
	rooms, roomGrid := dg.generateGrid(dungeon.ID)

	// Generate spanning tree for connectivity, then add extra doors
	connections := dg.generateConnections(roomGrid)

	return dungeon, rooms, connections, nil
}

// generateGrid creates a 5x5 grid of rooms
func (dg *DungeonGenerator) generateGrid(dungeonID string) ([]*game.Room, map[coord]*game.Room) {
	rooms := make([]*game.Room, 0, GridSize*GridSize)
	roomGrid := make(map[coord]*game.Room)

	for y := 0; y < GridSize; y++ {
		for x := 0; x < GridSize; x++ {
			room := &game.Room{
				ID:          generateID(),
				DungeonID:   dungeonID,
				Name:        dg.generateRoomName(),
				Description: dg.generateRoomDescription(x, y),
				IsEntrance:  x == 0 && y == 0,
				IsExit:      x == GridSize-1 && y == GridSize-1,
				X:           x,
				Y:           y,
			}
			rooms = append(rooms, room)
			roomGrid[coord{x, y}] = room
		}
	}

	return rooms, roomGrid
}

// generateConnections creates room connections using spanning tree + extra doors
func (dg *DungeonGenerator) generateConnections(roomGrid map[coord]*game.Room) []*game.RoomConnection {
	// Track which edges exist (bidirectional)
	edges := make(map[coord]map[string]bool) // coord -> direction -> exists
	for c := range roomGrid {
		edges[c] = make(map[string]bool)
	}

	// Step 1: Generate spanning tree using randomized Prim's algorithm
	visited := make(map[coord]bool)
	frontier := make([]edge, 0)

	// Start from entrance (0,0)
	start := coord{0, 0}
	visited[start] = true
	dg.addFrontierEdges(&frontier, start, visited)

	for len(frontier) > 0 {
		// Pick random edge from frontier
		idx := dg.random.Intn(len(frontier))
		e := frontier[idx]
		// Remove from frontier
		frontier = append(frontier[:idx], frontier[idx+1:]...)

		if visited[e.to] {
			continue
		}

		// Add this edge to the tree
		visited[e.to] = true
		edges[e.from][e.direction] = true
		edges[e.to][oppositeDir(e.direction)] = true

		// Add new frontier edges
		dg.addFrontierEdges(&frontier, e.to, visited)
	}

	// Step 2: Count doors per room and add extras to hit distribution
	// Target: 25% 1-door, 50% 2-door, 25% 3-door
	doorCounts := make(map[coord]int)
	for c := range roomGrid {
		doorCounts[c] = len(edges[c])
	}

	// Get all possible edges that don't exist yet
	possibleEdges := dg.getPossibleEdges(roomGrid, edges)
	dg.shuffleEdges(possibleEdges)

	// Add edges to reach target distribution
	for _, e := range possibleEdges {
		fromDoors := doorCounts[e.from]
		toDoors := doorCounts[e.to]

		// Don't exceed 3 doors per room
		if fromDoors >= 3 || toDoors >= 3 {
			continue
		}

		// Decide if we should add this edge based on distribution goals
		// Rooms with 1 door should sometimes get more, rooms with 2 might get a 3rd
		shouldAdd := false
		if fromDoors == 1 && dg.random.Float32() < 0.75 {
			shouldAdd = true // 1-door rooms often need more
		} else if fromDoors == 2 && dg.random.Float32() < 0.25 {
			shouldAdd = true // Some 2-door rooms become 3-door
		}

		if shouldAdd {
			edges[e.from][e.direction] = true
			edges[e.to][oppositeDir(e.direction)] = true
			doorCounts[e.from]++
			doorCounts[e.to]++
		}
	}

	// Step 3: Convert edges map to RoomConnection slice
	connections := make([]*game.RoomConnection, 0)
	for c, dirs := range edges {
		room := roomGrid[c]
		for dir := range dirs {
			neighbor := getNeighbor(c, dir)
			neighborRoom := roomGrid[neighbor]
			if neighborRoom != nil {
				connections = append(connections, &game.RoomConnection{
					ID:              generateID(),
					RoomID:          room.ID,
					Direction:       dir,
					ConnectedRoomID: neighborRoom.ID,
				})
			}
		}
	}

	return connections
}

// addFrontierEdges adds all edges from a coord to unvisited neighbors
func (dg *DungeonGenerator) addFrontierEdges(frontier *[]edge, c coord, visited map[coord]bool) {
	directions := []string{"north", "south", "east", "west"}
	for _, dir := range directions {
		neighbor := getNeighbor(c, dir)
		if neighbor.x >= 0 && neighbor.x < GridSize && neighbor.y >= 0 && neighbor.y < GridSize {
			if !visited[neighbor] {
				*frontier = append(*frontier, edge{from: c, to: neighbor, direction: dir})
			}
		}
	}
}

// getPossibleEdges returns all edges that could be added but don't exist
func (dg *DungeonGenerator) getPossibleEdges(roomGrid map[coord]*game.Room, edges map[coord]map[string]bool) []edge {
	possible := make([]edge, 0)
	directions := []string{"north", "east"} // Only check two directions to avoid duplicates

	for c := range roomGrid {
		for _, dir := range directions {
			neighbor := getNeighbor(c, dir)
			if roomGrid[neighbor] != nil && !edges[c][dir] {
				possible = append(possible, edge{from: c, to: neighbor, direction: dir})
			}
		}
	}
	return possible
}

// shuffleEdges randomizes edge order
func (dg *DungeonGenerator) shuffleEdges(edges []edge) {
	for i := len(edges) - 1; i > 0; i-- {
		j := dg.random.Intn(i + 1)
		edges[i], edges[j] = edges[j], edges[i]
	}
}

// getNeighbor returns the coord in the given direction
func getNeighbor(c coord, dir string) coord {
	switch dir {
	case "north":
		return coord{c.x, c.y + 1}
	case "south":
		return coord{c.x, c.y - 1}
	case "east":
		return coord{c.x + 1, c.y}
	case "west":
		return coord{c.x - 1, c.y}
	}
	return c
}

// oppositeDir returns the opposite direction
func oppositeDir(dir string) string {
	switch dir {
	case "north":
		return "south"
	case "south":
		return "north"
	case "east":
		return "west"
	case "west":
		return "east"
	}
	return ""
}

// generateRoomName creates a random room name
func (dg *DungeonGenerator) generateRoomName() string {
	adjectives := []string{"Dark", "Dusty", "Ancient", "Forgotten", "Cursed", "Silent", "Echoing", "Gloomy", "Damp", "Musty"}
	nouns := []string{"Chamber", "Hall", "Corridor", "Vault", "Crypt", "Passage", "Alcove", "Sanctum", "Den", "Lair"}

	adj := adjectives[dg.random.Intn(len(adjectives))]
	noun := nouns[dg.random.Intn(len(nouns))]

	return fmt.Sprintf("%s %s", adj, noun)
}

// generateRoomDescription creates a description based on position
func (dg *DungeonGenerator) generateRoomDescription(x, y int) string {
	distance := x + y // Manhattan distance from entrance

	if x == 0 && y == 0 {
		return "The entrance to the dungeon. Faint light filters in from behind you."
	}
	if x == GridSize-1 && y == GridSize-1 {
		return "A grand chamber with an ornate door leading to freedom!"
	}

	descriptions := []string{
		"Cold stone walls surround you. Water drips somewhere in the darkness.",
		"Cobwebs hang from the ceiling. The air smells of decay.",
		"Torches flicker weakly on the walls, casting dancing shadows.",
		"Bones are scattered across the floor. Something died here.",
		"Strange runes are carved into the walls, pulsing with faint light.",
		"The ceiling is low here, forcing you to crouch slightly.",
		"Claw marks score the stone walls. Something large passed through.",
		"A cold draft blows through, carrying whispers from deeper within.",
	}

	// Use distance to weight toward scarier descriptions
	idx := dg.random.Intn(len(descriptions))
	if distance > 4 && dg.random.Float32() < 0.5 {
		idx = dg.random.Intn(3) + 5 // Prefer scarier ones
	}

	return descriptions[idx]
}

// MonsterTemplate defines a monster type
type MonsterTemplate struct {
	Name        string
	Description string
	BaseHP      int
	BaseDamage  int
	MinDiff     int // minimum difficulty to spawn
}

// ItemTemplate defines an item type
type ItemTemplate struct {
	Name        string
	Description string
	Type        string
	Damage      int
	Armor       int
	Healing     int
}

var monsterTemplates = []MonsterTemplate{
	{Name: "Rat", Description: "A large, mangy rat with beady red eyes.", BaseHP: 5, BaseDamage: 2, MinDiff: 0},
	{Name: "Goblin", Description: "A small, green-skinned creature with a wicked grin.", BaseHP: 10, BaseDamage: 4, MinDiff: 1},
	{Name: "Skeleton", Description: "The animated bones of a long-dead warrior.", BaseHP: 15, BaseDamage: 5, MinDiff: 2},
	{Name: "Orc", Description: "A hulking brute with tusks and a massive club.", BaseHP: 25, BaseDamage: 8, MinDiff: 3},
	{Name: "Wraith", Description: "A shadowy figure that chills you to the bone.", BaseHP: 20, BaseDamage: 7, MinDiff: 4},
}

var itemTemplates = []ItemTemplate{
	{Name: "Health Potion", Description: "A red vial that restores health.", Type: "consumable", Healing: 10},
	{Name: "Greater Health Potion", Description: "A large red vial that restores significant health.", Type: "consumable", Healing: 20},
	{Name: "Rusty Sword", Description: "An old sword, still sharp enough to cut.", Type: "weapon", Damage: 3},
	{Name: "Short Sword", Description: "A well-balanced blade.", Type: "weapon", Damage: 5},
	{Name: "Wooden Shield", Description: "A simple wooden shield that provides basic protection.", Type: "armor", Armor: 2},
	{Name: "Iron Shield", Description: "A sturdy iron shield.", Type: "armor", Armor: 4},
}

// PopulateRoom adds monsters, items, and traps to a room
func (dg *DungeonGenerator) PopulateRoom(room *game.Room, difficulty int) ([]*game.Monster, []*game.Item, []*game.Trap) {
	monsters := make([]*game.Monster, 0)
	items := make([]*game.Item, 0)
	traps := make([]*game.Trap, 0)

	// Entrance room is safe - no monsters
	if room.IsEntrance {
		// Give the player a starting health potion
		startPotion := &game.Item{
			ID:          generateID(),
			Name:        "Health Potion",
			Description: "A red vial that restores health.",
			Type:        "consumable",
			Healing:     10,
			RoomID:      &room.ID,
		}
		items = append(items, startPotion)
		return monsters, items, traps
	}

	// Exit room has no monsters (victory condition)
	if room.IsExit {
		return monsters, items, traps
	}

	// Spawn monsters based on difficulty (Manhattan distance from entrance)
	eligibleMonsters := make([]MonsterTemplate, 0)
	for _, mt := range monsterTemplates {
		if mt.MinDiff <= difficulty {
			eligibleMonsters = append(eligibleMonsters, mt)
		}
	}

	// 70% chance of monsters in non-entrance/exit rooms
	if len(eligibleMonsters) > 0 && dg.random.Float32() < 0.7 {
		// Spawn 1-2 monsters
		numMonsters := 1
		if difficulty >= 3 && dg.random.Float32() < 0.4 {
			numMonsters = 2
		}

		for i := 0; i < numMonsters; i++ {
			// Pick a random eligible monster
			idx := dg.random.Intn(len(eligibleMonsters))
			template := eligibleMonsters[idx]

			// Scale HP and damage with difficulty
			scaleFactor := 1.0 + float64(difficulty)*0.15
			monster := &game.Monster{
				ID:          generateID(),
				Name:        template.Name,
				Description: template.Description,
				HP:          int(float64(template.BaseHP) * scaleFactor),
				MaxHP:       int(float64(template.BaseHP) * scaleFactor),
				Damage:      int(float64(template.BaseDamage) * scaleFactor),
				RoomID:      room.ID,
				IsAlive:     true,
			}
			monsters = append(monsters, monster)
		}
	}

	// Chance to spawn an item (25% base + 5% per difficulty)
	itemChance := 0.25 + float64(difficulty)*0.05
	if itemChance > 0.5 {
		itemChance = 0.5
	}

	if dg.random.Float64() < itemChance {
		// Pick a random item
		template := itemTemplates[dg.random.Intn(len(itemTemplates))]
		item := &game.Item{
			ID:          generateID(),
			Name:        template.Name,
			Description: template.Description,
			Type:        template.Type,
			Damage:      template.Damage,
			Armor:       template.Armor,
			Healing:     template.Healing,
			RoomID:      &room.ID,
		}
		items = append(items, item)
	}

	return monsters, items, traps
}

// GetRoomDifficulty calculates difficulty based on Manhattan distance from entrance
func GetRoomDifficulty(room *game.Room) int {
	return room.X + room.Y
}
