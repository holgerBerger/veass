package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"os"
)

type PanelModel interface {
	GetCell(x, y int) (rune, tcell.Style, []rune, int)
}

type Panel struct {
	title       string // title, in top line, e.g. file name
	firstline   int    // number of first line of buffer being displayed
	cursorline  int    // line of cursor, always visible, cursor is all line
	selectstart int    // first line of a selection
	selectend   int    // last line of selection
}

type View struct {
	split bool // true if two panels should be displayed

}

func RunGUI() {
	screen, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	quit := make(chan struct{})

	go func() {
		for {
			ev := screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyCtrlL:
					screen.Sync()
				}
			case *tcell.EventResize:
				screen.Sync()
			}
		}
	}()

	<-quit

	screen.Fini()
}
