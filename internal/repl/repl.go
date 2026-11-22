// Package repl lit les commandes du joueur et les exécute.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/dancodeur/grimhold/internal/ansi"
	"github.com/dancodeur/grimhold/internal/game"
	"github.com/dancodeur/grimhold/internal/open5e"
)

// MonsterAPI est ce dont le REPL a besoin de l'API Open5e (le vrai client, ou un faux dans les tests).
type MonsterAPI interface {
	GetMonster(name string) (open5e.Monster, error)
	ListMonsters(pageURL string) (open5e.MonsterPage, error)
	Prefetch(names []string) error
}

// cliCommand est une commande du jeu. arguments ne sert qu'à l'affichage de help.
type cliCommand struct {
	name        string
	arguments   string
	description string
	callback    func(session *Session, args []string) error
}

// Session regroupe tout ce dont les commandes ont besoin.
type Session struct {
	game     *game.Game
	api      MonsterAPI
	savePath string
	out      io.Writer
	commands map[string]cliCommand

	nextPage     string // adresses de pagination du dernier bestiary affiché
	previousPage string
	finished     bool
}

func New(dungeon *game.Game, api MonsterAPI, savePath string, out io.Writer) *Session {
	return &Session{
		game:     dungeon,
		api:      api,
		savePath: savePath,
		out:      out,
		commands: newCommands(),
	}
}

// Run lit une commande par ligne jusqu'à "quit" ou la fin de l'entrée.
func (s *Session) Run(in io.Reader) error {
	fmt.Fprintln(s.out, ansi.Bold("Welcome to Grimhold.")+" Type help to see the commands.")

	scanner := bufio.NewScanner(in)
	for !s.finished {
		fmt.Fprint(s.out, "grimhold > ")
		if !scanner.Scan() {
			fmt.Fprintln(s.out)
			break
		}
		s.execute(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading input: %w", err)
	}
	return nil
}

// execute lance la commande écrite sur une ligne.
// Les erreurs sont affichées au joueur, elles ne font jamais planter le programme.
func (s *Session) execute(line string) {
	words := parseInput(line)
	if len(words) == 0 {
		return
	}

	command, found := s.commands[words[0]]
	if !found {
		fmt.Fprintf(s.out, "Unknown command %q. Type help to see the commands.\n", words[0])
		return
	}

	if err := command.callback(s, words[1:]); err != nil {
		fmt.Fprintln(s.out, ansi.Red(err.Error()))
	}
}

// parseInput met la ligne en minuscules et la découpe en mots.
func parseInput(line string) []string {
	return strings.Fields(strings.ToLower(line))
}
