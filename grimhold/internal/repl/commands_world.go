package repl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dancodeur/grimhold/internal/ansi"
)

func commandMap(s *Session, args []string) error {
	if err := checkNoArgs(args, "map"); err != nil {
		return err
	}

	current := s.game.CurrentRegion()
	fmt.Fprintln(s.out, "Known regions:")
	for _, region := range s.game.World.Regions {
		marker := " "
		if region.ID == current.ID {
			marker = ">"
		}
		fmt.Fprintf(s.out, "%s %s [%s] - danger: %s\n", marker, region.Name, region.ID, ansi.DangerLabel(region.Danger))
		fmt.Fprintf(s.out, "    exits: %s\n", strings.Join(region.Exits, ", "))
	}
	return nil
}

func commandTravel(s *Session, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: travel <region>")
	}

	// "travel ashen marsh" est compris comme "ashen-marsh".
	regionID := strings.Join(args, "-")
	message, err := s.game.Travel(regionID)
	if err != nil {
		return err
	}
	fmt.Fprintln(s.out, message)

	s.scoutCurrentRegion()
	return nil
}

func commandLook(s *Session, args []string) error {
	if err := checkNoArgs(args, "look"); err != nil {
		return err
	}

	region := s.game.CurrentRegion()
	fmt.Fprintln(s.out, ansi.Bold(region.Name))
	fmt.Fprintln(s.out, region.Description)
	fmt.Fprintf(s.out, "Exits: %s\n", strings.Join(region.Exits, ", "))
	fmt.Fprintf(s.out, "Danger: %s\n", ansi.DangerLabel(region.Danger))
	return nil
}

// scoutCurrentRegion télécharge à l'avance les ennemis de la région,
// pour que explore réponde tout de suite.
func (s *Session) scoutCurrentRegion() {
	foes := s.game.CurrentRegion().Foes
	if len(foes) == 0 {
		return
	}
	if err := s.api.Prefetch(foes); err != nil {
		// Souvent la même panne pour chaque ennemi : on affiche seulement la première.
		firstError, _, _ := strings.Cut(err.Error(), "\n")
		fmt.Fprintln(s.out, ansi.Dim("(could not scout ahead: "+firstError+")"))
	}
}
