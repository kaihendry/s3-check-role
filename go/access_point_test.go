package main

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestAccessPointS3Access(t *testing.T) {
	ctx := context.TODO()
	bucketName := "s3-check-role-2025"
	accessPointAlias := "s3-check-role-2025-a-ns86askpr5cwp5kqkmjrmznbmpjaaeuw2b-s3alias"
	roleArn := "arn:aws:iam::407461997746:role/foo-via-access-point"

	tests := []struct {
		name            string
		operation       func(context.Context, *s3.Client) error
		expectAccessErr bool
	}{
		{
			name: "List top level Bucket Contents should fail",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name: "Listing foo/ directly from bucket should fail",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
					Prefix: aws.String("foo/"),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name: "Get foo/test.txt directly from bucket should fail",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("foo/test.txt"),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name: "List foo/ via access point should succeed",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(accessPointAlias),
					Prefix: aws.String("foo/"),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name: "Get foo/test.txt via access point should succeed",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(accessPointAlias),
					Key:    aws.String("foo/test.txt"),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name: "List bar/ via access point should fail",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(accessPointAlias),
					Prefix: aws.String("bar/"),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name: "Get bar/test.txt via access point should fail",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(accessPointAlias),
					Key:    aws.String("bar/test.txt"),
				})
				return err
			},
			expectAccessErr: true,
		},
	}

	client, err := getS3ClientForRole(ctx, roleArn)
	if err != nil {
		t.Fatalf("Failed to create S3 client for role ARN %s: %v", roleArn, err)
	}

	for _, tc := range tests {
		tc := tc // Capture tc for use in closure
		t.Run(tc.name, func(t *testing.T) {
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
