package main

import (
	"bytes"
	"fmt"
)

/*

	data structire to store a file line by line
	organized as a Slice of pointers to blocks,
	top level data structure has unlimited size,
	data blocks it points to are fixed in size (lineblocksize)
	This allows for fast access to certain lines and is somehow
	space efficient.
	it is a "insert once and read often but never update" format.

*/

const lineblocksize = 1024 // number of lines in one block
const initialblocks = 1024 // initial number of slots for lineblocks

// FileBuffer is the top level data structure to store a file, refering to LineBlocks
type FileBuffer struct {
	name       string
	lineblocks []*LineBlock
	appendin   int
}

// LineBlock is the second level structure holding the file lines
type LineBlock struct {
	lines               [lineblocksize]string
	len                 int // lines in this list
	firstline, lastline int // lines in files
}

// NewFileBuffer returns a new FileBuffer and initilaizes with one lineblock
func NewFileBuffer(name string) *FileBuffer {
	var (
		filebuffer       FileBuffer
		initiallineblock LineBlock
	)

	// get space for <initialblocks> blocks, and allocate one block for <lineblocksize> lines
	filebuffer.lineblocks = make([]*LineBlock, 0, initialblocks)
	filebuffer.name = name
	filebuffer.lineblocks = append(filebuffer.lineblocks, &initiallineblock)
	filebuffer.appendin = 0

	//fmt.Println(len(filebuffer.lineblocks))

	return &filebuffer
}

// Addline adds a line to a FileBuffer, appending new LineBlocks if needed
func (f *FileBuffer) Addline(linenr int, line string) {

	// we need a new block
	if f.lineblocks[f.appendin].len >= lineblocksize {
		var newlineblock LineBlock
		f.lineblocks = append(f.lineblocks, &newlineblock)
		f.appendin++
		// some progress bar every 50K lines for large files
		// guess nobody will ever notice :-)
		if f.appendin%50 == 0 {
			fmt.Print(".")
		}
	}

	// insert line and increment counters
	lb := &f.lineblocks[f.appendin]
	if (*lb).len == 0 {
		(*lb).firstline = linenr
	}
	(*lb).lines[(*lb).len] = expandtabs(line)
	(*lb).lastline = linenr
	(*lb).len++
}

// GetLine returns given Linenumber as string
func (f *FileBuffer) GetLine(linenr int) string {
	block := (linenr - 1) / lineblocksize
	lb := &f.lineblocks[block]
	if (*lb).firstline > linenr || (*lb).lastline < linenr {
		panic("Internal error!")
	}
	return (*lb).lines[linenr-(*lb).firstline]
}

// expandtabs replaces tab with spaces, we assume tabs of width 8 here
func expandtabsSlow(line string) string {
	var result string
	var spaces int
	for pos := 0; pos < len(line); pos++ {
		if line[pos] == '\t' {
			spaces = (((pos / 8) + 1) * 8) - pos
			for i := 0; i < spaces; i++ {
				result += " "
			}
		} else {
			result += string(line[pos])
		}
	}
	return result
}

// expandtabs replaces tab with spaces, we assume tabs of width 8 here
func expandtabs(line string) string {
	var result bytes.Buffer // this is 4x faster than string concatenation
	var spaces int
	for pos := 0; pos < len(line); pos++ {
		if line[pos] == '\t' {
			spaces = (((pos / 8) + 1) * 8) - pos
			for i := 0; i < spaces; i++ {
				result.WriteString(" ")
			}
		} else {
			result.WriteRune(rune(line[pos]))
		}
	}
	return result.String()
}
