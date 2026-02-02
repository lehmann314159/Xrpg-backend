package game

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Combat constants
const (
	D20               = 20 // Twenty-sided die for attack rolls
	D6                = 6  // Six-sided die for damage rolls
	BaseDefense       = 10 // Base armor class / defense value
	CriticalThreshold = 5  // Minimum d6 roll for critical hit (5 or 6)
	MinDamage         = 1  // Minimum damage on a hit
)

// combatRandom is the seeded random source for combat
var (
	combatRandom *rand.Rand
	combatMu     sync.Mutex
)

func init() {
	// Initialize with a default seed based on current time
	combatRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// SetCombatSeed sets the random seed for combat, making it reproducible
func SetCombatSeed(seed int64) {
	combatMu.Lock()
	defer combatMu.Unlock()
	combatRandom = rand.New(rand.NewSource(seed))
}

// Combat handles turn-based combat mechanics

// RollDice simulates a dice roll (e.g., d20) using the seeded random source
func RollDice(sides int) int {
	combatMu.Lock()
	defer combatMu.Unlock()
	return combatRandom.Intn(sides) + 1
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
	attackRoll := RollDice(D20) + (attackerStrength / 2)

	// Defense: base + armor bonus
	defense := BaseDefense + defenderArmor

	result := &CombatResult{}

	if attackRoll >= defense {
		// Hit! Roll damage
		damage := RollDamage(1, D6, attackerStrength/2)
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
// weaponBonus is extra damage from equipped weapon
// armorBonus is extra defense from equipped armor
// Returns updated combat state, enhanced result for frontend, and whether combat continues
func ExecuteCombatTurn(player *Character, monster *Monster, playerAction string, weaponBonus int, armorBonus int) (*CombatResult, *EnhancedCombatResult, bool) {
	result := &CombatResult{
		AttackerHP: player.HP,
		DefenderHP: monster.HP,
	}

	enhanced := &EnhancedCombatResult{}

	// Calculate player's damage bonus (base strength + equipped weapon)
	playerDamageBonus := (player.Strength / 2) + weaponBonus

	// Player attacks monster
	attackRoll := RollDice(D20) + (player.Dexterity / 2)
	monsterDefense := BaseDefense

	playerAttack := &AttackResult{
		AttackerName: player.Name,
		TargetName:   monster.Name,
	}

	if attackRoll >= monsterDefense {
		// Hit! Roll damage - track the d6 roll for critical detection
		damageRoll := RollDice(D6)
		damage := damageRoll + playerDamageBonus
		if damage < MinDamage {
			damage = MinDamage
		}

		playerAttack.WasHit = true
		playerAttack.Damage = damage
		playerAttack.WasCritical = damageRoll >= CriticalThreshold

		monster.HP -= damage
		result.DefenderDamage = damage

		if monster.HP <= 0 {
			monster.HP = 0
			monster.IsAlive = false
			result.DefenderHP = 0
			result.DefenderDied = true
			playerAttack.RemainingHP = 0
			enhanced.PlayerAttack = playerAttack
			enhanced.EnemyDefeated = true

			if playerAttack.WasCritical {
				result.Message = fmt.Sprintf("CRITICAL HIT! You strike the %s for %d damage! The %s collapses!",
					monster.Name, damage, monster.Name)
			} else {
				result.Message = fmt.Sprintf("You strike the %s for %d damage! The %s collapses!",
					monster.Name, damage, monster.Name)
			}
			return result, enhanced, false // Combat ends
		}

		result.DefenderHP = monster.HP
		playerAttack.RemainingHP = monster.HP
		if playerAttack.WasCritical {
			result.Message = fmt.Sprintf("CRITICAL HIT! You strike the %s for %d damage! (%d/%d HP)",
				monster.Name, damage, monster.HP, monster.MaxHP)
		} else {
			result.Message = fmt.Sprintf("You strike the %s for %d damage! (%d/%d HP)",
				monster.Name, damage, monster.HP, monster.MaxHP)
		}
	} else {
		playerAttack.WasHit = false
		playerAttack.Damage = 0
		playerAttack.RemainingHP = monster.HP
		result.Message = fmt.Sprintf("You swing at the %s but miss!", monster.Name)
	}
	enhanced.PlayerAttack = playerAttack

	// Monster counter-attacks
	monsterAttackRoll := RollDice(D20)
	// Player defense includes dexterity and equipped armor
	playerDefense := BaseDefense + (player.Dexterity / 2) + armorBonus

	enemyAttack := &AttackResult{
		AttackerName: monster.Name,
		TargetName:   player.Name,
	}

	if monsterAttackRoll >= playerDefense {
		// Monster hits - roll d6 for damage variance and critical detection
		damageRoll := RollDice(D6)
		monsterDamage := monster.Damage + (damageRoll - 3) // -2 to +3 variance
		if monsterDamage < MinDamage {
			monsterDamage = MinDamage
		}

		enemyAttack.WasHit = true
		enemyAttack.Damage = monsterDamage
		enemyAttack.WasCritical = damageRoll >= CriticalThreshold

		player.TakeDamage(monsterDamage)
		result.AttackerDamage = monsterDamage
		result.AttackerHP = player.HP
		enemyAttack.RemainingHP = player.HP

		if !player.IsAlive {
			result.AttackerDied = true
			enhanced.EnemyAttack = enemyAttack
			enhanced.PlayerDied = true

			if enemyAttack.WasCritical {
				result.Message += fmt.Sprintf(" CRITICAL HIT! The %s strikes back for %d damage! You have fallen...",
					monster.Name, monsterDamage)
			} else {
				result.Message += fmt.Sprintf(" The %s strikes back for %d damage! You have fallen...",
					monster.Name, monsterDamage)
			}
			return result, enhanced, false // Combat ends
		}

		if enemyAttack.WasCritical {
			result.Message += fmt.Sprintf(" CRITICAL HIT! The %s strikes back for %d damage! (HP: %d/%d)",
				monster.Name, monsterDamage, player.HP, player.MaxHP)
		} else {
			result.Message += fmt.Sprintf(" The %s strikes back for %d damage! (HP: %d/%d)",
				monster.Name, monsterDamage, player.HP, player.MaxHP)
		}
	} else {
		enemyAttack.WasHit = false
		enemyAttack.Damage = 0
		enemyAttack.RemainingHP = player.HP
		result.Message += fmt.Sprintf(" The %s tries to attack but misses!", monster.Name)
	}
	enhanced.EnemyAttack = enemyAttack

	return result, enhanced, true // Combat continues
}
