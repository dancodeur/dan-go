package open5e

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

const goblinJSON = `{
  "slug": "goblin", "name": "Goblin", "size": "Small", "type": "humanoid", "alignment": "neutral evil",
  "armor_class": 15, "hit_points": 7, "hit_dice": "2d6", "challenge_rating": "1/4",
  "strength": 8, "dexterity": 14, "constitution": 10, "intelligence": 10, "wisdom": 8, "charisma": 8,
  "actions": [
    {"name": "Scimitar", "desc": "Melee Weapon Attack.", "attack_bonus": 4, "damage_dice": "1d6+2"},
    {"name": "Nimble Escape", "desc": "Disengage or Hide."}
  ]
}`

// Certains monstres ont "actions": "" au lieu d'une liste.
const mummyJSON = `{"slug": "mummy", "name": "Mummy", "armor_class": 11, "hit_points": 58,
  "challenge_rating": "3", "actions": ""}`

// newTestClient lance un faux serveur Open5e et renvoie un client branché dessus,
// avec un compteur du nombre de requêtes reçues.
func newTestClient(t *testing.T) (*Client, *atomic.Int64) {
	t.Helper()
	var requests atomic.Int64
	var server *httptest.Server

	mux := http.NewServeMux()
	mux.HandleFunc("/monsters/goblin/", func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		fmt.Fprint(w, goblinJSON)
	})
	mux.HandleFunc("/monsters/mummy/", func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		fmt.Fprint(w, mummyJSON)
	})
	mux.HandleFunc("/monsters/", func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/monsters/" {
			http.NotFound(w, r)
			return
		}
		query := r.URL.Query()
		switch {
		case query.Get("search") == "gob":
			fmt.Fprintf(w, `{"count": 1, "next": null, "previous": null, "results": [%s]}`, goblinJSON)
		case query.Get("search") != "":
			fmt.Fprint(w, `{"count": 0, "next": null, "previous": null, "results": []}`)
		case query.Get("page") == "2":
			fmt.Fprintf(w, `{"count": 2, "next": null, "previous": "%s/monsters/?limit=1", "results": [%s]}`, server.URL, mummyJSON)
		default:
			fmt.Fprintf(w, `{"count": 2, "next": "%s/monsters/?page=2", "previous": null, "results": [%s]}`, server.URL, goblinJSON)
		}
	})

	server = httptest.NewServer(mux)
	t.Cleanup(server.Close)

	cache := newCache(time.Minute, time.Now)
	return NewClient(server.URL, cache), &requests
}

func TestGetMonsterDecodesStats(t *testing.T) {
	client, _ := newTestClient(t)

	monster, err := client.GetMonster("goblin")
	if err != nil {
		t.Fatal(err)
	}

	if monster.Name != "Goblin" || monster.ArmorClass != 15 || monster.HitPoints != 7 {
		t.Errorf("unexpected stats: %+v", monster)
	}
	if monster.ChallengeValue() != 0.25 {
		t.Errorf("challenge value = %v, want 0.25", monster.ChallengeValue())
	}
	if len(monster.Actions) != 2 || monster.Actions[0].DamageDice != "1d6+2" || monster.Actions[0].AttackBonus != 4 {
		t.Errorf("unexpected actions: %+v", monster.Actions)
	}
}

func TestGetMonsterWithEmptyActionsString(t *testing.T) {
	client, _ := newTestClient(t)

	monster, err := client.GetMonster("mummy")
	if err != nil {
		t.Fatalf("a monster with \"actions\": \"\" should decode: %v", err)
	}
	if len(monster.Actions) != 0 {
		t.Errorf("actions = %v, want none", monster.Actions)
	}
}

func TestGetMonsterUsesTheCache(t *testing.T) {
	client, requests := newTestClient(t)

	for i := 0; i < 3; i++ {
		if _, err := client.GetMonster("goblin"); err != nil {
			t.Fatal(err)
		}
	}
	if requests.Load() != 1 {
		t.Errorf("server received %d requests, want 1", requests.Load())
	}
}

func TestGetMonsterFallsBackToSearch(t *testing.T) {
	client, _ := newTestClient(t)

	monster, err := client.GetMonster("gob") // le slug "gob" n'existe pas
	if err != nil {
		t.Fatal(err)
	}
	if monster.Name != "Goblin" {
		t.Errorf("name = %q, want Goblin", monster.Name)
	}
}

func TestGetMonsterNotFound(t *testing.T) {
	client, _ := newTestClient(t)

	_, err := client.GetMonster("dragonzilla")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestListMonstersFollowsPagination(t *testing.T) {
	client, _ := newTestClient(t)

	first, err := client.ListMonsters("")
	if err != nil {
		t.Fatal(err)
	}
	if first.Count != 2 || first.Next == "" || first.Previous != "" {
		t.Fatalf("unexpected first page: %+v", first)
	}

	second, err := client.ListMonsters(first.Next)
	if err != nil {
		t.Fatal(err)
	}
	if second.Results[0].Name != "Mummy" || second.Next != "" || second.Previous == "" {
		t.Errorf("unexpected second page: %+v", second)
	}
}

func TestPrefetchFillsTheCache(t *testing.T) {
	client, requests := newTestClient(t)

	if err := client.Prefetch([]string{"goblin", "mummy"}); err != nil {
		t.Fatal(err)
	}
	downloaded := requests.Load()

	if _, err := client.GetMonster("goblin"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetMonster("mummy"); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != downloaded {
		t.Error("monsters should have been served from the cache after Prefetch")
	}
}

func TestPrefetchReportsErrors(t *testing.T) {
	client, _ := newTestClient(t)

	if err := client.Prefetch([]string{"goblin", "dragonzilla"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want an error wrapping ErrNotFound", err)
	}
}
