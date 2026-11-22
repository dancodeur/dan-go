package game

import (
	"math/rand"
	"testing"
)

func TestParseDice(t *testing.T) {
	tests := []struct {
		text    string
		want    Dice
		wantErr bool
	}{
		{text: "d20", want: Dice{Count: 1, Sides: 20}},
		{text: "2d6", want: Dice{Count: 2, Sides: 6}},
		{text: "1d8+2", want: Dice{Count: 1, Sides: 8, Bonus: 2}},
		{text: "3d4 - 1", want: Dice{Count: 3, Sides: 4, Bonus: -1}},
		{text: "", wantErr: true},
		{text: "20", wantErr: true},
		{text: "xd6", wantErr: true},
		{text: "2d", wantErr: true},
		{text: "0d6", wantErr: true},
		{text: "2d0", wantErr: true},
		{text: "2d6+x", wantErr: true},
	}

	for _, test := range tests {
		got, err := ParseDice(test.text)
		if test.wantErr {
			if err == nil {
				t.Errorf("ParseDice(%q): expected an error, got %v", test.text, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseDice(%q): unexpected error: %v", test.text, err)
			continue
		}
		if got != test.want {
			t.Errorf("ParseDice(%q) = %v, want %v", test.text, got, test.want)
		}
	}
}

func TestRollStaysInRange(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	dice := Dice{Count: 2, Sides: 6, Bonus: 1} // de 3 à 13

	for i := 0; i < 500; i++ {
		total := dice.Roll(rng)
		if total < 3 || total > 13 {
			t.Fatalf("roll %d is outside the range 3..13", total)
		}
	}
}

func TestDiceString(t *testing.T) {
	if got := (Dice{Count: 1, Sides: 8, Bonus: 1}).String(); got != "1d8+1" {
		t.Errorf("got %q, want 1d8+1", got)
	}
	if got := (Dice{Count: 2, Sides: 6, Bonus: -2}).String(); got != "2d6-2" {
		t.Errorf("got %q, want 2d6-2", got)
	}
}
