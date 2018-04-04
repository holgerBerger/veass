package main

/*
  read a source file into a buffer

	(c) Holger Berger 2018
*/

import (
	"bufio"
	"fmt"
	"os"
)

type Sourcefile struct {
	filebuffer *FileBuffer
}

func NewSourceFile(filename string) (*Sourcefile, error) {
	var (
		newfile Sourcefile
		err     error
	)

	// construct the object
	newfile.filebuffer = NewFileBuffer(filename)

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
		strline := fmt.Sprintf("%7d: %s", linecount, line[:len(line)-1])

		// append to buffer
		newfile.filebuffer.Addline(linecount, strline)

		// now we are done, push up linenumber
		linecount++
	}
	ifile.Close()
	return &newfile, nil
}
