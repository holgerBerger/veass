package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
)

var app = &views.Application{}
var window = &mainWindow{}

type mainWindow struct {
	main   *views.CellView
	keybar *views.SimpleStyledText
	status *views.SimpleStyledTextBar
	model  *model
	views.Panel
}

type model struct {
	x, y       int
	enab, hide bool
}

func (m *model) GetCell(x, y int) (rune, tcell.Style, []rune, int) {
	dig := []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	var ch rune
	style := tcell.StyleDefault
	if x >= 60 || y >= 15 {
		return ch, style, nil, 1
	}
	colors := []tcell.Color{
		tcell.ColorWhite,
		tcell.ColorGreen,
		tcell.ColorMaroon,
		tcell.ColorNavy,
		tcell.ColorOlive,
	}

	ch = dig[(x)%len(dig)]
	style = style.
		Foreground(colors[(y)%len(colors)]).
		Background(tcell.ColorBlack)

	return ch, style, nil, 1
}

func (m *model) GetBounds() (int, int) {
	return 10, 10
}

func (m *model) MoveCursor(offx, offy int) {
	m.x += offx
	m.y += offy
	//m.limitCursor()
}

func (m *model) SetCursor(x int, y int) {
	m.x = x
	m.y = y

	//m.limitCursor()
}

func (m *model) GetCursor() (int, int, bool, bool) {
	return m.x, m.y, m.enab, !m.hide
}

func (a *mainWindow) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlL:
			app.Refresh()
			return true
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'Q', 'q':
				app.Quit()
				return true

			}
		}
	}
	return a.Panel.HandleEvent(ev)
}

func RunGUI() {
	window.model = &model{}

	window.main = views.NewCellView()
	window.main.SetModel(window.model)
	window.main.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorBlack))
	app.SetRootWidget(window)
	if e := app.Run(); e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		os.Exit(1)
	}
}
