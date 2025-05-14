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

func testS3Access(t *testing.T, roleArn string, bucketName string) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		t.Fatalf("failed to load configuration, %v", err)
	}

	// Create STS client
	stsClient := sts.NewFromConfig(cfg)

	// Create credentials provider for assuming the role
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleArn)

	// Create new config with role credentials
	roleCfg := cfg.Copy()
	roleCfg.Credentials = aws.NewCredentialsCache(provider)

	// Create S3 client with role credentials and proper endpoint resolution mode
	s3Client := s3.NewFromConfig(roleCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	tests := []struct {
		name          string
		operation     func() error
		shouldSucceed bool
	}{
		{
			name: "ListBucket",
			operation: func() error {
				_, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
				})
				return err
			},
			shouldSucceed: true, // Both roles should be able to list
		},
		{
			name: "GetObject from foo/",
			operation: func() error {
				_, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("foo/test.txt"),
				})
				return err
			},
			shouldSucceed: strings.Contains(roleArn, "S3ReadOnlyRole"), // Only S3ReadOnlyRole should access foo/
		},
		{
			name: "GetObject from bar/",
			operation: func() error {
				_, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("bar/test.txt"),
				})
				return err
			},
			shouldSucceed: strings.Contains(roleArn, "dp-bar-consumer-rp"), // Only dp-bar-consumer-rp should access bar/
		},
		{
			name: "PutObject attempt",
			operation: func() error {
				_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("test.txt"),
					Body:   strings.NewReader("test"),
				})
				return err
			},
			shouldSucceed: false, // Neither role should be able to write
		},
	}

	for _, tt := range tests {
		testName := roleArn + " - " + tt.name
		t.Run(testName, func(t *testing.T) {
			err := tt.operation()
			if tt.shouldSucceed && err != nil {
				t.Errorf("%s: expected success, got error: %v", tt.name, err)
			}
			if !tt.shouldSucceed && err == nil {
				t.Errorf("%s: expected error, got success", tt.name)
			}
		})
	}
}

func TestActualS3Access(t *testing.T) {
	bucketName := "s3-check-role-2025"
	roles := []string{
		"arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
		"arn:aws:iam::407461997746:role/dp-bar-consumer-rp",
	}

	for _, roleArn := range roles {
		t.Run(roleArn, func(t *testing.T) {
			testS3Access(t, roleArn, bucketName)
		})
	}
}
