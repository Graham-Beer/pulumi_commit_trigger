package Role

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const CodeBuildPolicy = pulumi.String(`{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": ["sts:AssumeRole"],
			"Effect": "Allow",
			"Principal": {
				"Service": ["codebuild.amazonaws.com"]
			}
		}
	]
}`)

func NewCodeBuildRole(ctx *pulumi.Context, policyarn pulumi.StringOutput) (*iam.Role, error) {
	return iam.NewRole(ctx, "codeBuildRole", &iam.RoleArgs{
		AssumeRolePolicy: CodeBuildPolicy,
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/AWSCodeBuildAdminAccess"),
			pulumi.String("arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"),
			pulumi.String("arn:aws:iam::aws:policy/AmazonS3FullAccess"),
			pulumi.StringOutput(policyarn),
		},
	})
}
