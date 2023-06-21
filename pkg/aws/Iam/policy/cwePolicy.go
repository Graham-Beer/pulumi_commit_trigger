package Policy

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func StartCodeBuild(cbResource pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.Sprintf(`{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": [
                    "codebuild:StartBuild"
                ],
                "Resource": [
                    "%s"
                ]
            }
        ]
    }`, cbResource)
}

func NewPolicyForCWE(ctx *pulumi.Context, codebuildrole pulumi.StringOutput) (*iam.Policy, error) {
	// Start CodeBuild Policy
	pol := StartCodeBuild(codebuildrole)

	// return generated policy
	return iam.NewPolicy(ctx, "cwe-policy", &iam.PolicyArgs{
		Name:        pulumi.String("CodeBuild-Invoke-Role-For-Cloudwatch-Events2"),
		Path:        pulumi.String("/"),
		Description: pulumi.String("Policy to allow CodeBuild to be invoked by Cloudwatch Events."),
		Policy:      pulumi.StringInput(pol),
	})
}
