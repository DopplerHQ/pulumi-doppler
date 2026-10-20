//go:build dotnet || all
// +build dotnet all

package examples

import (
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestCoreCs(t *testing.T) {
	test := getBaseOptions(t).
		With(integration.ProgramTestOptions{
			// Published to the local feed (PULUMI_LOCAL_NUGET) by `make install_dotnet_sdk`.
			Dependencies: []string{"DopplerHQ.Doppler"},
		}).
		With(coreTestOptions(t, "core-cs"))

	integration.ProgramTest(t, &test)
}
