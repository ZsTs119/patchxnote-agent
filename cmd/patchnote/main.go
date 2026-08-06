package main

import (
	"fmt"
	"os"

	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
