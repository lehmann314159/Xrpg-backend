package generator

import (
	"crypto/rand"
	"fmt"
	mrand "math/rand"

	"github.com/yourusername/dungeon-crawler/internal/game"
)

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

// GenerateDungeon creates a new procedural dungeon
func (dg *DungeonGenerator) GenerateDungeon(depth int) (*game.Dungeon, []*game.Room, []*game.RoomConnection, error) {
	dungeon := &game.Dungeon{
		ID:    generateID(),
		Seed:  dg.seed,
		Depth: depth,
	}

	// TODO: Implement actual dungeon generation
	// For now, create a simple linear dungeon
	rooms := dg.generateLinearDungeon(dungeon.ID, 5+depth)
	connections := dg.connectRooms(rooms)

	return dungeon, rooms, connections, nil
}

// generateLinearDungeon creates a simple linear sequence of rooms
func (dg *DungeonGenerator) generateLinearDungeon(dungeonID string, numRooms int) []*game.Room {
	rooms := make([]*game.Room, numRooms)

	for i := 0; i < numRooms; i++ {
		room := &game.Room{
			ID:          generateID(),
			DungeonID:   dungeonID,
			Name:        dg.generateRoomName(),
			Description: "A dark room awaits description", // TODO: Use LLM for flavor
			IsEntrance:  i == 0,
			IsExit:      i == numRooms-1,
			X:           i,
			Y:           0,
		}
		rooms[i] = room
	}

	return rooms
}

// connectRooms creates connections between adjacent rooms
func (dg *DungeonGenerator) connectRooms(rooms []*game.Room) []*game.RoomConnection {
	connections := make([]*game.RoomConnection, 0)

	for i := 0; i < len(rooms)-1; i++ {
		// Forward connection
		connections = append(connections, &game.RoomConnection{
			ID:              generateID(),
			RoomID:          rooms[i].ID,
			Direction:       "east",
			ConnectedRoomID: rooms[i+1].ID,
		})

		// Backward connection
		connections = append(connections, &game.RoomConnection{
			ID:              generateID(),
			RoomID:          rooms[i+1].ID,
			Direction:       "west",
			ConnectedRoomID: rooms[i].ID,
		})
	}

	return connections
}

// generateRoomName creates a random room name
func (dg *DungeonGenerator) generateRoomName() string {
	adjectives := []string{"Dark", "Dusty", "Ancient", "Forgotten", "Cursed", "Silent", "Echoing"}
	nouns := []string{"Chamber", "Hall", "Corridor", "Vault", "Crypt", "Passage", "Alcove"}

	adj := adjectives[dg.random.Intn(len(adjectives))]
	noun := nouns[dg.random.Intn(len(nouns))]

	return fmt.Sprintf("%s %s", adj, noun)
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
	Healing     int
}

var monsterTemplates = []MonsterTemplate{
	{Name: "Rat", Description: "A large, mangy rat with beady red eyes.", BaseHP: 5, BaseDamage: 2, MinDiff: 0},
	{Name: "Goblin", Description: "A small, green-skinned creature with a wicked grin.", BaseHP: 10, BaseDamage: 4, MinDiff: 1},
	{Name: "Skeleton", Description: "The animated bones of a long-dead warrior.", BaseHP: 15, BaseDamage: 5, MinDiff: 2},
	{Name: "Orc", Description: "A hulking brute with tusks and a massive club.", BaseHP: 25, BaseDamage: 8, MinDiff: 3},
}

var itemTemplates = []ItemTemplate{
	{Name: "Health Potion", Description: "A red vial that restores health.", Type: "consumable", Healing: 10},
	{Name: "Greater Health Potion", Description: "A large red vial that restores significant health.", Type: "consumable", Healing: 20},
	{Name: "Rusty Sword", Description: "An old sword, still sharp enough to cut.", Type: "weapon", Damage: 3},
	{Name: "Short Sword", Description: "A well-balanced blade.", Type: "weapon", Damage: 5},
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

	// Spawn monsters based on difficulty
	// Higher difficulty = more/stronger monsters
	eligibleMonsters := make([]MonsterTemplate, 0)
	for _, mt := range monsterTemplates {
		if mt.MinDiff <= difficulty {
			eligibleMonsters = append(eligibleMonsters, mt)
		}
	}

	if len(eligibleMonsters) > 0 {
		// Spawn 1-2 monsters
		numMonsters := 1
		if difficulty >= 2 && dg.random.Float32() < 0.3 {
			numMonsters = 2
		}

		for i := 0; i < numMonsters; i++ {
			// Pick a random eligible monster, weighted toward harder ones at higher difficulty
			idx := dg.random.Intn(len(eligibleMonsters))
			template := eligibleMonsters[idx]

			// Scale HP and damage with difficulty
			scaleFactor := 1.0 + float64(difficulty)*0.1
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

	// Chance to spawn an item (30% base + 10% per difficulty)
	itemChance := 0.3 + float64(difficulty)*0.1
	if itemChance > 0.7 {
		itemChance = 0.7
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
			Healing:     template.Healing,
			RoomID:      &room.ID,
		}
		items = append(items, item)
	}

	return monsters, items, traps
}
