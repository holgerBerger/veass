package main

/*

	text user interface

	this draws everything on screen

	we have
	+---------------------------+
	|top												|
	|														|
	|														|
	=topbar======================
	|middle (optional)					|
	|														|
	=middlebar===================
	|bottom											|
	+----------------------------

	top is for assembly (can take focus and input)
	middle for source code (can take focus and input)
	bottom for messages (read only)

	this is a hack, unproper model/view architecture

	ncurses sucks, therefor we stick to 8 colors atm.
	ncurses is very hard coded here, tcell was to slow
	for remote usage

*/

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	gc "github.com/rthornton128/goncurses"
)

// TuiT is top level data structure
type TuiT struct {
	scr        *gc.Window
	maxx, maxy int
	top        *gc.Window
	topbar     *gc.Window
	middle     *gc.Window
	middlebar  *gc.Window
	bottom     *gc.Window

	toptopline int // file coordinate, 1 in beginning, line number of first line on screen
	toplines   int // number of lines of top panel (size, has to be updated in resize)
	topcursor  int // screen coordinate of cursor line (0-(toplines-2))

	topmodel   PanelModel
	topbartext string

	ops       *Opstable
	explainre *regexp.Regexp
}

// NewTui constructs a user interface, inits ncurses, colors etc
func NewTui() *TuiT {
	var newtui TuiT
	var err error

	newtui.ops = NewOpstable()
	newtui.explainre = regexp.MustCompile(`^\s+(.+?)[\[\s].+$`)

	newtui.scr, err = gc.Init()
	if err != nil {
		panic(err)
	}

	err = gc.StartColor()
	if err != nil {
		panic(err)
	}

	// gc.UseDefaultColors() // do not invert

	if gc.CanChangeColor() {
		gc.InitColor(gc.C_WHITE, 1000, 1000, 1000)
	}

	// FIXME make those constants named
	gc.InitPair(1, gc.C_WHITE, gc.C_BLACK)   // 1 = Black on White, normal text
	gc.InitPair(2, gc.C_BLACK, gc.C_YELLOW)  // 2 = Black on yellow, selection
	gc.InitPair(3, gc.C_BLUE, gc.C_BLACK)    // 3 = Blue on black, comments
	gc.InitPair(4, gc.C_RED, gc.C_BLACK)     // 4 = Red on black, labels
	gc.InitPair(5, gc.C_CYAN, gc.C_BLACK)    // 5 = Green on black, directives
	gc.InitPair(6, gc.C_MAGENTA, gc.C_BLACK) // 6 = Magenta on black, local labels

	newtui.maxy, newtui.maxx = newtui.scr.MaxYX()

	gc.Raw(true)
	gc.Echo(false)
	gc.Cursor(0)

	newtui.scr.Keypad(true)

	newtui.toplines = newtui.maxy - 5 // size of bottom window, no middle
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

	newtui.top.Color(1)
	newtui.top.Erase()
	newtui.bottom.Erase()

	newtui.top.ScrollOk(true)

	// draw empty topbar
	newtui.topbar.AttrOn(gc.A_REVERSE)
	newtui.topbar.Print(fmt.Sprintf("%-*s", newtui.maxx, newtui.topbartext))
	newtui.topbar.AttrOff(gc.A_REVERSE)

	newtui.scr.NoutRefresh()
	newtui.top.NoutRefresh()
	newtui.topbar.NoutRefresh()
	newtui.bottom.NoutRefresh()
	gc.Update()

	// FIXME broken resize code
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGWINCH)

	go func(c chan os.Signal) {
		for {
			_ = <-c
			newtui.Resize()
		}
	}(c)

	return &newtui
}

// Resize should be called on window resize
func (t *TuiT) Resize() {
	t.maxy, t.maxx = t.scr.MaxYX()
	t.toplines = t.maxy - 5 // FIXME, no middle here
	t.Refresh()
}

// Refresh draws everything, can be used for paging or resize
func (t *TuiT) Refresh() {
	t.refreshtop()
	t.refreshtopbar()
	gc.Update()
}

// drawline, y in screen coordinates
func (t *TuiT) drawline(y int) {
	for x := 0; x < mini(t.maxx, t.topmodel.GetLineLen(y+t.toptopline)); x++ {
		r, color, attr := t.topmodel.GetCell(x, y+t.toptopline)

		t.top.AttrOn(attr | gc.A_DIM)
		t.top.ColorOn(color)

		if y == t.topcursor {
			// t.top.AttrOn(gc.A_REVERSE)
			t.top.AttrOn(gc.A_BOLD)
		}
		t.top.MovePrint(y, x, string(r))

		t.top.AttrOff(attr | gc.A_DIM)
		// t.top.AttrOff(gc.A_REVERSE)
		t.top.AttrOff(gc.A_BOLD)
	}
	t.top.ClearToEOL()
	/* old code for inverse cursor, not needed for bold cursor
	// draw end of line after string
	if y == t.topcursor {
		//t.top.AttrOn(gc.A_REVERSE)
		t.top.AttrOn(gc.A_BOLD)
		l := mini(t.maxx, t.topmodel.GetLineLen(y+t.toptopline))
		t.top.MovePrint(y, l, fmt.Sprintf("%-*s", t.maxx-l, ""))
		//t.top.AttrOff(gc.A_REVERSE)
		t.top.AttrOff(gc.A_BOLD)
	} else {
		t.top.ClearToEOL()
	}
	*/
}

