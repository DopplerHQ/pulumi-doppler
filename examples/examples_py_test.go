//go:build python || all
// +build python all

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestCorePy(t *testing.T) {
	test := getBaseOptions(t).
		With(integration.ProgramTestOptions{
			// Built by `make build_python`.
			Dependencies: []string{filepath.Join("..", "sdk", "python", "bin")},
		}).
		With(coreTestOptions(t, "core-py"))

	integration.ProgramTest(t, &test)
}
