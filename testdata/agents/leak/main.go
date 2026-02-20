package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

// leak agent eats memory until killed
func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Read init message
	scanner.Scan()

	// Read task message
	scanner.Scan()

	// Allocate memory in a loop, send heartbeats so we don't
	// killed by heartbeat timeout before RSS enforcement kicks in
	var sink [][]byte
	for i := 0; ; i++ {
		// Allocate 1MB per iteration
		chunk := make([]byte, 1024*1024)
		sink = append(sink, chunk)

		fmt.Fprintf(os.Stdout, `{"type":"heartbeat","v":1,"id":"test","state":"running","tool":"bash","detail":"leaking","rss_mb":%d,"tokens_in":0,"tokens_out":0,"elapsed_s":%d}`+"\n", len(sink), i)

		time.Sleep(50 * time.Millisecond)
	}

	// Keep the compiler happy - unreachable but prevents "sink unused" error
	_ = sink
}
