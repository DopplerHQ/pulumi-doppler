import * as fs from "fs";
import * as pulumi from "@pulumi/pulumi";
import * as doppler from "@dopplerhq/pulumi-doppler";

// values.json holds everything that changes between test steps; step2/ replaces it.
const values = JSON.parse(fs.readFileSync("values.json", "utf8"));
const config = new pulumi.Config();
// An existing workplace user to add to the group and project. Optional.
const memberEmail = config.get("memberEmail");

// Doppler names are unique per workplace, so suffix them with the stack name.
const suffix = pulumi.getStack().toLowerCase().replace(/[^a-z0-9]/g, "").slice(-8);
const prefix = `pt-ts-${suffix}`;

const tag = new doppler.Tag("tag", { name: prefix, color: values.tagColor });

const project = new doppler.Project("project", {
    name: prefix,
    description: values.projectDescription,
});
new doppler.ProjectTag("project-tag", { project: project.name, tagSlug: tag.slug });

// Create every environment used here: workplaces differ in which defaults new projects get.
const qa = new doppler.Environment("qa", { project: project.name, slug: "qa", name: "QA" });
const uat = new doppler.Environment("uat", { project: project.name, slug: "uat", name: "UAT" });
const feature = new doppler.BranchConfig("qa-feature", {
    project: project.name,
    environment: qa.slug,
    name: "qa_feature",
});

const apiKey = new doppler.Secret("api-key", {
    project: project.name,
    config: qa.slug,
    name: "API_KEY",
    value: values.apiKey,
});
// Refreshing a restricted secret only works with a service token (issue #15).
new doppler.Secret("db-password", {
    project: project.name,
    config: qa.slug,
    name: "DB_PASSWORD",
    value: "hunter2",
    visibility: "restricted",
});
new doppler.SecretNote("api-key-note", {
    project: project.name,
    name: apiKey.name,
    note: "Owned by the provider integration test",
});

const serviceToken = new doppler.ServiceToken("service-token", {
    project: project.name,
    config: qa.slug,
    name: prefix,
    access: "read",
});

const sa = new doppler.ServiceAccount("service-account", { name: prefix, workplaceRole: "collaborator" });
const saToken = new doppler.ServiceAccountToken("service-account-token", {
    serviceAccountSlug: sa.slug,
    name: prefix,
});
new doppler.ServiceAccountIdentity("service-account-identity", {
    serviceAccountSlug: sa.slug,
    name: prefix,
    ttlSeconds: 600,
    configOidc: {
        discoveryUrl: "https://token.actions.githubusercontent.com",
        claims: [
            { key: "aud", values: ["https://github.com/DopplerHQ"] },
            { key: "sub", values: ["repo:DopplerHQ/pulumi-doppler:ref:refs/heads/main"] },
        ],
    },
});

const group = new doppler.Group("group", { name: prefix });
const projectRole = new doppler.ProjectRole("project-role", {
    name: `${prefix} log viewer`,
    // enclave_config_logs alone is rejected; it requires secrets read.
    permissions: ["enclave_config_logs", "enclave_project_config_secrets_read"],
});
new doppler.WorkplaceRole("workplace-role", { name: `${prefix} logs`, permissions: ["logs"] });

new doppler.projectmember.Group("project-group", {
    project: project.name,
    groupSlug: group.slug,
    role: "viewer",
    environments: [qa.slug],
});
new doppler.projectmember.ServiceAccount("project-service-account", {
    project: project.name,
    serviceAccountSlug: sa.slug,
    role: projectRole.identifier,
    environments: [qa.slug],
});

if (memberEmail) {
    const user = doppler.getUserOutput({ email: memberEmail });
    new doppler.GroupMember("group-member", { groupSlug: group.slug, userSlug: user.slug });
    new doppler.projectmember.User("project-user", {
        project: project.name,
        userSlug: user.slug,
        role: "viewer",
    });
}

new doppler.Webhook("webhook", {
    project: project.name,
    name: values.webhookName,
    url: "https://example.com/doppler-webhook",
    // Webhooks are always created enabled; upstream ignores enabled=false on create, so only
    // the update step disables it.
    enabled: values.webhookEnabled,
    enabledConfigs: [feature.name],
    authentication: { type: "Bearer", token: "integration-test-token" },
});

// Trusted IPs lock a config to those ranges for every caller, the provider included. Keep
// 0.0.0.0/0 and ::/0 in the list so the test can still read and delete the config.
new doppler.TrustedIps("trusted-ips", {
    project: project.name,
    config: feature.name,
    trustedIps: values.trustedIps,
});

new doppler.ChangeRequestPolicy("change-request-policy", {
    name: prefix,
    description: "Integration test policy",
    rules: { disallowSelfReview: true, requiredReviewers: [{ count: 1 }] },
    targets: { projects: [{ projectName: project.name, environmentSlugs: [uat.slug] }] },
});

// Functions. getSecrets reads after the secrets exist because its inputs come from them.
const secrets = doppler.getSecretsOutput({ project: apiKey.project, config: apiKey.config });
const environments = doppler.getEnvironmentsOutput({ project: qa.project });
const foundGroup = doppler.getGroupOutput({ slug: group.slug });

export const projectName = project.name;
export const environmentSlugs = environments.environments.apply(es => es.map(e => e.slug).sort());
export const apiKeyFromFunction = pulumi.unsecret(secrets.map.apply(m => m["API_KEY"]));
export const groupName = foundGroup.name;
export const serviceTokenIsSecret = pulumi.isSecret(serviceToken.key);
export const serviceAccountTokenIsSecret = pulumi.isSecret(saToken.apiKey);
