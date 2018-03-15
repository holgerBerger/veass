package main

import (
	"fmt"
	"os"
)

func main() {
	var (
		//assemblerfile *AssemblerFile
		err error
	)

	filename := os.Args[1]
	if filename[len(filename)-2:] == ".s" {
		_, err = NewAssemblerFile(filename)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("unknown file type")
		os.Exit(1)
	}

	RunGUI()
}
