package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// Dice représente un jet de dés comme 2d6+3 : 2 dés à 6 faces, plus 3.
type Dice struct {
	Count int `json:"count"`
	Sides int `json:"sides"`
	Bonus int `json:"bonus"`
}

// ParseDice lit une notation comme "d20", "2d6" ou "1d8+2".
func ParseDice(text string) (Dice, error) {
	cleaned := strings.ReplaceAll(strings.ToLower(text), " ", "")

	countText, rest, found := strings.Cut(cleaned, "d")
	if !found {
		return Dice{}, fmt.Errorf("invalid dice %q: missing 'd'", text)
	}

	count := 1
	if countText != "" {
		parsedCount, err := strconv.Atoi(countText)
		if err != nil {
			return Dice{}, fmt.Errorf("invalid dice %q: bad dice count: %w", text, err)
		}
		count = parsedCount
	}

	// Après le "d" : le nombre de faces, puis un bonus optionnel (+2 ou -1).
	sidesText, bonusText := rest, ""
	if position := strings.IndexAny(rest, "+-"); position >= 0 {
		sidesText, bonusText = rest[:position], rest[position:]
	}

	sides, err := strconv.Atoi(sidesText)
	if err != nil {
		return Dice{}, fmt.Errorf("invalid dice %q: bad number of sides: %w", text, err)
	}

	bonus := 0
	if bonusText != "" {
		bonus, err = strconv.Atoi(bonusText)
		if err != nil {
			return Dice{}, fmt.Errorf("invalid dice %q: bad bonus: %w", text, err)
		}
	}

	if count < 1 || sides < 1 {
		return Dice{}, fmt.Errorf("invalid dice %q: count and sides must be at least 1", text)
	}
	return Dice{Count: count, Sides: sides, Bonus: bonus}, nil
}

// Roll lance les dés et renvoie le total, bonus compris.
func (d Dice) Roll(rng *rand.Rand) int {
	total := d.Bonus
	if d.Sides < 1 {
		return total
	}
	for range d.Count {
		total += rng.Intn(d.Sides) + 1
	}
	return total
}

// String affiche les dés au format 1d8+2.
func (d Dice) String() string {
	text := fmt.Sprintf("%dd%d", d.Count, d.Sides)
	if d.Bonus > 0 {
		text += fmt.Sprintf("+%d", d.Bonus)
	} else if d.Bonus < 0 {
		text += fmt.Sprintf("%d", d.Bonus)
	}
	return text
}
