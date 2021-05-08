package monitor

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type Line struct {
	ts      float32
	n       int
	IP      string
	port    int
	Command string
}

func (r *RedisServer) Monitor(ctx context.Context) (chan Line, error) {
	// +1619454979.381488 [1 172.29.1.2:57676] "brpop"
	line := regexp.MustCompile(`^\+(\d+\.\d+) \[(\d+) ([\d.]+):(\d+)] "(.*?)"`)

	conn, err := net.Dial("tcp", r.address)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)
	if r.password != "" {
		fmt.Fprintf(conn, "AUTH %s\n", r.password)
		resp, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(resp, "+OK") {
			return nil, fmt.Errorf("auth failed, bad password")
		}
	}
	_, err = fmt.Fprintln(conn, "MONITOR")
	if err != nil {
		return nil, err
	}
	lines := make(chan Line)
	go func() {
		for {
			resp, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("monitor can't read", err)
				break
			}
			l := line.FindStringSubmatch(resp)
			if len(l) != 6 {
				continue
			}
			ts, err := strconv.ParseFloat(l[1], 32)
			if err != nil {
				fmt.Println("monitor", l, err)
				break
			}
			n, err := strconv.Atoi(l[2])
			if err != nil {
				fmt.Println("monitor", l, err)
				break
			}
			port, err := strconv.Atoi(l[4])
			if err != nil {
				fmt.Println("monitor", l, err)
				break
			}
			lines <- Line{
				ts:      float32(ts),
				n:       n,
				IP:      l[3],
				port:    port,
				Command: strings.ToUpper(l[5]),
			}
		}
	}()

	return lines, nil
}
