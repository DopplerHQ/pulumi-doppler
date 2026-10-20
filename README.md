# Doppler Provider

This provider lets you manage [Doppler](https://doppler.com/) resources.

## Installing

This package is available for several languages/platforms:

### Node.js (JavaScript/TypeScript)

To use from JavaScript or TypeScript in Node.js, install using either `npm`:

```bash
npm install @dopplerhq/pulumi-doppler
```

or `yarn`:

```bash
yarn add @dopplerhq/pulumi-doppler
```

### Python

To use from Python, install using `pip`:

```
$ pip install pulumi-doppler
```

### Go

To use from Go, use `go get` to grab the latest version of the library:

```bash
go get github.com/DopplerHQ/pulumi-doppler/sdk/go/...
```

### .NET

To use from Dotnet, use `dotnet add package` to install into your project. You must specify the version if it is a pre-release version.


```
$ dotnet add package DopplerHQ.Doppler
```

## Configuration

The following configuration points are available for the `doppler` provider:

- `doppler:dopplerToken` (environment: `DOPPLER_TOKEN`) - A Doppler personal or service token. Use a service account token for automation; refreshing secrets with restricted visibility requires one.
- `doppler:oidcIdentity` (environment: `DOPPLER_OIDC_IDENTITY`) - The service account identity ID for OIDC authentication, used with one of:
  - `doppler:oidcToken` (environment: `DOPPLER_OIDC_TOKEN`) - A JWT for OIDC authentication.
  - `doppler:oidcTokenFile` (environment: `DOPPLER_OIDC_TOKEN_FILE`) - A path to a file containing that JWT.
- `doppler:host` (environment: `DOPPLER_API_HOST`) - The Doppler API host. Defaults to `https://api.doppler.com`.
- `doppler:verifyTls` (environment: `DOPPLER_VERIFY_TLS`) - Whether to verify TLS. Defaults to `true`.

Only one of `dopplerToken` or OIDC may be configured.

## Reference

For detailed reference documentation, please visit [the Pulumi registry](https://www.pulumi.com/registry/packages/doppler/api-docs/).
