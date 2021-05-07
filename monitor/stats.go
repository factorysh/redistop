package monitor

import (
	"fmt"
	"strings"
)

type Stats struct {
	redisConn *RedisConn
}

func (r *RedisServer) Stats() (*Stats, error) {
	c, err := r.Conn()
	if err != nil {
		return nil, err
	}
	return &Stats{c}, nil
}

func (s *Stats) Values() (map[string]string, error) {
	_, err := fmt.Fprintln(s.redisConn.conn, "INFO STATS")
	if err != nil {
		return nil, err
	}
	r := make(map[string]string)
	for {
		resp, err := s.redisConn.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if resp == "\r\n" {
			break
		}
		kv := strings.Split(resp, ":")
		if len(kv) > 1 {
			r[kv[0]] = strings.Trim(kv[1], " \n\r")
		}
	}
	return r, nil
}
