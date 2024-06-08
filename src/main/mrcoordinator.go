package main

//
// start the coordinator process, which is implemented
// in ../mr/coordinator.go
//
// go run mrcoordinator.go 3 pg*.txt
//
//

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"6.5840/mr"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: mrcoordinator inputfiles...\n")
		os.Exit(1)
	}

	n_reducer, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid number of reducer: %s\n", os.Args[1])
		os.Exit(1)
	}
	m := mr.MakeCoordinator(os.Args[2:], n_reducer)
	for m.Done() == false {
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)
}
