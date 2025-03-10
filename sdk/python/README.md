# Doppler Provider

This provider lets you manage [Doppler](https://doppler.com/) resources.

## Installing

This package is available for several languages/platforms:

### Node.js (JavaScript/TypeScript)

To use from JavaScript or TypeScript in Node.js, install using either `npm`:

```bash
npm install @pulumiverse/doppler
```

or `yarn`:

```bash
yarn add @pulumiverse/doppler
```

### Python

To use from Python, install using `pip`:

```
$ pip install pulumiverse-doppler
```

### Go

To use from Go, use `go get` to grab the latest version of the library:

```bash
go get github.com/DopplerHQ/pulumi-doppler/sdk/go/...
```

### .NET

To use from Dotnet, use `dotnet add package` to install into your project. You must specify the version if it is a pre-release version.


```
$ dotnet add package Pulumiverse.Doppler
```

## Configuration

The following configuration points are available for the `doppler` provider:

- `doppler:token` - You need to provide a service account, client, or other token type for this provider to communicate successfully with the Doppler API.

## Reference

For detailed reference documentation, please visit [the Pulumi registry](https://www.pulumi.com/registry/packages/doppler/api-docs/).
