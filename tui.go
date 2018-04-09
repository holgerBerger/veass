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
	|middle
	(optional)					|
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

	(c) Holger Berger 2018
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

	focus int // 0=top 1=middle

	toptopline int          // file coordinate, 1 in beginning, line number of first line on screen
	toplines   int          // number of lines of top panel (size, has to be updated in resize)
	topcursor  int          // screen coordinate of cursor line (0-(toplines-2))
	topmarked  map[int]bool // marked lines in file coordinates (so we do not have to care of scrolling)
	topmodel   PanelModel   // the data model for top view

	middletopline int          // file coordinate, 1 in beginning, line number of first line on screen
	middlelines   int          // number of lines of middle panel (size, has to be updated in resize)
	middlecursor  int          // screen coordinate of cursor line (0-(middlelines-2))
	middlemarked  map[int]bool // marked lines in file coordinates (so we do not have to care of scrolling)
	middlemodel   PanelModel   // the data model for middle view

	bottomlines int // size of bottom window

	ops          *Opstable
	explainre    *regexp.Regexp
	searchinput  bool
	searchstring string
	searchdir    int
}

// NewTui constructs a user interface, inits ncurses, colors etc
func NewTui() *TuiT {
	var newtui TuiT
	var err error

	newtui.focus = 0
	newtui.topmarked = make(map[int]bool)
	newtui.middlemarked = make(map[int]bool)

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
	gc.InitPair(1, gc.C_WHITE, gc.C_BLACK)   // 1 = White on Black, normal text
	gc.InitPair(2, gc.C_BLUE, gc.C_YELLOW)   // 2 = Blue on yellow, selection
	gc.InitPair(3, gc.C_BLUE, gc.C_BLACK)    // 3 = Blue on black, comments
	gc.InitPair(4, gc.C_RED, gc.C_BLACK)     // 4 = Red on black, labels
	gc.InitPair(5, gc.C_CYAN, gc.C_BLACK)    // 5 = Green on black, directives
	gc.InitPair(6, gc.C_MAGENTA, gc.C_BLACK) // 6 = Magenta on black, local labels
	gc.InitPair(7, gc.C_RED, gc.C_WHITE)     // 7 = Red on white, active tab

	newtui.maxy, newtui.maxx = newtui.scr.MaxYX()

	gc.Raw(true)
	gc.Echo(false)
	gc.Cursor(0)

	newtui.scr.Keypad(true)

	newtui.bottomlines = 5 // size of bottom window

	newtui.middlelines = 0 // size of middle window

	newtui.toplines = newtui.maxy - 1 - newtui.bottomlines - newtui.middlelines - 1

	newtui.toptopline = 1
	newtui.topcursor = 0 // cursor is in screen coordinates
	newtui.top, err = gc.NewWindow(newtui.toplines, newtui.maxx, 0, 0)
	if err != nil {
		panic(err)
	}
	newtui.top.Keypad(true)

	newtui.topbar, err = gc.NewWindow(1, newtui.maxx, newtui.toplines, 0)
	if err != nil {
		panic(err)
	}

	newtui.middletopline = 1
	newtui.middlecursor = 0
	newtui.middle, err = gc.NewWindow(newtui.middlelines, newtui.maxx, newtui.toplines+1, 0)
	if err != nil {
		panic(err)
	}
	newtui.middle.Keypad(true)

	newtui.middlebar, err = gc.NewWindow(1, newtui.maxx, newtui.toplines+1+newtui.middlelines, 0)
	if err != nil {
		panic(err)
	}

	newtui.bottom, err = gc.NewWindow(newtui.bottomlines, newtui.maxx, newtui.maxy-newtui.bottomlines, 0)
	if err != nil {
		panic(err)
	}

	newtui.top.Color(1)
	newtui.top.Erase()
	newtui.middle.Color(1)
	newtui.middle.Erase()
	newtui.bottom.Erase()

	newtui.top.ScrollOk(true)
	newtui.middle.ScrollOk(true)

	// draw empty topbar
	newtui.topbar.AttrOn(gc.A_REVERSE)
	newtui.topbar.Print(fmt.Sprintf("%-*s", newtui.maxx, ""))
	newtui.topbar.AttrOff(gc.A_REVERSE)

	// draw empty middle
	newtui.middlebar.AttrOn(gc.A_REVERSE)
	newtui.middlebar.Print(fmt.Sprintf("%-*s", newtui.maxx, " <no source>"))
	newtui.middlebar.AttrOff(gc.A_REVERSE)

	newtui.scr.NoutRefresh()
	newtui.top.NoutRefresh()
	newtui.topbar.NoutRefresh()
	newtui.middle.NoutRefresh()
	newtui.middlebar.NoutRefresh()
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

	newtui.help()

	return &newtui
}

