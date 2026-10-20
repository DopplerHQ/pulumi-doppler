CHANGELOG
=========

## HEAD (Unreleased)

This is the first release maintained by Doppler. The SDKs are published under new package names, so existing programs need to switch packages.

### Breaking changes

The SDK packages and the provider plugin have moved from Pulumiverse to Doppler:

| Language | Before | Now |
|---|---|---|
| Node.js | `@pulumiverse/doppler` | `@dopplerhq/pulumi-doppler` |
| Python | `pulumiverse_doppler` | `pulumi_doppler` |
| Go | `github.com/pulumiverse/pulumi-doppler/sdk/go/doppler` | `github.com/DopplerHQ/pulumi-doppler/sdk/go/doppler` |
| .NET | `Pulumiverse.Doppler` (namespace `Pulumiverse.Doppler`) | `DopplerHQ.Doppler` (namespace `DopplerHQ.Doppler`) |

The provider plugin now downloads from `github://api.github.com/DopplerHQ`. Resource types are unchanged (`doppler:index/project:Project` and so on), so switching packages leaves existing stacks' resources in place.

The Go SDK requires Go 1.26.6 or newer, the minimum of the Pulumi Go SDK it builds on.

### Changes

- Upgrade terraform-provider-doppler from v1.10.0 to v1.22.0: 51 new resources and 3 new functions, including change request policies, project tags, trusted IPs, workplace roles, service account identities, secret notes, rotated secrets (`rotatedSecret` module) and their integrations, Azure Vault, GCP Secret Manager, GitHub Codespaces, Dependabot and Agents secrets syncs, and the `getEnvironments` and `getGroup` functions.
- Provider configuration: add `host`, `verifyTls`, `oidcIdentity`, `oidcToken` and `oidcTokenFile`. `dopplerToken` is now a secret.
- Upgrade pulumi-terraform-bridge to v3.141.0 and Pulumi to v3.270.0.
- In the .NET SDK, `TrustedIps.TrustedIps` is `TrustedIps.TrustedIpList`, because C# forbids a member named after its class.
- Requests to the Doppler API identify the provider as `pulumi-doppler/<version>` alongside the upstream Terraform provider, instead of as Terraform alone.
- SDK readmes credit the upstream provider's actual license, Apache 2.0, instead of MPL 2.0.

---
