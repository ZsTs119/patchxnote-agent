package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "setup":
		fmt.Println("patchnote setup: not implemented yet")
	case "login":
		fmt.Println("patchnote login: not implemented yet")
	case "mcp":
		fmt.Println("patchnote mcp: not implemented yet")
	default:
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Println(`PatchNote Agent

Usage:
  patchnote setup
  patchnote login
  patchnote mcp`)
}
