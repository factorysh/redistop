package main

import (
	"context"
	"fmt"
	"os"

	"github.com/factorysh/redistop/monitor"
)

func main() {
	lines, err := monitor.Monitor(context.TODO(), os.Args[1], os.Args[2])
	if err != nil {
		panic(err)
	}
	for line := range lines {
		fmt.Println(line)
	}
}
