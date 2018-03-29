package main

/*

  read an assembler file into a buffer
  as we want to be fast and are not (yet) interested in details
  and want to safe memory as well, we only store the associated
  source lines with it.

  as we want to be able to jump from assembler to source and
  from source to assembler, we store an index to find
  assembler lines associated with a line quickly.
	From assembler to sourceline can be done with backward search
	in assembler

  locations is in assembler as
  (see https://sourceware.org/binutils/docs-2.18/as/LNS-directives.html#LNS-directives)

  .file idx "path"
  .loc idx line column

*/

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type loctuple struct {
	fileid, linenr int
}

// AssemblerFile is the class to represent the file and its locations tables
type AssemblerFile struct {
	filebuffer    *FileBuffer        // the associated buffer storing the file
	filenametable []string           // table of filename
	loctable      map[loctuple][]int // table mapping loctuple to linenumber in assembler file
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

		// process the line, search for .file and .loc and <# line nr>
		if strline[0] == '#' {
			if strings.Index(strline, "# line") != -1 {
				// FIXME
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
					newfile.loctable[loctuple{fileid, linenr}] = append(newfile.loctable[loctuple{fileid, linenr}], linecount)
				} else if strings.Index(strline[pos:], ".file ") != -1 {
					// collect filenames, the without index is current file and in position 0
					flds := strings.Fields(strline[pos:])
					if len(flds) > 2 {
						newfile.filenametable = append(newfile.filenametable, flds[2][1:len(flds[2])-1])
						expandfilename(flds[2][1 : len(flds[2])-1])
					} else {
						newfile.filenametable = append(newfile.filenametable, flds[1][1:len(flds[1])-1])
						expandfilename(flds[1][1 : len(flds[1])-1])
					}
				}
			}
		}

		// now we are done, push up linenumber
		linecount++
	}

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
	fmt.Println("did not find source", fn)
	return fn
}
