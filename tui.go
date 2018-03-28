package main

import (
	"fmt"

	gc "github.com/rthornton128/goncurses"
)

type TuiT struct {
	scr        *gc.Window
	maxx, maxy int
	top        *gc.Window
	topbar     *gc.Window
	middle     *gc.Window
	middlebar  *gc.Window
	bottom     *gc.Window

	toptopline int // file coordinate
	toplines   int // number of lines of top panel
	topcursor  int // screen coordinate of cursor line

	topmodel   PanelModel
	topbartext string
}

func NewTui() *TuiT {
	var newtui TuiT
	var err error

	newtui.scr, err = gc.Init()
	if err != nil {
		panic(err)
	}

	err = gc.StartColor()
	if err != nil {
		panic(err)
	}

	// gc.UseDefaultColors() // do not invert

	/*
		if !gc.CanChangeColor() {
			panic("can not change colors!")
		}
	*/

	gc.InitPair(1, gc.C_WHITE, gc.C_BLACK)  // 1 = Black on White, normal text
	gc.InitPair(2, gc.C_BLACK, gc.C_YELLOW) // 2 = Black on yellow, selection

	newtui.scr.SetBackground(gc.C_WHITE)

	newtui.maxy, newtui.maxx = newtui.scr.MaxYX()

	gc.Raw(true)
	gc.Echo(false)
	gc.Cursor(0)

	newtui.toplines = newtui.maxy - 5
	newtui.toptopline = 1
	newtui.topcursor = 0 // cursor is in screen coordinates
	newtui.top, err = gc.NewWindow(newtui.maxy-5, newtui.maxx, 0, 0)
	if err != nil {
		panic(err)
	}
	newtui.top.Keypad(true)

	newtui.topbar, err = gc.NewWindow(1, newtui.maxx, newtui.maxy-5, 0)
	if err != nil {
		panic(err)
	}

	newtui.bottom, err = gc.NewWindow(4, newtui.maxx, newtui.maxy-4, 0)
	if err != nil {
		panic(err)
	}

	newtui.top.Erase()
	newtui.bottom.Erase()

	newtui.top.ScrollOk(true)

	newtui.topbar.AttrOn(gc.A_REVERSE)
	newtui.topbar.Print(fmt.Sprintf("%-*s", newtui.maxx, newtui.topbartext))
	newtui.topbar.AttrOff(gc.A_REVERSE)

	newtui.scr.NoutRefresh()
	newtui.top.NoutRefresh()
	newtui.topbar.NoutRefresh()
	newtui.bottom.NoutRefresh()
	gc.Update()

	return &newtui
}

func (t *TuiT) Refresh() {
	t.refreshtop()
	gc.Update()
}

// drawline, y in screen coordinates
func (t *TuiT) drawline(y int) {
	for x := 0; x < mini(t.maxx, t.topmodel.GetLineLen(y+t.toptopline)); x++ {
		r, color, attr := t.topmodel.GetCell(x, y+t.toptopline)

		t.top.AttrOn(attr | gc.A_DIM)
		t.top.ColorOn(color)

		if y == t.topcursor {
			t.top.AttrOn(gc.A_REVERSE)
		}
		t.top.MovePrint(y, x, string(r))

		t.top.AttrOff(attr | gc.A_DIM)
		t.top.AttrOff(gc.A_REVERSE)
	}
	// draw end of line after string
	if y == t.topcursor {
		t.top.AttrOn(gc.A_REVERSE)
		l := mini(t.maxx, t.topmodel.GetLineLen(y+t.toptopline))
		t.top.MovePrint(y, l, fmt.Sprintf("%-*s", t.maxx-l, ""))
		t.top.AttrOff(gc.A_REVERSE)
	} else {
		t.top.ClearToEOL()
	}
}

func (t *TuiT) refreshtop() {
	for y := 0; y < t.toplines; y++ {
		t.drawline(y)
	}
	t.top.NoutRefresh()

	t.topbar.AttrOn(gc.A_REVERSE)
	t.topbar.Print(fmt.Sprintf("%-*s", t.maxx, t.topbartext))
	t.topbar.AttrOff(gc.A_REVERSE)

	t.topbar.NoutRefresh()
}

// move cursor DOWN
func (t *TuiT) sdowntop() {
	updated := false
	if t.topcursor < t.toplines-2 {
		t.topcursor++
		t.drawline(t.topcursor - 1)
		t.drawline(t.topcursor)
		updated = true
	} else {
		if t.toptopline+t.toplines < t.topmodel.GetNrLines() {
			t.top.Scroll(1)
			t.toptopline++
			t.drawline(t.toplines - 1) // new line
			t.drawline(t.toplines - 2) // new cursor line
			t.drawline(t.toplines - 3) // old cursor line
			updated = true
		}
	}
	if updated {
		t.top.NoutRefresh()
		gc.Update()
	}
}

// move cursor UP
func (t *TuiT) suptop() {
	updated := false
	if t.topcursor > 0 {
		t.topcursor--
		t.drawline(t.topcursor + 1)
		t.drawline(t.topcursor)
		updated = true
	} else {
		if t.toptopline > 1 {
			t.toptopline--
			t.top.Scroll(-1)
			t.drawline(0)
			t.drawline(1)
			updated = true
		}
	}
	if updated {
		t.top.NoutRefresh()
		gc.Update()
	}
}

func (t *TuiT) Run() {
	t.Refresh()
main:
	for {
		switch t.top.GetChar() {
		case 'q':
			break main
		case gc.KEY_DOWN:
			t.sdowntop()
		case gc.KEY_UP:
			t.suptop()
		}
	}
	gc.End()
}

func mini(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
