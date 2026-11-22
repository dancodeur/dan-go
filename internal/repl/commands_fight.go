package repl

import (
	"fmt"
	"sort"

	"github.com/dancodeur/grimhold/internal/ansi"
)

func commandExplore(s *Session, args []string) error {
	if err := checkNoArgs(args, "explore"); err != nil {
		return err
	}

	message, err := s.game.Explore()
	if err != nil {
		return err
	}
	fmt.Fprintln(s.out, message)
	s.printFoeBar()
	return nil
}

func commandFight(s *Session, args []string) error {
	if err := checkNoArgs(args, "fight"); err != nil {
		return err
	}

	message, err := s.game.FightRound()
	if err != nil {
		return err
	}
	fmt.Fprintln(s.out, message)
	s.printCombatStatus()
	return nil
}

func commandFlee(s *Session, args []string) error {
	if err := checkNoArgs(args, "flee"); err != nil {
		return err
	}

	message, err := s.game.Flee()
	if err != nil {
		return err
	}
	fmt.Fprintln(s.out, message)
	s.printCombatStatus()
	return nil
}

func commandRest(s *Session, args []string) error {
	if err := checkNoArgs(args, "rest"); err != nil {
		return err
	}

	message, err := s.game.Rest()
	if err != nil {
		return err
	}
	fmt.Fprintln(s.out, message)
	return nil
}

func commandCharacter(s *Session, args []string) error {
	if err := checkNoArgs(args, "character"); err != nil {
		return err
	}

	character := s.game.State.Character
	fmt.Fprintln(s.out, ansi.Bold("Fiche du personnage"))
	fmt.Fprintf(s.out, "  Niveau   %d (XP %d/%d)\n", character.Level, character.XP, character.XPForNextLevel())
	fmt.Fprintf(s.out, "  PV       %s\n", ansi.HPBar(character.HP, character.MaxHP))
	fmt.Fprintf(s.out, "  Attaque  +%d\n", character.AttackBonus)
	fmt.Fprintf(s.out, "  Armure   %d\n", character.ArmorClass)
	fmt.Fprintf(s.out, "  Arme     %s (%s)\n", character.Weapon.Name, character.Weapon.Damage)
	fmt.Fprintf(s.out, "  Or       %d\n", s.game.State.Gold)
	return nil
}

func commandSlain(s *Session, args []string) error {
	if err := checkNoArgs(args, "slain"); err != nil {
		return err
	}

	slain := s.game.State.Slain
	if len(slain) == 0 {
		fmt.Fprintln(s.out, "Tu n'as encore vaincu aucun monstre.")
		return nil
	}

	names := make([]string, 0, len(slain))
	for name := range slain {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Fprintln(s.out, ansi.Bold("Monstres vaincus"))
	for _, name := range names {
		fmt.Fprintf(s.out, "  %-20s x%d\n", name, slain[name])
	}
	return nil
}

// printFoeBar affiche la vie de l'ennemi, s'il y en a un.
func (s *Session) printFoeBar() {
	foe := s.game.Foe
	if foe == nil {
		return
	}
	fmt.Fprintf(s.out, "%-14s %s\n", foe.Monster.Name, ansi.HPBar(foe.HP, foe.MaxHP))
}

// printCombatStatus affiche la vie du joueur, puis celle de l'ennemi s'il est encore là.
func (s *Session) printCombatStatus() {
	character := s.game.State.Character
	fmt.Fprintf(s.out, "%-14s %s\n", "Toi", ansi.HPBar(character.HP, character.MaxHP))
	s.printFoeBar()
}
