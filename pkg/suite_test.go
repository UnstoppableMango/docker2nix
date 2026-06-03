package docker2nix_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDockerToNix(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Docker2Nix Suite")
}
