package open5e

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// Action est une attaque ou une capacité d'un monstre.
type Action struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
	AttackBonus int    `json:"attack_bonus"`
	DamageDice  string `json:"damage_dice"`
	DamageBonus int    `json:"damage_bonus"`
}

// ActionList est la liste des actions d'un monstre.
// L'API renvoie parfois "" au lieu d'une liste quand il n'y a aucune action.
type ActionList []Action

func (list *ActionList) UnmarshalJSON(data []byte) error {
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte(`"`)) {
		*list = nil
		return nil
	}

	var actions []Action
	if err := json.Unmarshal(data, &actions); err != nil {
		return err
	}
	*list = actions
	return nil
}

// Monster est la fiche d'un monstre renvoyée par l'API.
type Monster struct {
	Slug            string     `json:"slug"`
	Name            string     `json:"name"`
	Size            string     `json:"size"`
	Type            string     `json:"type"`
	Alignment       string     `json:"alignment"`
	ArmorClass      int        `json:"armor_class"`
	HitPoints       int        `json:"hit_points"`
	HitDice         string     `json:"hit_dice"`
	ChallengeRating string     `json:"challenge_rating"`
	Strength        int        `json:"strength"`
	Dexterity       int        `json:"dexterity"`
	Constitution    int        `json:"constitution"`
	Intelligence    int        `json:"intelligence"`
	Wisdom          int        `json:"wisdom"`
	Charisma        int        `json:"charisma"`
	Actions         ActionList `json:"actions"`
}

// MonsterPage est une page de résultats. Next et Previous sont vides
// quand il n'y a pas de page suivante ou précédente.
type MonsterPage struct {
	Count    int       `json:"count"`
	Next     string    `json:"next"`
	Previous string    `json:"previous"`
	Results  []Monster `json:"results"`
}

// ChallengeValue transforme "1/4" en 0.25 et "2" en 2.
// Elle renvoie 0 si le format n'est pas reconnu.
func (m Monster) ChallengeValue() float64 {
	top, bottom, isFraction := strings.Cut(strings.TrimSpace(m.ChallengeRating), "/")

	topValue, err := strconv.ParseFloat(top, 64)
	if err != nil {
		return 0
	}
	if !isFraction {
		return topValue
	}

	bottomValue, err := strconv.ParseFloat(bottom, 64)
	if err != nil || bottomValue == 0 {
		return 0
	}
	return topValue / bottomValue
}