// Resize should be called on window resize
func (t *TuiT) Resize() {
	t.maxy, t.maxx = t.scr.MaxYX()
	t.toplines = t.maxy - 1 - t.bottomlines - t.middlelines - 1

	t.top.Resize(t.toplines, t.maxx)
	t.topbar.MoveWindow(t.toplines, 0)

	t.middle.Resize(t.middlelines, t.maxx)
	t.middle.MoveWindow(t.toplines+1, 0)

	t.middlebar.MoveWindow(t.toplines+1+t.middlelines, 0)

	t.bottom.Resize(t.bottomlines, t.maxx)
	t.Refreshtopall()
}

// Refresh everything
func (t *TuiT) Refresh() {

	t.middle.NoutRefresh()

	// FIXME do some real redraw here
	t.middlebar.AttrOn(gc.A_REVERSE)
	t.middlebar.Print(fmt.Sprintf("%-*s", t.maxx, " <no source>"))
	t.middlebar.AttrOff(gc.A_REVERSE)

	t.bottom.Erase()
	t.bottom.NoutRefresh()
	t.Refreshtopall()
}

// Refreshtopall draws everything, can be used for paging or resize
func (t *TuiT) Refreshtopall() {
	t.refreshtop()
	t.refreshtopbar()
	gc.Update()
}

// Refreshmiddleall draws everything, can be used for paging or resize
func (t *TuiT) Refreshmiddleall() {
	t.refreshmiddle()
	t.refreshmiddlebar()
	gc.Update()
}

// drawlinetop, y in screen coordinates
func (t *TuiT) drawlinetop(y int) {
	for x := 0; x < mini(t.maxx, t.topmodel.GetLineLen(y+t.toptopline)); x++ {
		r, color, attr := t.topmodel.GetCell(x, y+t.toptopline)
		t.top.AttrOn(attr)
		t.top.ColorOn(color)
		_, ok := t.topmarked[y+t.toptopline]
		if color == 1 && ok {
			t.top.ColorOn(2)
		}

		if y == t.topcursor {
			t.top.AttrOn(gc.A_BOLD)
		}
		t.top.MovePrint(y, x, string(r))
		t.top.AttrOff(attr)
		t.top.AttrOff(gc.A_BOLD)
		t.top.AttrOff(gc.A_REVERSE) // selection
	}
	t.top.ClearToEOL()
}

// drawlinemiddle, y in screen coordinates
func (t *TuiT) drawlinemiddle(y int) {
	for x := 0; x < mini(t.maxx, t.middlemodel.GetLineLen(y+t.middletopline)); x++ {
		r, color, attr := t.middlemodel.GetCell(x, y+t.middletopline)
		t.middle.AttrOn(attr)
		t.middle.ColorOn(color)
		_, ok := t.middlemarked[y+t.middletopline]
		if color == 1 && ok {
			t.middle.ColorOn(2)
		}

		if y == t.middlecursor {
			t.middle.AttrOn(gc.A_BOLD)
		}
		t.middle.MovePrint(y, x, string(r))
		t.middle.AttrOff(attr)
		t.middle.AttrOff(gc.A_BOLD)
		t.middle.AttrOff(gc.A_REVERSE) // selection
	}
	t.middle.ClearToEOL()
}

