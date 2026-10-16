package main

import (
	"fmt"
	"io"
	"os"
)

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: releaserail <healthcheck>")
		return 1
	}

	switch args[0] {
	case "healthcheck":
		fmt.Fprintln(stdout, "ok")
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		return 1
	}
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
