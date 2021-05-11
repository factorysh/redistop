package cli

import (
	"bytes"
	"context"
	"fmt"
	_log "log"
	"strconv"
	"sync"
	"time"

	"github.com/factorysh/redistop/monitor"
	"github.com/factorysh/redistop/stats"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const freq = 2 // Stats per commands and per IPs, every freq seconds

type App struct {
	header     *widgets.Table
	graph      *widgets.Sparkline
	graphBox   *widgets.SparklineGroup
	cmds       *widgets.Table
	ips        *widgets.Table
	pile       *Pile
	keyspaces  *widgets.Table
	errorPanel *widgets.Paragraph
}

func Top(host, password string) error {
	_log.Printf("Connecting to redis://%s\n", host)
	redis, err := monitor.Redis(host, password)
	if err != nil {
		return err
	}

	if err := ui.Init(); err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	app := &App{}

	width, height := ui.TerminalDimensions()
	var myWidth int
	if width >= 120 {
		myWidth = 120
	} else {
		myWidth = 80
	}

	infos, err := redis.Info()
	if err != nil {
		return err
	}
	app.header = widgets.NewTable()
	app.header.Title = fmt.Sprintf("Redis Top -[ v%s/%s pid: %s port: %s hz: %s uptime: %sd ]",
		infos["redis_version"],
		infos["multiplexing_api"],
		infos["process_id"],
		infos["tcp_port"],
		infos["hz"],
		infos["uptime_in_days"],
	)
	app.header.Rows = make([][]string, 1)
	if myWidth > 80 {
		app.header.Rows[0] = make([]string, 6)
	} else {
		app.header.Rows[0] = make([]string, 4)
	}
	app.header.Rows[0][0] = ""
	app.header.SetRect(0, 0, myWidth, 3)
	ui.Render(app.header)

	app.graph = widgets.NewSparkline()
	app.graphBox = widgets.NewSparklineGroup(app.graph)
	fatGraphY := 8
	if height > 40 {
		fatGraphY = 16
	}
	app.graphBox.SetRect(0, 3, myWidth, fatGraphY)
	ui.Render(app.graphBox)

	app.cmds = widgets.NewTable()
	app.cmds.RowSeparator = false
	app.cmds.Title = "By command/s"
	app.cmds.ColumnWidths = []int{30, 10}
	app.cmds.SetRect(0, fatGraphY, 40, height-3)

	app.ips = widgets.NewTable()
	app.ips.RowSeparator = false
	app.ips.Title = "By IP/s"
	app.ips.SetRect(41, fatGraphY, 80, height-3)

	app.pile = NewPile(81, fatGraphY, 39)

	app.keyspaces = widgets.NewTable()
	app.pile.Add(app.keyspaces)
	app.keyspaces.RowSeparator = false
	app.keyspaces.Title = "Keyspace"
	app.keyspaces.Rows = make([][]string, 2)

	errorPanel := widgets.NewParagraph()
	errorPanel.Title = "Error"
	errorPanel.SetRect(0, height-3, myWidth, height)
	ui.Render(errorPanel)

	log := &Logger{
		block: errorPanel,
	}

	if myWidth > 80 {

		memories := widgets.NewTable()
		app.pile.Add(memories)
		memories.RowSeparator = false
		memories.Title = "Memory"
		memories.Rows = make([][]string, 4)

		app.pile.ComputePosition()

		go func() {
			for {
				m, err := redis.Memory()
				if err != nil {
					log.Printf("Memory Error : %s", err.Error())
				} else {
					app.header.Rows[0][4] = fmt.Sprintf("keys: %d", m.KeysCount)
					app.header.Rows[0][5] = fmt.Sprintf("mem: %s", DisplayUnit(float64(m.PeakAllocated)))
					memories.Rows = m.Table()
				}
				kv, err := redis.Info()
				if err != nil {
					log.Printf("Info Memory Error : %s", err.Error())
				} else {
					memories.Title = fmt.Sprintf("Memory [ %s ]", kv["maxmemory_policy"])
				}

				if len(memories.Rows) > 0 && len(memories.Rows[0]) > 0 {
					ui.Render(memories)
				}
				time.Sleep(5 * time.Second)
			}
		}()

	}

	statz := stats.New()
	lock := sync.Mutex{}

	lines, monitorErrors := redis.Monitor(context.TODO(), func(ok bool) {
		if ok {
			ui.Render(app.graphBox)
		} else {
			msg := "Not connected"
			argh := widgets.NewParagraph()
			argh.SetRect(20, 6, myWidth-20, 11)
			buff := &bytes.Buffer{}
			buff.WriteRune('\n')
			for i := 0; i < (myWidth-40-len(msg))/2; i++ {
				buff.WriteRune(' ')
			}
			buff.WriteString(msg)
			argh.Text = buff.String()
			argh.TextStyle.Fg = ui.ColorRed
			argh.Block.BorderStyle.Fg = ui.ColorRed
			ui.Render(argh)
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
		var cpu *monitor.CPU
		for {
			kv, err := redis.Info()
			if err != nil {
				log.Printf("Info Error : %s", err.Error())
			} else {
				if kv["instantaneous_ops_per_sec"] == "" {
					app.header.Rows[0][1] = "️0 op"
				} else {
					ops, err := strconv.ParseFloat(kv["instantaneous_ops_per_sec"], 32)
					if err != nil {
						log.Printf("Float parse error: %s %s", kv["instantaneous_ops_per_sec"], err)
						app.header.Rows[0][1] = "☠️"
					} else {
						app.header.Rows[0][1] = fmt.Sprintf("%s ops/s", DisplayUnit(ops))
					}
				}
				iips, err := strconv.ParseFloat(kv["instantaneous_input_kbps"], 32)
				if err != nil {
					log.Printf("Float parse error: %s %s", kv["instantaneous_input_kbps"], err)
					app.header.Rows[0][2] = "☠️"
				} else {
					app.header.Rows[0][2] = fmt.Sprintf("in: %sb/s", DisplayUnit(iips))
				}
				iops, err := strconv.ParseFloat(kv["instantaneous_output_kbps"], 32)
				if err != nil {
					log.Printf("Float parse error: %s %s", kv["instantaneous_output_kbps"], err)
					app.header.Rows[0][3] = "☠️"
				} else {
					app.header.Rows[0][3] = fmt.Sprintf("out: %sb/s", DisplayUnit(iops))
				}
			}

			if myWidth > 80 {
				app.keyspaces.Rows[0] = []string{"hits", kv["keyspace_hits"]}
				app.keyspaces.Rows[1] = []string{"misess", kv["keyspace_misses"]}
				ui.Render(app.keyspaces)
			}

			kv, err = redis.Info()
			if err != nil {
				log.Printf("CPU Error : %s", err.Error())
			} else {
				sys, err := strconv.ParseFloat(kv["used_cpu_sys"], 64)
				if err != nil {
					log.Printf("%s %s", kv["used_cpu_sys"], err.Error())
				} else {
					user, err := strconv.ParseFloat(kv["used_cpu_user"], 64)
					if err != nil {
						log.Printf("%s %s", kv["used_cpu_sys"], err.Error())
					} else {
						if cpu == nil {
							cpu = monitor.NewCPU(sys, user)
						} else {
							s, u := cpu.Tick(sys, user)
							app.header.Rows[0][0] = fmt.Sprintf("s: %.1f%% u: %.1f%%", s, u)
						}
					}
				}
			}

			ui.Render(app.header)

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
				ui.Render(app.cmds, app.ips, app.graphBox)
			}
		}
	}()

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}

	return nil

}