// refreshtopbar draws the status bar of top, but does not trigger screen update
func (t *TuiT) refreshtopbar() {
	t.topbar.Erase()
	if t.focus == 0 {
		t.topbar.ColorOn(7)
	} else {
		t.topbar.AttrOn(gc.A_REVERSE)
		t.topbar.ColorOn(1)
	}
	t.topbar.Print(fmt.Sprintf("%-*s", t.maxx, " "+t.topmodel.GetFilename()+" in global symbol: "+t.topmodel.GetSymbol(t.toptopline+t.topcursor)))
	t.topbar.MovePrint(0, t.maxx-20, fmt.Sprintf("%d/%d", t.toptopline+t.topcursor, t.topmodel.GetNrLines()))
	t.topbar.AttrOff(gc.A_REVERSE)
	t.topbar.AttrOff(gc.A_BOLD)

	// marker markers <>
	// show in topbar if there is marks before < or behind > current view
	if len(t.topmarked) > 0 {
		min := t.topmodel.GetNrLines() + 1
		max := 0
		for i := range t.topmarked {
			if i > max {
				max = i
			} else if i < min {
				min = i
			}
		}
		if min < t.toptopline {
			t.topbar.ColorOn(2)
			t.topbar.MovePrint(0, t.maxx-24, "<")
		}
		if max > t.toptopline+t.toplines {
			t.topbar.ColorOn(2)
			t.topbar.MovePrint(0, t.maxx-23, ">")
		}
	}

	t.topbar.NoutRefresh()
}

// refreshmiddlebar draws the status bar of middle, but does not trigger screen update
func (t *TuiT) refreshmiddlebar() {
	t.middlebar.Erase()

	if t.focus == 1 {
		//t.middlebar.AttrOn(gc.A_BOLD)
		t.middlebar.ColorOn(7)
	} else {
		t.middlebar.AttrOn(gc.A_REVERSE)
		t.middlebar.ColorOn(1)
	}
	if t.middlelines > 0 {
		t.middlebar.Print(fmt.Sprintf("%-*s", t.maxx, " "+t.middlemodel.GetFilename()))
		t.middlebar.MovePrint(0, t.maxx-20, fmt.Sprintf("%d/%d", t.middletopline+t.middlecursor, t.middlemodel.GetNrLines()))
	} else {
		t.middlebar.Print(fmt.Sprintf("%-*s", t.maxx, " <no source>"))
	}
	t.middlebar.AttrOff(gc.A_REVERSE)
	t.middlebar.AttrOff(gc.A_BOLD)
	t.middlebar.NoutRefresh()
}

// full redraw of top windows
func (t *TuiT) refreshtop() {
	for y := 0; y < t.toplines; y++ {
		t.drawlinetop(y)
	}
	t.top.NoutRefresh()
}

// full redraw of middle windows
func (t *TuiT) refreshmiddle() {
	for y := 0; y < t.middlelines; y++ {
		t.drawlinemiddle(y)
	}
	t.middle.NoutRefresh()
}

