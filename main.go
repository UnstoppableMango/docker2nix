package main

import (
	"github.com/unmango/go/cli"
	"github.com/unstoppablemango/docker2nix/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		cli.Fail(err)
	}
}
