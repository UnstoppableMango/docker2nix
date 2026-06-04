package e2e_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var binary string

var _ = BeforeSuite(func() {
	var err error
	binary, err = gexec.Build("github.com/unstoppablemango/docker2nix")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func runGenerate(ctx context.Context, stdin string, args ...string) *gexec.Session {
	GinkgoHelper()
	cmd := exec.CommandContext(ctx, binary, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

// nixParse runs the given Nix expression through nix-instantiate --parse to
// verify it is syntactically valid. The expression is wrapped in a lambda so
// that free variables (dockerTools, nix2container) don't cause errors.
func nixParse(ctx context.Context, expr string) *gexec.Session {
	GinkgoHelper()
	wrapped := "{ dockerTools ? null, nix2container ? null }:\n(" + expr + ")"
	cmd := exec.CommandContext(ctx, "nix-instantiate", "--parse", "-")
	cmd.Stdin = strings.NewReader(wrapped)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

var _ = Describe("docker2nix", func() {
	It("should handle multi-stage Dockerfiles", func(ctx context.Context) {
		dockerfile := `FROM golang:1.23 AS builder
WORKDIR /src
COPY . .
RUN go build -o /app/main .

FROM ubuntu:24.04
COPY --from=builder /app/main /app/main
CMD ["/app/main"]
`

		session := runGenerate(ctx, dockerfile)

		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out).To(gbytes.Say(`dockerTools\.buildLayeredImage`))
		Expect(session.Out).To(gbytes.Say("builder"))
		Expect(session.Out).To(gbytes.Say(`name = "ubuntu"`))
		parse := nixParse(ctx, string(session.Out.Contents()))
		Eventually(parse).Should(gexec.Exit(0))
	})

	It("should accept a Dockerfile path as argument", func(ctx context.Context) {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "Dockerfile")
		Expect(os.WriteFile(path, []byte("FROM ubuntu:24.04\n"), 0o644)).To(Succeed())

		session := runGenerate(ctx, "", path)

		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out).To(gbytes.Say(`dockerTools\.buildLayeredImage`))
		Expect(session.Out).To(gbytes.Say(`name = "ubuntu"`))
		Expect(session.Out).To(gbytes.Say(`tag = "24\.04"`))
		parse := nixParse(ctx, string(session.Out.Contents()))
		Eventually(parse).Should(gexec.Exit(0))
	})

	It("should return an error for an invalid Dockerfile", func(ctx context.Context) {
		session := runGenerate(ctx, "NOT A VALID DOCKERFILE INSTRUCTION\n")

		Eventually(session).Should(gexec.Exit(1))

		Expect(session.Err).To(gbytes.Say(".+"))
	})

	It("should emit nix2container.buildImage with --format nix2container", func(ctx context.Context) {
		session := runGenerate(ctx, "FROM ubuntu:24.04\n", "--format", "nix2container")

		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out).To(gbytes.Say(`nix2container\.buildImage`))
		Expect(session.Out).To(gbytes.Say(`name = "ubuntu"`))
		Expect(session.Out).To(gbytes.Say(`tag = "24\.04"`))
		parse := nixParse(ctx, string(session.Out.Contents()))
		Eventually(parse).Should(gexec.Exit(0))
	})

	It("should default to dockerTools.buildLayeredImage without --format flag", func(ctx context.Context) {
		session := runGenerate(ctx, "FROM ubuntu:24.04\n")

		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out).To(gbytes.Say(`dockerTools\.buildLayeredImage`))
		parse := nixParse(ctx, string(session.Out.Contents()))
		Eventually(parse).Should(gexec.Exit(0))
	})

	It("should produce valid Nix with ENV and CMD", func(ctx context.Context) {
		session := runGenerate(ctx, `FROM ubuntu:24.04
ENV FOO=bar BAZ=qux
CMD ["/app/main", "--flag"]
`)

		Eventually(session).Should(gexec.Exit(0))
		parse := nixParse(ctx, string(session.Out.Contents()))
		Eventually(parse).Should(gexec.Exit(0))
	})

	It("should produce valid Nix for nix2container with ENV and CMD", func(ctx context.Context) {
		session := runGenerate(ctx, `FROM ubuntu:24.04
ENV FOO=bar
CMD ["/app/main"]
`, "--format", "nix2container")

		Eventually(session).Should(gexec.Exit(0))
		parse := nixParse(ctx, string(session.Out.Contents()))
		Eventually(parse).Should(gexec.Exit(0))
	})

	It("should produce valid Nix for nix2container multi-stage", func(ctx context.Context) {
		session := runGenerate(ctx, `FROM golang:1.23 AS builder
WORKDIR /src

FROM ubuntu:24.04
CMD ["/app/main"]
`, "--format", "nix2container")

		Eventually(session).Should(gexec.Exit(0))
		parse := nixParse(ctx, string(session.Out.Contents()))
		Eventually(parse).Should(gexec.Exit(0))
	})
})
