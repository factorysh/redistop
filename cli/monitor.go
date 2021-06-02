package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/factorysh/redistop/circular"
	"github.com/factorysh/redistop/stats"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) MonitorLoop() {

	statz := stats.New()
	lock := sync.Mutex{}

	lines, monitorErrors := a.redis.Monitor(context.TODO(), func(ok bool) {
		if !ok {
			a.ui.Alert("Not connected")
		}
	})

	go func() {
		for {
			err := <-monitorErrors
			a.ui.Alert(fmt.Sprintf("%v", err))
		}
	}()

	go func() {
		for line := range lines {
			lock.Lock()
			statz.Feed(line)
			lock.Unlock()
		}
	}()

	go func() {
		scale := float64(a.config.Frequency) / float64(time.Second)
		values := circular.NewCircular(250, scale)
		for {
			time.Sleep(a.config.Frequency)
			a.ui.monitorIsReady = true

			a.ui.grid.RemoveItem(a.ui.splash)
			a.ui.grid.AddItem(a.ui.cmds, 2, 0, 1, 1, 0, 0, false).
				AddItem(a.ui.ips, 2, 1, 1, 1, 0, 0, false)

			lock.Lock()
			s := stats.Count(statz.Commands)
			ip := stats.Count(statz.Ips)
			statz.Reset()
			lock.Unlock()
			for _, i := range s {
				values.Add(i.V)
			}
			_, _, w, _ := a.ui.graph.GetInnerRect()
			vv := values.LastValues(w - 7)
			var m float64 = 0
			for _, v := range vv {
				if v > m {
					m = v
				}
			}
			values.Next()
			a.ui.app.QueueUpdateDraw(func() {
				a.ui.graph.SetSeries(vv)
				a.ui.graph.SetTitle(fmt.Sprintf("Commands [current: %.1f max: %.1f]",
					vv[len(vv)-1],
					m,
				))
				a.ui.cmds.Clear()
				size := len(s)
				_, _, w, _ := a.ui.cmds.GetInnerRect()
				if size > 0 {
					for i, kv := range s {
						a.ui.cmds.SetCell(size-i-1, 0,
							tview.NewTableCell(fmt.Sprintf("%-*s", w/2, kv.K)).SetAttributes(tcell.AttrBold))
						a.ui.cmds.SetCell(size-i-1, 1,
							tview.NewTableCell(fmt.Sprintf("%.1f", float64(kv.V)/scale)).
								SetAlign(tview.AlignRight))
					}
				}

				a.ui.ips.Clear()
				size = len(ip)
				_, _, w, _ = a.ui.ips.GetInnerRect()
				if size > 0 {
					for i, kv := range ip {
						a.ui.ips.SetCell(size-i-1, 0,
							tview.NewTableCell(fmt.Sprintf("%-*s", w/2, kv.K)).SetAttributes(tcell.AttrItalic))
						a.ui.ips.SetCellSimple(size-i-1, 1, fmt.Sprintf("%.1f", float64(kv.V)/scale))
					}
				}
			})

		}
	}()

}
