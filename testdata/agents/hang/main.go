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

	// Send one heartbeat so Leopold knows we started
	fmt.Println(`{"type":"heartbeat","v":1,"id":"test","state":"running","tool":"bash","detail":"starting","rss_mb":10,"tokens_in":0,"tokens_out":0,"elapsed_s":0}`)

	// Then go silent. Leopold's heartbeat timeout should kill the agent.
	time.Sleep(10 * time.Minute)
}
