package cli

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/factorysh/redistop/monitor"
	"github.com/factorysh/redistop/stats"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const freq = 2 // Stats per commands and per IPs, every freq seconds

func Top(host, password string) error {
	log.Printf("Connecting to redis://%s\n", host)

	redis := monitor.Redis(host, password)
	lines, err := redis.Monitor(context.TODO())
	if err != nil {
		return err
	}
	if err := ui.Init(); err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	_, height := ui.TerminalDimensions()

	p := widgets.NewTable()
	p.Title = "Redis Top"
	p.Rows = make([][]string, 1)
	p.Rows[0] = make([]string, 4)
	p.Rows[0][0] = host
	p.SetRect(0, 0, 80, 3)
	ui.Render(p)

	graph := widgets.NewSparkline()
	graphBox := widgets.NewSparklineGroup(graph)
	fatGraphY := 8
	if height > 40 {
		fatGraphY = 16
	}
	graphBox.SetRect(0, 3, 80, fatGraphY)

	cmds := widgets.NewTable()
	cmds.RowSeparator = false
	cmds.Title = "By command/s"
	cmds.ColumnWidths = []int{30, 10}
	cmds.SetRect(0, fatGraphY, 40, height)

	ips := widgets.NewTable()
	ips.RowSeparator = false
	ips.Title = "By IP/s"
	ips.SetRect(41, fatGraphY, 80, height)

	statz := stats.New()
	lock := sync.Mutex{}
	go func() {
		for line := range lines {
			lock.Lock()
			statz.Feed(line)
			lock.Unlock()
		}
	}()
	redisStats, err := redis.Stats()
	if err != nil {
		log.Fatalf("Statz %p", err)
	}
	go func() {
		for {
			kv, err := redisStats.Values()
			if err != nil {
				log.Printf("Stats Error : %p", err)
				continue
			}
			p.Rows[0][1] = fmt.Sprintf("%s ops/s", kv["instantaneous_ops_per_sec"])
			p.Rows[0][2] = fmt.Sprintf("in: %s kps", kv["instantaneous_input_kbps"])
			p.Rows[0][3] = fmt.Sprintf("out: %s kps", kv["instantaneous_output_kbps"])
			ui.Render(p)
			time.Sleep(time.Second)
		}
	}()
	go func() {
		poz := 0
		maxValues := 78
		values := make([]int, maxValues)
		for {
			time.Sleep(freq * time.Second)

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
			graphBox.Title = fmt.Sprintf("Commands [current: %d max: %d]", total/freq, m/freq)

			size := len(s)
			cmds.Rows = make([][]string, size)
			if size > 0 {
				for i, kv := range s {
					cmds.Rows[size-i-1] = []string{kv.K, fmt.Sprintf("%.1f", float64(kv.V)/freq)}
				}
			}

			size = len(ip)
			ips.Rows = make([][]string, size)
			if size > 0 {
				for i, kv := range ip {
					ips.Rows[size-i-1] = []string{kv.K, fmt.Sprintf("%.1f", float64(kv.V)/freq)}
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

	return nil

}
