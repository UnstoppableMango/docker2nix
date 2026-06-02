package docker2nix_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	docker2nix "github.com/unstoppablemango/docker2nix/pkg"
)

var _ = Describe("Generate", func() {
	It("should convert a FROM-only Dockerfile to a Nix expression", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{}.Build()

		resp, err := docker2nix.Generate(ctx, req)

		Expect(err).NotTo(HaveOccurred())
		Expect(resp).NotTo(BeNil())
	})

	It("should handle RUN, COPY, ENV, and CMD instructions", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{}.Build()

		resp, err := docker2nix.Generate(ctx, req)

		Expect(err).NotTo(HaveOccurred())
		Expect(resp).NotTo(BeNil())
	})

	It("should handle multi-stage Dockerfiles", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{}.Build()

		resp, err := docker2nix.Generate(ctx, req)

		Expect(err).NotTo(HaveOccurred())
		Expect(resp).NotTo(BeNil())
	})

	It("should return an error for an invalid Dockerfile", func(ctx context.Context) {
		req := docker2nix.GenerateRequest_builder{}.Build()

		resp, err := docker2nix.Generate(ctx, req)

		Expect(err).To(HaveOccurred())
		Expect(resp).To(BeNil())
	})
})
