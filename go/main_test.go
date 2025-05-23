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

// test prefixes for objects
var (
	fooPrefix = "datalake/TEMP/TRN/foo/" // MUST have trailing slash!!
	barPrefix = "bar/"
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

func TestAccessPointS3AccessMain(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name            string
		roleArn         string
		bucket          string
		operation       func(context.Context, *s3.Client, string) error
		expectAccessErr bool
	}{
		{
			name:    "List parent bucket should not succeed",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-access-point",
			bucket:  "s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client, bucket string) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucket),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name:    "List foo/ via parent bucket should not succeed",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-access-point",
			bucket:  "s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client, bucket string) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucket),
					Prefix: aws.String("foo/"),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name:    "List foo/ via access point should succeed",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-access-point",
			bucket:  "s3-check-role-2025-a-ardis4yekbwq7db1eewhxzxcfzwryeuw2b-s3alias",
			operation: func(ctx context.Context, client *s3.Client, bucket string) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucket),
					Prefix: aws.String(fooPrefix),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name:    "Get foo/test.txt via access point should succeed",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-access-point",
			bucket:  "s3-check-role-2025-a-ardis4yekbwq7db1eewhxzxcfzwryeuw2b-s3alias",
			operation: func(ctx context.Context, client *s3.Client, bucket string) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucket),
					Key:    aws.String(fooPrefix + "test.txt"),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name:    "List bar/ via access point should fail",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-access-point",
			bucket:  "s3-check-role-2025-a-ardis4yekbwq7db1eewhxzxcfzwryeuw2b-s3alias",
			operation: func(ctx context.Context, client *s3.Client, bucket string) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucket),
					Prefix: aws.String(barPrefix),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name:    "Get bar/test.txt via access point should fail",
			roleArn: "arn:aws:iam::407461997746:role/foo-via-access-point",
			bucket:  "s3-check-role-2025-a-ardis4yekbwq7db1eewhxzxcfzwryeuw2b-s3alias",
			operation: func(ctx context.Context, client *s3.Client, bucket string) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucket),
					Key:    aws.String(barPrefix + "test.txt"),
				})
				return err
			},
			expectAccessErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture tc for use in closure
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Assumed role ARN: %s", tc.roleArn)
			client, err := getS3ClientForRole(ctx, tc.roleArn)
			if err != nil {
				t.Fatalf("Failed to create S3 client for role ARN %s: %v", tc.roleArn, err)
			}
			err = tc.operation(ctx, client, tc.bucket)
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
