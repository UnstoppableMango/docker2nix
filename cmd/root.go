package cmd

import "github.com/spf13/cobra"

var root = &cobra.Command{
	Use:   "docker2nix",
	Short: "A tool to convert Dockerfiles to Nix expressions",
	Long:  `docker2nix is a tool to convert Dockerfiles to Nix expressions.`,
}

func Execute() error {
	return root.Execute()
}
