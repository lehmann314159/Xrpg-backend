package game

import (
	"crypto/rand"
	"fmt"
	"time"
)

// CharacterService handles character operations
type CharacterService struct {
	// TODO: Add database reference
}

// NewCharacter creates a new character
func NewCharacter(name string) *Character {
	return &Character{
		ID:        generateID(),
		Name:      name,
		HP:        20,
		MaxHP:     20,
		Strength:  10,
		Dexterity: 10,
		IsAlive:   true,
		CreatedAt: time.Now(),
	}
}

// generateID creates a simple random ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// TakeDamage applies damage to a character
func (c *Character) TakeDamage(damage int) {
	c.HP -= damage
	if c.HP <= 0 {
		c.HP = 0
		c.IsAlive = false
		now := time.Now()
		c.DiedAt = &now
	}
}

// Heal restores HP to a character
func (c *Character) Heal(amount int) {
	c.HP += amount
	if c.HP > c.MaxHP {
		c.HP = c.MaxHP
	}
}

// CanMove checks if character can move in a direction
func (c *Character) CanMove() error {
	if !c.IsAlive {
		return fmt.Errorf("character is dead")
	}
	return nil
}

// EquipItem equips an item to the character
func (c *Character) EquipItem(item *Item) error {
	if item.Type != "weapon" && item.Type != "armor" {
		return fmt.Errorf("cannot equip item of type: %s", item.Type)
	}
	// TODO: Implement equipment slots and stat bonuses
	return nil
}

// UseConsumable uses a consumable item
func (c *Character) UseConsumable(item *Item) error {
	if item.Type != "consumable" {
		return fmt.Errorf("item is not consumable")
	}
	
	if item.Healing > 0 {
		c.Heal(item.Healing)
	}
	
	// TODO: Handle other consumable effects
	return nil
}
