//go:build nodejs || all
// +build nodejs all

package examples

import (
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestCoreTs(t *testing.T) {
	test := getBaseOptions(t).
		With(integration.ProgramTestOptions{
			// Linked by `make install_nodejs_sdk`.
			Dependencies: []string{"@dopplerhq/pulumi-doppler"},
		}).
		With(coreTestOptions(t, "core-ts"))

	integration.ProgramTest(t, &test)
}
