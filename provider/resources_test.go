package doppler

import (
	"strings"
	"testing"

	"github.com/DopplerHQ/terraform-provider-doppler/doppler"

	"github.com/DopplerHQ/pulumi-doppler/provider/pkg/version"
)

func TestUserAgentIdentifiesPulumi(t *testing.T) {
	// Normally set by ldflags; Provider() needs a parseable version.
	version.Version = "0.10.0"
	Provider()
	// Upstream prefixes this with "terraform-provider-doppler/". Test binaries carry no
	// dependency build info, so the upstream version reads "unknown" here.
	if !strings.HasSuffix(doppler.ProviderVersion, " pulumi-doppler/0.10.0") {
		t.Fatalf("ProviderVersion = %q, want <upstream version> pulumi-doppler/0.10.0", doppler.ProviderVersion)
	}
}
