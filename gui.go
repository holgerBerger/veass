package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
)

type PanelModel interface {
	GetCell(x, y int) (rune, tcell.Style)
	GetNrLines() int
	GetLineLen(line int) int
}

type Panel struct {
	title       string      // title, in top line, e.g. file name
	firstline   int         // number of first line of buffer being displayed
	cursorline  int         // line of cursor, always visible, cursor is all line
	selectstart int         // first line of a selection
	selectend   int         // last line of selection
	model       *PanelModel // model for this panel
}

type FocusT int

const (
	FocusLeft FocusT = iota
	FocusRight
	FocusBottom
)

type View struct {
	split  bool          // true: two panels displayed side by side (always a bottom panel)
	left   *Panel        // left panel
	right  *Panel        // right panel
	bottom *Panel        // bottom panel spawning full width
	focus  FocusT        // curren tinput focus
	screen *tcell.Screen // screen to draw to
}

// Newview just creates view, does not much
func NewView(s *tcell.Screen) *View {
	newview := &View{split: false}
	newview.screen = s
	return newview
}

func (v *View) Redraw() {
	w, h := (*v.screen).Size()
	hw := w // halfwidth for split
	ls := 0 // left start
	if v.split {
		hw = w / 2
		ls = hw + 1
	}
	// left panel
	// title bar
	style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorSilver)
	for x := 0; x < hw; x++ {
		(*v.screen).SetContent(x, 0, ' ', nil, style)
	}
	// title text
	for x := 0; x < len(v.left.title); x++ {
		(*v.screen).SetContent(x+1, 0, rune(v.left.title[x]), nil, style)
	}

	if v.split {
		for y := 0; y < h-5; y++ {
			(*v.screen).SetContent(hw, y, ' ', nil, style)
		}
		// right panel
		// title bar
		style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorSilver)
		for x := ls; x < w; x++ {
			(*v.screen).SetContent(x, 0, ' ', nil, style)
		}
		// title text
		for x := 0; x < len(v.right.title); x++ {
			(*v.screen).SetContent(x+1+ls, 0, rune(v.right.title[x]), nil, style)
		}
	}

	// bottom title
	for x := 0; x < w; x++ {
		(*v.screen).SetContent(x, h-5, ' ', nil, style)
	}
	// title text
	for x := 0; x < len(v.bottom.title); x++ {
		(*v.screen).SetContent(x+1, h-5, rune(v.bottom.title[x]), nil, style)
	}

	// left panel body
	// FIXME
	for y := 1; y < h-5; y++ {
		lnr := y + (*v.left).firstline
		for x := 0; x < mini(hw, (*v.left.model).GetLineLen(lnr)); x++ {
			r, s := (*v.left.model).GetCell(x, lnr)
			(*v.screen).SetContent(x, y, r, nil, s)
		}
		if mini(hw, (*v.left.model).GetLineLen(lnr)) < hw {
			for x := mini(hw, (*v.left.model).GetLineLen(lnr)); x < hw; x++ {
				(*v.screen).SetContent(x, y, ' ', nil, tcell.StyleDefault)
			}
		}
	}

}

func mini(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func RunGUI(assemblermodel PanelModel) {
	screen, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	quit := make(chan struct{})

	view := NewView(&screen)
	leftPanel := &Panel{title: "LEFT", model: &assemblermodel}
	rightPanel := &Panel{title: "RIGHT"}
	bottomPanel := &Panel{title: "information"}
	view.left = leftPanel
	view.right = rightPanel
	view.bottom = bottomPanel
	view.split = true

	if e = screen.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	screen.Clear()
	screen.Show()

	go func() {
		for {
			ev := screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyDown:
					view.left.firstline++
					view.Redraw()
					screen.Show()
				case tcell.KeyUp:
					view.left.firstline--
					view.Redraw()
					screen.Show()
				case tcell.KeyCtrlL:
					screen.Sync()
					view.Redraw()
				}
			case *tcell.EventResize:
				screen.Clear()
				view.Redraw()
				screen.Sync()
			}
		}
	}()

	<-quit

	screen.Fini()
}
