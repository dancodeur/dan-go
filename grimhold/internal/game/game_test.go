package game

import (
	"errors"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/dancodeur/grimhold/internal/open5e"
	"github.com/dancodeur/grimhold/internal/world"
)

// fakeSource sert des monstres connus d'avance, sans réseau.
type fakeSource map[string]open5e.Monster

func (f fakeSource) GetMonster(name string) (open5e.Monster, error) {
	monster, found := f[name]
	if !found {
		return open5e.Monster{}, open5e.ErrNotFound
	}
	return monster, nil
}

// Une cible que le joueur touche toujours et qui meurt en un coup.
func weakMonster() open5e.Monster {
	return open5e.Monster{
		Name: "Dummy", ArmorClass: 1, HitPoints: 1, ChallengeRating: "1/4",
		Actions: open5e.ActionList{{Name: "Poke", AttackBonus: 1, DamageDice: "1d4"}},
	}
}

// Un monstre que le joueur ne touche jamais et qui touche toujours très fort.
func deadlyMonster() open5e.Monster {
	return open5e.Monster{
		Name: "Doom", ArmorClass: 100, HitPoints: 100, ChallengeRating: "10",
		Actions: open5e.ActionList{{Name: "Crush", AttackBonus: 50, DamageDice: "10d10"}},
	}
}

func newTestGame(monster open5e.Monster) *Game {
	testWorld := &world.World{
		Start: "camp",
		Regions: []world.Region{
			{ID: "camp", Name: "Camp", Exits: []string{"woods"}},
			{ID: "woods", Name: "Woods", Danger: 1, Foes: []string{"dummy"}, Exits: []string{"camp"}},
			{ID: "cave", Name: "Cave"},
		},
	}
	source := fakeSource{"dummy": monster}
	return New(testWorld, source, rand.New(rand.NewSource(42)))
}

func TestTravelOnlyAlongExits(t *testing.T) {
	dungeon := newTestGame(weakMonster())

	if _, err := dungeon.Travel("cave"); err == nil {
		t.Error("travel to a region without a connecting exit should fail")
	}
	if _, err := dungeon.Travel("nowhere"); err == nil {
		t.Error("travel to an unknown region should fail")
	}
	if _, err := dungeon.Travel("woods"); err != nil {
		t.Fatalf("travel to a connected region failed: %v", err)
	}
	if dungeon.State.Location != "woods" {
		t.Errorf("location = %q, want woods", dungeon.State.Location)
	}
}

func TestExploreStartsAFightThatBlocksOtherActions(t *testing.T) {
	dungeon := newTestGame(weakMonster())
	if _, err := dungeon.Travel("woods"); err != nil {
		t.Fatal(err)
	}

	if _, err := dungeon.Explore(); err != nil {
		t.Fatalf("explore failed: %v", err)
	}
	if dungeon.Foe == nil {
		t.Fatal("explore should start a fight in a region with foes")
	}

	if _, err := dungeon.Travel("camp"); !errors.Is(err, ErrInFight) {
		t.Errorf("travel during a fight: got %v, want ErrInFight", err)
	}
	if _, err := dungeon.Rest(); !errors.Is(err, ErrInFight) {
		t.Errorf("rest during a fight: got %v, want ErrInFight", err)
	}
}

func TestExploreInAQuietRegion(t *testing.T) {
	dungeon := newTestGame(weakMonster())

	if _, err := dungeon.Explore(); err != nil {
		t.Fatal(err)
	}
	if dungeon.Foe != nil {
		t.Error("a region without foes should not start a fight")
	}
}

func TestWinningAFight(t *testing.T) {
	dungeon := newTestGame(weakMonster())
	dungeon.Travel("woods")
	dungeon.Explore()

	if _, err := dungeon.FightRound(); err != nil {
		t.Fatal(err)
	}

	if dungeon.Foe != nil {
		t.Error("the fight should be over")
	}
	if dungeon.State.Slain["Dummy"] != 1 {
		t.Errorf("slain count = %d, want 1", dungeon.State.Slain["Dummy"])
	}
	if dungeon.State.Character.XP != 25 {
		t.Errorf("XP = %d, want 25 (challenge 1/4)", dungeon.State.Character.XP)
	}
}

func TestLosingAFightSendsYouBackToTheStart(t *testing.T) {
	dungeon := newTestGame(deadlyMonster())
	dungeon.Travel("woods")
	dungeon.Explore()
	dungeon.State.Character.HP = 5

	if _, err := dungeon.FightRound(); err != nil {
		t.Fatal(err)
	}

	if dungeon.Foe != nil {
		t.Error("the fight should be over")
	}
	if dungeon.State.Location != "camp" {
		t.Errorf("location = %q, want camp", dungeon.State.Location)
	}
	if dungeon.State.Character.HP != 1 {
		t.Errorf("HP = %d, want 1", dungeon.State.Character.HP)
	}
}

func TestFightWithoutFoe(t *testing.T) {
	dungeon := newTestGame(weakMonster())

	if _, err := dungeon.FightRound(); !errors.Is(err, ErrNotInFight) {
		t.Errorf("fight: got %v, want ErrNotInFight", err)
	}
	if _, err := dungeon.Flee(); !errors.Is(err, ErrNotInFight) {
		t.Errorf("flee: got %v, want ErrNotInFight", err)
	}
}

func TestRestOncePerVisit(t *testing.T) {
	dungeon := newTestGame(weakMonster())
	dungeon.State.Character.HP = 3

	if _, err := dungeon.Rest(); err != nil {
		t.Fatal(err)
	}
	if dungeon.State.Character.HP != dungeon.State.Character.MaxHP {
		t.Error("rest should heal fully")
	}
	if _, err := dungeon.Rest(); err == nil {
		t.Error("a second rest in the same visit should fail")
	}

	dungeon.Travel("woods")
	if _, err := dungeon.Rest(); err != nil {
		t.Errorf("rest after travelling should work: %v", err)
	}
}

func TestGainXPRaisesLevel(t *testing.T) {
	character := NewCharacter()

	if levels := character.GainXP(50); levels != 1 {
		t.Errorf("levels gained = %d, want 1", levels)
	}
	if character.Level != 2 || character.MaxHP != 38 || character.AttackBonus != 5 {
		t.Errorf("after level up: %+v", character)
	}

	character.GainXP(950) // total 1000 XP : niveau 5
	if character.Level != 5 {
		t.Errorf("level = %d, want 5", character.Level)
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	dungeon := newTestGame(weakMonster())
	dungeon.Travel("woods")
	dungeon.State.Gold = 12
	dungeon.State.Slain["Dummy"] = 3

	if err := dungeon.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded := newTestGame(weakMonster())
	if err := loaded.Load(path); err != nil {
		t.Fatal(err)
	}
	if loaded.State.Location != "woods" || loaded.State.Gold != 12 || loaded.State.Slain["Dummy"] != 3 {
		t.Errorf("loaded state does not match: %+v", loaded.State)
	}
}

func TestLoadWithoutSaveFile(t *testing.T) {
	dungeon := newTestGame(weakMonster())

	err := dungeon.Load(filepath.Join(t.TempDir(), "missing.json"))
	if !errors.Is(err, ErrNoSave) {
		t.Errorf("got %v, want ErrNoSave", err)
	}
}

func TestSaveRefusedDuringFight(t *testing.T) {
	dungeon := newTestGame(weakMonster())
	dungeon.Travel("woods")
	dungeon.Explore()

	err := dungeon.Save(filepath.Join(t.TempDir(), "save.json"))
	if !errors.Is(err, ErrInFight) {
		t.Errorf("got %v, want ErrInFight", err)
	}
}
