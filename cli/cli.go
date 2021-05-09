package cli

import (
	"context"
	"fmt"
	"log"
	"strconv"
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

	redis, err := monitor.Redis(host, password)
	if err != nil {
		return err
	}
	lines, err := redis.Monitor(context.TODO())
	if err != nil {
		return err
	}
	if err := ui.Init(); err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	width, height := ui.TerminalDimensions()
	var myWidth int
	if width >= 120 {
		myWidth = 120
	} else {
		myWidth = 80
	}

	infoServer, err := redis.InfoServer()
	if err != nil {
		return err
	}
	p := widgets.NewTable()
	p.Title = fmt.Sprintf("Redis Top -[ v%s/%s pid: %s port: %s hz: %s uptime: %sd]",
		infoServer["redis_version"],
		infoServer["multiplexing_api"],
		infoServer["process_id"],
		infoServer["tcp_port"],
		infoServer["hz"],
		infoServer["uptime_in_days"],
	)
	p.Rows = make([][]string, 1)
	if myWidth > 80 {
		p.Rows[0] = make([]string, 6)
	} else {
		p.Rows[0] = make([]string, 4)
	}
	p.Rows[0][0] = fmt.Sprintf("")
	p.SetRect(0, 0, myWidth, 3)
	ui.Render(p)

	graph := widgets.NewSparkline()
	graphBox := widgets.NewSparklineGroup(graph)
	fatGraphY := 8
	if height > 40 {
		fatGraphY = 16
	}
	graphBox.SetRect(0, 3, myWidth, fatGraphY)
	ui.Render(graphBox)

	cmds := widgets.NewTable()
	cmds.RowSeparator = false
	cmds.Title = "By command/s"
	cmds.ColumnWidths = []int{30, 10}
	cmds.SetRect(0, fatGraphY, 40, height)

	ips := widgets.NewTable()
	ips.RowSeparator = false
	ips.Title = "By IP/s"
	ips.SetRect(41, fatGraphY, 80, height)

	pile := NewPile(81, fatGraphY, 39)

	keyspaces := widgets.NewTable()
	pile.Add(keyspaces)
	keyspaces.RowSeparator = false
	keyspaces.Title = "Keyspace"
	keyspaces.Rows = make([][]string, 2)

	if myWidth > 80 {

		memories := widgets.NewTable()
		pile.Add(memories)
		memories.RowSeparator = false
		memories.Title = "Memory"
		memories.Rows = make([][]string, 4)

		pile.ComputePosition()

		go func() {
			for {
				m, err := redis.Memory()
				if err != nil {
					log.Printf("Memory Error : %s", err.Error())
					continue
				}
				p.Rows[0][4] = fmt.Sprintf("keys: %d", m.KeysCount)
				p.Rows[0][5] = fmt.Sprintf("mem: %s", DisplayUnit(float64(m.PeakAllocated)))
				memories.Rows = m.Table()
				if len(memories.Rows) > 0 && len(memories.Rows[0]) > 0 {
					ui.Render(memories)
				}
				time.Sleep(5 * time.Second)
			}
		}()

	}

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
		for {
			kv, err := redis.Stats()
			if err != nil {
				log.Printf("Stats Error : %s", err.Error())
				continue
			}
			if kv["instantaneous_ops_per_sec"] == "" {
				p.Rows[0][1] = "️0 op"
			} else {
				ops, err := strconv.ParseFloat(kv["instantaneous_ops_per_sec"], 32)
				if err != nil {
					log.Printf("Float parse error: %s %s", kv["instantaneous_ops_per_sec"], err)
					p.Rows[0][1] = "☠️"
				} else {
					p.Rows[0][1] = fmt.Sprintf("%s ops/s", DisplayUnit(ops))
				}
			}
			iips, err := strconv.ParseFloat(kv["instantaneous_input_kbps"], 32)
			if err != nil {
				log.Printf("Float parse error: %s %s", kv["instantaneous_input_kbps"], err)
				p.Rows[0][2] = "☠️"
			} else {
				p.Rows[0][2] = fmt.Sprintf("in: %sb/s", DisplayUnit(iips))
			}
			iops, err := strconv.ParseFloat(kv["instantaneous_output_kbps"], 32)
			if err != nil {
				log.Printf("Float parse error: %s %s", kv["instantaneous_output_kbps"], err)
				p.Rows[0][3] = "☠️"
			} else {
				p.Rows[0][3] = fmt.Sprintf("out: %sb/s", DisplayUnit(iops))
			}
			ui.Render(p)

			if myWidth > 80 {
				keyspaces.Rows[0] = []string{"hits", kv["keyspace_hits"]}
				keyspaces.Rows[1] = []string{"misess", kv["keyspace_misses"]}
				ui.Render(keyspaces)
			}
			time.Sleep(time.Second)
		}
	}()
	go func() {
		poz := 0
		maxValues := myWidth - 2
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
