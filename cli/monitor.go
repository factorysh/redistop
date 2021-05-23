package cli

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/factorysh/redistop/stats"
	ui "github.com/gizak/termui/v3"
)

func (a *App) MonitorLoop() {

	statz := stats.New()
	lock := sync.Mutex{}

	lines, monitorErrors := a.redis.Monitor(context.TODO(), func(ok bool) {
		if ok {
			ui.Render(a.ui.graphBox)
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
		poz := 0
		maxValues := a.ui.myWidth - 2
		values := make([]int, maxValues)
		for {
			time.Sleep(a.config.Frequency)

			a.ui.splash.Text = ""
			a.ui.splash.Border = false
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
			a.ui.graph.Data = make([]float64, maxValues)
			m := 0
			for i := 0; i < maxValues; i++ {
				j := i + poz
				if j >= maxValues {
					j -= maxValues
				}
				a.ui.graph.Data[i] = float64(values[j])
				if values[i] > m {
					m = values[i]
				}
			}
			a.ui.graphBox.Title = fmt.Sprintf("Commands [current: %.1f max: %.1f]",
				float64(total)/float64(a.config.Frequency/time.Second),
				float64(m)/float64(a.config.Frequency/time.Second))

			size := len(s)
			a.ui.cmds.Rows = make([][]string, size)
			if size > 0 {
				for i, kv := range s {
					a.ui.cmds.Rows[size-i-1] = []string{kv.K, fmt.Sprintf("%.1f", float64(kv.V)/float64(a.config.Frequency/time.Second))}
				}
			}

			size = len(ip)
			a.ui.ips.Rows = make([][]string, size)
			if size > 0 {
				for i, kv := range ip {
					a.ui.ips.Rows[size-i-1] = []string{kv.K, fmt.Sprintf("%.1f", float64(kv.V)/float64(a.config.Frequency/time.Second))}
				}
			}

			if len(a.ui.ips.Rows) > 0 {
				ui.Render(a.ui.splash, a.ui.cmds, a.ui.ips, a.ui.graphBox)
			}
		}
	}()

}
