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
