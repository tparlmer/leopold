package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Read init message
	scanner.Scan()

	// Read task message
	scanner.Scan()

	// Send a heartbeat
	fmt.Println(`{"type":"heartbeat","v":1,"id":"test","state":"running","tool":"bash","detail":"working","rss_mb":10,"tokens_in":100,"tokens_out":50,"elapsed_s":1}`)

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Send another heartbeat
	fmt.Println(`{"type":"heartbeat","v":1,"id":"test","state":"running","tool":"file_write","detail":"writing code","rss_mb":12,"tokens_in":500,"tokens_out":200,"elapsed_s":2}`)

	// Complete successfully
	fmt.Println(`{"type":"complete","v":1,"id":"test","state":"done","summary":"task completed","files_changed":["main.go"],"tokens_in":1000,"tokens_out":400,"elapsed_s":3}`)
}
