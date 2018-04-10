package main

/*
	the AssemblerModel implements a model part of a view/model architecture
	it delivers character + attributes to caller
	this model decides how a line appears on screen, character by character
	it will be called for each screen coordinate (valid in text coordinates and in text coordinate)
	to deliver characters from top left to lover right

		(c) Holger Berger 2018
*/

import (
	"regexp"

	"github.com/rthornton128/goncurses"
	gc "github.com/rthornton128/goncurses"
)

// AssemblerModel implements PanelModel to allow a viewer to get characters and attributes of certain coordinates in file
type AssemblerModel struct {
	assemblerfile *AssemblerFile // reference to prepared data
	file          *FileBuffer    // reference to underleying data
	lastline      string         // caching: buffer last line
	lastlinenr    int            // caching: buffer number of last line
	lastcolor     int16          // caching: color number of last call
	recomment     *regexp.Regexp // optimization: precompiled regexps
	relabel       *regexp.Regexp
	relocallabel  *regexp.Regexp
	redirective   *regexp.Regexp
	reregister1   *regexp.Regexp
	reregister2   *regexp.Regexp
	rematch1      [][]int // indices of register matches, 2 directions
	rematch2      [][]int
}

// NewAssemblerModel creates a model for the view into an assemblefile
func NewAssemblerModel(assemblerfile *AssemblerFile) *AssemblerModel {
	var na AssemblerModel
	na.assemblerfile = assemblerfile
	na.file = assemblerfile.filebuffer
	na.recomment = regexp.MustCompile("^#.*")
	na.relabel = regexp.MustCompile(`^\S+:.*`)
	na.relocallabel = regexp.MustCompile(`^\.\S+:.*`)
	na.redirective = regexp.MustCompile(`^\s+\..*`)
	return &na
}

// GetCell returns character, color and attribute for a given coordinate in file coordinates, (first line = 1)
func (a *AssemblerModel) GetCell(x, y int) (rune, int16, goncurses.Char) {
	if y != a.lastlinenr || a.lastlinenr == 0 {
		a.lastline = a.file.GetLine(y)
		a.lastlinenr = y

		// FIXME make those color constants names constants

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

		// match registers, we attach an end marker | here, to properly match cases where a register is last argument
		if a.reregister1 != nil {
			a.rematch1 = a.reregister1.FindAllStringIndex(a.lastline+"|", -1)
		} else {
			a.rematch1 = nil
		}
		if a.reregister2 != nil {
			a.rematch2 = a.reregister2.FindAllStringIndex(a.lastline+"|", -1)
		} else {
			a.rematch2 = nil
		}
	}

	// normal instructions
	if a.lastcolor == 1 && (a.reregister1 != nil || a.reregister2 != nil) {
		for _, ii := range a.rematch1 {
			if x >= ii[0] && x < ii[1] {
				return rune(a.lastline[x]), 4, gc.A_BOLD
			}
		}
		for _, ii := range a.rematch2 {
			if x >= ii[0] && x < ii[1] {
				return rune(a.lastline[x]), 6, gc.A_BOLD
			}
		}
	}
	return rune(a.lastline[x]), a.lastcolor, 0
}

// GetNrLines returns the number of lines in the file
func (a *AssemblerModel) GetNrLines() int {
	// FIXME optimize, could be pushed to filebuffer
	block := a.file.appendin
	return a.file.lineblocks[block].lastline
}

// GetLineLen returns the length of the line (file coordinates, first =  1)
func (a *AssemblerModel) GetLineLen(line int) int {
	if line <= a.GetNrLines() {
		return len(a.file.GetLine(line))
	}
	return 0
}

// GetLine returns line without anyu processing (tabs expanded)
func (a *AssemblerModel) GetLine(line int) string {
	return a.file.GetLine(line)
}

// GetFilename returns the filename
func (a *AssemblerModel) GetFilename() string {
	return a.file.name
}

// GetSymbol returns the global symbol preceeding a line
func (a *AssemblerModel) GetSymbol(line int) string {
	if line < len(a.assemblerfile.index) {
		return a.assemblerfile.index[line].symbol
	}
	return ""
}

// GetPosition returns filename and position in source for line
func (a *AssemblerModel) GetPosition(line int) (string, int) {
	fileid := a.assemblerfile.index[line].loc.fileid
	return a.assemblerfile.filenametable[fileid], a.assemblerfile.index[line].loc.linenr
}

// SetRegexp sets the regular expressions for register highlightning
func (a *AssemblerModel) SetRegexp(r1, r2 *regexp.Regexp) {
	a.reregister1 = r1
	a.reregister2 = r2
}
