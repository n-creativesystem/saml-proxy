package main

import (
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		cmd := args[0]
		switch strings.ToLower(cmd) {
		case "ca":
			generateCA()
		case "server":
			generateSever()
		}
	}
}
