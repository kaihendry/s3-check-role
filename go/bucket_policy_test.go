package main

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestBucketPolicyS3Access(t *testing.T) {
	ctx := context.TODO()
	bucketName := "s3-check-role-2025"
	roleArn := "arn:aws:iam::407461997746:role/foo-via-bucket-policy"

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
			name: "Listing /foo/ should succeed",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
					Prefix: aws.String("foo/"),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name: "Get /foo/test.txt should succeed",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("foo/test.txt"),
				})
				return err
			},
			expectAccessErr: false,
		},
		{
			name: "List /bar/ should fail",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
					Prefix: aws.String("bar/"),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name: "Get object outside /foo should fail",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("bar/test.txt"),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name: "List with parent path /fo should fail",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
					Prefix: aws.String("fo"),
				})
				return err
			},
			expectAccessErr: true,
		},
		{
			name: "List with similar path /foobar should fail",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
					Prefix: aws.String("foobar/"),
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
