# Examples and integration tests

`core-ts`, `core-py`, `core-go` and `core-cs` are the same Pulumi program in each SDK language. Each one creates about 24 resources that need nothing but a Doppler workplace (projects, environments, configs, secrets, tokens, service accounts, groups, roles, members, webhooks, trusted IPs, tags and a change request policy) and calls every function.

The tests in this directory run each program with [ProgramTest](https://pkg.go.dev/github.com/pulumi/pulumi/pkg/v3/testing/integration):

1. Create everything from `values.json` and check the stack outputs.
2. Apply `step2/values.json`, which updates resources in place, and check the outputs again.
3. Refresh after each step. Any change fails the test, so a resource that drifts from what it wrote is caught.
4. Destroy.

## Running

The tests create and delete real resources in the workplace that the token belongs to. Use a test workplace and a service account token with workplace admin access. Every name is prefixed with `pt-<lang>-` and a suffix from the stack name.

```bash
make build install_sdks   # provider binary in bin/ plus all four SDKs

export DOPPLER_TOKEN=dp.sa.xxx
# Optional: an existing workplace user to add to the test group and project.
export DOPPLER_TEST_MEMBER_EMAIL=someone@example.com

cd examples
go test -v -count=1 -timeout 60m -parallel 1 -tags=all -run TestCore .
```

Without `DOPPLER_TOKEN` the tests skip. Use `-tags=nodejs`, `python`, `go` or `dotnet` to run one language.

Prefer running the languages one at a time (`-parallel 1`). With several programs writing secrets at once, the Doppler API intermittently answers secret writes with a 500, and the provider doesn't retry those. The tests retry failed steps up to three times to absorb this, since CI runs one job per language in parallel.

## In CI

CI passes the `DOPPLER_TOKEN` secret and the optional `DOPPLER_TEST_MEMBER_EMAIL` variable from `.ci-mgmt.yaml` to the test jobs, on pull requests from this repository and nightly. Fork pull requests get no secrets, so their tests skip.
