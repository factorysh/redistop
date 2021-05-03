package main

import (
	"fmt"
	"log"
	"os"

	"github.com/factorysh/redistop/cli"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			fmt.Println(`RedisTop top for Redis, group by command and client IP

Redis is local, without auth:

redistop

Redis is somewhere :

  redistop localhost:6379

Redis has a password:

  redistop localhost:6379 password

`)
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

	log.Fatal(cli.Top(host, password))

}
