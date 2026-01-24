package game

import (
	"fmt"
	"math/rand"
)

// Combat handles turn-based combat mechanics

// RollDice simulates a dice roll (e.g., d20)
func RollDice(sides int) int {
	return rand.Intn(sides) + 1
}

// RollDamage calculates damage with dice (e.g., 2d6 + modifier)
func RollDamage(numDice, diceSides, modifier int) int {
	total := modifier
	for i := 0; i < numDice; i++ {
		total += RollDice(diceSides)
	}
	if total < 0 {
		return 0
	}
	return total
}

// CalculateAttack performs a turn-based attack
func CalculateAttack(attackerStrength, defenderArmor int) *CombatResult {
	// Attack roll: d20 + strength modifier
	attackRoll := RollDice(20) + (attackerStrength / 2)
	
	// Defense: 10 + armor bonus
	defense := 10 + defenderArmor
	
	result := &CombatResult{}
	
	if attackRoll >= defense {
		// Hit! Roll damage
		damage := RollDamage(1, 6, attackerStrength/2)
		result.DefenderDamage = damage
		result.Message = "Hit!"
	} else {
		// Miss
		result.DefenderDamage = 0
		result.Message = "Miss!"
	}
	
	return result
}

// ExecuteCombatTurn executes one full turn of combat
// Returns updated combat state and whether combat continues
func ExecuteCombatTurn(player *Character, monster *Monster, playerAction string) (*CombatResult, bool) {
	result := &CombatResult{
		AttackerHP: player.HP,
		DefenderHP: monster.HP,
	}

	// Calculate player's weapon damage bonus
	playerDamageBonus := player.Strength / 2

	// Player attacks monster
	attackRoll := RollDice(20) + (player.Dexterity / 2)
	monsterDefense := 10 // Base monster AC

	if attackRoll >= monsterDefense {
		// Hit! Roll damage
		damage := RollDamage(1, 6, playerDamageBonus)
		if damage < 1 {
			damage = 1 // Minimum 1 damage on hit
		}
		monster.HP -= damage
		result.DefenderDamage = damage

		if monster.HP <= 0 {
			monster.HP = 0
			monster.IsAlive = false
			result.DefenderHP = 0
			result.DefenderDied = true
			result.Message = fmt.Sprintf("You strike the %s for %d damage! The %s collapses!",
				monster.Name, damage, monster.Name)
			return result, false // Combat ends
		}

		result.DefenderHP = monster.HP
		result.Message = fmt.Sprintf("You strike the %s for %d damage! (%d/%d HP)",
			monster.Name, damage, monster.HP, monster.MaxHP)
	} else {
		result.Message = fmt.Sprintf("You swing at the %s but miss!", monster.Name)
	}

	// Monster counter-attacks
	monsterAttackRoll := RollDice(20)
	playerDefense := 10 + (player.Dexterity / 2)

	if monsterAttackRoll >= playerDefense {
		// Monster hits
		monsterDamage := monster.Damage
		// Add some variance
		variance := RollDice(3) - 2 // -1 to +1
		monsterDamage += variance
		if monsterDamage < 1 {
			monsterDamage = 1
		}

		player.TakeDamage(monsterDamage)
		result.AttackerDamage = monsterDamage
		result.AttackerHP = player.HP

		if !player.IsAlive {
			result.AttackerDied = true
			result.Message += fmt.Sprintf(" The %s strikes back for %d damage! You have fallen...",
				monster.Name, monsterDamage)
			return result, false // Combat ends
		}

		result.Message += fmt.Sprintf(" The %s strikes back for %d damage! (HP: %d/%d)",
			monster.Name, monsterDamage, player.HP, player.MaxHP)
	} else {
		result.Message += fmt.Sprintf(" The %s tries to attack but misses!", monster.Name)
	}

	return result, true // Combat continues
}
