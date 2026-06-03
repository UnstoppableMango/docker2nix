package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/unmango/go/cli"
	docker2nix "github.com/unstoppablemango/docker2nix/pkg"
)

var generateCmd = &cobra.Command{
	Use:   "generate [Dockerfile]",
	Short: "Generate a Nix expression from a Dockerfile",
	Long:  `Generate a Nix expression from a Dockerfile.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var r io.Reader
		if len(args) == 1 {
			f, err := os.Open(args[0])
			if err != nil {
				cli.Fail(err)
			}
			defer f.Close()
			r = f
		} else {
			r = cmd.InOrStdin()
		}

		content, err := io.ReadAll(r)
		if err != nil {
			cli.Fail(err)
		}

		s := string(content)
		req := docker2nix.GenerateRequest_builder{Dockerfile: &s}
		resp, err := docker2nix.Generate(cmd.Context(), req.Build())
		if err != nil {
			cli.Fail(err)
		}

		fmt.Fprint(cmd.OutOrStdout(), resp.GetNix())
	},
}

func init() {
	root.AddCommand(generateCmd)
}
