package main

import (
	"fmt"
	"log"
	"os"

	"github.com/factorysh/redistop/cli"
	"github.com/factorysh/redistop/version"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			fmt.Printf(`RedisTop %s
top for Redis, group by command and client IP

Usage:
  redistop [[localhost:6379] password]

You can set REDISTOP_PASSWORD
`, version.Version())
			return
		}
	}
	host := "localhost:6379"
	if len(os.Args) > 1 {
		host = os.Args[1]
	}
	var password string
	if len(os.Args) > 2 {
		password = os.Args[2]
	}
	p := os.Getenv("REDISTOP_PASSWORD")
	if p != "" {
		password = p
	}

	log.Fatal(cli.Top(host, password))

}
