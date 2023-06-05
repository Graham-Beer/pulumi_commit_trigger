package main

import (
	"os"

	Policy "CommitTrigger/pkg/aws/Iam/policy"
	Role "CommitTrigger/pkg/aws/Iam/role"
	Codebuilder "CommitTrigger/pkg/aws/codebuild"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codecommit"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codepipeline"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
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
