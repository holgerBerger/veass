package main

import "github.com/rthornton128/goncurses"

// PanelModel is the interface a view uses to get data from underlying data model
type PanelModel interface {
	GetCell(x, y int) (rune, int16, goncurses.Char) // char, color, style
	GetNrLines() int
	GetLineLen(line int) int
	GetLine(line int) string
}
