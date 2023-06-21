package events

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type EventTargetArgs struct {
	Arn     pulumi.StringInput
	RoleArn pulumi.StringInput
	Rule    pulumi.StringInput
}

const template = `{"environmentVariablesOverride": [
	{"name":"pullRequestId","type":"PLAINTEXT","value":<pullRequestId>},
	{"name":"sourceCommit","type":"PLAINTEXT","value":<sourceCommit>},
	{"name":"destinationCommit","type":"PLAINTEXT","value":<destinationCommit>},
	{"name":"repositoryName","type":"PLAINTEXT","value":<repositoryName>},
	{"name":"revisionId","type":"PLAINTEXT","value":<revisionId>},
	{"name":"sourceReference","type":"PLAINTEXT","value":<sourceReference>},
	{"name":"region","type":"PLAINTEXT","value":<region>},
	{"name":"account","type":"PLAINTEXT","value":<account>},
	{"name":"id","type":"PLAINTEXT","value":<id>}]
  }`

var InputPaths = pulumi.StringMap{
	"sourceReference":   pulumi.String("$.detail.sourceReference"),
	"revisionId":        pulumi.String("$.detail.revisionId"),
	"sourceVersion":     pulumi.String("$.detail.sourceCommit"),
	"destinationCommit": pulumi.String("$.detail.destinationCommit"),
	"pullRequestId":     pulumi.String("$.detail.pullRequestId"),
	"repositoryName":    pulumi.String("$.detail.repositoryNames[0]"),
	"sourceCommit":      pulumi.String("$.detail.sourceCommit"),
	"region":            pulumi.String("$.region"),
	"account":           pulumi.String("$.account"),
	"id":                pulumi.String("$.id"),
}

// Create an EventBridge target that triggers the CodeBuild project
func NewEventTarget(ctx *pulumi.Context, args EventTargetArgs) (*cloudwatch.EventTarget, error) {
	return cloudwatch.NewEventTarget(ctx, "codebuild-trigger", &cloudwatch.EventTargetArgs{
		Arn: args.Arn,
		InputTransformer: &cloudwatch.EventTargetInputTransformerArgs{
			InputPaths:    InputPaths,
			InputTemplate: pulumi.String(template),
		},
		RoleArn: args.Arn,
		Rule:    args.Rule,
	})
}
