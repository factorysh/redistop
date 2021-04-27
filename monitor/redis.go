package monitor

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type RedisConn struct {
	conn   net.Conn
	reader *bufio.Reader
}

type RedisServer struct {
	address  string
	password string
}

func Redis(address, password string) *RedisServer {
	return &RedisServer{
		address:  address,
		password: password,
	}
}

func (r *RedisServer) Conn() (*RedisConn, error) {
	conn, err := net.Dial("tcp", r.address)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintln(conn, "PING")
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(resp, "-NOAUTH") {
		fmt.Fprintf(conn, "AUTH %s\n", r.password)
		resp, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(resp, "+OK") {
			return nil, fmt.Errorf("AUTH not ok : %s", resp)
		}
	}
	return &RedisConn{
		conn:   conn,
		reader: reader}, nil
}
