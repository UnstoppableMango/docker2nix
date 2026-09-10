package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/unmango/go/cli"
	docker2nix "github.com/unstoppablemango/docker2nix/pkg"
)

var rootFormat string

var root = &cobra.Command{
	Use:   "docker2nix [Dockerfile]",
	Short: "Convert a Dockerfile to a Nix expression",
	Long:  `docker2nix converts Dockerfiles to Nix expressions.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var r io.Reader
		if len(args) == 1 {
			f, err := os.Open(args[0])
			if err != nil {
				cli.Fail(err)
			}
			defer func() { _ = f.Close() }()
			r = f
		} else {
			r = cmd.InOrStdin()
		}

		content, err := io.ReadAll(r)
		if err != nil {
			cli.Fail(err)
		}

		var format docker2nix.Format
		switch rootFormat {
		case "docker-tools":
			format = docker2nix.FORMAT_DOCKER_TOOLS
		case "nix2container":
			format = docker2nix.FORMAT_NIX2CONTAINER
		default:
			cli.Fail(fmt.Errorf("unsupported format: %q", rootFormat))
		}

		s := string(content)
		req := docker2nix.GenerateRequest_builder{Dockerfile: &s, Format: &format}
		resp, err := docker2nix.Generate(cmd.Context(), req.Build())
		if err != nil {
			cli.Fail(err)
		}

		if _, err := fmt.Fprint(cmd.OutOrStdout(), resp.GetNix()); err != nil {
			cli.Fail(err)
		}
	},
}

func init() {
	root.Flags().StringVar(&rootFormat, "format", "docker-tools", "Output format: docker-tools, nix2container")
}

// Execute runs the root command.
func Execute() error {
	return root.Execute()
}
