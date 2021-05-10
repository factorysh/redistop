package monitor

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Line struct {
	ts      float32
	n       int
	IP      string
	port    int
	Command string
}

func (r *RedisServer) Monitor(ctx context.Context) (chan Line, chan error) {
	// +1619454979.381488 [1 172.29.1.2:57676] "brpop"
	line := regexp.MustCompile(`^\+(\d+\.\d+) \[(\d+) ([\d.]+):(\d+)] "(.*?)"`)

	lines := make(chan Line)
	errors := make(chan error)

	go func() {
		for {
			conn, err := net.Dial("tcp", r.address)
			if err != nil {
				errors <- err
				time.Sleep(5 * time.Second)
				continue
			}
			reader := bufio.NewReader(conn)
			if r.password != "" {
				fmt.Fprintf(conn, "AUTH %s\n", r.password)
				resp, err := reader.ReadString('\n')
				if err != nil {
					errors <- err
					continue
				}
				if !strings.HasPrefix(resp, "+OK") {
					errors <- fmt.Errorf("auth failed, bad password")
					continue
				}
			}
			_, err = fmt.Fprintln(conn, "MONITOR")
			if err != nil {
				errors <- err
				continue
			}
			for {
				resp, err := reader.ReadString('\n')
				if err != nil {
					errors <- fmt.Errorf("monitor can't read %v", err)
					break
				}
				l := line.FindStringSubmatch(resp)
				if len(l) != 6 {
					continue
				}
				ts, err := strconv.ParseFloat(l[1], 32)
				if err != nil {
					errors <- fmt.Errorf("monitor %v %v", l, err)
					break
				}
				n, err := strconv.Atoi(l[2])
				if err != nil {
					errors <- fmt.Errorf("monitor %v %v", l, err)
					break
				}
				port, err := strconv.Atoi(l[4])
				if err != nil {
					errors <- fmt.Errorf("monitor %v %v", l, err)
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
		}
	}()

	return lines, errors
}
