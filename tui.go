package main

import (
	gc "github.com/rthornton128/goncurses"
	"time"
)

type TuiT struct {
	scr        *gc.Window
	maxx, maxy int
	top        *gc.Window
	topbar     *gc.Window
	middle     *gc.Window
	middlebar  *gc.Window
	bottom     *gc.Window
}

func NewTui() *TuiT {
	var newtui TuiT
	var err error

	newtui.scr, err = gc.Init()
	if err != nil {
		panic(err)
	}

	newtui.maxy, newtui.maxx = newtui.scr.MaxYX()

	newtui.scr.Print(newtui.maxy, newtui.maxx)

	gc.Raw(true)
	gc.Echo(false)
	gc.Cursor(0)

	newtui.top, err = gc.NewWindow(newtui.maxx, newtui.maxy/2, 0, 0)
	if err != nil {
		panic(err)
	}
	newtui.topbar, err = gc.NewWindow(newtui.maxx, 1, 0, newtui.maxy/2)
	if err != nil {
		panic(err)
	}

	//newtui.top.MovePrint(0, 0, "top")
	//newtui.topbar.MovePrint(0, 0, "topbar")
	newtui.top.Print("top")
	newtui.topbar.Print("topbar")

	newtui.top.Box(0, 0)

	if gc.MouseOk() {
		newtui.scr.MovePrint(newtui.maxy-1, 0, "WARN: Mouse support not detected.")
	} else {
		newtui.scr.MovePrint(newtui.maxy-1, 0, "OK: Mouse support detected.")
	}

	newtui.top.NoutRefresh()
	newtui.topbar.NoutRefresh()
	newtui.scr.Refresh()
	gc.Update()

	return &newtui
}

func (t *TuiT) Run() {
	time.Sleep(2 * time.Second)
	gc.End()
}
