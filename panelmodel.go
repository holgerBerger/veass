package main

import "github.com/rthornton128/goncurses"

type PanelModel interface {
	GetCell(x, y int) (rune, int16, goncurses.Char) // char, color, style
	GetNrLines() int
	GetLineLen(line int) int
}
