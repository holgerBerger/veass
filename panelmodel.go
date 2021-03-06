package main

// 	(c) Holger Berger 2018

import (
	"regexp"

	"github.com/rthornton128/goncurses"
)

// PanelModel is the interface a view uses to get data from underlying data model
type PanelModel interface {
	GetCell(x, y int) (rune, int16, goncurses.Char) // char, color, style
	GetNrLines() int
	GetLineLen(line int) int
	GetLine(line int) string
	GetFilename() string
	GetSymbol(line int) string
	GetPosition(line int) (string, int)
	SetRegexp(r1, r2 *regexp.Regexp)
}
