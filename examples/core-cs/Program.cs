using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text.Json;
using Pulumi;
using DopplerHQ.Doppler;
using DopplerHQ.Doppler.Inputs;
using ProjectMember = DopplerHQ.Doppler.ProjectMember;

return await Deployment.RunAsync(() =>
{
    // values.json holds everything that changes between test steps; step2/ replaces it.
    var values = JsonDocument.Parse(File.ReadAllText("values.json")).RootElement;
    string Str(string key) => values.GetProperty(key).GetString()!;
    var config = new Pulumi.Config();
    // An existing workplace user to add to the group and project. Optional.
    var memberEmail = config.Get("memberEmail");

    // Doppler names are unique per workplace, so suffix them with the stack name.
    var suffix = System.Text.RegularExpressions.Regex.Replace(Deployment.Instance.StackName.ToLowerInvariant(), "[^a-z0-9]", "");
    suffix = suffix.Length > 8 ? suffix[^8..] : suffix;
    var prefix = $"pt-cs-{suffix}";

    var tag = new Tag("tag", new TagArgs { Name = prefix, Color = Str("tagColor") });

    var project = new Project("project", new ProjectArgs
    {
        Name = prefix,
        Description = Str("projectDescription"),
    });
    new ProjectTag("project-tag", new ProjectTagArgs { Project = project.Name, TagSlug = tag.Slug });

    // Create every environment used here: workplaces differ in which defaults new projects get.
    var qa = new DopplerHQ.Doppler.Environment("qa", new EnvironmentArgs { Project = project.Name, Slug = "qa", Name = "QA" });
    var uat = new DopplerHQ.Doppler.Environment("uat", new EnvironmentArgs { Project = project.Name, Slug = "uat", Name = "UAT" });
    var feature = new BranchConfig("qa-feature", new BranchConfigArgs
    {
        Project = project.Name,
        Environment = qa.Slug,
        Name = "qa_feature",
    });

    var apiKey = new Secret("api-key", new SecretArgs
    {
        Project = project.Name,
        Config = qa.Slug,
        Name = "API_KEY",
        Value = Str("apiKey"),
    });
    // Refreshing a restricted secret only works with a service token (issue #15).
    new Secret("db-password", new SecretArgs
    {
        Project = project.Name,
        Config = qa.Slug,
        Name = "DB_PASSWORD",
        Value = "hunter2",
        Visibility = "restricted",
    });
    new SecretNote("api-key-note", new SecretNoteArgs
    {
        Project = project.Name,
        Name = apiKey.Name,
        Note = "Owned by the provider integration test",
    });

    var serviceToken = new ServiceToken("service-token", new ServiceTokenArgs
    {
        Project = project.Name,
        Config = qa.Slug,
        Name = prefix,
        Access = "read",
    });

    var sa = new ServiceAccount("service-account", new ServiceAccountArgs { Name = prefix, WorkplaceRole = "collaborator" });
    var saToken = new ServiceAccountToken("service-account-token", new ServiceAccountTokenArgs
    {
        ServiceAccountSlug = sa.Slug,
        Name = prefix,
    });
    new ServiceAccountIdentity("service-account-identity", new ServiceAccountIdentityArgs
    {
        ServiceAccountSlug = sa.Slug,
        Name = prefix,
        TtlSeconds = 600,
        ConfigOidc = new ServiceAccountIdentityConfigOidcArgs
        {
            DiscoveryUrl = "https://token.actions.githubusercontent.com",
            Claims =
            {
                new ServiceAccountIdentityConfigOidcClaimArgs { Key = "aud", Values = { "https://github.com/DopplerHQ" } },
                new ServiceAccountIdentityConfigOidcClaimArgs { Key = "sub", Values = { "repo:DopplerHQ/pulumi-doppler:ref:refs/heads/main" } },
            },
        },
    });

    var group = new Group("group", new GroupArgs { Name = prefix });
    var projectRole = new ProjectRole("project-role", new ProjectRoleArgs
    {
        Name = $"{prefix} log viewer",
        // enclave_config_logs alone is rejected; it requires secrets read.
        Permissions = { "enclave_config_logs", "enclave_project_config_secrets_read" },
    });
    new WorkplaceRole("workplace-role", new WorkplaceRoleArgs { Name = $"{prefix} logs", Permissions = { "logs" } });

    new ProjectMember.Group("project-group", new ProjectMember.GroupArgs
    {
        Project = project.Name,
        GroupSlug = group.Slug,
        Role = "viewer",
        Environments = { qa.Slug },
    });
    new ProjectMember.ServiceAccount("project-service-account", new ProjectMember.ServiceAccountArgs
    {
        Project = project.Name,
        ServiceAccountSlug = sa.Slug,
        Role = projectRole.Identifier,
        Environments = { qa.Slug },
    });

    if (!string.IsNullOrEmpty(memberEmail))
    {
        var user = GetUser.Invoke(new GetUserInvokeArgs { Email = memberEmail });
        new GroupMember("group-member", new GroupMemberArgs { GroupSlug = group.Slug, UserSlug = user.Apply(u => u.Slug) });
        new ProjectMember.User("project-user", new ProjectMember.UserArgs
        {
            Project = project.Name,
            UserSlug = user.Apply(u => u.Slug),
            Role = "viewer",
        });
    }

    new Webhook("webhook", new WebhookArgs
    {
        Project = project.Name,
        Name = Str("webhookName"),
        Url = "https://example.com/doppler-webhook",
        // Webhooks are always created enabled; upstream ignores enabled=false on create, so only
        // the update step disables it.
        Enabled = values.GetProperty("webhookEnabled").GetBoolean(),
        EnabledConfigs = { feature.Name },
        Authentication = new WebhookAuthenticationArgs { Type = "Bearer", Token = "integration-test-token" },
    });

    // Trusted IPs lock a config to those ranges for every caller, the provider included. Keep
    // 0.0.0.0/0 and ::/0 in the list so the test can still read and delete the config.
    // C# names the trustedIps property TrustedIpList: a member can't share its class's name.
    var trustedIps = new TrustedIps("trusted-ips", new TrustedIpsArgs
    {
        Project = project.Name,
        Config = feature.Name,
        TrustedIpList = values.GetProperty("trustedIps").EnumerateArray().Select(e => e.GetString()!).ToList(),
    });

    new ChangeRequestPolicy("change-request-policy", new ChangeRequestPolicyArgs
    {
        Name = prefix,
        Description = "Integration test policy",
        Rules = new ChangeRequestPolicyRulesArgs
        {
            DisallowSelfReview = true,
            RequiredReviewers = { new ChangeRequestPolicyRulesRequiredReviewerArgs { Count = 1 } },
        },
        Targets = new ChangeRequestPolicyTargetsArgs
        {
            Projects =
            {
                new ChangeRequestPolicyTargetsProjectArgs { ProjectName = project.Name, EnvironmentSlugs = { uat.Slug } },
            },
        },
    });

    // Functions. GetSecrets reads after the secrets exist because its inputs come from them.
    var secrets = GetSecrets.Invoke(new GetSecretsInvokeArgs { Project = apiKey.Project, Config = apiKey.Config });
    var environments = GetEnvironments.Invoke(new GetEnvironmentsInvokeArgs { Project = qa.Project });
    var foundGroup = GetGroup.Invoke(new GetGroupInvokeArgs { Slug = group.Slug });

    return new Dictionary<string, object?>
    {
        ["projectName"] = project.Name,
        ["environmentSlugs"] = environments.Apply(r => r.Environments.Select(e => e.Slug).OrderBy(s => s, System.StringComparer.Ordinal).ToList()),
        ["apiKeyFromFunction"] = Output.Unsecret(secrets.Apply(r => r.Map["API_KEY"])),
        ["groupName"] = foundGroup.Apply(g => g.Name),
        ["serviceTokenIsSecret"] = Output.Create(Output.IsSecretAsync(serviceToken.Key)),
        ["serviceAccountTokenIsSecret"] = Output.Create(Output.IsSecretAsync(saToken.ApiKey)),
        ["trustedIpCount"] = trustedIps.TrustedIpList.Apply(l => l.Length),
    };
});
