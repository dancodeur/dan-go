package repl

import (
	"fmt"
	"sort"
	"strings"
)

// newCommands enregistre toutes les commandes. La map est construite ici
// (et non dans une variable globale) car help doit pouvoir la lire.
func newCommands() map[string]cliCommand {
	list := []cliCommand{
		{name: "help", description: "Affiche cette aide", callback: commandHelp},
		{name: "map", description: "Liste les régions connues", callback: commandMap},
		{name: "travel", arguments: "<region>", description: "Voyage vers une région voisine", callback: commandTravel},
		{name: "look", description: "Décrit les environs", callback: commandLook},
		{name: "explore", description: "Explore la région (risque de rencontre)", callback: commandExplore},
		{name: "fight", description: "Combat l'ennemi pendant un tour", callback: commandFight},
		{name: "flee", description: "Tente de fuir l'ennemi", callback: commandFlee},
		{name: "rest", description: "Soigne complètement (une fois par visite)", callback: commandRest},
		{name: "character", description: "Affiche la fiche du personnage", callback: commandCharacter},
		{name: "slain", description: "Liste les monstres vaincus", callback: commandSlain},
		{name: "bestiary", description: "Liste les monstres de l'API Open5e", callback: commandBestiary},
		{name: "next", description: "Page suivante du bestiaire", callback: commandNextPage},
		{name: "prev", description: "Page précédente du bestiaire", callback: commandPreviousPage},
		{name: "inspect", arguments: "<monster>", description: "Affiche la fiche d'un monstre", callback: commandInspect},
		{name: "save", description: "Sauvegarde la partie", callback: commandSave},
		{name: "load", description: "Charge la partie sauvegardée", callback: commandLoad},
		{name: "quit", description: "Quitte le jeu", callback: commandQuit},
	}

	commands := make(map[string]cliCommand, len(list))
	for _, command := range list {
		commands[command.name] = command
	}
	return commands
}

// checkNoArgs refuse les arguments en trop pour les commandes qui n'en prennent pas.
func checkNoArgs(args []string, commandName string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: %s (this command takes no arguments)", commandName)
	}
	return nil
}

func commandHelp(s *Session, args []string) error {
	if err := checkNoArgs(args, "help"); err != nil {
		return err
	}

	names := make([]string, 0, len(s.commands))
	for name := range s.commands {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Fprintln(s.out, "Grimhold — un jeu d'exploration de donjon")
	fmt.Fprintln(s.out)
	fmt.Fprintln(s.out, "Commandes :")
	for _, name := range names {
		command := s.commands[name]
		usage := strings.TrimSpace(command.name + " " + command.arguments)
		fmt.Fprintf(s.out, "  %-20s %s\n", usage, command.description)
	}
	return nil
}

func commandSave(s *Session, args []string) error {
	if err := checkNoArgs(args, "save"); err != nil {
		return err
	}
	if err := s.game.Save(s.savePath); err != nil {
		return err
	}
	fmt.Fprintf(s.out, "Game saved to %s.\n", s.savePath)
	return nil
}

func commandLoad(s *Session, args []string) error {
	if err := checkNoArgs(args, "load"); err != nil {
		return err
	}
	if err := s.game.Load(s.savePath); err != nil {
		return err
	}
	fmt.Fprintf(s.out, "Game loaded. You are at %s.\n", s.game.CurrentRegion().Name)
	return nil
}

func commandQuit(s *Session, args []string) error {
	if err := checkNoArgs(args, "quit"); err != nil {
		return err
	}
	fmt.Fprintln(s.out, "Farewell, wanderer.")
	s.finished = true
	return nil
}
