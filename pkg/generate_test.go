package docker2nix_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	docker2nix "github.com/unstoppablemango/docker2nix/pkg"
)

var _ = Describe("Generate", func() {
	It("should convert a FROM-only Dockerfile to a Nix expression", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new("FROM ubuntu:24.04\n"),
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring("dockerTools.buildLayeredImage"))
		Expect(resp.GetNix()).To(ContainSubstring(`name = "ubuntu"`))
		Expect(resp.GetNix()).To(ContainSubstring(`tag = "24.04"`))
	})

	It("should handle RUN, COPY, ENV, and CMD instructions", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new(`FROM ubuntu:24.04
RUN apt-get update && apt-get install -y curl
COPY . /app
ENV FOO=bar
CMD ["/app/main"]
`),
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring("dockerTools.buildLayeredImage"))
		Expect(resp.GetNix()).To(ContainSubstring(`name = "ubuntu"`))
		Expect(resp.GetNix()).To(ContainSubstring("runAsRoot"))
		Expect(resp.GetNix()).To(ContainSubstring("apt-get update && apt-get install -y curl"))
		Expect(resp.GetNix()).To(ContainSubstring(`"FOO=bar"`))
		Expect(resp.GetNix()).To(ContainSubstring(`"/app/main"`))
	})

	It("should render a single RUN instruction as runAsRoot", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new(`FROM ubuntu:24.04
RUN apt-get update
`),
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring("runAsRoot = ''"))
		Expect(resp.GetNix()).To(ContainSubstring("apt-get update"))
	})

	It("should render multiple RUN instructions into a single runAsRoot", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new(`FROM ubuntu:24.04
RUN apt-get update
RUN apt-get install -y curl
`),
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring("runAsRoot = ''"))
		Expect(resp.GetNix()).To(ContainSubstring("apt-get update"))
		Expect(resp.GetNix()).To(ContainSubstring("apt-get install -y curl"))
	})

	It("should render exec-form RUN as joined args in runAsRoot", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new(`FROM ubuntu:24.04
RUN ["apt-get", "install", "-y", "curl"]
`),
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring("runAsRoot = ''"))
		Expect(resp.GetNix()).To(ContainSubstring("apt-get install -y curl"))
	})

	It("should handle multi-stage Dockerfiles", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new(`FROM golang:1.23 AS builder
WORKDIR /src
COPY . .
RUN go build -o /app/main .

FROM ubuntu:24.04
COPY --from=builder /app/main /app/main
CMD ["/app/main"]
`),
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring("dockerTools.buildLayeredImage"))
		Expect(resp.GetNix()).To(ContainSubstring("builder"))
		Expect(resp.GetNix()).To(ContainSubstring(`name = "ubuntu"`))
	})

	It("should use the last CMD when multiple CMD instructions are present", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new(`FROM ubuntu:24.04
CMD ["/first"]
CMD ["/last"]
`),
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring(`"/last"`))
		Expect(resp.GetNix()).NotTo(ContainSubstring(`"/first"`))
	})

	It("should parse registry:port/image:tag without splitting on the registry colon", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new("FROM localhost:5000/myimg:v1\n"),
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring(`name = "localhost:5000/myimg"`))
		Expect(resp.GetNix()).To(ContainSubstring(`tag = "v1"`))
	})

	It("should sanitize hyphenated stage names to valid Nix identifiers", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new(`FROM golang:1.23 AS my-builder
WORKDIR /src

FROM ubuntu:24.04
COPY --from=my-builder /src/app /app
`),
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring("my_builder"))
		Expect(resp.GetNix()).NotTo(ContainSubstring("my-builder"))
	})

	It("should escape Nix interpolation sequences in ENV values", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			// $$ in Dockerfile produces a literal $ in the value
			Dockerfile: new(`FROM ubuntu:24.04
ENV MSG=$${inject}
`),
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring(`\${inject}`))
	})

	It("should return an error for an invalid Dockerfile", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new("NOT A VALID DOCKERFILE INSTRUCTION\n"),
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).To(HaveOccurred())
		Expect(resp).To(BeNil())
	})
})

var _ = Describe("Generate docker-tools explicit", func() {
	format := docker2nix.FORMAT_DOCKER_TOOLS

	It("should convert a FROM-only Dockerfile to a dockerTools expression", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new("FROM ubuntu:24.04\n"),
			Format:     &format,
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring("dockerTools.buildLayeredImage"))
		Expect(resp.GetNix()).NotTo(ContainSubstring("nix2container"))
		Expect(resp.GetNix()).To(ContainSubstring(`name = "ubuntu"`))
		Expect(resp.GetNix()).To(ContainSubstring(`tag = "24.04"`))
	})
})

var _ = Describe("Generate unknown format", func() {
	It("should return an error for an unsupported format value", func(ctx context.Context) {
		format := docker2nix.Format(99)
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new("FROM ubuntu:24.04\n"),
			Format:     &format,
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported format"))
		Expect(resp).To(BeNil())
	})
})

var _ = Describe("Generate nix2container", func() {
	format := docker2nix.FORMAT_NIX2CONTAINER

	It("should convert a FROM-only Dockerfile to a nix2container expression", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new("FROM ubuntu:24.04\n"),
			Format:     &format,
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring("nix2container.buildImage"))
		Expect(resp.GetNix()).NotTo(ContainSubstring("dockerTools"))
		Expect(resp.GetNix()).To(ContainSubstring(`name = "ubuntu"`))
		Expect(resp.GetNix()).To(ContainSubstring(`tag = "24.04"`))
	})

	It("should use lowercase env and cmd config keys", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new(`FROM ubuntu:24.04
ENV FOO=bar
CMD ["/app/main"]
`),
			Format: &format,
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring("nix2container.buildImage"))
		Expect(resp.GetNix()).To(ContainSubstring(`env = [`))
		Expect(resp.GetNix()).To(ContainSubstring(`cmd = [`))
		Expect(resp.GetNix()).NotTo(ContainSubstring(`Env = [`))
		Expect(resp.GetNix()).NotTo(ContainSubstring(`Cmd = [`))
		Expect(resp.GetNix()).To(ContainSubstring(`"FOO=bar"`))
		Expect(resp.GetNix()).To(ContainSubstring(`"/app/main"`))
	})

	It("should handle multi-stage Dockerfiles", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{
			Dockerfile: new(`FROM golang:1.23 AS builder
WORKDIR /src
COPY . .
RUN go build -o /app/main .

FROM ubuntu:24.04
COPY --from=builder /app/main /app/main
CMD ["/app/main"]
`),
			Format: &format,
		}

		resp, err := docker2nix.Generate(ctx, req.Build())

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetNix()).To(ContainSubstring("nix2container.buildImage"))
		Expect(resp.GetNix()).To(ContainSubstring("builder"))
		Expect(resp.GetNix()).To(ContainSubstring(`name = "ubuntu"`))
	})
})
