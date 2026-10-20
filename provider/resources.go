// Copyright 2016-2018, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doppler

import (
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"

	// Allow embedding bridge-metadata.json in the provider.
	_ "embed"

	"github.com/DopplerHQ/terraform-provider-doppler/doppler"

	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge/tokens"
	shim "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfshim"
	shimv2 "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfshim/sdk-v2"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"

	"github.com/DopplerHQ/pulumi-doppler/provider/pkg/version"
)

// all of the token components used below.
const (
	// This variable controls the default name of the package in the package
	// registries for nodejs and python:
	mainPkg = "doppler"
	// modules:
	mainMod = "index" // the doppler module
	// githubOrg is Doppler's GitHub org, also used as the .NET root namespace.
	githubOrg = "DopplerHQ"
)

//go:embed cmd/pulumi-resource-doppler/bridge-metadata.json
var metadata []byte

// preConfigureCallback is called before the providerConfigure function of the underlying provider.
// It should validate that the provider can be configured, and provide actionable errors in the case
// it cannot be. Configuration variables can be read from `vars` using the `stringValue` function -
// for example `stringValue(vars, "accessKey")`.
func preConfigureCallback(_ resource.PropertyMap, _ shim.ResourceConfig) error {
	return nil
}

// Provider returns additional overlaid schema and metadata associated with the provider..
func Provider() tfbridge.ProviderInfo {
	// Upstream sends "terraform-provider-doppler/<ProviderVersion>" as its User-Agent, which
	// would identify every Pulumi request as Terraform. Append our own product so it reads
	// "terraform-provider-doppler/1.22.0 pulumi-doppler/<version>".
	upstreamVersion := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/DopplerHQ/terraform-provider-doppler" {
				upstreamVersion = strings.TrimPrefix(dep.Version, "v")
			}
		}
	}
	doppler.ProviderVersion = fmt.Sprintf("%s pulumi-doppler/%s", upstreamVersion, version.Version)

	// Instantiate the Terraform provider
	p := shimv2.NewProvider(doppler.Provider())

	oldConfigType := tfbridge.MakeResource(mainPkg, mainMod, "Config").String()

	// Create a Pulumi provider mapping
	prov := tfbridge.ProviderInfo{
		P:    p,
		Name: "doppler",
		// DisplayName is a way to be able to change the casing of the provider
		// name when being displayed on the Pulumi registry
		DisplayName: "Doppler",
		// Shown in the Pulumi Registry. Must match a name in pulumi/registry's
		// publisher-names.json, or registry publishing fails.
		Publisher: "Doppler",
		// Doppler's gradient logomark. Keep it a PNG: the .NET codegen downloads this URL into
		// the NuGet package icon, and NuGet accepts only PNG and JPEG icons.
		LogoURL: "https://raw.githubusercontent.com/DopplerHQ/pulumi-doppler/main/docs/logo.png",
		// PluginDownloadURL is an optional URL used to download the Provider
		// for use in Pulumi programs
		// e.g https://github.com/org/pulumi-provider-name/releases/
		PluginDownloadURL: "github://api.github.com/DopplerHQ",
		Description:       "A Pulumi package for managing Doppler projects, secrets, and access.",
		// category/cloud tag helps with categorizing the package in the Pulumi Registry.
		// For all available categories, see `Keywords` in
		// https://www.pulumi.com/docs/guides/pulumi-packages/schema/#package.
		Keywords:   []string{"pulumi", "doppler", "secrets", "secrets-management", "category/cloud"},
		License:    "Apache-2.0",
		Homepage:   "https://www.doppler.com",
		Repository: "https://github.com/DopplerHQ/pulumi-doppler",
		Version:    version.Version,
		// Without this the bridge assumes the upstream provider is MPL-2.0 and says so in
		// every SDK README.
		TFProviderLicense: tfbridge.SetProviderLicense(tfbridge.Apache20LicenseType),
		// The GitHub Org for the provider - defaults to `terraform-providers`. Note that this
		// should match the TF provider module's require directive, not any replace directives.
		GitHubOrg:    githubOrg,
		MetadataInfo: tfbridge.NewProviderMetadata(metadata),
		Config: map[string]*tfbridge.SchemaInfo{
			// Upstream doesn't mark the token Sensitive, so Pulumi would store it in plaintext.
			"doppler_token": {Secret: tfbridge.True()},
		},
		PreConfigureCallback: preConfigureCallback,
		Resources: map[string]*tfbridge.ResourceInfo{
			"doppler_config": {
				Tok: tfbridge.MakeResource(mainPkg, mainMod, "BranchConfig"),
				Aliases: []tfbridge.AliasInfo{
					{
						Type: &oldConfigType,
					},
				},
			},
			"doppler_trusted_ips": {
				Fields: map[string]*tfbridge.SchemaInfo{
					// C# forbids a member named after its enclosing type (TrustedIps).
					"trusted_ips": {CSharpName: "TrustedIpList"},
				},
			},
		},
		DataSources: map[string]*tfbridge.DataSourceInfo{
			"doppler_environments": {
				Fields: map[string]*tfbridge.SchemaInfo{
					// The bridge pluralizes the upstream "list" attribute to "lists".
					"list": {Name: "environments"},
				},
			},
		},
		JavaScript: &tfbridge.JavaScriptInfo{
			PackageName:          "@dopplerhq/pulumi-doppler",
			RespectSchemaVersion: true,
			// See the documentation for tfbridge.OverlayInfo for how to lay out this
			// section, or refer to the AWS provider. Delete this section if there are
			// no overlay files.
			//Overlay: &tfbridge.OverlayInfo{},
		},
		Python: &tfbridge.PythonInfo{
			PackageName: "pulumi_doppler",
			// List any Python dependencies and their version ranges
			Requires: map[string]string{
				"pulumi": ">=3.0.0,<4.0.0",
			},
			PyProject:            struct{ Enabled bool }{true},
			RespectSchemaVersion: true,
		},
		Golang: &tfbridge.GolangInfo{
			ImportBasePath: filepath.Join(
				fmt.Sprintf("github.com/DopplerHQ/pulumi-%[1]s/sdk/", mainPkg),
				tfbridge.GetModuleMajorVersion(version.Version),
				"go",
				mainPkg,
			),
			GenerateResourceContainerTypes: true,
			RespectSchemaVersion:           true,
		},
		CSharp: &tfbridge.CSharpInfo{
			RootNamespace: githubOrg,
			PackageReferences: map[string]string{
				"Pulumi": "3.*",
			},
			RespectSchemaVersion: true,
		},
	}

	prov.SetAutonaming(255, "-")
	prov.MustComputeTokens(
		tokens.KnownModules(
			"doppler_",
			mainMod,
			[]string{
				"integration_",
				"project_member_",
				"rotated_secret_",
				"secrets_sync_",
			},
			tokens.MakeStandard(mainPkg)))
	// Record token history in bridge-metadata.json so later renames get aliases automatically.
	prov.MustApplyAutoAliases()

	return prov
}
