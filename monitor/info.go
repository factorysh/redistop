package monitor

import "github.com/mediocregopher/radix/v3"

func (r *RedisServer) Info() (map[string]string, error) {
	var bulk string
	err := r.pool.Do(radix.Cmd(&bulk, "INFO"))
	if err != nil {
		return nil, err
	}
	return BulkTable(bulk)
}
