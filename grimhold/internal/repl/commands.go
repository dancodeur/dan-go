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
		{name: "help", description: "Show this message", callback: commandHelp},
		{name: "map", description: "List the known regions", callback: commandMap},
		{name: "travel", arguments: "<region>", description: "Travel to a connected region", callback: commandTravel},
		{name: "look", description: "Describe your surroundings", callback: commandLook},
		{name: "explore", description: "Wander the region (risk an encounter)", callback: commandExplore},
		{name: "fight", description: "Fight the foe for one round", callback: commandFight},
		{name: "flee", description: "Try to run away from the foe", callback: commandFlee},
		{name: "rest", description: "Heal fully (once per region visit)", callback: commandRest},
		{name: "character", description: "Show your character sheet", callback: commandCharacter},
		{name: "slain", description: "List the monsters you have slain", callback: commandSlain},
		{name: "bestiary", description: "List the monsters of the Open5e API", callback: commandBestiary},
		{name: "next", description: "Next page of the bestiary", callback: commandNextPage},
		{name: "prev", description: "Previous page of the bestiary", callback: commandPreviousPage},
		{name: "inspect", arguments: "<monster>", description: "Show a monster's stat block", callback: commandInspect},
		{name: "save", description: "Save the game", callback: commandSave},
		{name: "load", description: "Load the saved game", callback: commandLoad},
		{name: "quit", description: "Leave the game", callback: commandQuit},
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

	fmt.Fprintln(s.out, "Grimhold — a dungeon crawler")
	fmt.Fprintln(s.out)
	fmt.Fprintln(s.out, "Commands:")
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
