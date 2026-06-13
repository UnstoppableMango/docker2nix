package docker2nix

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	docker2nixv1alpha1 "github.com/unstoppablemango/docker2nix/pkg/docker2nix/v1alpha1"
)

func Generate(_ context.Context, req *docker2nixv1alpha1.GenerateRequest) (*docker2nixv1alpha1.GenerateResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
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

	var nix string
	switch req.GetFormat() {
	case docker2nixv1alpha1.Format_FORMAT_UNSPECIFIED, docker2nixv1alpha1.Format_FORMAT_DOCKER_TOOLS:
		nix = renderNix(stages)
	case docker2nixv1alpha1.Format_FORMAT_NIX2CONTAINER:
		nix = renderNix2Container(stages)
	default:
		return nil, fmt.Errorf("unsupported format: %v", req.GetFormat())
	}
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
			fmt.Fprintf(&b, "    %s = ", sanitizeNixIdent(s.Name))
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
	runs := extractRuns(s)

	b.WriteString("dockerTools.buildLayeredImage {\n")
	fmt.Fprintf(b, "%s  name = \"%s\";\n", prefix, name)
	fmt.Fprintf(b, "%s  tag = \"%s\";\n", prefix, tag)

	if len(runs) > 0 {
		fmt.Fprintf(b, "%s  runAsRoot = ''\n", prefix)
		for _, r := range runs {
			fmt.Fprintf(b, "%s    %s\n", prefix, nixEscapeIndented(r))
		}
		fmt.Fprintf(b, "%s  '';\n", prefix)
	}

	if len(envs) > 0 || len(cmds) > 0 {
		fmt.Fprintf(b, "%s  config = {\n", prefix)
		if len(envs) > 0 {
			fmt.Fprintf(b, "%s    Env = [\n", prefix)
			for _, e := range envs {
				fmt.Fprintf(b, "%s      \"%s\"\n", prefix, nixEscape(e))
			}
			fmt.Fprintf(b, "%s    ];\n", prefix)
		}
		if len(cmds) > 0 {
			fmt.Fprintf(b, "%s    Cmd = [\n", prefix)
			for _, c := range cmds {
				fmt.Fprintf(b, "%s      \"%s\"\n", prefix, nixEscape(c))
			}
			fmt.Fprintf(b, "%s    ];\n", prefix)
		}
		fmt.Fprintf(b, "%s  };\n", prefix)
	}

	fmt.Fprintf(b, "%s}", prefix)
}

func writeAttrs(b *strings.Builder, s instructions.Stage, prefix string) {
	name, tag := parseImage(s.BaseName)
	envs := extractEnvs(s)
	cmds := extractCmds(s)
	runs := extractRuns(s)

	fmt.Fprintf(b, "%s{\n", prefix)
	fmt.Fprintf(b, "%s  name = \"%s\";\n", prefix, name)
	fmt.Fprintf(b, "%s  tag = \"%s\";\n", prefix, tag)

	if len(runs) > 0 {
		fmt.Fprintf(b, "%s  runAsRoot = ''\n", prefix)
		for _, r := range runs {
			fmt.Fprintf(b, "%s    %s\n", prefix, nixEscapeIndented(r))
		}
		fmt.Fprintf(b, "%s  '';\n", prefix)
	}

	if len(envs) > 0 || len(cmds) > 0 {
		fmt.Fprintf(b, "%s  config = {\n", prefix)
		if len(envs) > 0 {
			fmt.Fprintf(b, "%s    Env = [\n", prefix)
			for _, e := range envs {
				fmt.Fprintf(b, "%s      \"%s\"\n", prefix, nixEscape(e))
			}
			fmt.Fprintf(b, "%s    ];\n", prefix)
		}
		if len(cmds) > 0 {
			fmt.Fprintf(b, "%s    Cmd = [\n", prefix)
			for _, c := range cmds {
				fmt.Fprintf(b, "%s      \"%s\"\n", prefix, nixEscape(c))
			}
			fmt.Fprintf(b, "%s    ];\n", prefix)
		}
		fmt.Fprintf(b, "%s  };\n", prefix)
	}

	fmt.Fprintf(b, "%s}", prefix)
}

func parseImage(baseName string) (name, tag string) {
	// Handle digest references (e.g. ubuntu@sha256:abc)
	if name, digest, ok := strings.Cut(baseName, "@"); ok {
		return name, digest
	}
	// Find the tag after the last colon that follows the last slash,
	// so registry:port in the host portion is not mistaken for a tag.
	lastSlash := strings.LastIndex(baseName, "/")
	rest := baseName[lastSlash+1:]
	if i := strings.LastIndex(rest, ":"); i >= 0 {
		return baseName[:lastSlash+1+i], rest[i+1:]
	}
	return baseName, "latest"
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
	// Docker uses the last CMD instruction; iterate all to find it.
	var result []string
	for _, cmd := range s.Commands {
		if c, ok := cmd.(*instructions.CmdCommand); ok {
			result = c.CmdLine
		}
	}
	return result
}

func extractRuns(s instructions.Stage) []string {
	var runs []string
	for _, cmd := range s.Commands {
		if r, ok := cmd.(*instructions.RunCommand); ok {
			if r.PrependShell {
				runs = append(runs, r.CmdLine[0])
			} else {
				runs = append(runs, strings.Join(r.CmdLine, " "))
			}
		}
	}
	return runs
}

