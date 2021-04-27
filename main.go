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
	for _, b := range ui.BARS {
		fmt.Println(string(b))
	}
	cmds := widgets.NewSparklineGroup()
	cmds.Title = "By command"
	cmds.SetRect(0, 3, 40, 60)

	ips := widgets.NewSparklineGroup()
	ips.Title = "By IP"
	ips.SetRect(41, 3, 80, 60)

	statz := stats.New(10)
	lock := sync.Mutex{}
	go func() {
		for line := range lines {
			lock.Lock()
			statz.Feed(line)
			lock.Unlock()
		}
	}()
	go func() {
		for {
			time.Sleep(2 * time.Second)

			lock.Lock()
			s := statz.Commands.Values()
			ip := statz.Ips.Values()
			statz.Next()
			lock.Unlock()

			size := len(s)
			cmds.Sparklines = make([]*widgets.Sparkline, size)
			if size > 0 {
				for i, kv := range s {
					cmds.Sparklines[size-i-1] = widgets.NewSparkline()
					//cmds.Sparklines[size-i-1].MaxHeight = 5
					cmds.Sparklines[size-i-1].Data = kv.V
					cmds.Sparklines[size-i-1].Title = fmt.Sprintf("%s %d", kv.K, int(kv.V[len(kv.V)-1]))
				}
			}

			size = len(ip)
			ips.Sparklines = make([]*widgets.Sparkline, size)
			if size > 0 {
				for i, kv := range ip {
					ips.Sparklines[size-i-1] = widgets.NewSparkline()
					ips.Sparklines[size-i-1].MaxHeight = 2
					ips.Sparklines[size-i-1].Data = kv.V
					ips.Sparklines[size-i-1].Title = kv.K
				}
			}

			ui.Render(cmds, ips)
		}
	}()

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}

}
