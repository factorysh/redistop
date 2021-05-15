package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/factorysh/redistop/monitor"
	"github.com/factorysh/redistop/stats"
	ui "github.com/gizak/termui/v3"
)

func MonitorLoop(redis *monitor.RedisServer, app *App, log *Logger) {

	statz := stats.New()
	lock := sync.Mutex{}

	lines, monitorErrors := redis.Monitor(context.TODO(), func(ok bool) {
		if ok {
			ui.Render(app.graphBox)
		} else {
			app.Alert("Not connected")
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
		poz := 0
		maxValues := app.myWidth - 2
		values := make([]int, maxValues)
		for {
			time.Sleep(freq * time.Second)

			app.splash.Text = ""
			app.splash.Border = false
			lock.Lock()
			s := stats.Count(statz.Commands)
			ip := stats.Count(statz.Ips)
			statz.Reset()
			lock.Unlock()
			total := 0
			for _, i := range s {
				total += i.V
			}
			values[poz] = total
			poz++
			if poz >= maxValues {
				poz = 0
			}
			app.graph.Data = make([]float64, maxValues)
			m := 0
			for i := 0; i < maxValues; i++ {
				j := i + poz
				if j >= maxValues {
					j -= maxValues
				}
				app.graph.Data[i] = float64(values[j])
				if values[i] > m {
					m = values[i]
				}
			}
			app.graphBox.Title = fmt.Sprintf("Commands [current: %d max: %d]", total/freq, m/freq)

			size := len(s)
			app.cmds.Rows = make([][]string, size)
			if size > 0 {
				for i, kv := range s {
					app.cmds.Rows[size-i-1] = []string{kv.K, fmt.Sprintf("%.1f", float64(kv.V)/freq)}
				}
			}

			size = len(ip)
			app.ips.Rows = make([][]string, size)
			if size > 0 {
				for i, kv := range ip {
					app.ips.Rows[size-i-1] = []string{kv.K, fmt.Sprintf("%.1f", float64(kv.V)/freq)}
				}
			}

			if len(app.ips.Rows) > 0 {
				ui.Render(app.splash, app.cmds, app.ips, app.graphBox)
			}
		}
	}()

}
