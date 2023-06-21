package Policy

import (
	"encoding/json"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GitPol() (string, error) {
	pol, err := json.Marshal(map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Action": []string{
					"codecommit:GitPull",
					"codecommit:GitPush",
				},
				"Effect":   "Allow",
				"Resource": "*",
			},
		},
	})
	if err != nil {
		return "", err
	}

	return string(pol), nil
}

func NewGitPolicy(ctx *pulumi.Context) (*iam.Policy, error) {
	// GitPol() to generate policy to allow pull and push
	pol, err := GitPol()
	if err != nil {
		return nil, err
	}

	// return generated policy
	return iam.NewPolicy(ctx, "git-policy", &iam.PolicyArgs{
		Name:        pulumi.String("GitPullPolicy2"),
		Path:        pulumi.String("/"),
		Description: pulumi.String("Policy to allow git pull"),
		Policy:      pulumi.String(pol),
	})
}
