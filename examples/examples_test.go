package examples

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/assert"
)

func getCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.FailNow()
	}

	return cwd
}

// getBaseOptions configures a test that creates real resources in the Doppler workplace that
// DOPPLER_TOKEN belongs to, and skips the test when it isn't set. Set
// DOPPLER_TEST_MEMBER_EMAIL to an existing workplace user to also test group and project
// membership for users.
func getBaseOptions(t *testing.T) integration.ProgramTestOptions {
	t.Helper()
	if os.Getenv("DOPPLER_TOKEN") == "" {
		t.Skip("DOPPLER_TOKEN is not set; skipping tests that create Doppler resources")
	}
	binPath, err := filepath.Abs("../bin")
	if err != nil {
		t.Fatal(err)
	}
	config := map[string]string{}
	if email := os.Getenv("DOPPLER_TEST_MEMBER_EMAIL"); email != "" {
		config["memberEmail"] = email
	}
	return integration.ProgramTestOptions{
		LocalProviders: []integration.LocalDependency{{Package: "doppler", Path: binPath}},
		Config:         config,
		// With several programs writing secrets at once, the Doppler API intermittently
		// answers with a 500, and upstream doesn't retry those. CI runs one job per
		// language in parallel. A retried update resumes where it failed; real drift fails
		// every attempt, so this doesn't hide it.
		RetryFailedSteps: true,
	}
}

// coreTestOptions runs a core-<lang> program: create, then step2/ updates values in place.
// Refreshes must find no changes, so a provider that drifts from what it wrote fails here.
func coreTestOptions(t *testing.T, dir string) integration.ProgramTestOptions {
	t.Helper()
	return integration.ProgramTestOptions{
		Dir:                    filepath.Join(getCwd(t), dir),
		ExpectRefreshChanges:   false,
		ExtraRuntimeValidation: validateCoreOutputs("value-1"),
		EditDirs: []integration.EditDir{{
			Dir:                    filepath.Join(getCwd(t), dir, "step2"),
			Additive:               true,
			ExtraRuntimeValidation: validateCoreOutputs("value-2"),
		}},
	}
}

func validateCoreOutputs(apiKey string) func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
	return func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
		out := stack.Outputs
		assert.Equal(t, apiKey, out["apiKeyFromFunction"], "getSecrets should return the API_KEY just written")
		assert.Equal(t, []interface{}{"qa", "uat"}, out["environmentSlugs"])
		assert.Equal(t, out["projectName"], out["groupName"], "getGroup should find the group by slug")
		assert.Equal(t, true, out["serviceTokenIsSecret"], "service token keys must be secret")
		assert.Equal(t, true, out["serviceAccountTokenIsSecret"], "service account API keys must be secret")
		fmt.Printf("validated outputs for %s\n", out["projectName"])
	}
}
