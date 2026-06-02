package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	runGenerate := func(stdin string, args ...string) *gexec.Session {
		cmd := exec.Command(binary, append([]string{"generate"}, args...)...)
		if stdin != "" {
			cmd.Stdin = nil
			r, w, err := os.Pipe()
			Expect(err).NotTo(HaveOccurred())
			_, err = w.WriteString(stdin)
			Expect(err).NotTo(HaveOccurred())
			w.Close()
			cmd.Stdin = r
		}
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		return session
	}

	It("should convert a FROM-only Dockerfile to a Nix expression", func() {
		session := runGenerate("FROM ubuntu:24.04\n")
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("ubuntu"))
		Expect(string(session.Out.Contents())).To(ContainSubstring("24.04"))
	})

	It("should handle RUN, COPY, ENV, and CMD instructions", func() {
		dockerfile := `FROM ubuntu:24.04
RUN apt-get update && apt-get install -y curl
COPY . /app
ENV FOO=bar
CMD ["/app/main"]
`
		session := runGenerate(dockerfile)
		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).NotTo(BeEmpty())
	})

	It("should handle multi-stage Dockerfiles", func() {
		dockerfile := `FROM golang:1.23 AS builder
WORKDIR /src
COPY . .
RUN go build -o /app/main .

FROM ubuntu:24.04
COPY --from=builder /app/main /app/main
CMD ["/app/main"]
`
		session := runGenerate(dockerfile)
		Eventually(session).Should(gexec.Exit(0))
		out := string(session.Out.Contents())
		Expect(out).To(ContainSubstring("builder"))
		Expect(out).To(ContainSubstring("ubuntu"))
	})

	It("should accept a Dockerfile path as argument", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "Dockerfile")
		Expect(os.WriteFile(path, []byte("FROM ubuntu:24.04\n"), 0o644)).To(Succeed())

		session := runGenerate("", path)
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("ubuntu"))
	})

	It("should return an error for an invalid Dockerfile", func() {
		session := runGenerate("NOT A VALID DOCKERFILE INSTRUCTION\n")
		Eventually(session).ShouldNot(gexec.Exit(0))
		Expect(session.Err.Contents()).NotTo(BeEmpty())
	})
})
