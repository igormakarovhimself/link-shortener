package main

import "os"

func main() {
	os.Exit(1)
	panic("fatal")
}

func helper() {
	os.Exit(1)     // want "os.Exit call"
	panic("fatal") // want "panic call"
}
