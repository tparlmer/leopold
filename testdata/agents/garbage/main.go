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

	// Spew non-JSON garbage
	fmt.Println("this is not json")
	fmt.Println("neither is this {{{")
	fmt.Println(`{"type": but malformed`)
	fmt.Println("SEGFAULT CORE DUMPED just kidding")

	time.Sleep(50 * time.Millisecond)

	// Exit without sending a complete message
	os.Exit(0)
}
