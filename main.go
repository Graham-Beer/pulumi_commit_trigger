package main

import (
	"os"

	Policy "CommitTrigger/pkg/aws/Iam/policy"
	Role "CommitTrigger/pkg/aws/Iam/role"
	Events "CommitTrigger/pkg/aws/cloudwatch"
	Codebuilder "CommitTrigger/pkg/aws/codebuild"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codecommit"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		var cb *Codebuilder.Codebuild
		if err := cfg.TryObject("Codebuild", &cb); err != nil {
			return err
		}

		gitPol, err := Policy.NewGitPolicy(ctx)
		if err != nil {
			return err
		}

		// Create an IAM role for AWS CodeBuild
		codeBuildRole, err := Role.NewCodeBuildRole(ctx, gitPol.Arn)
		if err != nil {
			return err
		}

		// Create a CodeCommit repository
		repo, err := codecommit.NewRepository(ctx, "my-repo", &codecommit.RepositoryArgs{
			RepositoryName: pulumi.String("my-repo-2"),
			DefaultBranch:  pulumi.String("main"),
		})
		if err != nil {
			return err
		}

		// Buildspec YAML content
		buildspecYaml, err := os.ReadFile("buildspec.yaml")
		if err != nil {
			return err
		}

		cbProj := Codebuilder.NewProjectConfig{
			Cbuild: Codebuilder.Codebuild{
				ComputeType: cb.ComputeType,
				Image:       cb.Image,
				Type:        cb.Type,
			},
			Repo: Codebuilder.Repository{
				Name: repo.RepositoryName,
			},
			Build: Codebuilder.File{
				Buildspec: buildspecYaml,
			},
			Srole: Codebuilder.ServiceRole{
				Role: codeBuildRole.Arn,
			},
		}

		// Create a CodeBuild project
		project, err := Codebuilder.NewCbProject(ctx, cbProj)
		if err != nil {
			return err
		}

		// Polices and roles
		cwpolicy, err := Policy.NewPolicyForCWE(ctx, project.Arn)
		if err != nil {
			return err
		}

		cwrole, err := Role.NewCwBuildRole(ctx, cwpolicy.Arn)
		if err != nil {
			return err
		}

		// Define the event rule pattern for pull request events in CodeCommit
		cwEventPat := Events.CreateEventPattern(repo.Arn)

		// Create an EventBridge rule for the pull request events
		CwNewRule, err := Events.CwNewEventRule(ctx, Events.EventArgs{
			RoleArn:      pulumi.StringOutput(cwrole.Arn),
			EventPattern: pulumi.StringOutput(cwEventPat),
		})
		if err != nil {
			return err
		}

		TargetArgs := Events.EventTargetArgs{
			Arn:     project.Arn,
			RoleArn: cwrole.Arn,
			Rule:    CwNewRule.Name,
		}

		_, err = Events.NewEventTarget(ctx, TargetArgs)
		if err != nil {
			return err
		}

		ctx.Export("repositoryName", repo.RepositoryName)
		ctx.Export("codebuildProjectName", project.Name)
		return nil
	})
}
