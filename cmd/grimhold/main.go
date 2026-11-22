// Grimhold est un jeu d'exploration de donjon en mode texte.
// Auteur : Dan David Elenga (LDF A5).
package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/dancodeur/grimhold/internal/ansi"
	"github.com/dancodeur/grimhold/internal/game"
	"github.com/dancodeur/grimhold/internal/open5e"
	"github.com/dancodeur/grimhold/internal/repl"
	"github.com/dancodeur/grimhold/internal/world"
)

const (
	cacheLifetime   = 5 * time.Minute
	cacheCleanEvery = time.Minute
)

func main() {
	worldPath := flag.String("world", "world.json", "chemin du fichier du monde")
	savePath := flag.String("save", "save.json", "chemin du fichier de sauvegarde")
	apiURL := flag.String("api", open5e.DefaultBaseURL, "adresse de base de l'API Open5e")
	continueGame := flag.Bool("continue", false, "charge la sauvegarde au démarrage")
	noColor := flag.Bool("no-color", false, "désactive les couleurs ANSI")
	flag.Parse()

	ansi.Enabled = !*noColor && outputIsTerminal()

	if err := run(*worldPath, *savePath, *apiURL, *continueGame); err != nil {
		fmt.Fprintln(os.Stderr, "grimhold:", err)
		os.Exit(1)
	}
}

func run(worldPath, savePath, apiURL string, continueGame bool) error {
	gameWorld, err := world.Load(worldPath)
	if err != nil {
		return err
	}

	cache := open5e.NewCache(cacheLifetime, cacheCleanEvery)
	defer cache.Close()
	client := open5e.NewClient(apiURL, cache)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	dungeon := game.New(gameWorld, client, rng)

	if continueGame {
		err = dungeon.Load(savePath)
		switch {
		case errors.Is(err, game.ErrNoSave):
			fmt.Println("Aucune sauvegarde trouvée, nouvelle partie.")
		case err != nil:
			return fmt.Errorf("chargement de la sauvegarde : %w", err)
		}
	}

	session := repl.New(dungeon, client, savePath, os.Stdout)
	return session.Run(os.Stdin)
}

// outputIsTerminal évite d'écrire des codes de couleur dans un fichier ou un pipe.
func outputIsTerminal() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	info, err := os.Stdout.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
