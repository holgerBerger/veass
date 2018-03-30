package main

/*
	the SourceModel implements a model part of a view/model architecture
	it delivers character + attributes to caller
	this model decides how a line appears on screen, character by character
	it will be called for each screen coordinate (valid in text coordinates and in text coordinate)
	to deliver characters from top left to lover right

		(c) Holger Berger 2018
*/

import (
	"regexp"

	"github.com/rthornton128/goncurses"
)

// SourceModel implements PanelModel to allow a viewer to get characters and attributes of certain coordinates in file
type SourceModel struct {
	sourcefile *SourceFile // reference to prepared data
	file       *FileBuffer // reference to underleying data
	lastline   string      // caching: buffer last line
	lastlinenr int         // caching: buffer number of last line
	lastcolor  int16       // caching: color number of last call
}

// NewSourceModel creates a model for the view into an sourcefile
func NewSourceModel(sourcefile *SourceFile) *SourceModel {
	var na SourceModel
	na.sourcefile = sourcefile
	na.file = sourcefile.filebuffer
	return &na
}

// GetCell returns character, color and attribute for a given coordinate in file coordinates, (first line = 1)
func (a SourceModel) GetCell(x, y int) (rune, int16, goncurses.Char) {
	if y != a.lastlinenr || a.lastlinenr == 0 {
		a.lastline = a.file.GetLine(y)
		a.lastlinenr = y

		// FIXME make those color constants names constants

		a.lastcolor = 1

	}
	return rune(a.lastline[x]), a.lastcolor, 0
}

// GetNrLines returns the number of lines in the file
func (a SourceModel) GetNrLines() int {
	// FIXME optimize, could be pushed to filebuffer
	block := a.file.appendin
	return a.file.lineblocks[block].lastline
}

// GetLineLen returns the length of the line (file coordinates, first =  1)
func (a SourceModel) GetLineLen(line int) int {
	if line <= a.GetNrLines() {
		return len(a.file.GetLine(line))
	}
	return 0
}

// GetLine returns line without anyu processing (tabs expanded)
func (a SourceModel) GetLine(line int) string {
	return a.file.GetLine(line)
}

// GetFilename returns the filename
func (a SourceModel) GetFilename() string {
	return a.file.name
}

// GetSymbol returns the global symbol precedding a line
func (a SourceModel) GetSymbol(line int) string {
	// this is a dummy
	return ""
}

// GetPosition returns filename and position in source for line
func (a SourceModel) GetPosition(line int) (string, int) {
	// this is a dummy
	return "", 0
}
