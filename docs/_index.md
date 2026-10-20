---
title: Doppler
meta_desc: Use the Doppler provider for Pulumi to manage Doppler projects, environments, configs, secrets, service accounts, and integrations.
layout: package
---

The Doppler provider for Pulumi lets you manage [Doppler](https://www.doppler.com/) resources as part of your Pulumi programs: projects, environments, configs, secrets, service accounts and tokens, groups and roles, integrations, secrets syncs, and rotated secrets.

## Installation

{{< chooser language "typescript,python,go,csharp,java,yaml,hcl" >}}
{{% choosable language typescript %}}

```bash
npm install @dopplerhq/pulumi-doppler
```

{{% /choosable %}}
{{% choosable language python %}}

```bash
pip install pulumi_doppler
```

{{% /choosable %}}
{{% choosable language go %}}

```bash
go get github.com/DopplerHQ/pulumi-doppler/sdk/go/doppler
```

{{% /choosable %}}
{{% choosable language csharp %}}

```bash
dotnet add package DopplerHQ.Doppler
```

{{% /choosable %}}
{{% choosable language java %}}

There is no published Java SDK. Generate one locally and add it to your project:

```bash
pulumi package add doppler
```

{{% /choosable %}}
{{% choosable language yaml %}}

```bash
pulumi package add doppler
```

{{% /choosable %}}
{{% choosable language hcl %}}

Declare the provider in your program, then run `pulumi install`:

```hcl
terraform {
  required_providers {
    doppler = {
      source  = "DopplerHQ/doppler"
      version = "~> 1.22"
    }
  }
}
```

```bash
pulumi install
```

{{% /choosable %}}
{{< /chooser >}}

## Example Usage

This program creates a project with a `ci` environment and sets a secret in that environment's root config, which shares the environment's slug. Which other environments a new project starts with depends on your workplace's default environment settings.

{{< chooser language "typescript,python,go,csharp,yaml" >}}
{{% choosable language typescript %}}

```typescript
import * as doppler from "@dopplerhq/pulumi-doppler";

const project = new doppler.Project("backend", {
    name: "backend",
    description: "Backend services",
});

const ci = new doppler.Environment("ci", {
    project: project.name,
    slug: "ci",
    name: "CI",
});

new doppler.Secret("database-url", {
    project: project.name,
    config: ci.slug,
    name: "DATABASE_URL",
    value: "postgres://db.example.com:5432/app",
});
```

{{% /choosable %}}
{{% choosable language python %}}

```python
import pulumi_doppler as doppler

project = doppler.Project("backend",
    name="backend",
    description="Backend services")

ci = doppler.Environment("ci",
    project=project.name,
    slug="ci",
    name="CI")

doppler.Secret("database-url",
    project=project.name,
    config=ci.slug,
    name="DATABASE_URL",
    value="postgres://db.example.com:5432/app")
```

{{% /choosable %}}
{{% choosable language go %}}

```go
package main

import (
	"github.com/DopplerHQ/pulumi-doppler/sdk/go/doppler"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		project, err := doppler.NewProject(ctx, "backend", &doppler.ProjectArgs{
			Name:        pulumi.String("backend"),
			Description: pulumi.String("Backend services"),
		})
		if err != nil {
			return err
		}
		ci, err := doppler.NewEnvironment(ctx, "ci", &doppler.EnvironmentArgs{
			Project: project.Name,
			Slug:    pulumi.String("ci"),
			Name:    pulumi.String("CI"),
		})
		if err != nil {
			return err
		}
		_, err = doppler.NewSecret(ctx, "database-url", &doppler.SecretArgs{
			Project: project.Name,
			Config:  ci.Slug,
			Name:    pulumi.String("DATABASE_URL"),
			Value:   pulumi.String("postgres://db.example.com:5432/app"),
		})
		return err
	})
}
```

{{% /choosable %}}
{{% choosable language csharp %}}

```csharp
using Pulumi;
using DopplerHQ.Doppler;

return await Deployment.RunAsync(() =>
{
    var project = new Project("backend", new ProjectArgs
    {
        Name = "backend",
        Description = "Backend services",
    });

    var ci = new DopplerHQ.Doppler.Environment("ci", new EnvironmentArgs
    {
        Project = project.Name,
        Slug = "ci",
        Name = "CI",
    });

    new Secret("database-url", new SecretArgs
    {
        Project = project.Name,
        Config = ci.Slug,
        Name = "DATABASE_URL",
        Value = "postgres://db.example.com:5432/app",
    });
});
```

{{% /choosable %}}
{{% choosable language yaml %}}

```yaml
name: doppler-example
runtime: yaml
resources:
  backend:
    type: doppler:Project
    properties:
      name: backend
      description: Backend services
  ci:
    type: doppler:Environment
    properties:
      project: ${backend.name}
      slug: ci
      name: CI
  database-url:
    type: doppler:Secret
    properties:
      project: ${backend.name}
      config: ${ci.slug}
      name: DATABASE_URL
      value: postgres://db.example.com:5432/app
```

{{% /choosable %}}
{{< /chooser >}}

Authenticate with a [service account](https://docs.doppler.com/docs/service-accounts) token:

```bash
pulumi config set --secret doppler:dopplerToken dp.sa.xxx
```

## Configuration

Configure either `dopplerToken` or OIDC (`oidcIdentity` with one of `oidcToken` or `oidcTokenFile`), not both.

- `dopplerToken` (Secret) — A Doppler personal or service token. May also be set with the `DOPPLER_TOKEN` environment variable. Use a service account token for automation: with a personal token the provider can't read secrets with [restricted visibility](https://docs.doppler.com/docs/secret-visibility), so refreshing them fails. Can't be combined with the OIDC options.
- `oidcIdentity` — The ID (a UUID) of the [service account identity](https://docs.doppler.com/docs/service-account-identities) to authenticate as with OIDC. May also be set with the `DOPPLER_OIDC_IDENTITY` environment variable. Requires `oidcToken` or `oidcTokenFile`.
- `oidcToken` (Secret) — A JWT for OIDC authentication. May also be set with the `DOPPLER_OIDC_TOKEN` environment variable. Can't be combined with `oidcTokenFile`.
- `oidcTokenFile` — A path to a file containing a JWT for OIDC authentication, such as a Kubernetes projected service account token. May also be set with the `DOPPLER_OIDC_TOKEN_FILE` environment variable. Can't be combined with `oidcToken`.
- `host` — The Doppler API host. May also be set with the `DOPPLER_API_HOST` environment variable. Defaults to `https://api.doppler.com`.
- `verifyTls` — Whether to verify TLS. May also be set with the `DOPPLER_VERIFY_TLS` environment variable. Defaults to `true`.

### Configuration examples

With a service account token:

```bash
pulumi config set --secret doppler:dopplerToken dp.sa.xxx
```

With OIDC, for example from a Kubernetes pod using a projected service account token:

```bash
pulumi config set doppler:oidcIdentity 00000000-0000-0000-0000-000000000000
pulumi config set doppler:oidcTokenFile /var/run/secrets/tokens/doppler
```
