// Package world charge la carte du monde depuis world.json.
package world

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
)

// Region est un lieu du monde, avec ses ennemis possibles et ses sorties.
type Region struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Danger      int      `json:"danger"`
	Foes        []string `json:"foes"`
	Exits       []string `json:"exits"`
}

// World est la liste des régions et le point de départ.
type World struct {
	Start   string   `json:"start"`
	Regions []Region `json:"regions"`
}

// HasExit indique si on peut aller directement vers la région donnée.
func (r Region) HasExit(regionID string) bool {
	return slices.Contains(r.Exits, regionID)
}

// Load lit et vérifie un fichier world.json.
func Load(path string) (*World, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading world file: %w", err)
	}

	var world World
	if err := json.Unmarshal(data, &world); err != nil {
		return nil, fmt.Errorf("decoding world file %s: %w", path, err)
	}
	if err := world.validate(); err != nil {
		return nil, fmt.Errorf("invalid world file %s: %w", path, err)
	}
	return &world, nil
}

// Region cherche une région par son identifiant.
func (w *World) Region(regionID string) (Region, bool) {
	for _, region := range w.Regions {
		if region.ID == regionID {
			return region, true
		}
	}
	return Region{}, false
}

// validate s'assure que le départ et toutes les sorties existent.
func (w *World) validate() error {
	if _, found := w.Region(w.Start); !found {
		return fmt.Errorf("start region %q does not exist", w.Start)
	}
	for _, region := range w.Regions {
		for _, exit := range region.Exits {
			if _, found := w.Region(exit); !found {
				return fmt.Errorf("region %q has an unknown exit %q", region.ID, exit)
			}
		}
	}
	return nil
}
