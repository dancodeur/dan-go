package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Save écrit l'état de la partie dans un fichier JSON.
// On écrit d'abord dans un fichier temporaire, puis on le renomme :
// l'ancienne sauvegarde n'est jamais laissée à moitié écrite.
func (g *Game) Save(path string) error {
	if err := g.blockedByFight("save"); err != nil {
		return err
	}

	data, err := json.MarshalIndent(g.State, "", "  ")
	if err != nil {
		return fmt.Errorf("encodage de la sauvegarde : %w", err)
	}

	temporaryPath := path + ".tmp"
	if err := os.WriteFile(temporaryPath, data, 0o644); err != nil {
		return fmt.Errorf("écriture de la sauvegarde : %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("remplacement du fichier de sauvegarde : %w", err)
	}
	return nil
}

// Load remplace l'état de la partie par celui du fichier.
// Elle renvoie ErrNoSave si le fichier n'existe pas.
func (g *Game) Load(path string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return ErrNoSave
	}
	if err != nil {
		return fmt.Errorf("lecture de la sauvegarde : %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("lecture de la sauvegarde %s : %w", path, err)
	}
	if _, found := g.World.Region(state.Location); !found {
		return fmt.Errorf("sauvegarde %s : région inconnue %q", path, state.Location)
	}
	if state.Slain == nil {
		state.Slain = map[string]int{}
	}

	g.State = state
	g.Foe = nil
	return nil
}
