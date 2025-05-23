package main

import (
	"context"

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
