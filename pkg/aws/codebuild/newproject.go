package Codebuilder

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codebuild"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Codebuild struct {
	ComputeType string
	Image       string
	Type        string
}

type Repository struct {
	Name pulumi.StringOutput
}

type File struct {
	Buildspec []byte
}

type ServiceRole struct {
	Role pulumi.StringOutput
}

type NewProjectConfig struct {
	Cbuild Codebuild
	Repo   Repository
	Build  File
	Srole  ServiceRole
}

// Create a CodeBuild project
func NewCbProject(ctx *pulumi.Context, np NewProjectConfig) (*codebuild.Project, error) {
	return codebuild.NewProject(ctx, "my-codebuild", &codebuild.ProjectArgs{
		Name:         pulumi.String("my-codebuild"),
		BadgeEnabled: pulumi.Bool(true),
		Artifacts: codebuild.ProjectArtifactsArgs{
			Type: pulumi.String("NO_ARTIFACTS"),
		},
		Environment: codebuild.ProjectEnvironmentArgs{
			ComputeType: pulumi.String(np.Cbuild.ComputeType),
			Image:       pulumi.String(np.Cbuild.Image),
			Type:        pulumi.String(np.Cbuild.Type),
		},
		Source: codebuild.ProjectSourceArgs{
			Buildspec: pulumi.String(np.Build.Buildspec),
			Type:      pulumi.String("CODECOMMIT"),
			Location: pulumi.Sprintf("https://git-codecommit.%s.amazonaws.com/v1/repos/%s",
				"eu-west-1", np.Repo.Name),
		},
		ServiceRole: pulumi.StringInput(np.Srole.Role),

		Tags: pulumi.StringMap{
			"ManagedBy": pulumi.String("Pulumi"),
		},
		ProjectVisibility: pulumi.String("PUBLIC_READ"),
	})
}