// refreshtopbar draws the status bar of top, but does not trigger screen update
func (t *TuiT) refreshtopbar() {
	t.topbar.Erase()
	t.topbar.AttrOn(gc.A_REVERSE)
	t.topbar.Color(1)
	t.topbar.Print(fmt.Sprintf("%-*s", t.maxx, " "+t.topmodel.GetFilename()+" in global symbol: "+t.topmodel.GetSymbol(t.toptopline+t.topcursor)))
	t.topbar.MovePrint(0, t.maxx-20, fmt.Sprintf("%d/%d", t.toptopline+t.topcursor, t.topmodel.GetNrLines()))
	t.topbar.AttrOff(gc.A_REVERSE)
	t.topbar.NoutRefresh()
}

// full redraw of top windows
func (t *TuiT) refreshtop() {
	for y := 0; y < t.toplines; y++ {
		t.drawline(y)
	}
	t.top.NoutRefresh()
}

// move cursor DOWN top window
func (t *TuiT) sdowntop() {
	updated := false
	if t.topcursor < t.toplines-2 && t.topcursor < t.topmodel.GetNrLines()-1 {
		t.topcursor++
		t.drawline(t.topcursor - 1)
		t.drawline(t.topcursor)
		updated = true
	} else {
		if t.toptopline+t.toplines < t.topmodel.GetNrLines()+2 {
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
		t.refreshtopbar()
		gc.Update()
	}
}

// move cursor UP top window
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
		t.refreshtopbar()
		gc.Update()
	}
}

// page down top window
func (t *TuiT) pagedowntop() {
	t.toptopline = mini(t.topmodel.GetNrLines()-t.toplines+1, t.toptopline+t.toplines)
	t.Refresh()
}

// page up top window
func (t *TuiT) pageuptop() {
	t.toptopline = maxi(1, t.toptopline-t.toplines)
	t.Refresh()
}

// explain an assembly instruction
func (t *TuiT) explain() {
	line := t.topmodel.GetLine(t.toptopline + t.topcursor)
	// find main explanation
	m := t.explainre.FindStringSubmatch(line)
	if m == nil {
		// for lines without spaces at end
		r := regexp.MustCompile(`^\s+(.+?)$`)
		m = r.FindStringSubmatch(line)
		if m == nil {
			return // bail out for lines not matching
		}
	}
	t.bottom.Erase()
	//t.bottom.Println("try <", m[1], "> ")
	e := t.ops.getops(m[1])
	if e != "" {
		t.bottom.Println(e)
	} else {
		// FIXME what about . in first position?
		tokens := strings.Split(m[1], ".")
	outer:
		for i := len(tokens); i >= 1; i-- {
			o := tokens[0]
			for j := 1; j < i; j++ {
				o = o + "." + tokens[j]
			}
			if o != "" {
				// t.bottom.Println("try <", o, "> ")
				e := t.ops.getops(o)
				if e != "" {
					t.bottom.Println(e)
					break outer
				}
			}
		}
	}
	// explain suffixes
	tokens := strings.Fields(line)
	first := true
	for suffix := range suffixes {
		if strings.Index(tokens[0]+".", suffix+".") >= 0 {
			if !first {
				t.bottom.Print(", ")
			}
			t.bottom.Print(suffix, ":", suffixes[suffix])
			first = false
		}
	}
	// explain registers
	if !first {
		t.bottom.Println()
	}
	first = true
	for register := range registers {
		if strings.Index(line, register) >= 0 {
			if !first {
				t.bottom.Print(", ")
			}
			t.bottom.Print(register, ":", registers[register])
			first = false
		}
	}

	t.bottom.NoutRefresh()
	gc.Update()
}

// Run is the UI main event loop
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
		case gc.KEY_PAGEDOWN:
			t.pagedowntop()
		case gc.KEY_PAGEUP:
			t.pageuptop()
		case gc.KEY_RESIZE:
			t.bottom.Println("resize!")
		case gc.KEY_RETURN:
			t.explain()
		}
	}
	gc.End()
}

// min of 2 int
func mini(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max of 2 int
func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}
