package main

import (
	"github.com/gdamore/tcell"
)

type assemblerModelT struct {
	file       *FileBuffer
	lastline   string // buffer last line
	lastlinenr int    // buffer numbe rof last lien
}

func (a assemblerModelT) GetCell(x, y int) (rune, tcell.Style) {
	if y != a.lastlinenr || a.lastlinenr == 0 {
		a.lastline = a.file.GetLine(y)
		a.lastlinenr = y
	}
	return rune(a.lastline[x]), tcell.StyleDefault
}

func (a assemblerModelT) GetNrLines() int {
	block := a.file.appendin
	return a.file.lineblocks[block].lastline
}

func (a assemblerModelT) GetLineLen(line int) int {
	return len(a.file.GetLine(line))
}
