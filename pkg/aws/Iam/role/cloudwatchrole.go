package Role

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const CwBuildPolicy = pulumi.String(`{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "events.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}`)

func NewCwBuildRole(ctx *pulumi.Context, policyarn pulumi.StringOutput) (*iam.Role, error) {
	return iam.NewRole(ctx, "cloudWatchRole", &iam.RoleArgs{
		AssumeRolePolicy: CwBuildPolicy,
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/AWSCodeBuildDeveloperAccess"),
			pulumi.StringOutput(policyarn),
		},
	})
}
