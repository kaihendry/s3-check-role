package main

import (
	"context"
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
	bucketName := "s3-check-role-2025"

	tests := []struct {
		name          string
		roleArn       string
		operation     func(context.Context, *s3.Client) error
		shouldSucceed bool
	}{
		{
			name:    "List top level Bucket Contents should fail",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-bucket-policy",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
				})
				return err
			},
			shouldSucceed: false,
		},
		{
			name:    "Listing /foo/ should succeed",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-bucket-policy",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
					Prefix: aws.String("foo/"),
				})
				return err
			},
			shouldSucceed: true,
		},
		// check if role can get foo/test.txt
		{
			name:    "Get /foo/test.txt should succeed",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-bucket-policy",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("foo/test.txt"),
				})
				return err
			},
			shouldSucceed: true,
		},
		// check that arn:aws:iam::407461997746:role/foo-via-access-point cannot access s3://s3-check-role-2025/ or s3://s3-check-role-2025/foo/test.txt
		{
			name:    "List top level Bucket Contents should fail",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-access-point",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
				})
				return err
			},
			shouldSucceed: false,
		},
		{
			name:    "Listing /foo/ should fail",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-access-point",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
					Prefix: aws.String("foo/"),
				})
				return err
			},
			shouldSucceed: false,
		},
		// check if role can get foo/test.txt
		{
			name:    "Get /foo/test.txt should fail",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-access-point",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("foo/test.txt"),
				})
				return err
			},
			shouldSucceed: false,
		},
		{
			name:    "List via access point should succeed",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-access-point",
			operation: func(ctx context.Context, client *s3.Client) error {
				accessPointAlias := "s3-check-role-2025-a-ns86askpr5cwp5kqkmjrmznbmpjaaeuw2b-s3alias"
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(accessPointAlias),
				})
				return err
			},
			shouldSucceed: true,
		},
		{
			name:    "Get /foo/test.txt via access point should succeed",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-access-point",
			operation: func(ctx context.Context, client *s3.Client) error {
				accessPointAlias := "s3-check-role-2025-a-ns86askpr5cwp5kqkmjrmznbmpjaaeuw2b-s3alias"
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(accessPointAlias),
					Key:    aws.String("foo/test.txt"),
				})
				return err
			},
			shouldSucceed: true,
		},
	}

	clients := make(map[string]*s3.Client)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reuse client for same role
			client, exists := clients[tt.roleArn]
			if !exists {
				var err error
				client, err = getS3ClientForRole(ctx, tt.roleArn)
				if err != nil {
					t.Fatalf("Failed to create S3 client for role ARN %s: %v", tt.roleArn, err)
				}
				clients[tt.roleArn] = client
			}

			err := tt.operation(ctx, client)
			if tt.shouldSucceed && err != nil {
				t.Errorf("Expected success but got error: %v", err)
			}
			if !tt.shouldSucceed && err == nil {
				t.Errorf("Expected error but operation succeeded")
			}
		})
	}
}