func nixEscapeIndented(s string) string {
	s = strings.ReplaceAll(s, "''", "'''")
	s = strings.ReplaceAll(s, "${", "''${")
	return s
}

func nixEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "${", "\\${")
	return s
}

func renderNix2Container(stages []instructions.Stage) string {
	var b strings.Builder
	final := stages[len(stages)-1]

	var bindings []instructions.Stage
	for _, s := range stages[:len(stages)-1] {
		if s.Name != "" {
			bindings = append(bindings, s)
		}
	}

	if len(bindings) > 0 {
		b.WriteString("nix2container.buildImage\n")
		b.WriteString("  (let\n")
		for _, s := range bindings {
			fmt.Fprintf(&b, "    %s = ", sanitizeNixIdent(s.Name))
			writeStageNix2Container(&b, s, "    ")
			b.WriteString(";\n")
		}
		b.WriteString("  in\n")
		writeAttrsNix2Container(&b, final, "  ")
		b.WriteString(")\n")
	} else {
		writeStageNix2Container(&b, final, "")
		b.WriteString("\n")
	}

	return b.String()
}

func writeStageNix2Container(b *strings.Builder, s instructions.Stage, prefix string) {
	name, tag := parseImage(s.BaseName)
	envs := extractEnvs(s)
	cmds := extractCmds(s)
	runs := extractRuns(s)

	b.WriteString("nix2container.buildImage {\n")
	fmt.Fprintf(b, "%s  name = \"%s\";\n", prefix, name)
	fmt.Fprintf(b, "%s  tag = \"%s\";\n", prefix, tag)

	if len(runs) > 0 {
		fmt.Fprintf(b, "%s  layers = [\n", prefix)
		fmt.Fprintf(b, "%s    (nix2container.buildLayer {\n", prefix)
		fmt.Fprintf(b, "%s      deps = [(pkgs.runCommand \"run-layer\" {} ''\n", prefix)
		for _, r := range runs {
			fmt.Fprintf(b, "%s        %s\n", prefix, nixEscapeIndented(r))
		}
		fmt.Fprintf(b, "%s      '')];\n", prefix)
		fmt.Fprintf(b, "%s    })\n", prefix)
		fmt.Fprintf(b, "%s  ];\n", prefix)
	}

	if len(envs) > 0 || len(cmds) > 0 {
		fmt.Fprintf(b, "%s  config = {\n", prefix)
		if len(envs) > 0 {
			fmt.Fprintf(b, "%s    env = [\n", prefix)
			for _, e := range envs {
				fmt.Fprintf(b, "%s      \"%s\"\n", prefix, nixEscape(e))
			}
			fmt.Fprintf(b, "%s    ];\n", prefix)
		}
		if len(cmds) > 0 {
			fmt.Fprintf(b, "%s    cmd = [\n", prefix)
			for _, c := range cmds {
				fmt.Fprintf(b, "%s      \"%s\"\n", prefix, nixEscape(c))
			}
			fmt.Fprintf(b, "%s    ];\n", prefix)
		}
		fmt.Fprintf(b, "%s  };\n", prefix)
	}

	fmt.Fprintf(b, "%s}", prefix)
}

func writeAttrsNix2Container(b *strings.Builder, s instructions.Stage, prefix string) {
	name, tag := parseImage(s.BaseName)
	envs := extractEnvs(s)
	cmds := extractCmds(s)
	runs := extractRuns(s)

	fmt.Fprintf(b, "%s{\n", prefix)
	fmt.Fprintf(b, "%s  name = \"%s\";\n", prefix, name)
	fmt.Fprintf(b, "%s  tag = \"%s\";\n", prefix, tag)

	if len(runs) > 0 {
		fmt.Fprintf(b, "%s  layers = [\n", prefix)
		fmt.Fprintf(b, "%s    (nix2container.buildLayer {\n", prefix)
		fmt.Fprintf(b, "%s      deps = [(pkgs.runCommand \"run-layer\" {} ''\n", prefix)
		for _, r := range runs {
			fmt.Fprintf(b, "%s        %s\n", prefix, nixEscapeIndented(r))
		}
		fmt.Fprintf(b, "%s      '')];\n", prefix)
		fmt.Fprintf(b, "%s    })\n", prefix)
		fmt.Fprintf(b, "%s  ];\n", prefix)
	}

	if len(envs) > 0 || len(cmds) > 0 {
		fmt.Fprintf(b, "%s  config = {\n", prefix)
		if len(envs) > 0 {
			fmt.Fprintf(b, "%s    env = [\n", prefix)
			for _, e := range envs {
				fmt.Fprintf(b, "%s      \"%s\"\n", prefix, nixEscape(e))
			}
			fmt.Fprintf(b, "%s    ];\n", prefix)
		}
		if len(cmds) > 0 {
			fmt.Fprintf(b, "%s    cmd = [\n", prefix)
			for _, c := range cmds {
				fmt.Fprintf(b, "%s      \"%s\"\n", prefix, nixEscape(c))
			}
			fmt.Fprintf(b, "%s    ];\n", prefix)
		}
		fmt.Fprintf(b, "%s  };\n", prefix)
	}

	fmt.Fprintf(b, "%s}", prefix)
}

func sanitizeNixIdent(s string) string {
	var b strings.Builder
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_':
			b.WriteRune(r)
		case r >= '0' && r <= '9' && i > 0:
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}
