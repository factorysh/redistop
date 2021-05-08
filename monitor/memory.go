package monitor

import (
	"fmt"
	"strconv"

	"github.com/mediocregopher/radix/v3"
)

type MemoryStats struct {
	PeakAllocated      int64
	DatasetBytes       int64
	KeysCount          int64
	Fragmentation      float64
	ReplicationBacklog int64
}

func (r *RedisServer) Memory() (*MemoryStats, error) {
	m := &MemoryStats{}
	var mem map[string]interface{}
	err := r.pool.Do(radix.Cmd(&mem, "MEMORY", "STATS"))
	if err != nil {
		return nil, err
	}
	for k, v := range mem {
		//fmt.Printf("%s => %v\n", k, v)
		switch k {
		case "peak.allocated":
			vv, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("not an int : %v", v)
			}
			m.PeakAllocated = vv
		case "dataset.bytes":
			vv, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("not an int : %v", v)
			}
			m.DatasetBytes = vv
		case "keys.count":
			vv, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("not an int : %v", v)
			}
			m.KeysCount = vv
		case "fragmentation":
			vv, ok := v.([]byte)
			if !ok {
				return nil, fmt.Errorf("not a string : %v", v)
			}
			vvv, err := strconv.ParseFloat(string(vv), 64)
			if err != nil {
				return nil, err
			}
			m.Fragmentation = vvv
		case "replication.backlog":
			vv, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("not an int : %v", v)
			}
			m.ReplicationBacklog = vv
		}
	}
	return m, err
}

func (m *MemoryStats) Table() [][]string {
	return [][]string{
		{"peak allocated", fmt.Sprintf("%d", m.PeakAllocated)},
		{"dataset", fmt.Sprintf("%d bytes", m.DatasetBytes)},
		{"keys", fmt.Sprintf("%d", m.KeysCount)},
		{"fragmentation", fmt.Sprintf("%.2f", m.Fragmentation)},
		{"repl.backlog", fmt.Sprintf("%d", m.ReplicationBacklog)},
	}
}
