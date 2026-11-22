// Package ansi fournit des couleurs de terminal écrites à la main.
package ansi

import (
	"fmt"
	"strings"
)

// Enabled permet de couper les couleurs (option --no-color, sortie redirigée, tests).
var Enabled = true

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
)

const barWidth = 10

func paint(code, text string) string {
	if !Enabled {
		return text
	}
	return code + text + reset
}

func Bold(text string) string   { return paint(bold, text) }
func Dim(text string) string    { return paint(dim, text) }
func Red(text string) string    { return paint(red, text) }
func Green(text string) string  { return paint(green, text) }
func Yellow(text string) string { return paint(yellow, text) }

// HPBar dessine une barre de vie, par exemple [######----] 18/30.
// Elle est verte, jaune ou rouge selon la vie restante.
func HPBar(current, maxHP int) string {
	if maxHP < 1 {
		maxHP = 1
	}
	current = min(max(current, 0), maxHP)

	filled := current * barWidth / maxHP
	bar := strings.Repeat("#", filled) + strings.Repeat("-", barWidth-filled)
	text := fmt.Sprintf("[%s] %d/%d", bar, current, maxHP)

	percent := current * 100 / maxHP
	switch {
	case percent >= 60:
		return Green(text)
	case percent >= 30:
		return Yellow(text)
	default:
		return Red(text)
	}
}

// DangerLabel traduit un niveau de danger en mot coloré.
func DangerLabel(level int) string {
	switch level {
	case 0:
		return Green("sûr")
	case 1:
		return Yellow("inquiétant")
	case 2:
		return Red("dangereux")
	default:
		return Bold(Red("mortel"))
	}
}
