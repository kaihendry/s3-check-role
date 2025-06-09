package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// test prefixes for objects

func getS3ClientDefault(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg), nil
}

func printCurrentIdentity(ctx context.Context, t *testing.T) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Logf("Failed to load AWS config: %v", err)
		return
	}
	stsClient := sts.NewFromConfig(cfg)
	output, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		t.Logf("Failed to get caller identity: %v", err)
		return
	}
	t.Logf("Current AWS Identity: ARN=%s, Account=%s, UserId=%s", aws.ToString(output.Arn), aws.ToString(output.Account), aws.ToString(output.UserId))
}

func getAPAlias() string {
	alias := os.Getenv("AP_ALIAS")
	if alias == "" {
		panic("AP_ALIAS environment variable not set")
	}
	return alias
}

func TestAccessPointS3AccessMain_NoAssumeRole(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name            string
		bucket          string
		itemKeyOrPrefix string // Added for verbose logging
		operation       func(context.Context, *s3.Client, string) error
		expectAccessErr bool
	}{
		{
			name:            "List parent bucket should succeed (no assume role)",
			bucket:          "s3-check-role-2025",
			itemKeyOrPrefix: "",
			operation: func(ctx context.Context, client *s3.Client, bucket string) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucket),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name:            "List $prefix/ via parent bucket should succeed (no assume role)",
			bucket:          "s3-check-role-2025",
			itemKeyOrPrefix: fooPrefix,
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
			name:            "List $prefix/ via access point should not succeed (no assume role)",
			bucket:          getAPAlias(),
			itemKeyOrPrefix: fooPrefix,
			operation: func(ctx context.Context, client *s3.Client, bucket string) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucket),
					Prefix: aws.String(fooPrefix),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name:            "Get $prefix/test.txt via access point should succeed (no assume role)",
			bucket:          getAPAlias(),
			itemKeyOrPrefix: fooPrefix + "test.txt",
			operation: func(ctx context.Context, client *s3.Client, bucket string) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucket),
					Key:    aws.String(fooPrefix + "test.txt"),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name:            "List bar/ via access point should fail (no assume role)",
			bucket:          getAPAlias(),
			itemKeyOrPrefix: barPrefix,
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
			name:            "Get bar/test.txt via access point should fail (no assume role)",
			bucket:          getAPAlias(),
			itemKeyOrPrefix: barPrefix + "test.txt",
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

	printCurrentIdentity(ctx, t)

	client, err := getS3ClientDefault(ctx)
	if err != nil {
		t.Fatalf("Failed to create default S3 client: %v", err)
	}

	for _, tc := range tests {
		tc := tc // Capture tc for closure
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing S3 Bucket: %s", tc.bucket)
			if tc.itemKeyOrPrefix != "" {
				t.Logf("Target Item (Key/Prefix): %s", tc.itemKeyOrPrefix)
			} else {
				t.Logf("Target Item (Key/Prefix): <bucket root>")
			}
			err := tc.operation(ctx, client, tc.bucket)
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
