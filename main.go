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

	/*
		f1, err := os.Create("cpuprofile")
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f1); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	*/

	args, err := flags.Parse(&opts)

	if len(args) < 1 {
		fmt.Println("veass version", version)
		fmt.Println("usage: veass <file.s>")
		os.Exit(0)
	}

	filename := args[0]
	if filename[len(filename)-2:] == ".s" {
		assemblerfile, err = NewAssemblerFile(filename)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("unknown file type")
		os.Exit(1)
	}

	assemblermodel := NewAssemblerModel(assemblerfile)

	tui := NewTui()

	tui.topmodel = assemblermodel

	tui.Run()

	/*
		f2, err := os.Create("memprofile")
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f2); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f2.Close()
	*/
}
