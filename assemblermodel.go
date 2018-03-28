package main

import (
	"github.com/rthornton128/goncurses"
	"regexp"
)

type assemblerModel struct {
	file         *FileBuffer
	lastline     string // buffer last line
	lastlinenr   int    // buffer number of last line
	lastcolor    int16
	recomment    *regexp.Regexp
	relabel      *regexp.Regexp
	relocallabel *regexp.Regexp
	redirective  *regexp.Regexp
}

func NewAssemblerModel(filebuffer *FileBuffer) *assemblerModel {
	var na assemblerModel
	na.file = filebuffer
	na.recomment = regexp.MustCompile("^#.*")
	na.relabel = regexp.MustCompile(`^\S+:.*`)
	na.relocallabel = regexp.MustCompile(`^\.\S+:.*`)
	na.redirective = regexp.MustCompile(`^\s+\..*`)
	return &na
}

func (a assemblerModel) GetCell(x, y int) (rune, int16, goncurses.Char) {
	if y != a.lastlinenr || a.lastlinenr == 0 {
		a.lastline = a.file.GetLine(y)
		a.lastlinenr = y

		a.lastcolor = 1

		// comments
		m := a.recomment.FindAllStringSubmatch(a.lastline, -1)
		if m != nil {
			a.lastcolor = 3
		}
		// labels
		m = a.relabel.FindAllStringSubmatch(a.lastline, -1)
		if m != nil {
			a.lastcolor = 4
		}
		// local labels
		m = a.relocallabel.FindAllStringSubmatch(a.lastline, -1)
		if m != nil {
			a.lastcolor = 6
		}
		// directives
		m = a.redirective.FindAllStringSubmatch(a.lastline, -1)
		if m != nil {
			a.lastcolor = 5
		}
	}
	return rune(a.lastline[x]), a.lastcolor, 0
}

func (a assemblerModel) GetNrLines() int {
	block := a.file.appendin
	return a.file.lineblocks[block].lastline
}

func (a assemblerModel) GetLineLen(line int) int {
	if line <= a.GetNrLines() {
		return len(a.file.GetLine(line))
	} else {
		return 0
	}
}
