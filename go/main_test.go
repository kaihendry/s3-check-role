package main

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func getS3ClientForRole(ctx context.Context, roleArn string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleArn)
	roleCfg := cfg.Copy()
	roleCfg.Credentials = aws.NewCredentialsCache(provider)

	return s3.NewFromConfig(roleCfg), nil
}

func TestS3Access(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name            string
		roleArn         string
		bucketName      string
		operation       func(context.Context, *s3.Client) error
		expectAccessErr bool
	}{
		{
			name:       "List top level Bucket Contents should fail",
			roleArn:    "arn:aws:iam::407461997746:role/foo-via-bucket-policy",
			bucketName: "s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String("s3-check-role-2025"),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name:       "Listing /foo/ should succeed",
			roleArn:    "arn:aws:iam::407461997746:role/foo-via-bucket-policy",
			bucketName: "s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String("s3-check-role-2025"),
					Prefix: aws.String("foo/"),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name:       "Get /foo/test.txt should succeed",
			roleArn:    "arn:aws:iam::407461997746:role/foo-via-bucket-policy",
			bucketName: "s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String("s3-check-role-2025"),
					Key:    aws.String("foo/test.txt"),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name:       "List top level Bucket Contents should fail",
			roleArn:    "arn:aws:iam::407461997746:role/foo-via-access-point",
			bucketName: "s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String("s3-check-role-2025"),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name:       "Listing /foo/ should fail",
			roleArn:    "arn:aws:iam::407461997746:role/foo-via-access-point",
			bucketName: "s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String("s3-check-role-2025"),
					Prefix: aws.String("foo/"),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name:       "Get /foo/test.txt should fail",
			roleArn:    "arn:aws:iam::407461997746:role/foo-via-access-point",
			bucketName: "s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String("s3-check-role-2025"),
					Key:    aws.String("foo/test.txt"),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name:       "List /foo via access point should succeed",
			roleArn:    "arn:aws:iam::407461997746:role/foo-via-access-point",
			bucketName: "s3-check-role-2025-a-ns86askpr5cwp5kqkmjrmznbmpjaaeuw2b-s3alias",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String("s3-check-role-2025-a-ns86askpr5cwp5kqkmjrmznbmpjaaeuw2b-s3alias"),
					Prefix: aws.String("foo/"),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name:       "Get /foo/test.txt via access point should succeed",
			roleArn:    "arn:aws:iam::407461997746:role/foo-via-access-point",
			bucketName: "s3-check-role-2025-a-ns86askpr5cwp5kqkmjrmznbmpjaaeuw2b-s3alias",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String("s3-check-role-2025-a-ns86askpr5cwp5kqkmjrmznbmpjaaeuw2b-s3alias"),
					Key:    aws.String("foo/test.txt"),
				})
				return err
			},
			expectAccessErr: false,
		},
	}

	clients := make(map[string]*s3.Client)
	for _, tc := range tests {
		tc := tc // Capture tc for use in closure
		t.Run(tc.name, func(t *testing.T) {
			// Reuse client for same role
			client, exists := clients[tc.roleArn]
			if !exists {
				var err error
				client, err = getS3ClientForRole(ctx, tc.roleArn)
				if err != nil {
					t.Fatalf("Failed to create S3 client for role ARN %s: %v", tc.roleArn, err)
				}
				clients[tc.roleArn] = client
			}

			err := tc.operation(ctx, client)

			if tc.expectAccessErr {
				if err == nil {
					t.Error("Expected access denied error but operation succeeded")
				} else if !strings.Contains(err.Error(), "AccessDenied") {
					t.Errorf("Expected AccessDenied error but got: %v", err)
				}
			} else if err != nil {
				t.Errorf("Expected operation to succeed but got error: %v", err)
			}
		})
	}
}
