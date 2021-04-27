package monitor

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func RedisConn(address, password string) (net.Conn, *bufio.Reader, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, nil, err
	}

	_, err = fmt.Fprintln(conn, "PING")
	if err != nil {
		return nil, nil, err
	}
	reader := bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		return nil, nil, err
	}
	if strings.HasPrefix(resp, "-NOAUTH") {
		fmt.Fprintf(conn, "AUTH %s\n", password)
		resp, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, err
		}
		if !strings.HasPrefix(resp, "+OK") {
			return nil, nil, fmt.Errorf("AUTH not ok : %s", resp)
		}
	}
	return conn, reader, nil
}
