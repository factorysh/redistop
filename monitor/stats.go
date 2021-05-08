package monitor

import (
	"strings"

	"github.com/mediocregopher/radix/v3"
)

func (r *RedisServer) Stats() (map[string]string, error) {
	s := make(map[string]string)
	var stats string
	err := r.pool.Do(radix.Cmd(&stats, "INFO", "STATS"))
	if err != nil {
		return nil, err
	}
	lines := strings.Split(stats, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		values := strings.Split(line, ":")
		if len(values) > 1 {
			s[values[0]] = values[1][:len(values[1])-1]
		}
	}
	return s, nil
}
