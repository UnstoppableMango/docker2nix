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

var _ = Describe("docker2nix generate", func() {
	runGenerate := func(ctx context.Context, stdin string, args ...string) *gexec.Session {
		cmd := exec.CommandContext(ctx, binary, append([]string{"generate"}, args...)...)
		if stdin != "" {
			cmd.Stdin = strings.NewReader(stdin)
		}
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		return session
	}

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
		Expect(session.Out).To(gbytes.Say("builder"))
		Expect(session.Out).To(gbytes.Say("ubuntu"))
	})

	It("should accept a Dockerfile path as argument", func(ctx context.Context) {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "Dockerfile")
		Expect(os.WriteFile(path, []byte("FROM ubuntu:24.04\n"), 0o644)).To(Succeed())

		session := runGenerate(ctx, "", path)

		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out).To(gbytes.Say("ubuntu"))
	})

	It("should return an error for an invalid Dockerfile", func(ctx context.Context) {
		session := runGenerate(ctx, "NOT A VALID DOCKERFILE INSTRUCTION\n")

		Eventually(session).ShouldNot(gexec.Exit(0))

		Expect(session.Err).To(gbytes.Say(".+"))
	})
})
