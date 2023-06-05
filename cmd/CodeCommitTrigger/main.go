package main

import (
	"os"

	Policy "CommitTrigger/pkg/aws/Iam/policy"
	Role "CommitTrigger/pkg/aws/Iam/role"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codebuild"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codecommit"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codepipeline"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

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
			RepositoryName: pulumi.String("my-repo"),
		})
		if err != nil {
			return err
		}

		codepipelineBucket, err := s3.NewBucketV2(ctx, "codepipelineBucket", &s3.BucketV2Args{
			ForceDestroy: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// Buildspec YAML content
		buildspecYaml, err := os.ReadFile("buildspec.yaml")
		if err != nil {
			return err
		}

		// Create a CodeBuild project
		project, err := codebuild.NewProject(ctx, "my-codebuild", &codebuild.ProjectArgs{
			Name: pulumi.String("my-codebuild"),
			Artifacts: codebuild.ProjectArtifactsArgs{
				Type: pulumi.String("NO_ARTIFACTS"),
			},
			Environment: codebuild.ProjectEnvironmentArgs{
				ComputeType: pulumi.String("BUILD_GENERAL1_SMALL"),
				Image:       pulumi.String("aws/codebuild/amazonlinux2-x86_64-standard:5.0"),
				Type:        pulumi.String("LINUX_CONTAINER"),
			},
			Source: codebuild.ProjectSourceArgs{
				Buildspec: pulumi.String(buildspecYaml),
				Type:      pulumi.String("CODECOMMIT"),
				Location: pulumi.Sprintf("https://git-codecommit.%s.amazonaws.com/v1/repos/%s",
					"eu-west-1", repo.RepositoryName),
			},
			ServiceRole: pulumi.StringInput(codeBuildRole.Arn),

			Tags: pulumi.StringMap{
				"ManagedBy": pulumi.String("Pulumi"),
			},
		})
		if err != nil {
			return err
		}

		// Create pipeline role
		pipelineRole, err := iam.NewRole(ctx, "pipeline", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
"Version": "2012-10-17",
"Statement": [
	{
		"Action": "sts:AssumeRole",
		"Principal": {
			"Service": "codepipeline.amazonaws.com"
		},
		"Effect": "Allow",
		"Sid": ""
	}
]
}`),
		})
		if err != nil {
			return err
		}

		// Attach policy to pipeline role
		_, err = iam.NewRolePolicyAttachment(ctx, "pipelineAttachment", &iam.RolePolicyAttachmentArgs{
			Role:      pipelineRole.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AdministratorAccess"),
		})
		if err != nil {
			return err
		}

		// Create a CodePipeline triggered by the CodeCommit repository
		_, err = codepipeline.NewPipeline(ctx, "my-pipeline", &codepipeline.PipelineArgs{
			ArtifactStores: codepipeline.PipelineArtifactStoreArray{
				codepipeline.PipelineArtifactStoreArgs{
					Location: codepipelineBucket.Bucket,
					Type:     pulumi.String("S3"),
				},
			},
			RoleArn: pipelineRole.Arn,
			Stages: codepipeline.PipelineStageArray{
				&codepipeline.PipelineStageArgs{
					// Source stage
					Name: pulumi.String("CodeCommit"),
					Actions: codepipeline.PipelineStageActionArray{
						codepipeline.PipelineStageActionArgs{
							Name:     pulumi.String("SourceAction"),
							Category: pulumi.String("Source"),
							Owner:    pulumi.String("AWS"),
							Provider: pulumi.String("CodeCommit"),
							Version:  pulumi.String("1"),
							OutputArtifacts: pulumi.StringArray{
								pulumi.String("source_output"),
							},
							Configuration: pulumi.StringMap{
								"RepositoryName": repo.RepositoryName,
								"BranchName":     pulumi.String("main"),
							},
						},
					},
				},
				&codepipeline.PipelineStageArgs{
					Actions: codepipeline.PipelineStageActionArray{
						&codepipeline.PipelineStageActionArgs{
							Configuration: pulumi.StringMap{
								"ProjectName": project.Name,
							},
							InputArtifacts: pulumi.StringArray{
								pulumi.String("source_output"),
							},
							Name: pulumi.String("CodeBuild"),
							OutputArtifacts: pulumi.StringArray{
								pulumi.String("output"),
							},
							Category: pulumi.String("Build"),
							Owner:    pulumi.String("AWS"),
							Provider: pulumi.String("CodeBuild"),
							RunOrder: pulumi.Int(1),
							Version:  pulumi.String("1"),
						},
					},
					Name: pulumi.String("Build"),
				},
			},
			Name: pulumi.String("my-pipeline"),
		})
		if err != nil {
			return err
		}

		ctx.Export("repositoryCloneUrlHttp", repo.CloneUrlHttp)
		ctx.Export("repositoryCloneUrlSsh", repo.CloneUrlSsh)
		ctx.Export("codeBuildProjectName", project.Name)
		return nil
	})
}
