package app

import "fmt"

func printBanner() {
	fmt.Println("===================================")
	fmt.Println(" XDL | X/Twitter Media Downloader ")
	fmt.Println("===================================")
}

func statusf(format string, args ...any) {
	fmt.Printf("[xdl] "+format+"\n", args...)
}
