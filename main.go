package main

/*
	main

	read assembly file
	start gui

*/

import (
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
)

var version string

var opts struct {
	Sourcedirs string `long:"sourcedirs" short:"s" description:"comma seperated list of directories to search for source files"`
}

func main() {
	var (
		assemblerfile *AssemblerFile
		err           error
	)

	args, err := flags.Parse(&opts)

	if len(args) < 1 {
		fmt.Println("veass version", version)
		fmt.Println("usage: veass <file.s>")
		os.Exit(0)
	}

	filename := args[0]
	if filename[len(filename)-2:] == ".s" {
		fmt.Println("reading source file and bulding index, this can take a few seconds")
		assemblerfile, err = NewAssemblerFile(filename)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("unknown file type")
		os.Exit(1)
	}

	assemblermodel := NewAssemblerModel(assemblerfile.filebuffer)

	tui := NewTui()

	tui.topmodel = assemblermodel

	tui.Run()

}
