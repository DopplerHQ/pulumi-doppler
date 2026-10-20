//go:build go || all
// +build go all

package examples

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/require"
)

func TestCoreGo(t *testing.T) {
	sdkPath, err := filepath.Abs("../sdk")
	require.NoError(t, err)
	goDepRoot := os.Getenv("PULUMI_GO_DEP_ROOT")
	if goDepRoot == "" {
		goDepRoot, err = filepath.Abs("../..")
		require.NoError(t, err)
	}

	test := getBaseOptions(t).
		With(integration.ProgramTestOptions{
			Dependencies: []string{fmt.Sprintf("github.com/DopplerHQ/pulumi-doppler/sdk=%s", sdkPath)},
			Env:          []string{fmt.Sprintf("PULUMI_GO_DEP_ROOT=%s", goDepRoot)},
		}).
		With(coreTestOptions(t, "core-go"))

	integration.ProgramTest(t, &test)
}
