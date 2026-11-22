package repl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dancodeur/grimhold/internal/ansi"
	"github.com/dancodeur/grimhold/internal/open5e"
)

func commandBestiary(s *Session, args []string) error {
	if err := checkNoArgs(args, "bestiary"); err != nil {
		return err
	}
	return s.showBestiaryPage("")
}

func commandNextPage(s *Session, args []string) error {
	if err := checkNoArgs(args, "next"); err != nil {
		return err
	}
	if s.nextPage == "" {
		return errors.New("pas de page suivante (lance d'abord bestiary, ou tu es à la dernière page)")
	}
	return s.showBestiaryPage(s.nextPage)
}

func commandPreviousPage(s *Session, args []string) error {
	if err := checkNoArgs(args, "prev"); err != nil {
		return err
	}
	if s.previousPage == "" {
		return errors.New("pas de page précédente (lance d'abord bestiary, ou tu es à la première page)")
	}
	return s.showBestiaryPage(s.previousPage)
}

func commandInspect(s *Session, args []string) error {
	if len(args) == 0 {
		return errors.New("usage : inspect <monster>")
	}

	monster, err := s.api.GetMonster(strings.Join(args, " "))
	if errors.Is(err, open5e.ErrNotFound) {
		return fmt.Errorf("aucun monstre nommé %q dans le bestiaire", strings.Join(args, " "))
	}
	if err != nil {
		return err
	}

	s.printStatBlock(monster)
	return nil
}

// showBestiaryPage affiche une page de monstres et retient les adresses des pages voisines.
func (s *Session) showBestiaryPage(pageURL string) error {
	page, err := s.api.ListMonsters(pageURL)
	if err != nil {
		return err
	}
	s.nextPage = page.Next
	s.previousPage = page.Previous

	fmt.Fprintf(s.out, "%s (%d monstres)\n", ansi.Bold("Bestiaire"), page.Count)
	for _, monster := range page.Results {
		fmt.Fprintf(s.out, "  %-28s FP %-5s PV %-4d CA %d\n",
			monster.Name, monster.ChallengeRating, monster.HitPoints, monster.ArmorClass)
	}

	if page.Next != "" {
		fmt.Fprintln(s.out, "Tape next pour la page suivante.")
	}
	if page.Previous != "" {
		fmt.Fprintln(s.out, "Tape prev pour la page précédente.")
	}
	return nil
}

func (s *Session) printStatBlock(monster open5e.Monster) {
	fmt.Fprintln(s.out, ansi.Bold(monster.Name))
	fmt.Fprintf(s.out, "%s %s, %s\n", monster.Size, monster.Type, monster.Alignment)
	fmt.Fprintf(s.out, "Classe d'armure : %d\n", monster.ArmorClass)
	fmt.Fprintf(s.out, "Points de vie : %d (%s)\n", monster.HitPoints, monster.HitDice)
	fmt.Fprintf(s.out, "Facteur de puissance : %s\n", monster.ChallengeRating)
	fmt.Fprintf(s.out, "FOR %d  DEX %d  CON %d  INT %d  SAG %d  CHA %d\n",
		monster.Strength, monster.Dexterity, monster.Constitution,
		monster.Intelligence, monster.Wisdom, monster.Charisma)

	if len(monster.Actions) == 0 {
		return
	}
	fmt.Fprintln(s.out, "Actions :")
	for _, action := range monster.Actions {
		fmt.Fprintf(s.out, "  %s: %s\n", ansi.Bold(action.Name), action.Description)
	}
}
