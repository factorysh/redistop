package monitor

import (
	"time"

	"github.com/mediocregopher/radix/v3"
)

type RedisServer struct {
	address  string
	password string
	pool     *radix.Pool
}

func Redis(address, password string) (*RedisServer, error) {
	r := &RedisServer{
		address:  address,
		password: password,
	}
	var err error
	r.pool, err = r.makePool()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RedisServer) makePool() (*radix.Pool, error) {
	opts := []radix.DialOpt{
		radix.DialConnectTimeout(2 * time.Second),
	}
	if r.password != "" {
		opts = append(opts, radix.DialAuthPass(r.password))
	}
	p, err := radix.NewPool("tcp", r.address, 1, radix.PoolConnFunc(func(network, addr string) (radix.Conn, error) {
		conn, err := radix.Dial("tcp", r.address, opts...)
		if err != nil {
			return nil, err
		}
		var pong string
		err = conn.Do(radix.Cmd(&pong, "PING"))
		if err != nil {
			return nil, err
		}
		return conn, nil
	}))
	if err != nil {
		return nil, err
	}
	return p, nil
}
