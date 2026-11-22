package world

import (
	"os"
	"path/filepath"
	"testing"
)

// Le vrai world.json du projet doit toujours être valide.
func TestShippedWorldIsValid(t *testing.T) {
	world, err := Load("../../world.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(world.Regions) < 5 {
		t.Errorf("world has %d regions, want at least 5", len(world.Regions))
	}
}

func TestLoadRejectsBrokenWorlds(t *testing.T) {
	tests := map[string]string{
		"unknown start":     `{"start": "nowhere", "regions": [{"id": "a", "name": "A"}]}`,
		"unknown exit":      `{"start": "a", "regions": [{"id": "a", "name": "A", "exits": ["b"]}]}`,
		"not valid json":    `{"start": `,
		"empty world":       `{}`,
		"wrong field types": `{"start": 3, "regions": "none"}`,
	}

	for name, content := range tests {
		path := filepath.Join(t.TempDir(), "world.json")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("expected an error for a missing file")
	}
}

func TestHasExit(t *testing.T) {
	region := Region{Exits: []string{"a", "b"}}

	if !region.HasExit("a") || region.HasExit("c") {
		t.Error("HasExit gives the wrong answer")
	}
}
