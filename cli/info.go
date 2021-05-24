package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/factorysh/redistop/monitor"
	ui "github.com/gizak/termui/v3"
)

func (a *App) InfoLoop() {

	go func() {
		var cpu *monitor.CPU
		for {
			kv, err := a.redis.Info()
			if err != nil {
				a.log.Printf("Info Error : %s", err.Error())
			} else {
				a.ui.app.QueueUpdate(func() {
					if kv["instantaneous_ops_per_sec"] == "" {
						a.ui.header.GetCell(0, 1).Text = "️0 op"
					} else {
						ops, err := strconv.ParseFloat(kv["instantaneous_ops_per_sec"], 32)
						if err != nil {
							a.log.Printf("Float parse error: %s %s", kv["instantaneous_ops_per_sec"], err)
							a.ui.header.GetCell(0, 1).Text = "☠️"
						} else {
							a.ui.header.GetCell(0, 1).Text = fmt.Sprintf("%s ops/s", DisplayUnit(ops))
						}
					}
					iips, err := strconv.ParseFloat(kv["instantaneous_input_kbps"], 32)
					if err != nil {
						a.log.Printf("Float parse error: %s %s", kv["instantaneous_input_kbps"], err)
						a.ui.header.GetCell(0, 2).Text = "☠️"
					} else {
						a.ui.header.GetCell(0, 2).Text = fmt.Sprintf("in: %sb/s", DisplayUnit(iips))
					}
					iops, err := strconv.ParseFloat(kv["instantaneous_output_kbps"], 32)
					if err != nil {
						a.log.Printf("Float parse error: %s %s", kv["instantaneous_output_kbps"], err)
						a.ui.header.GetCell(0, 3).Text = "☠️"
					} else {
						a.ui.header.GetCell(0, 3).Text = fmt.Sprintf("out: %sb/s", DisplayUnit(iops))
					}
				})

				sys, err := strconv.ParseFloat(kv["used_cpu_sys"], 64)
				if err != nil {
					a.log.Printf("%s %s", kv["used_cpu_sys"], err.Error())
				} else {
					user, err := strconv.ParseFloat(kv["used_cpu_user"], 64)
					if err != nil {
						a.log.Printf("%s %s", kv["used_cpu_sys"], err.Error())
					} else {
						if cpu == nil {
							cpu = monitor.NewCPU(sys, user)
						} else {
							s, u := cpu.Tick(sys, user)
							a.ui.app.QueueUpdateDraw(func() {
								a.ui.header.GetCell(0, 0).Text = fmt.Sprintf("s: %.1f%% u: %.1f%%", s, u)
							})
						}
					}
				}

				a.ui.keyspaces.Rows[0] = []string{
					small("hits", kv["keyspace_hits"]),
					small("misses", kv["keyspace_misses"]),
				}

				a.ui.pubsub.Rows[0] = []string{
					small("channels", kv["pubsub_channels"]),
					small("patterns", kv["pubsub_patterns"]),
				}

				a.ui.clients.Rows[0] = []string{
					small("connected", kv["connected_clients"]),
					small("blocked", kv["blocked_clients"]),
				}
				a.ui.clients.Rows[1] = []string{
					small("tracking", kv["tracking_clients"]),
					"",
				}

				a.ui.persistence.Rows[0] = []string{"status", ""}
				if kv["loading"] == "1" {
					a.ui.persistence.Rows[0][1] = "loading"
				} else {
					if kv["rdb_bgsave_in_progress"] == "1" {
						a.ui.persistence.Rows[0][1] = "rdb_bgsave_in_progress"
					} else {
						if kv["aof_rewrite_in_progress"] == "1" {
							a.ui.persistence.Rows[0][1] = "aof_rewrite_in_progress"
						}
					}
				}
				a.ui.persistence.Rows[1] = []string{"rdb_changes_since_last_save", kv["rdb_changes_since_last_save"]}
				a.ui.persistence.Rows[2] = []string{"rdb_last_save_time", kv["rdb_last_save_time"]}

				if a.ui.myWidth > 80 {
					ui.Render(a.ui.keyspaces, a.ui.pubsub, a.ui.clients, a.ui.persistence)
				}
			}

			time.Sleep(time.Second)
		}
	}()

}

func small(left, right string) string {
	return fmt.Sprintf("%s%*s", left, 18-len(left), right)
}
