package main

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/DopplerHQ/pulumi-doppler/sdk/go/doppler"
	"github.com/DopplerHQ/pulumi-doppler/sdk/go/doppler/projectmember"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// values.json holds everything that changes between test steps; step2/ replaces it.
type stepValues struct {
	ProjectDescription string   `json:"projectDescription"`
	APIKey             string   `json:"apiKey"`
	TagColor           string   `json:"tagColor"`
	TrustedIPs         []string `json:"trustedIps"`
	WebhookName        string   `json:"webhookName"`
	WebhookEnabled     bool     `json:"webhookEnabled"`
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		raw, err := os.ReadFile("values.json")
		if err != nil {
			return err
		}
		var values stepValues
		if err := json.Unmarshal(raw, &values); err != nil {
			return err
		}
		// An existing workplace user to add to the group and project. Optional.
		memberEmail := config.Get(ctx, "memberEmail")

		// Doppler names are unique per workplace, so suffix them with the stack name.
		suffix := regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(strings.ToLower(ctx.Stack()), "")
		if len(suffix) > 8 {
			suffix = suffix[len(suffix)-8:]
		}
		prefix := "pt-go-" + suffix

		tag, err := doppler.NewTag(ctx, "tag", &doppler.TagArgs{
			Name:  pulumi.String(prefix),
			Color: pulumi.String(values.TagColor),
		})
		if err != nil {
			return err
		}

		project, err := doppler.NewProject(ctx, "project", &doppler.ProjectArgs{
			Name:        pulumi.String(prefix),
			Description: pulumi.String(values.ProjectDescription),
		})
		if err != nil {
			return err
		}
		if _, err := doppler.NewProjectTag(ctx, "project-tag", &doppler.ProjectTagArgs{
			Project: project.Name,
			TagSlug: tag.Slug,
		}); err != nil {
			return err
		}

		// Create every environment used here: workplaces differ in which defaults new projects get.
		qa, err := doppler.NewEnvironment(ctx, "qa", &doppler.EnvironmentArgs{
			Project: project.Name,
			Slug:    pulumi.String("qa"),
			Name:    pulumi.String("QA"),
		})
		if err != nil {
			return err
		}
		uat, err := doppler.NewEnvironment(ctx, "uat", &doppler.EnvironmentArgs{
			Project: project.Name,
			Slug:    pulumi.String("uat"),
			Name:    pulumi.String("UAT"),
		})
		if err != nil {
			return err
		}
		feature, err := doppler.NewBranchConfig(ctx, "qa-feature", &doppler.BranchConfigArgs{
			Project:     project.Name,
			Environment: qa.Slug,
			Name:        pulumi.String("qa_feature"),
		})
		if err != nil {
			return err
		}

		apiKey, err := doppler.NewSecret(ctx, "api-key", &doppler.SecretArgs{
			Project: project.Name,
			Config:  qa.Slug,
			Name:    pulumi.String("API_KEY"),
			Value:   pulumi.String(values.APIKey),
		})
		if err != nil {
			return err
		}
		// Refreshing a restricted secret only works with a service token (issue #15).
		if _, err := doppler.NewSecret(ctx, "db-password", &doppler.SecretArgs{
			Project:    project.Name,
			Config:     qa.Slug,
			Name:       pulumi.String("DB_PASSWORD"),
			Value:      pulumi.String("hunter2"),
			Visibility: pulumi.String("restricted"),
		}); err != nil {
			return err
		}
		if _, err := doppler.NewSecretNote(ctx, "api-key-note", &doppler.SecretNoteArgs{
			Project: project.Name,
			Name:    apiKey.Name,
			Note:    pulumi.String("Owned by the provider integration test"),
		}); err != nil {
			return err
		}

		serviceToken, err := doppler.NewServiceToken(ctx, "service-token", &doppler.ServiceTokenArgs{
			Project: project.Name,
			Config:  qa.Slug,
			Name:    pulumi.String(prefix),
			Access:  pulumi.String("read"),
		})
		if err != nil {
			return err
		}

		sa, err := doppler.NewServiceAccount(ctx, "service-account", &doppler.ServiceAccountArgs{
			Name:          pulumi.String(prefix),
			WorkplaceRole: pulumi.String("collaborator"),
		})
		if err != nil {
			return err
		}
		saToken, err := doppler.NewServiceAccountToken(ctx, "service-account-token", &doppler.ServiceAccountTokenArgs{
			ServiceAccountSlug: sa.Slug,
			Name:               pulumi.String(prefix),
		})
		if err != nil {
			return err
		}
		if _, err := doppler.NewServiceAccountIdentity(ctx, "service-account-identity", &doppler.ServiceAccountIdentityArgs{
			ServiceAccountSlug: sa.Slug,
			Name:               pulumi.String(prefix),
			TtlSeconds:         pulumi.Int(600),
			ConfigOidc: &doppler.ServiceAccountIdentityConfigOidcArgs{
				DiscoveryUrl: pulumi.String("https://token.actions.githubusercontent.com"),
				Claims: doppler.ServiceAccountIdentityConfigOidcClaimArray{
					&doppler.ServiceAccountIdentityConfigOidcClaimArgs{
						Key:    pulumi.String("aud"),
						Values: pulumi.StringArray{pulumi.String("https://github.com/DopplerHQ")},
					},
					&doppler.ServiceAccountIdentityConfigOidcClaimArgs{
						Key:    pulumi.String("sub"),
						Values: pulumi.StringArray{pulumi.String("repo:DopplerHQ/pulumi-doppler:ref:refs/heads/main")},
					},
				},
			},
		}); err != nil {
			return err
		}

		group, err := doppler.NewGroup(ctx, "group", &doppler.GroupArgs{Name: pulumi.String(prefix)})
		if err != nil {
			return err
		}
		projectRole, err := doppler.NewProjectRole(ctx, "project-role", &doppler.ProjectRoleArgs{
			Name: pulumi.String(prefix + " log viewer"),
			// enclave_config_logs alone is rejected; it requires secrets read.
			Permissions: pulumi.StringArray{
				pulumi.String("enclave_config_logs"),
				pulumi.String("enclave_project_config_secrets_read"),
			},
		})
		if err != nil {
			return err
		}
		if _, err := doppler.NewWorkplaceRole(ctx, "workplace-role", &doppler.WorkplaceRoleArgs{
			Name:        pulumi.String(prefix + " logs"),
			Permissions: pulumi.StringArray{pulumi.String("logs")},
		}); err != nil {
			return err
		}

		if _, err := projectmember.NewGroup(ctx, "project-group", &projectmember.GroupArgs{
			Project:      project.Name,
			GroupSlug:    group.Slug,
			Role:         pulumi.String("viewer"),
			Environments: pulumi.StringArray{qa.Slug},
		}); err != nil {
			return err
		}
		if _, err := projectmember.NewServiceAccount(ctx, "project-service-account", &projectmember.ServiceAccountArgs{
			Project:            project.Name,
			ServiceAccountSlug: sa.Slug,
			Role:               projectRole.Identifier,
			Environments:       pulumi.StringArray{qa.Slug},
		}); err != nil {
			return err
		}

		if memberEmail != "" {
			user := doppler.GetUserOutput(ctx, doppler.GetUserOutputArgs{Email: pulumi.String(memberEmail)})
			if _, err := doppler.NewGroupMember(ctx, "group-member", &doppler.GroupMemberArgs{
				GroupSlug: group.Slug,
				UserSlug:  user.Slug(),
			}); err != nil {
				return err
			}
			if _, err := projectmember.NewUser(ctx, "project-user", &projectmember.UserArgs{
				Project:  project.Name,
				UserSlug: user.Slug(),
				Role:     pulumi.String("viewer"),
			}); err != nil {
				return err
			}
		}

		if _, err := doppler.NewWebhook(ctx, "webhook", &doppler.WebhookArgs{
			Project: project.Name,
			Name:    pulumi.String(values.WebhookName),
			Url:     pulumi.String("https://example.com/doppler-webhook"),
			// Webhooks are always created enabled; upstream ignores enabled=false on create, so
			// only the update step disables it.
			Enabled:        pulumi.Bool(values.WebhookEnabled),
			EnabledConfigs: pulumi.StringArray{feature.Name},
			Authentication: &doppler.WebhookAuthenticationArgs{
				Type:  pulumi.String("Bearer"),
				Token: pulumi.String("integration-test-token"),
			},
		}); err != nil {
			return err
		}

		// Trusted IPs lock a config to those ranges for every caller, the provider included.
		// Keep 0.0.0.0/0 and ::/0 in the list so the test can still read and delete the config.
		if _, err := doppler.NewTrustedIps(ctx, "trusted-ips", &doppler.TrustedIpsArgs{
			Project:    project.Name,
			Config:     feature.Name,
			TrustedIps: pulumi.ToStringArray(values.TrustedIPs),
		}); err != nil {
			return err
		}

		if _, err := doppler.NewChangeRequestPolicy(ctx, "change-request-policy", &doppler.ChangeRequestPolicyArgs{
			Name:        pulumi.String(prefix),
			Description: pulumi.String("Integration test policy"),
			Rules: &doppler.ChangeRequestPolicyRulesArgs{
				DisallowSelfReview: pulumi.Bool(true),
				RequiredReviewers: doppler.ChangeRequestPolicyRulesRequiredReviewerArray{
					&doppler.ChangeRequestPolicyRulesRequiredReviewerArgs{Count: pulumi.Int(1)},
				},
			},
			Targets: &doppler.ChangeRequestPolicyTargetsArgs{
				Projects: doppler.ChangeRequestPolicyTargetsProjectArray{
					&doppler.ChangeRequestPolicyTargetsProjectArgs{
						ProjectName:      project.Name,
						EnvironmentSlugs: pulumi.StringArray{uat.Slug},
					},
				},
			},
		}); err != nil {
			return err
		}

		// Functions. GetSecrets reads after the secrets exist because its inputs come from them.
		secrets := doppler.GetSecretsOutput(ctx, doppler.GetSecretsOutputArgs{Project: apiKey.Project, Config: apiKey.Config})
		environments := doppler.GetEnvironmentsOutput(ctx, doppler.GetEnvironmentsOutputArgs{Project: qa.Project})
		foundGroup := doppler.LookupGroupOutput(ctx, doppler.LookupGroupOutputArgs{Slug: group.Slug})

		ctx.Export("projectName", project.Name)
		ctx.Export("environmentSlugs", environments.Environments().ApplyT(func(es []doppler.GetEnvironmentsEnvironment) []string {
			slugs := make([]string, len(es))
			for i, e := range es {
				slugs[i] = e.Slug
			}
			sort.Strings(slugs)
			return slugs
		}))
		ctx.Export("apiKeyFromFunction", pulumi.Unsecret(secrets.Map().MapIndex(pulumi.String("API_KEY"))))
		ctx.Export("groupName", foundGroup.Name())
		ctx.Export("serviceTokenIsSecret", pulumi.Bool(pulumi.IsSecret(serviceToken.Key)))
		ctx.Export("serviceAccountTokenIsSecret", pulumi.Bool(pulumi.IsSecret(saToken.ApiKey)))
		return nil
	})
}
