package docker2nix

import docker2nixv1alpha1 "github.com/unstoppablemango/docker2nix/pkg/docker2nix/v1alpha1"

type (
	Format                   = docker2nixv1alpha1.Format
	GenerateRequest          = docker2nixv1alpha1.GenerateRequest
	GenerateRequest_builder  = docker2nixv1alpha1.GenerateRequest_builder
	GenerateResponse         = docker2nixv1alpha1.GenerateResponse
	GenerateResponse_builder = docker2nixv1alpha1.GenerateResponse_builder
)

const (
	FORMAT_UNSPECIFIED   = docker2nixv1alpha1.Format_FORMAT_UNSPECIFIED
	FORMAT_DOCKER_TOOLS  = docker2nixv1alpha1.Format_FORMAT_DOCKER_TOOLS
	FORMAT_NIX2CONTAINER = docker2nixv1alpha1.Format_FORMAT_NIX2CONTAINER
)
