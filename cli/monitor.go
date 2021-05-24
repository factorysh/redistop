package cli

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/factorysh/redistop/circular"
	"github.com/factorysh/redistop/stats"
	"github.com/guptarohit/asciigraph"
)

func (a *App) MonitorLoop() {

	statz := stats.New()
	lock := sync.Mutex{}

	lines, monitorErrors := a.redis.Monitor(context.TODO(), func(ok bool) {
		if ok {
			a.ui.app.QueueUpdate(func() {

			})
		} else {
			a.ui.Alert("Not connected")
		}
	})

	go func() {
		for {
			err := <-monitorErrors
			log.Printf("%v", err)
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
		values := circular.NewCircular(118, scale)
		for {
			time.Sleep(a.config.Frequency)
			a.ui.monitorIsReady = true

			a.ui.splash.Text = ""
			a.ui.splash.Border = false
			lock.Lock()
			s := stats.Count(statz.Commands)
			ip := stats.Count(statz.Ips)
			statz.Reset()
			lock.Unlock()
			for _, i := range s {
				values.Add(i.V)
			}
			_, _, w, _ := a.ui.graph.GetInnerRect()
			vv := values.LastValues(w - 8)
			var m float64 = 0
			for _, v := range vv {
				if v > m {
					m = v
				}
			}
			values.Next()
			a.ui.app.QueueUpdate(func() {
				p := asciigraph.Plot(vv,
					asciigraph.Height(4),
				)
				a.ui.graph.SetText(p)
				a.ui.graph.SetTitle(fmt.Sprintf("Commands [current: %.1f max: %.1f]",
					vv[len(vv)-1],
					m,
				))
				size := len(s)
				a.ui.cmds.Rows = make([][]string, size)
				if size > 0 {
					for i, kv := range s {
						a.ui.cmds.Rows[size-i-1] = []string{kv.K, fmt.Sprintf("%.1f", float64(kv.V)/scale)}
					}
				}

				size = len(ip)
				a.ui.ips.Rows = make([][]string, size)
				if size > 0 {
					for i, kv := range ip {
						a.ui.ips.Rows[size-i-1] = []string{kv.K, fmt.Sprintf("%.1f", float64(kv.V)/scale)}
					}
				}
			})

		}
	}()

}
