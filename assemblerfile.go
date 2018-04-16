package main

/*

  read an assembler file into a buffer
  as we want to be fast and are not (yet) interested in details
  and want to safe memory as well, we only store the associated
  source lines with it, and store fileid/sourceline for each line.

  as we want to be able to jump from assembler to source and
  from source to assembler, we store an index to find
  assembler lines associated with a line quickly.
	From assembler to sourceline can be done with backward search
	in assembler

  locations is in assembler as
  (see https://sourceware.org/binutils/docs-2.18/as/LNS-directives.html#LNS-directives)

  .file idx "path"
  .loc idx line column

	(c) Holger Berger 2018
*/

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// used to jump from source into assembler
type loctuple struct {
	fileid, linenr int
}

// for each line we store this information, indexed by assembly file line number
type indextuple struct {
	loc    loctuple // source location, file and line#
	symbol string   // last=current global symbol
}

// AssemblerFile is the class to represent the file and its locations tables
type AssemblerFile struct {
	filebuffer    *FileBuffer        // the associated buffer storing the file
	filenametable []string           // table of filename
	loctable      map[loctuple][]int // table mapping loctuple to linenumber in assembler file
	index         []indextuple       // table of location information indexed by line number
}

// NewAssemblerFile reads a file into a filebuffer
func NewAssemblerFile(filename string) (*AssemblerFile, error) {
	var (
		newfile AssemblerFile
		err     error
	)

	// construct the object
	newfile.filebuffer = NewFileBuffer(filename)
	newfile.filenametable = make([]string, 0, 1024) // initial space for 1024 filenames
	newfile.loctable = make(map[loctuple][]int)

	ifile, err := os.Open(filename)
	if err != nil {
		return &newfile, err
	}

	reader := bufio.NewReaderSize(ifile, 1024*1024) // get a nice buffer

	fmt.Print("Reading file...")
	linecount := 1
	for {
		// read line from file, bail out at end of file
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		// remove last char \n
		// FIXME might need a check for some files last line? could crash...
		strline := string(line[:len(line)-1])

		// append to buffer
		newfile.filebuffer.Addline(linecount, strline)

		// now we are done, push up linenumber
		linecount++
	}

	// now we know how many lines we have
	newfile.index = make([]indextuple, linecount)

	// go over all lines again
	fmt.Println("\nIndexing file...")
	curloc := loctuple{}
	cursymbol := ""
	// process lines
	for cl := 1; cl < linecount; cl++ {
		strline := newfile.filebuffer.GetLine(cl)
		// process the line, search for .file and .loc and <# line nr>
		if len(strline) == 0 {
			continue
		}
		if strline[0] == '#' {
			if strings.Index(strline, "# line") != -1 {
				flds := strings.Fields(strline)
				linenr, err := strconv.Atoi(flds[2])
				if err != nil {
					panic(err)
				}
				_, ok := newfile.loctable[loctuple{0, linenr}]
				if !ok {
					newfile.loctable[loctuple{0, linenr}] = make([]int, 0, 16)
				}
				newfile.loctable[loctuple{0, linenr}] = append(newfile.loctable[loctuple{0, linenr}], cl)
				curloc = loctuple{0, linenr}
			}
		} else {
			// search for first non-blank character
			pos := 1
			for {
				if strline[pos] != ' ' && strline[pos] != '\t' {
					break
				}
				pos++
			}
			if strline[pos] == '.' {
				if strings.Index(strline[pos:], ".loc\t") != -1 || strings.Index(strline[pos:], ".loc ") != -1 {
					// store .loc lines, to find blocks for a location
					flds := strings.Fields(strline[pos:])
					fileid, err1 := strconv.Atoi(flds[1])
					linenr, err2 := strconv.Atoi(flds[2])
					if err1 != nil || err2 != nil {
						panic(err)
					}
					// check if location already appeared
					_, ok := newfile.loctable[loctuple{fileid, linenr}]
					if !ok {
						newfile.loctable[loctuple{fileid, linenr}] = make([]int, 0, 16)
					}
					newfile.loctable[loctuple{fileid, linenr}] = append(newfile.loctable[loctuple{fileid, linenr}], cl)
					curloc = loctuple{fileid, linenr}
				} else if strings.Index(strline[pos:], ".file ") != -1 {
					// collect filenames, the one without index is current file and in position 0
					flds := strings.Fields(strline[pos:])
					if len(flds) > 2 {
						newfile.filenametable = append(newfile.filenametable, expandfilename(flds[2][1:len(flds[2])-1]))
					} else {
						newfile.filenametable = append(newfile.filenametable, expandfilename(flds[1][1:len(flds[1])-1]))
					}
				} else if strings.Index(strline[pos:], ".globl ") != -1 {
					cursymbol = strings.Join(strings.Fields(strline[pos:])[1:], " ")
				} else if strings.Index(strline[pos:], ".byte") != -1 {
					flds := strings.Fields(strline[pos:])
					v, err := strconv.Atoi(flds[1])
					if err == nil {
						// BUG this is a hack, we access and modify underlying data structure
						block := (cl - 1) / lineblocksize
						newfile.filebuffer.lineblocks[block].lines[(cl-1)%lineblocksize] = fmt.Sprintf("        .byte %3d      # '%s'", v, string(rune(v)))
					}
				}
			} // lines with .
		}
		newfile.index[cl] = indextuple{curloc, cursymbol}
	} // loop process lines

	ifile.Close()
	return &newfile, nil
}

// expandfilename prepends searchpath, so we can later just open and read
func expandfilename(fn string) string {
	if _, err := os.Stat(fn); err == nil {
		fmt.Println("found source", fn)
		return fn
	}
	if fn[0] != '/' && len(opts.Sourcedirs) > 0 {
		sp := strings.Split(opts.Sourcedirs, ",")
		for _, p := range sp {
			testpath := p + "/" + fn
			//fmt.Println("trying ", testpath)
			if _, err := os.Stat(testpath); err == nil {
				fmt.Println("found source", fn, "at", testpath)
				return testpath
			}
		}
	}
	fmt.Println("could not find source", fn)
	return fn
}
