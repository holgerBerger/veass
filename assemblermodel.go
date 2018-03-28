package main

import "github.com/rthornton128/goncurses"

type assemblerModelT struct {
	file       *FileBuffer
	lastline   string // buffer last line
	lastlinenr int    // buffer numbe rof last lien
}

func (a assemblerModelT) GetCell(x, y int) (rune, int16, goncurses.Char) {
	if y != a.lastlinenr || a.lastlinenr == 0 {
		a.lastline = a.file.GetLine(y)
		a.lastlinenr = y
	}
	return rune(a.lastline[x]), 1, 0
}

func (a assemblerModelT) GetNrLines() int {
	block := a.file.appendin
	return a.file.lineblocks[block].lastline
}

func (a assemblerModelT) GetLineLen(line int) int {
	return len(a.file.GetLine(line))
}
