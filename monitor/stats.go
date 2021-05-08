package monitor

import (
	"github.com/mediocregopher/radix/v3"
)

func (r *RedisServer) Stats() (map[string]string, error) {
	var stats string
	err := r.pool.Do(radix.Cmd(&stats, "INFO", "STATS"))
	if err != nil {
		return nil, err
	}
	return BulkTable(stats)
}
