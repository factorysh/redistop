package cli

import (
	"fmt"
	_log "log"
	"strconv"
	"time"

	"github.com/factorysh/redistop/monitor"
	ui "github.com/gizak/termui/v3"
)

const freq = 2 // Stats per commands and per IPs, every freq seconds

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

	app := NewApp()

	infos, err := redis.Info()
	if err != nil {
		return err
	}

	app.header.Title = fmt.Sprintf("Redis Top -[ v%s/%s pid: %s port: %s hz: %s uptime: %sd ]",
		infos["redis_version"],
		infos["multiplexing_api"],
		infos["process_id"],
		infos["tcp_port"],
		infos["hz"],
		infos["uptime_in_days"],
	)
	ui.Render(app.header)
	ui.Render(app.graphBox)
	ui.Render(app.errorPanel)

	log := &Logger{
		block: app.errorPanel,
	}

	if app.myWidth > 80 {

		go func() {
			for {
				m, err := redis.Memory()
				if err != nil {
					log.Printf("Memory Error : %s", err.Error())
				} else {
					app.header.Rows[0][4] = fmt.Sprintf("keys: %d", m.KeysCount)
					app.header.Rows[0][5] = fmt.Sprintf("mem: %s", DisplayUnit(float64(m.PeakAllocated)))
					app.memories.Rows = m.Table()
				}
				kv, err := redis.Info()
				if err != nil {
					log.Printf("Info Memory Error : %s", err.Error())
				} else {
					app.memories.Title = fmt.Sprintf("Memory [ %s ]", kv["maxmemory_policy"])
				}

				if len(app.memories.Rows) > 0 && len(app.memories.Rows[0]) > 0 {
					ui.Render(app.memories)
				}
				time.Sleep(5 * time.Second)
			}
		}()

	}

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

			if app.myWidth > 80 {
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

	MonitorLoop(redis, app, log)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}

	return nil
}
