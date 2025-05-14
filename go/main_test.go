package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

func simulatePrincipalPolicy(client *iam.Client, roleArn string, action string, resourceArn string) (string, error) {
	input := &iam.SimulatePrincipalPolicyInput{
		PolicySourceArn: aws.String(roleArn),
		ActionNames:     []string{action},
		ResourceArns:    []string{resourceArn},
	}

	result, err := client.SimulatePrincipalPolicy(context.TODO(), input)
	if err != nil {
		return "", err
	}

	if len(result.EvaluationResults) < 1 {
		return "", fmt.Errorf("no evaluation results found")
	}

	if len(result.EvaluationResults) > 1 {
		return "", fmt.Errorf("multiple evaluation results found")
	}

	return string(result.EvaluationResults[0].EvalDecision), nil
}

func TestSimulatePrincipalPolicy(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		t.Fatalf("failed to load configuration, %v", err)
	}

	client := iam.NewFromConfig(cfg)

	tests := []struct {
		name           string
		roleArn        string
		action         string
		resourceArn    string
		expectedResult string
	}{
		{
			name:           "GetObject wrong prefix",
			roleArn:        "arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
			action:         "s3:GetObject",
			resourceArn:    "arn:aws:s3:::s3-check-role-2025/example.txt",
			expectedResult: "implicitDeny",
		},
		{
			name:           "GetObject correct prefix",
			roleArn:        "arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
			action:         "s3:GetObject",
			resourceArn:    "arn:aws:s3:::s3-check-role-2025/foo/example.txt",
			expectedResult: "allowed",
		},
		{
			name:           "Deny PutObject",
			roleArn:        "arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
			action:         "s3:PutObject",
			resourceArn:    "arn:aws:s3:::s3-check-role-2025/example.txt",
			expectedResult: "implicitDeny",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := simulatePrincipalPolicy(client, tt.roleArn, tt.action, tt.resourceArn)
			if err != nil {
				t.Fatalf("failed to simulate policy, %v", err)
			}

			if result != tt.expectedResult {
				t.Errorf("expected %s, got %s", tt.expectedResult, result)
			}
		})
	}
}
