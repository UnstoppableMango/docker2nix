package cmd

import "github.com/spf13/cobra"

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a Nix expression from a Dockerfile",
	Long:  `Generate a Nix expression from a Dockerfile.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	root.AddCommand(generateCmd)
}