// move cursor DOWN top window
func (t *TuiT) sdowntop() {
	updated := false
	if t.topcursor < t.toplines-2 && t.topcursor < t.topmodel.GetNrLines()-1 {
		t.topcursor++
		t.drawlinetop(t.topcursor - 1)
		t.drawlinetop(t.topcursor)
		updated = true
	} else {
		if t.toptopline+t.toplines < t.topmodel.GetNrLines()+2 {
			t.top.Scroll(1)
			t.toptopline++
			t.drawlinetop(t.toplines - 1) // new line
			t.drawlinetop(t.toplines - 2) // new cursor line
			t.drawlinetop(t.toplines - 3) // old cursor line
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
		t.drawlinetop(t.topcursor + 1)
		t.drawlinetop(t.topcursor)
		updated = true
	} else {
		if t.toptopline > 1 {
			t.toptopline--
			t.top.Scroll(-1)
			t.drawlinetop(0)
			t.drawlinetop(1)
			updated = true
		}
	}
	if updated {
		t.top.NoutRefresh()
		t.refreshtopbar()
		gc.Update()
	}
}

// move cursor DOWN middle window
func (t *TuiT) sdownmiddle() {
	updated := false
	if t.middlecursor < t.middlelines-2 && t.middlecursor < t.middlemodel.GetNrLines()-1 {
		t.middlecursor++
		t.drawlinemiddle(t.middlecursor - 1)
		t.drawlinemiddle(t.middlecursor)
		updated = true
	} else {
		if t.middletopline+t.middlelines < t.middlemodel.GetNrLines()+2 {
			t.middle.Scroll(1)
			t.middletopline++
			t.drawlinemiddle(t.middlelines - 1) // new line
			t.drawlinemiddle(t.middlelines - 2) // new cursor line
			t.drawlinemiddle(t.middlelines - 3) // old cursor line
			updated = true
		}
	}
	if updated {
		t.middle.NoutRefresh()
		t.refreshmiddlebar()
		gc.Update()
	}
}

// move cursor UP middle window
func (t *TuiT) supmiddle() {
	updated := false
	if t.middlecursor > 0 {
		t.middlecursor--
		t.drawlinemiddle(t.middlecursor + 1)
		t.drawlinemiddle(t.middlecursor)
		updated = true
	} else {
		if t.middletopline > 1 {
			t.middletopline--
			t.middle.Scroll(-1)
			t.drawlinemiddle(0)
			t.drawlinemiddle(1)
			updated = true
		}
	}
	if updated {
		t.middle.NoutRefresh()
		t.refreshmiddlebar()
		gc.Update()
	}
}

// page down top window
func (t *TuiT) pagedowntop() {
	if t.toptopline+t.toplines > t.topmodel.GetNrLines() {
		// this means all file is on screen, lets move cursor to end of files
		t.jumpendtop()
	} else {
		t.toptopline = mini(t.topmodel.GetNrLines()-t.toplines+1, t.toptopline+t.toplines)
		t.Refreshtopall()
	}
}

// page up top window
func (t *TuiT) pageuptop() {
	if t.toptopline == 1 {
		t.topcursor = 0
	}
	t.toptopline = maxi(1, t.toptopline-t.toplines)
	t.Refreshtopall()
}

// page down middle window
func (t *TuiT) pagedownmiddle() {
	if t.middletopline+t.middlelines > t.middlemodel.GetNrLines() {
		// this means all file is on screen, lets move cursor to end of files
		t.jumpendmiddle()
	} else {
		t.middletopline = mini(t.middlemodel.GetNrLines()-t.middlelines+1, t.middletopline+t.middlelines)
		t.Refreshmiddleall()
	}
}

// page up middle window
func (t *TuiT) pageupmiddle() {
	if t.middletopline == 1 {
		t.middlecursor = 0
	}
	t.middletopline = maxi(1, t.middletopline-t.middlelines)
	t.Refreshmiddleall()
}

func (t *TuiT) jumphometop() {
	t.toptopline = 1
	t.topcursor = 0
	t.Refreshtopall()
}

func (t *TuiT) jumpendtop() {
	t.top.Erase()
	t.toptopline = maxi(1, t.topmodel.GetNrLines()-t.toplines+2)
	t.topcursor = mini(t.topmodel.GetNrLines()-t.toptopline, t.toplines)
	t.Refreshtopall()
}

func (t *TuiT) jumphomemiddle() {
	t.middletopline = 1
	t.middlecursor = 0
	t.Refreshmiddleall()
}

func (t *TuiT) jumpendmiddle() {
	t.middle.Erase()
	t.middletopline = maxi(1, t.middlemodel.GetNrLines()-t.middlelines+2)
	t.middlecursor = mini(t.middlemodel.GetNrLines()-t.middletopline, t.middlelines)
	t.Refreshmiddleall()
}

func (t *TuiT) showlinetop(line int) {
	if line <= t.topmodel.GetNrLines() {
		t.toptopline = mini(line, t.topmodel.GetNrLines()-t.toplines+1)
		t.topcursor = 0
		t.Refreshtopall()
	}
}

func (t *TuiT) showlinemiddle(line int) {
	if line <= t.middlemodel.GetNrLines() {
		t.middletopline = line
		t.middlecursor = 0
		t.Refreshmiddleall()
	}
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

// help prints keyboard help
func (t *TuiT) help() {
	t.bottom.Erase()

	msg := []string{
		"<H>/<h>/<F1>: help, ",
		"<q>: quit, ",
		"<enter>: explain instruction/show in other view, ",
		"<home>: jump to top of file ",
		"<end>/<G>: jump to end of file, ",
		"<n/p>: jump to next/previous search/marks/global symbol ",
		"<i>: position info., ",
		"<c>: clear selection, ",
		"<space>/<backspace>: select/deselect line, ",
		"<m> select lines from same sourceline, ",
		"<v>: view sourcefile, ",
		"<V> close sourcefile, ",
		"<TAB>: change focus ",
		"</>/<?>: search forward/backwards",
		"<d>: highlight dependencies"}

	for _, m := range msg {
		t.printwithbreak(m, t.bottom)
	}

	t.bottom.NoutRefresh()
	gc.Update()
}

// print but break line before text if too long
func (t *TuiT) printwithbreak(text string, w *gc.Window) {
	_, x := w.CursorYX()
	_, width := w.MaxYX()
	if x+len(text) >= width && len(text) < width {
		w.Println("")
	}
	w.Print(text)
}

// print some position info, helps in case of longmangeld c++ names which do not fit into sttaus bar
func (t *TuiT) posinfo() {
	t.bottom.Erase()
	t.bottom.Println("in symbol", t.topmodel.GetSymbol(t.toptopline+t.topcursor))
	filename, linenr := t.topmodel.GetPosition(t.toptopline + t.topcursor)
	t.bottom.Println("produced for line", linenr, "in", filename)
	t.bottom.NoutRefresh()
	gc.Update()
}

// mark a single line in top
func (t *TuiT) marktop() {
	fileline := t.topcursor + t.toptopline
	t.topmarked[fileline] = true
	t.refreshtop()
	gc.Update()
}

// mark all lines which result from same source line as marked line
func (t *TuiT) markalltop() {
	fileline := t.topcursor + t.toptopline
	filename, line := t.topmodel.GetPosition(fileline)
	var fileid int
	for id, name := range assemblerfile.filenametable {
		if name == filename {
			fileid = id
			break
		}
	}
	// file 0 and file 1 are same in case of -g, loc contains 1,X
	if fileid == 0 && len(assemblerfile.filenametable) > 1 && assemblerfile.filenametable[1] == filename {
		fileid = 1
	}
	// we use loctable to quickly jump to .loc lines and search from there
	for _, l := range assemblerfile.loctable[loctuple{fileid, line}] {
		s := l
		for s < t.topmodel.GetNrLines() {
			cf, cl := t.topmodel.GetPosition(s)
			if cf != filename || cl != line {
				break
			}
			t.topmarked[s] = true
			s++
		}
	}
	t.refreshtop()
	gc.Update()
}

func (t *TuiT) showass() {
	filename := t.middlemodel.GetFilename()
	line := t.middletopline + t.middlecursor
	var fileid int
	for id, name := range assemblerfile.filenametable {
		if name == filename || strings.Index(filename, "/"+name) > 0 {
			fileid = id
			break
		}
	}
	// file 0 and file 1 are same in case of -g, loc contains 1,X
	if fileid == 0 && len(assemblerfile.filenametable) > 1 && assemblerfile.filenametable[1] == filename {
		fileid = 1
	}
	firstline := -1
	// we use loctable to quickly jump to .loc lines and search from there
	for _, l := range assemblerfile.loctable[loctuple{fileid, line}] {
		s := l
		for s < t.topmodel.GetNrLines() {
			cf, cl := t.topmodel.GetPosition(s)
			if cf != filename || cl != line {
				break
			}
			t.topmarked[s] = true
			if firstline == -1 {
				firstline = s
			}
			s++
		}
	}
	if firstline != -1 {
		t.showlinetop(firstline)
	}
	t.refreshtop()
	gc.Update()
}

// unmark a line in top
func (t *TuiT) unmarktop() {
	fileline := t.topcursor + t.toptopline
	_, ok := t.topmarked[fileline]
	if ok {
		delete(t.topmarked, fileline)
	}
	t.refreshtop()
	gc.Update()
}

// clear all marked lines in top
func (t *TuiT) clearmarktop() {
	t.topmarked = make(map[int]bool)
	t.refreshtop()
	t.refreshtopbar()
	gc.Update()
}

// clear all marked lines in middle
func (t *TuiT) clearmarkmiddle() {
	t.middlemarked = make(map[int]bool)
	t.refreshmiddle()
	gc.Update()
}

func (t *TuiT) jumpprevioustop() {
	currline := t.toptopline + t.topcursor
	if len(t.topmarked) > 0 {
		closest := -1
		for c := range t.topmarked {
			if c < currline && c > closest {
				closest = c
			}
		}
		if closest != -1 {
			if currline-closest > 1 {
				t.showlinetop(closest - 1) // FIXME why?
			} else {
				t.showlinetop(closest)
			}
		}
	} else {
		currsymbol := t.topmodel.GetSymbol(currline)
		l := currline
		for l > 1 && currsymbol == t.topmodel.GetSymbol(l) {
			l--
		}
		// now search again for start of block
		currsymbol = t.topmodel.GetSymbol(l)
		for l > 1 && currsymbol == t.topmodel.GetSymbol(l) {
			l--
		}
		if l != currline {
			t.showlinetop(l + 1)
		}
	}
}

func (t *TuiT) jumpnexttop() {
	currline := t.toptopline + t.topcursor
	if len(t.topmarked) > 0 {
		closest := t.topmodel.GetNrLines() + 1
		for c := range t.topmarked {
			if c > currline && c < closest {
				closest = c
			}
		}
		if closest != t.topmodel.GetNrLines()+1 {
			t.showlinetop(closest)
		}
	} else {
		currsymbol := t.topmodel.GetSymbol(currline)
		l := currline
		for l < t.topmodel.GetNrLines()-1 && currsymbol == t.topmodel.GetSymbol(l) {
			l++
		}
		if l != currline {
			t.showlinetop(l)
		}
	}
}

// highlight dependencies, highlight input and out registers of current line
func (t *TuiT) dependencies() {
	// search all registers
	re := regexp.MustCompile(`(%v\d+|%s\d+)|(?:(%[a-z]+)[^,\)])`)
	matches := re.FindAllString(t.topmodel.GetLine(t.toptopline+t.topcursor), -1)

	// first register is usually modified, therefor output
	// vst/vsc and st do not alter first register!
	// further registers are usually input

	instrwp := strings.Fields(t.topmodel.GetLine(t.toptopline + t.topcursor))[0]
	instr := strings.Split(instrwp, ".")

	output := ""
	input := ""

	if matches != nil {
		// those are instructions with first arg not being output
		if strings.Index("st vst vsc lvl", instr[0]) != -1 {
			input = input + "(" + matches[0] + ")"
		} else {
			output = output + "(" + matches[0] + ")"
		}

		for _, m := range matches[1:] {
			if input != "" {
				input = input + "|"
			}
			input = input + "(" + m + ")"
		}

		t.topmodel.SetRegexp(regexp.MustCompile(input), regexp.MustCompile(output))
		t.refreshtop()
		gc.Update()
	}
}

// search forward and backward depending on "dir" -1/1
func (t *TuiT) search(dir int) {
	if dir > 0 {
		if t.toptopline+t.topcursor >= t.topmodel.GetNrLines()-1 {
			t.toptopline = 1
			t.topcursor = 0
		}
		for linenr := t.toptopline + t.topcursor + 1; linenr < t.topmodel.GetNrLines(); linenr++ {
			if strings.Index(t.topmodel.GetLine(linenr), t.searchstring) != -1 {
				t.showlinetop(linenr)
				break
			}
		}
	} else {
		if t.toptopline+t.topcursor <= 2 {
			t.toptopline = maxi(1, t.topmodel.GetNrLines()-t.toplines+2)
			t.topcursor = mini(t.topmodel.GetNrLines()-t.toptopline, t.toplines)
		}
		for linenr := t.toptopline + t.topcursor - 1; linenr > 1; linenr-- {
			if strings.Index(t.topmodel.GetLine(linenr), t.searchstring) != -1 {
				t.showlinetop(linenr)
				break
			}
		}
	}
}

func (t *TuiT) opensourcefile() bool {
	var err error
	filename, _ := t.topmodel.GetPosition(t.toptopline + t.topcursor)
	sourcefile, err = NewSourceFile(filename)
	if err != nil {
		t.bottom.Erase()
		t.bottom.Println("could not open sourcefile", filename)
		t.bottom.NoutRefresh()
		gc.Update()
		return false
	}
	t.middlelines = t.toplines / 2
	t.Resize()
	t.middlemodel = NewSourceModel(sourcefile)
	t.refreshmiddle()
	return true
}

// Run is the UI main event loop
func (t *TuiT) Run() {
	t.Refreshtopall()
main:
	for {
		input := t.top.GetChar()
		if t.searchinput {
			switch input {
			case gc.KEY_RETURN:
				t.searchinput = false
				if t.searchstring != "" {
					t.search(t.searchdir)
				}
			case gc.KEY_BACKSPACE:
				t.searchstring = t.searchstring[:len(t.searchstring)-1]
			default:
				t.searchstring = t.searchstring + string(input)
			}
			t.bottom.Erase()
			t.bottom.Print([3]string{"?", "", "/"}[t.searchdir+1] + t.searchstring)
			t.bottom.NoutRefresh()
			gc.Update()
		} else {
			switch input {
			case 'q', 'Q':
				break main
			case gc.KEY_DOWN, 'j':
				if t.focus == 0 {
					t.sdowntop()
				} else {
					t.sdownmiddle()
				}
			case gc.KEY_UP, 'k':
				if t.focus == 0 {
					t.suptop()
				} else {
					t.supmiddle()
				}
			case gc.KEY_PAGEDOWN:
				if t.focus == 0 {
					t.pagedowntop()
				} else {
					t.pagedownmiddle()
				}
			case gc.KEY_PAGEUP:
				if t.focus == 0 {
					t.pageuptop()
				} else {
					t.pageupmiddle()
				}
			case gc.KEY_HOME:
				if t.focus == 0 {
					t.jumphometop()
				} else {
					t.jumphomemiddle()
				}
			case gc.KEY_END, 'G':
				if t.focus == 0 {
					t.jumpendtop()
				} else {
					t.jumpendmiddle()
				}
			case 'n':
				if t.searchstring != "" {
					t.search(t.searchdir)
				} else {
					t.jumpnexttop()
				}
			case 'p':
				if t.searchstring != "" {
					t.search(-1 * t.searchdir)
				} else {
					t.jumpprevioustop()
				}
			case gc.KEY_RESIZE:
				// FIXME does not work
				t.bottom.Println("resize!")
			case gc.KEY_TAB:
				if t.focus == 0 && t.middlelines > 0 {
					t.focus = 1
				} else {
					t.focus = 0
				}
				t.Refreshtopall()
				t.Refreshmiddleall()
			case gc.KEY_RETURN:
				if t.focus == 0 {
					if t.middlelines > 0 {
						filename, line := t.topmodel.GetPosition(t.toptopline + t.topcursor)
						if filename == t.middlemodel.GetFilename() {
							t.middlemarked[line] = true
							t.showlinemiddle(line)
						} else {
							t.bottom.Erase()
							t.bottom.Println("wrong file loaded, use <v> to load file.")
							t.bottom.NoutRefresh()
							gc.Update()
						}
					}
					t.explain()
				} else {
					/*
						if t.middlelines > 0 {
							_, line := t.middlemodel.GetPosition(t.middletopline + t.middlecursor)
							t.middlemarked[line] = true
							t.Refreshtopall()
							gc.Update()
						}
					*/
					t.showass()
				}
			case 'h', 'H', gc.KEY_F1:
				t.help()
			case 'i', 'I':
				t.posinfo()
			case ' ':
				t.marktop()
			case gc.KEY_BACKSPACE:
				t.unmarktop()
			case 'c':
				t.topmodel.SetRegexp(nil, nil)
				t.clearmarktop()
				t.clearmarkmiddle()
			case 'm':
				t.markalltop()
			case 'd':
				t.dependencies()
			case 'V':
				t.focus = 0
				t.middlelines = 0
				t.Resize()
				t.Refresh()
			case 'v':
				if t.opensourcefile() {
					t.Resize()
					_, line := t.topmodel.GetPosition(t.toptopline + t.topcursor)
					if line > 0 {
						t.middlemarked[line] = true
						t.showlinemiddle(line)
					}
					t.refreshmiddlebar()
					t.Refresh()
				}
			case 'R', 'r':
				// FIXME does not work!??!
				t.bottom.Erase()
				t.bottom.Refresh()
				t.Refreshtopall()
				t.Refreshmiddleall()
			case '/':
				t.searchinput = true
				t.searchstring = ""
				t.searchdir = 1
				t.bottom.Erase()
				t.bottom.Print("/" + t.searchstring)
				t.bottom.NoutRefresh()
				gc.Update()
			case '?':
				t.searchinput = true
				t.searchstring = ""
				t.searchdir = -1
				t.bottom.Erase()
				t.bottom.Print([3]string{"?", "", "/"}[t.searchdir+1] + t.searchstring)
				t.bottom.NoutRefresh()
				gc.Update()
			}
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
