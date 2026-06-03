package docker2nix

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	docker2nixv1alpha1 "github.com/unstoppablemango/docker2nix/pkg/docker2nix/v1alpha1"
)

func Generate(ctx context.Context, req *docker2nixv1alpha1.GenerateRequest) (*docker2nixv1alpha1.GenerateResponse, error) {
	result, err := parser.Parse(strings.NewReader(req.GetDockerfile()))
	if err != nil {
		return nil, err
	}

	stages, _, err := instructions.Parse(result.AST, nil)
	if err != nil {
		return nil, err
	}

	if len(stages) == 0 {
		return nil, fmt.Errorf("no stages in Dockerfile")
	}

	nix := renderNix(stages)
	resp := docker2nixv1alpha1.GenerateResponse_builder{Nix: &nix}
	return resp.Build(), nil
}

func renderNix(stages []instructions.Stage) string {
	var b strings.Builder
	final := stages[len(stages)-1]

	var bindings []instructions.Stage
	for _, s := range stages[:len(stages)-1] {
		if s.Name != "" {
			bindings = append(bindings, s)
		}
	}

	if len(bindings) > 0 {
		b.WriteString("dockerTools.buildLayeredImage\n")
		b.WriteString("  (let\n")
		for _, s := range bindings {
			b.WriteString("    " + s.Name + " = ")
			writeStage(&b, s, "    ")
			b.WriteString(";\n")
		}
		b.WriteString("  in\n")
		writeAttrs(&b, final, "  ")
		b.WriteString(")\n")
	} else {
		writeStage(&b, final, "")
		b.WriteString("\n")
	}

	return b.String()
}

func writeStage(b *strings.Builder, s instructions.Stage, prefix string) {
	name, tag := parseImage(s.BaseName)
	envs := extractEnvs(s)
	cmds := extractCmds(s)

	b.WriteString("dockerTools.buildLayeredImage {\n")
	b.WriteString(prefix + `  name = "` + name + `";` + "\n")
	b.WriteString(prefix + `  tag = "` + tag + `";` + "\n")

	if len(envs) > 0 || len(cmds) > 0 {
		b.WriteString(prefix + "  config = {\n")
		if len(envs) > 0 {
			b.WriteString(prefix + "    Env = [\n")
			for _, e := range envs {
				b.WriteString(prefix + `      "` + e + `";` + "\n")
			}
			b.WriteString(prefix + "    ];\n")
		}
		if len(cmds) > 0 {
			b.WriteString(prefix + "    Cmd = [\n")
			for _, c := range cmds {
				b.WriteString(prefix + `      "` + c + `";` + "\n")
			}
			b.WriteString(prefix + "    ];\n")
		}
		b.WriteString(prefix + "  };\n")
	}

	b.WriteString(prefix + "}")
}

func writeAttrs(b *strings.Builder, s instructions.Stage, prefix string) {
	name, tag := parseImage(s.BaseName)
	envs := extractEnvs(s)
	cmds := extractCmds(s)

	b.WriteString(prefix + "{\n")
	b.WriteString(prefix + `  name = "` + name + `";` + "\n")
	b.WriteString(prefix + `  tag = "` + tag + `";` + "\n")

	if len(envs) > 0 || len(cmds) > 0 {
		b.WriteString(prefix + "  config = {\n")
		if len(envs) > 0 {
			b.WriteString(prefix + "    Env = [\n")
			for _, e := range envs {
				b.WriteString(prefix + `      "` + e + `";` + "\n")
			}
			b.WriteString(prefix + "    ];\n")
		}
		if len(cmds) > 0 {
			b.WriteString(prefix + "    Cmd = [\n")
			for _, c := range cmds {
				b.WriteString(prefix + `      "` + c + `";` + "\n")
			}
			b.WriteString(prefix + "    ];\n")
		}
		b.WriteString(prefix + "  };\n")
	}

	b.WriteString(prefix + "}")
}

func parseImage(baseName string) (name, tag string) {
	parts := strings.SplitN(baseName, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], "latest"
}

func extractEnvs(s instructions.Stage) []string {
	var envs []string
	for _, cmd := range s.Commands {
		if e, ok := cmd.(*instructions.EnvCommand); ok {
			for _, kv := range e.Env {
				envs = append(envs, kv.Key+"="+kv.Value)
			}
		}
	}
	return envs
}

func extractCmds(s instructions.Stage) []string {
	for _, cmd := range s.Commands {
		if c, ok := cmd.(*instructions.CmdCommand); ok {
			return c.CmdLine
		}
	}
	return nil
}
