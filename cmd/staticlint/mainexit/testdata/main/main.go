package main

import "os"

func main() {
	os.Exit(9) // want "main must not contain os.Exit call"
}
