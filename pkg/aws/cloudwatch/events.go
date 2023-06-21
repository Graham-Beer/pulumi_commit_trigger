package events

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type EventArgs struct {
	RoleArn      pulumi.StringOutput
	EventPattern pulumi.StringOutput
}

func CreateEventPattern(arn pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.Sprintf(`{
		"source": [
		  "aws.codecommit"
		],
		"detail-type": [
		  "CodeCommit Pull Request State Change"
		],
		"resources": [
			"%s"
		],
		"detail": {
			"event": [
				"pullRequestCreated",
				"pullRequestUpdated",
				"pullRequestSourceBranchUpdated"
			]
		}
	}`, arn)
}

// Create an EventBridge rule for the pull request events
func CwNewEventRule(ctx *pulumi.Context, ev EventArgs) (*cloudwatch.EventRule, error) {
	return cloudwatch.NewEventRule(ctx, "pull-request-event-rule", &cloudwatch.EventRuleArgs{
		Name:         pulumi.String("pull-request-Codebuild-event-rule"),
		Description:  pulumi.String("Trigger CodeBuild project on pull request events from CodeCommit"),
		EventPattern: pulumi.StringPtrInput(ev.EventPattern),
		RoleArn:      pulumi.StringOutput(ev.RoleArn),
	})
}
