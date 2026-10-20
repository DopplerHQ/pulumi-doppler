import json
import re

import pulumi
import pulumi_doppler as doppler

# values.json holds everything that changes between test steps; step2/ replaces it.
with open("values.json") as f:
    values = json.load(f)
config = pulumi.Config()
# An existing workplace user to add to the group and project. Optional.
member_email = config.get("memberEmail")

# Doppler names are unique per workplace, so suffix them with the stack name.
suffix = re.sub(r"[^a-z0-9]", "", pulumi.get_stack().lower())[-8:]
prefix = f"pt-py-{suffix}"

tag = doppler.Tag("tag", name=prefix, color=values["tagColor"])

project = doppler.Project("project", name=prefix, description=values["projectDescription"])
doppler.ProjectTag("project-tag", project=project.name, tag_slug=tag.slug)

# Create every environment used here: workplaces differ in which defaults new projects get.
qa = doppler.Environment("qa", project=project.name, slug="qa", name="QA")
uat = doppler.Environment("uat", project=project.name, slug="uat", name="UAT")
feature = doppler.BranchConfig("qa-feature", project=project.name, environment=qa.slug, name="qa_feature")

api_key = doppler.Secret("api-key", project=project.name, config=qa.slug, name="API_KEY", value=values["apiKey"])
# Refreshing a restricted secret only works with a service token (issue #15).
doppler.Secret(
    "db-password",
    project=project.name,
    config=qa.slug,
    name="DB_PASSWORD",
    value="hunter2",
    visibility="restricted",
)
doppler.SecretNote(
    "api-key-note",
    project=project.name,
    name=api_key.name,
    note="Owned by the provider integration test",
)

service_token = doppler.ServiceToken("service-token", project=project.name, config=qa.slug, name=prefix, access="read")

sa = doppler.ServiceAccount("service-account", name=prefix, workplace_role="collaborator")
sa_token = doppler.ServiceAccountToken("service-account-token", service_account_slug=sa.slug, name=prefix)
doppler.ServiceAccountIdentity(
    "service-account-identity",
    service_account_slug=sa.slug,
    name=prefix,
    ttl_seconds=600,
    config_oidc=doppler.ServiceAccountIdentityConfigOidcArgs(
        discovery_url="https://token.actions.githubusercontent.com",
        claims=[
            doppler.ServiceAccountIdentityConfigOidcClaimArgs(key="aud", values=["https://github.com/DopplerHQ"]),
            doppler.ServiceAccountIdentityConfigOidcClaimArgs(
                key="sub", values=["repo:DopplerHQ/pulumi-doppler:ref:refs/heads/main"]
            ),
        ],
    ),
)

group = doppler.Group("group", name=prefix)
project_role = doppler.ProjectRole(
    "project-role",
    name=f"{prefix} log viewer",
    # enclave_config_logs alone is rejected; it requires secrets read.
    permissions=["enclave_config_logs", "enclave_project_config_secrets_read"],
)
doppler.WorkplaceRole("workplace-role", name=f"{prefix} logs", permissions=["logs"])

doppler.projectmember.Group(
    "project-group", project=project.name, group_slug=group.slug, role="viewer", environments=[qa.slug]
)
doppler.projectmember.ServiceAccount(
    "project-service-account",
    project=project.name,
    service_account_slug=sa.slug,
    role=project_role.identifier,
    environments=[qa.slug],
)

if member_email:
    user = doppler.get_user_output(email=member_email)
    doppler.GroupMember("group-member", group_slug=group.slug, user_slug=user.slug)
    doppler.projectmember.User("project-user", project=project.name, user_slug=user.slug, role="viewer")

doppler.Webhook(
    "webhook",
    project=project.name,
    name=values["webhookName"],
    url="https://example.com/doppler-webhook",
    # Webhooks are always created enabled; upstream ignores enabled=False on create, so only
    # the update step disables it.
    enabled=values["webhookEnabled"],
    enabled_configs=[feature.name],
    authentication=doppler.WebhookAuthenticationArgs(type="Bearer", token="integration-test-token"),
)

# Trusted IPs lock a config to those ranges for every caller, the provider included. Keep
# 0.0.0.0/0 and ::/0 in the list so the test can still read and delete the config.
doppler.TrustedIps("trusted-ips", project=project.name, config=feature.name, trusted_ips=values["trustedIps"])

doppler.ChangeRequestPolicy(
    "change-request-policy",
    name=prefix,
    description="Integration test policy",
    rules=doppler.ChangeRequestPolicyRulesArgs(
        disallow_self_review=True,
        required_reviewers=[doppler.ChangeRequestPolicyRulesRequiredReviewerArgs(count=1)],
    ),
    targets=doppler.ChangeRequestPolicyTargetsArgs(
        projects=[doppler.ChangeRequestPolicyTargetsProjectArgs(project_name=project.name, environment_slugs=[uat.slug])]
    ),
)

# Functions. get_secrets reads after the secrets exist because its inputs come from them.
secrets = doppler.get_secrets_output(project=api_key.project, config=api_key.config)
environments = doppler.get_environments_output(project=qa.project)
found_group = doppler.get_group_output(slug=group.slug)

pulumi.export("projectName", project.name)
pulumi.export("environmentSlugs", environments.environments.apply(lambda es: sorted(e.slug for e in es)))
pulumi.export("apiKeyFromFunction", pulumi.Output.unsecret(secrets.map.apply(lambda m: m["API_KEY"])))
pulumi.export("groupName", found_group.name)
pulumi.export("serviceTokenIsSecret", pulumi.Output.from_input(service_token.key.is_secret()))
pulumi.export("serviceAccountTokenIsSecret", pulumi.Output.from_input(sa_token.api_key.is_secret()))
