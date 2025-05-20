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

	return s3.NewFromConfig(roleCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
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
			name:    "S3ReadOnlyRole - List Bucket Contents",
			roleArn: "arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
				})
				return err
			},
			shouldSucceed: true,
		},
		{
			name:    "S3ReadOnlyRole - Get Object from foo/",
			roleArn: "arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("foo/test.txt"),
				})
				return err
			},
			shouldSucceed: true,
		},
		{
			name:    "S3ReadOnlyRole - Get Object from bar/",
			roleArn: "arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("bar/test.txt"),
				})
				return err
			},
			shouldSucceed: false,
		},
		{
			name:    "S3ReadOnlyRole - Put Object Attempt to foo/",
			roleArn: "arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.PutObject(ctx, &s3.PutObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("foo/test.txt"),
					Body:   strings.NewReader("test"),
				})
				return err
			},
			shouldSucceed: false,
		},
		{
			name:    "BarConsumerRole - List Bucket Contents",
			roleArn: "arn:aws:iam::407461997746:role/dp-bar-consumer-rp",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
				})
				return err
			},
			shouldSucceed: true,
		},
		{
			name:    "BarConsumerRole - Get Object from bar/",
			roleArn: "arn:aws:iam::407461997746:role/dp-bar-consumer-rp",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("bar/test.txt"),
				})
				return err
			},
			shouldSucceed: true,
		},
		{
			name:    "BarConsumerRole - Get Object from foo/",
			roleArn: "arn:aws:iam::407461997746:role/dp-bar-consumer-rp",
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
			name:    "BarConsumerRole - Put Object Attempt to bar/",
			roleArn: "arn:aws:iam::407461997746:role/dp-bar-consumer-rp",
			operation: func(ctx context.Context, client *s3.Client) error {
				_, err := client.PutObject(ctx, &s3.PutObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String("bar/test.txt"),
					Body:   strings.NewReader("test"),
				})
				return err
			},
			shouldSucceed: false,
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
