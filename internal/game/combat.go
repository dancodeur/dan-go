package game

import (
	"fmt"
	"strings"

	"github.com/dancodeur/grimhold/internal/open5e"
)

const (
	fleeDifficulty = 12 // il faut d20 + niveau >= 12 pour s'enfuir
	goldChance     = 60 // pourcentage de chances de trouver de l'or
	maxGoldFound   = 10
	xpPerChallenge = 100 // un monstre de FP 1 donne 100 XP
	minimumXP      = 10
)

// attack décrit une attaque, qu'elle vienne du joueur ou d'un monstre.
type attack struct {
	name        string
	attackBonus int
	damage      Dice
}

// FightRound joue un tour : le joueur attaque, puis l'ennemi riposte s'il est encore debout.
func (g *Game) FightRound() (string, error) {
	if g.Foe == nil {
		return "", ErrNotInFight
	}

	character := g.State.Character
	playerAttack := attack{
		name:        character.Weapon.Name,
		attackBonus: character.AttackBonus,
		damage:      character.Weapon.Damage,
	}

	damage, line := g.rollAttack("Tu", playerAttack, g.Foe.Monster.ArmorClass)
	g.Foe.HP -= damage
	lines := []string{line}

	if g.Foe.HP <= 0 {
		lines = append(lines, g.winFight()...)
	} else {
		lines = append(lines, g.foeStrike()...)
	}
	return strings.Join(lines, "\n"), nil
}

// Flee tente de fuir : un jet réussi met fin au combat, un échec donne un coup gratuit à l'ennemi.
func (g *Game) Flee() (string, error) {
	if g.Foe == nil {
		return "", ErrNotInFight
	}

	roll := g.rng.Intn(20) + 1 + g.State.Character.Level
	if roll >= fleeDifficulty {
		name := g.Foe.Monster.Name
		g.Foe = nil
		return fmt.Sprintf("Tu t'échappes de %s. (jet %d, il fallait %d)", name, roll, fleeDifficulty), nil
	}

	lines := []string{fmt.Sprintf("Tu n'arrives pas à fuir ! (jet %d, il fallait %d)", roll, fleeDifficulty)}
	lines = append(lines, g.foeStrike()...)
	return strings.Join(lines, "\n"), nil
}

// rollAttack lance d20 + bonus contre la classe d'armure.
// Elle renvoie les dégâts infligés (0 si raté) et une ligne de texte.
func (g *Game) rollAttack(attackerName string, strike attack, armorClass int) (int, string) {
	roll := g.rng.Intn(20) + 1
	total := roll + strike.attackBonus
	prefix := fmt.Sprintf("%s [%s] : jet %d + %d = %d contre CA %d.",
		attackerName, strike.name, roll, strike.attackBonus, total, armorClass)

	if total < armorClass {
		return 0, prefix + " Raté."
	}

	damage := max(strike.damage.Roll(g.rng), 0)
	return damage, fmt.Sprintf("%s Touché, %d dégâts.", prefix, damage)
}

// foeStrike fait attaquer l'ennemi avec une de ses attaques, tirée au hasard.
func (g *Game) foeStrike() []string {
	attacks := monsterAttacks(g.Foe.Monster)
	chosen := attacks[g.rng.Intn(len(attacks))]

	character := &g.State.Character
	damage, line := g.rollAttack(g.Foe.Monster.Name, chosen, character.ArmorClass)
	character.HP -= damage

	lines := []string{line}
	if character.HP <= 0 {
		lines = append(lines, g.loseFight()...)
	}
	return lines
}

// winFight donne l'XP, l'or éventuel et inscrit le monstre au bestiaire des victoires.
func (g *Game) winFight() []string {
	monster := g.Foe.Monster
	g.Foe = nil

	xp := max(int(monster.ChallengeValue()*xpPerChallenge), minimumXP)
	g.State.Slain[monster.Name]++
	levelsGained := g.State.Character.GainXP(xp)

	lines := []string{fmt.Sprintf("%s s'effondre ! Tu gagnes %d XP.", monster.Name, xp)}

	if g.rng.Intn(100) < goldChance {
		gold := 1 + g.rng.Intn(maxGoldFound)
		g.State.Gold += gold
		lines = append(lines, fmt.Sprintf("Tu trouves %d pièces d'or.", gold))
	}
	if levelsGained > 0 {
		character := g.State.Character
		lines = append(lines, fmt.Sprintf("Tu passes au niveau %d ! Tes PV max sont maintenant de %d.", character.Level, character.MaxHP))
	}
	return lines
}

// loseFight ramène le joueur au point de départ avec 1 PV.
// Le fichier de sauvegarde n'est pas touché.
func (g *Game) loseFight() []string {
	g.Foe = nil
	g.State.Location = g.World.Start
	g.State.Character.HP = 1
	g.State.Rested = false

	return []string{
		"Tout devient noir...",
		fmt.Sprintf("Tu te réveilles à %s avec 1 PV et l'orgueil blessé.", g.CurrentRegion().Name),
	}
}

// monsterAttacks garde les actions du monstre qui sont de vraies attaques.
// Sans aucune attaque exploitable, le monstre donne un coup désespéré.
func monsterAttacks(monster open5e.Monster) []attack {
	var attacks []attack
	for _, action := range monster.Actions {
		if action.DamageDice == "" || action.AttackBonus == 0 {
			continue
		}

		dice, err := ParseDice(action.DamageDice)
		if err != nil {
			continue // données illisibles : on ignore cette action
		}
		// Selon les données, le bonus est dans damage_dice ou dans damage_bonus.
		if dice.Bonus == 0 {
			dice.Bonus = action.DamageBonus
		}
		attacks = append(attacks, attack{name: action.Name, attackBonus: action.AttackBonus, damage: dice})
	}

	if len(attacks) == 0 {
		attacks = append(attacks, attack{name: "Coup désespéré", attackBonus: 2, damage: Dice{Count: 1, Sides: 4}})
	}
	return attacks
}
