package game

const (
	startingHP          = 30
	startingAttackBonus = 4
	startingArmorClass  = 13
	hpPerLevel          = 8
	xpPerLevelFactor    = 50
)

// Weapon est l'arme du joueur et ses dégâts.
type Weapon struct {
	Name   string `json:"name"`
	Damage Dice   `json:"damage"`
}

// Character est la fiche du joueur.
type Character struct {
	HP          int    `json:"hp"`
	MaxHP       int    `json:"max_hp"`
	Level       int    `json:"level"`
	XP          int    `json:"xp"`
	AttackBonus int    `json:"attack_bonus"`
	ArmorClass  int    `json:"armor_class"`
	Weapon      Weapon `json:"weapon"`
}

// NewCharacter crée un personnage de niveau 1.
func NewCharacter() Character {
	return Character{
		HP:          startingHP,
		MaxHP:       startingHP,
		Level:       1,
		AttackBonus: startingAttackBonus,
		ArmorClass:  startingArmorClass,
		Weapon:      Weapon{Name: "Rusty Sword", Damage: Dice{Count: 1, Sides: 8, Bonus: 1}},
	}
}

// XPForNextLevel donne l'XP total à atteindre pour passer au niveau suivant
// (niveau² × 50 : 50 pour le niveau 2, 200 pour le niveau 3, etc.).
func (c Character) XPForNextLevel() int {
	return c.Level * c.Level * xpPerLevelFactor
}

// GainXP ajoute de l'XP et renvoie le nombre de niveaux gagnés.
// Chaque niveau augmente la vie max et le bonus d'attaque.
func (c *Character) GainXP(amount int) int {
	c.XP += amount

	levelsGained := 0
	for c.XP >= c.XPForNextLevel() {
		c.Level++
		c.MaxHP += hpPerLevel
		c.HP += hpPerLevel
		c.AttackBonus++
		levelsGained++
	}
	return levelsGained
}
