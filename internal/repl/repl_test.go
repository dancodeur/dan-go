package repl

import (
	"bytes"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dancodeur/grimhold/internal/ansi"
	"github.com/dancodeur/grimhold/internal/game"
	"github.com/dancodeur/grimhold/internal/open5e"
	"github.com/dancodeur/grimhold/internal/world"
)

func TestMain(m *testing.M) {
	ansi.Enabled = false // pas de codes de couleur dans les textes comparés
	os.Exit(m.Run())
}

// fakeAPI répond toujours avec le même monstre faible, sans réseau.
type fakeAPI struct{}

var weakMonster = open5e.Monster{
	Name: "Dummy", Size: "Small", Type: "construct", ArmorClass: 1, HitPoints: 1, ChallengeRating: "1/4",
	Actions: open5e.ActionList{{Name: "Poke", Description: "A weak poke.", AttackBonus: 1, DamageDice: "1d4"}},
}

func (fakeAPI) GetMonster(name string) (open5e.Monster, error) {
	if name != "dummy" {
		return open5e.Monster{}, open5e.ErrNotFound
	}
	return weakMonster, nil
}

func (fakeAPI) ListMonsters(pageURL string) (open5e.MonsterPage, error) {
	return open5e.MonsterPage{Count: 1, Results: []open5e.Monster{weakMonster}}, nil
}

func (fakeAPI) Prefetch(names []string) error { return nil }

// play joue un script de commandes et renvoie tout ce qui a été affiché.
func play(t *testing.T, script string) string {
	t.Helper()

	testWorld := &world.World{
		Start: "camp",
		Regions: []world.Region{
			{ID: "camp", Name: "Camp", Description: "A quiet camp.", Exits: []string{"deep-woods"}},
			{ID: "deep-woods", Name: "Deep Woods", Description: "Dark.", Danger: 1, Foes: []string{"dummy"}, Exits: []string{"camp"}},
		},
	}
	dungeon := game.New(testWorld, fakeAPI{}, rand.New(rand.NewSource(1)))

	var output bytes.Buffer
	session := New(dungeon, fakeAPI{}, filepath.Join(t.TempDir(), "save.json"), &output)
	if err := session.Run(strings.NewReader(script)); err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func TestHelpListsEveryCommand(t *testing.T) {
	output := play(t, "help\n")

	for name := range newCommands() {
		if !strings.Contains(output, name) {
			t.Errorf("help does not mention %q", name)
		}
	}
}

func TestBadInputNeverCrashes(t *testing.T) {
	script := strings.Join([]string{
		"",
		"   ",
		"dance",
		"HELP me",
		"travel",
		"travel the moon",
		"look around",
		"fight",
		"flee",
		"next",
		"prev",
		"inspect",
		"inspect nothing",
		"load",
		"quit",
	}, "\n")

	output := play(t, script)

	for _, expected := range []string{
		`Commande inconnue "dance"`,
		"usage : help",
		"usage : travel <region>",
		"tu ne peux pas aller",
		"usage : look",
		"pas en combat",
		"pas de page suivante",
		"usage : inspect <monster>",
		`aucun monstre nommé "nothing"`,
		"aucune sauvegarde trouvée",
		"Adieu",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("output does not contain %q", expected)
		}
	}
}

func TestEndOfInputEndsTheSession(t *testing.T) {
	// Pas de "quit" : la fin de l'entrée (Ctrl+D) doit arrêter proprement la boucle.
	output := play(t, "look\n")

	if !strings.Contains(output, "Camp") {
		t.Errorf("look did not run: %q", output)
	}
}

func TestTravelAcceptsSpacesInRegionNames(t *testing.T) {
	output := play(t, "travel deep woods\nlook\n")

	if !strings.Contains(output, "Tu voyages vers Deep Woods") {
		t.Errorf("travel did not work: %q", output)
	}
}

func TestExploreAndFight(t *testing.T) {
	output := play(t, "travel deep-woods\nexplore\nfight\nslain\ncharacter\n")

	for _, expected := range []string{"Dummy surgit", "Dummy s'effondre", "Dummy", "x1", "Niveau"} {
		if !strings.Contains(output, expected) {
			t.Errorf("output does not contain %q:\n%s", expected, output)
		}
	}
}

func TestSaveAndLoadCommands(t *testing.T) {
	output := play(t, "travel deep-woods\nsave\ntravel camp\nload\nlook\n")

	if !strings.Contains(output, "Partie sauvegardée") || !strings.Contains(output, "Partie chargée. Tu es à Deep Woods.") {
		t.Errorf("save/load did not work:\n%s", output)
	}
}

func TestBestiaryAndInspect(t *testing.T) {
	output := play(t, "bestiary\ninspect dummy\n")

	if !strings.Contains(output, "Bestiaire (1 monstres)") || !strings.Contains(output, "A weak poke.") {
		t.Errorf("bestiary/inspect did not work:\n%s", output)
	}
}
