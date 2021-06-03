package cli

import (
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/guptarohit/asciigraph"

	"github.com/rivo/tview"
)

type AppUI struct {
	app            *tview.Application
	grid           *tview.Grid
	header         *tview.Table
	graph          *GraphBox
	splash         *tview.Box
	cmds           *tview.Table
	ips            *tview.Table
	memories       *tview.Table
	pile           *tview.Flex
	keyspaces      *tview.Table
	clients        *tview.Table
	persistence    *tview.Table
	pubsub         *tview.Table
	errorPanel     *tview.TextView
	monitorIsReady bool
}

type GraphBox struct {
	*tview.Box
	lock   *sync.Mutex
	series []float64
}

func NewGraphBox() *GraphBox {
	g := &GraphBox{
		Box:    tview.NewBox(),
		series: make([]float64, 0),
		lock:   &sync.Mutex{},
	}
	g.SetBorder(true)
	return g
}

func (g *GraphBox) SetSeries(series []float64) {
	g.series = series
}

func (g *GraphBox) Draw(screen tcell.Screen) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.DrawForSubclass(screen, g)
	if len(g.series) > 0 {
		x, y, width, height := g.GetInnerRect()
		p := asciigraph.Plot(g.series, asciigraph.Height(height-1))
		for i, line := range strings.Split(p, "\n") {
			fullLine := make([]rune, width)
			for j := 0; j < width; j++ {
				fullLine[j] = ' '
			}
			copy(fullLine, []rune(line))
			screen.SetContent(x, y+i, fullLine[0], fullLine[1:], tcell.StyleDefault)
		}
	}
}

func (g *GraphBox) GetInnerRect() (int, int, int, int) {
	x, y, width, height := g.GetRect()
	return x + 1, y + 1, width - 2, height - 2
}

const art = `
               _._
          _.-''__ ''-._
     _.-''    '.  '_.  ''-._
 .-'' .-'''.  '''\/    _.,_ ''-._
(    '      ,       .-'  | ',    )
|'-._'-...-' __...-.''-._|'' _.-'|
|    '-._   '._    /     _.-'    |
 '-._    '-._  '-./  _.-'    _.-'
|'-._'-._    '-.__.-'    _.-'_.-'|
|    '-._'-._        _.-'_.-'    |
 '-._    '-._'-.__.-'_.-'    _.-'
|'-._'-._    '-.__.-'    _.-'_.-'|
|    '-._'-._        _.-'_.-'    |
 '-._    '-._'-.__.-'_.-'    _.-'
     '-._    '-.__.-'    _.-'
         '-._        _.-'
             '-.__.-'
`

func NewAppUI() *AppUI {
	appUI := &AppUI{
		app:            tview.NewApplication(),
		monitorIsReady: false,
	}
	appUI.fundation()
	return appUI
}

func (a *AppUI) fundation() {
	a.header = tview.NewTable().SetFixed(1, 4)
	a.header.SetBorder(true)
	a.header.SetTitle("Redistop")
	for i := 0; i < 4; i++ {
		a.header.SetCell(0, i, tview.NewTableCell("*"))
	}
	a.graph = NewGraphBox()
	a.graph.SetBorder(true).SetMouseCapture(
		func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
			if action == tview.MouseLeftDoubleClick {
				_, _, _, h := a.graph.GetInnerRect()
				if h <= 5 {
					a.grid.SetRows(3, 12, 0, 1)
				} else {
					if h <= 12 {
						a.grid.SetRows(3, 17, 0, 1)
					} else {
						a.grid.SetRows(3, 7, 0, 1)
					}
				}
			}
			return action, event
		})

	a.splash = tview.NewBox()
	a.splash.SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
		xs := (width - 34) / 2
		ys := (height - 16) / 2
		for i, line := range strings.Split(art, "\n") {
			for j, l := range line {
				screen.SetCell(x+xs+j, y+ys+i, tcell.StyleDefault, l)
			}
		}
		return a.splash.GetInnerRect()
	})

	a.cmds = tview.NewTable()
	a.cmds.SetBorder(true)
	a.cmds.SetSeparator(tcell.RuneVLine)
	a.cmds.SetTitle("By command/s")

	a.ips = tview.NewTable()
	a.ips.SetBorder(true)
	a.ips.SetSeparator(tcell.RuneVLine)
	a.ips.SetTitle("By IP/s")

	a.errorPanel = tview.NewTextView().SetTextColor(tcell.GetColor("red")).SetMaxLines(1)

	a.pile = tview.NewFlex()
	a.pile.SetDirection(tview.FlexRow)

	a.grid = tview.NewGrid().SetRows(3, 7, 0, 1).SetColumns(0, 0, 40).
		AddItem(a.header, 0, 0, 1, 3, 0, 0, false).
		AddItem(a.graph, 1, 0, 1, 3, 0, 0, false).
		AddItem(a.splash, 2, 0, 1, 2, 0, 0, false).
		AddItem(a.pile, 2, 2, 1, 1, 0, 0, false).
		AddItem(a.errorPanel, 3, 0, 1, 3, 0, 0, false)

	a.app.SetRoot(a.grid, true).SetFocus(a.grid).EnableMouse(true)

	a.keyspaces = tview.NewTable()
	a.keyspaces.SetBorder(true)
	a.keyspaces.SetTitle("Keyspace")
	a.keyspaces.SetCellSimple(0, 0, "")
	a.keyspaces.SetCellSimple(0, 1, "")
	a.pile.AddItem(a.keyspaces, 3, 1, false)

	a.pubsub = tview.NewTable()
	a.pubsub.SetBorder(true)
	a.pubsub.SetTitle("Pubsub")
	a.pubsub.SetCellSimple(0, 0, "")
	a.pubsub.SetCellSimple(0, 1, "")
	a.pile.AddItem(a.pubsub, 3, 1, false)

	a.memories = tview.NewTable()
	a.memories.SetBorder(true)
	a.memories.SetTitle("Memory")
	for i := 0; i < 4; i++ {
		a.memories.SetCellSimple(i, 0, "")
		a.memories.SetCellSimple(i, 1, "")
	}
	a.pile.AddItem(a.memories, 6, 1, false)

	a.clients = tview.NewTable()
	a.clients.SetBorder(true)
	a.clients.SetTitle("Clients")
	for i := 0; i < 2; i++ {
		a.clients.SetCellSimple(i, 0, "")
		a.clients.SetCellSimple(i, 1, "")
	}
	a.pile.AddItem(a.clients, 4, 1, false)

	a.persistence = tview.NewTable()
	a.persistence.SetBorder(true)
	a.persistence.SetTitle("Persistance")
	for i := 0; i < 3; i++ {
		a.persistence.SetCellSimple(i, 0, "")
		a.persistence.SetCellSimple(i, 1, "")
	}
	a.pile.AddItem(a.persistence, 5, 1, false)

}

func (a *AppUI) Alert(msg string) {
	a.app.QueueUpdateDraw(func() {
		a.errorPanel.SetText(msg)
	})
}
