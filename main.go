package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/factorysh/redistop/monitor"
	"github.com/factorysh/redistop/stats"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func main() {
	lines, err := monitor.Monitor(context.TODO(), os.Args[1], os.Args[2])
	if err != nil {
		log.Fatalf("", err)
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p := widgets.NewParagraph()
	p.Title = "Redis Top"
	p.Text = fmt.Sprintf("redis://%s", os.Args[1])
	p.SetRect(0, 0, 80, 3)
	ui.Render(p)

	graph := widgets.NewSparkline()
	graphBox := widgets.NewSparklineGroup(graph)
	graphBox.SetRect(0, 3, 80, 8)

	cmds := widgets.NewTable()
	cmds.RowSeparator = false
	cmds.Title = "By command"
	cmds.ColumnWidths = []int{30, 10}
	cmds.SetRect(0, 8, 40, 40)

	ips := widgets.NewTable()
	ips.RowSeparator = false
	ips.Title = "By IP"
	ips.SetRect(41, 8, 80, 40)

	statz := stats.New()
	lock := sync.Mutex{}
	go func() {
		for line := range lines {
			lock.Lock()
			statz.Feed(line)
			lock.Unlock()
		}
	}()
	go func() {
		poz := 0
		maxValues := 78
		values := make([]int, maxValues)
		for {
			time.Sleep(2 * time.Second)

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
			if poz > maxValues {
				poz = 0
			}
			graph.Data = make([]float64, maxValues)
			m := 0
			for i := 0; i < maxValues; i++ {
				j := i + poz
				if j >= maxValues {
					j -= maxValues
				}
				graph.Data[i] = float64(values[j])
				if values[i] > m {
					m = values[i]
				}
			}
			graphBox.Title = fmt.Sprintf("Commands [max: %d]", m)

			size := len(s)
			cmds.Rows = make([][]string, size)
			if size > 0 {
				for i, kv := range s {
					cmds.Rows[size-i-1] = []string{kv.K, fmt.Sprintf("%d", kv.V)}
				}
			}

			size = len(ip)
			ips.Rows = make([][]string, size)
			if size > 0 {
				for i, kv := range ip {
					ips.Rows[size-i-1] = []string{kv.K, fmt.Sprintf("%d", kv.V)}
				}
			}

			ui.Render(cmds, ips, graphBox)
		}
	}()

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}

}
