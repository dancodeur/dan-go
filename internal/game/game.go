// Package game contient les règles : déplacements, rencontres, combat et sauvegarde.
package game

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/dancodeur/grimhold/internal/open5e"
	"github.com/dancodeur/grimhold/internal/world"
)

var (
	ErrInFight    = errors.New("utilise fight ou flee")
	ErrNotInFight = errors.New("tu n'es pas en combat, essaie explore")
	ErrNoSave     = errors.New("aucune sauvegarde trouvée")
)

// MonsterSource donne les fiches de monstres (le client Open5e, ou un faux dans les tests).
type MonsterSource interface {
	GetMonster(name string) (open5e.Monster, error)
}

// State est tout ce qui est écrit dans save.json.
type State struct {
	Location  string         `json:"location"`
	Character Character      `json:"character"`
	Gold      int            `json:"gold"`
	Slain     map[string]int `json:"slain"` // nom du monstre -> nombre de victoires
	Rested    bool           `json:"rested"`
}

// Foe est l'ennemi du combat en cours.
type Foe struct {
	Monster open5e.Monster
	HP      int
	MaxHP   int
}

// Game regroupe le monde, l'état du joueur et le combat en cours (Foe est nil hors combat).
type Game struct {
	World    *world.World
	State    State
	Foe      *Foe
	monsters MonsterSource
	rng      *rand.Rand
}

// New démarre une nouvelle partie au point de départ du monde.
func New(gameWorld *world.World, monsters MonsterSource, rng *rand.Rand) *Game {
	return &Game{
		World: gameWorld,
		State: State{
			Location:  gameWorld.Start,
			Character: NewCharacter(),
			Slain:     map[string]int{},
		},
		monsters: monsters,
		rng:      rng,
	}
}

// CurrentRegion renvoie la région où se trouve le joueur.
func (g *Game) CurrentRegion() world.Region {
	region, _ := g.World.Region(g.State.Location)
	return region
}

// Travel déplace le joueur vers une région voisine.
func (g *Game) Travel(regionID string) (string, error) {
	if err := g.blockedByFight("travel"); err != nil {
		return "", err
	}

	current := g.CurrentRegion()
	destination, found := g.World.Region(regionID)
	if !found || !current.HasExit(regionID) {
		return "", fmt.Errorf("tu ne peux pas aller vers %q depuis %s (sorties : %s)",
			regionID, current.Name, strings.Join(current.Exits, ", "))
	}

	g.State.Location = regionID
	g.State.Rested = false
	return fmt.Sprintf("Tu voyages vers %s. %s", destination.Name, destination.Description), nil
}

// Explore cherche une rencontre dans la région et démarre un combat.
func (g *Game) Explore() (string, error) {
	if err := g.blockedByFight("explore"); err != nil {
		return "", err
	}

	region := g.CurrentRegion()
	if len(region.Foes) == 0 {
		return "Tu erres un moment, mais rien ne bouge.", nil
	}

	monster, err := g.rollEncounter(region)
	if err != nil {
		return "", fmt.Errorf("exploration de %s : %w", region.Name, err)
	}

	hitPoints := max(monster.HitPoints, 1)
	g.Foe = &Foe{Monster: monster, HP: hitPoints, MaxHP: hitPoints}

	return fmt.Sprintf("%s surgit de l'ombre ! (FP %s, CA %d, PV %d)\nTape fight pour attaquer ou flee pour fuir.",
		monster.Name, monster.ChallengeRating, monster.ArmorClass, hitPoints), nil
}

// Rest soigne complètement le joueur, une seule fois par visite d'une région.
func (g *Game) Rest() (string, error) {
	if err := g.blockedByFight("rest"); err != nil {
		return "", err
	}
	if g.State.Rested {
		return "", errors.New("tu t'es déjà reposé ici, va d'abord ailleurs")
	}

	g.State.Character.HP = g.State.Character.MaxHP
	g.State.Rested = true
	return "Tu te reposes jusqu'à l'aube et te sens complètement guéri.", nil
}

// rollEncounter tire un ennemi au hasard dans la région. Plus la région est
// dangereuse, plus on tire de fois, et on garde le plus fort.
func (g *Game) rollEncounter(region world.Region) (open5e.Monster, error) {
	rolls := 1 + region.Danger/2

	var strongest open5e.Monster
	for roll := 0; roll < rolls; roll++ {
		name := region.Foes[g.rng.Intn(len(region.Foes))]
		monster, err := g.monsters.GetMonster(name)
		if err != nil {
			return open5e.Monster{}, err
		}
		if roll == 0 || monster.ChallengeValue() > strongest.ChallengeValue() {
			strongest = monster
		}
	}
	return strongest, nil
}

// blockedByFight renvoie une erreur si un combat est en cours.
func (g *Game) blockedByFight(action string) error {
	if g.Foe == nil {
		return nil
	}
	return fmt.Errorf("la commande %s est impossible pendant le combat contre %s : %w", action, g.Foe.Monster.Name, ErrInFight)
}
