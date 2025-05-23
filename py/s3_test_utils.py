import boto3
from typing import Callable, Any
from dataclasses import dataclass


# Define constants for ARNs, bucket name, and access point alias
BUCKET_NAME = "s3-check-role-2025"
ACCESS_POINT_ALIAS = "s3-check-role-2025-a-ns86askpr5cwp5kqkmjrmznbmpjaaeuw2b-s3alias"
ROLE_ARN_BUCKET_POLICY = "arn:aws:iam::407461997746:role/foo-via-bucket-policy"
ROLE_ARN_ACCESS_POINT = "arn:aws:iam::407461997746:role/foo-via-access-point"


@dataclass
class S3Test:
    name: str
    operation: Callable[[Any], None]
    expect_access_err: bool


def get_s3_client_for_role(role_arn: str) -> boto3.client:
    sts = boto3.client("sts")
    assumed_role = sts.assume_role(RoleArn=role_arn, RoleSessionName="s3_access_test")
    credentials = assumed_role["Credentials"]

    # Remove path style addressing
    return boto3.client(
        "s3",
        aws_access_key_id=credentials["AccessKeyId"],
        aws_secret_access_key=credentials["SecretAccessKey"],
        aws_session_token=credentials["SessionToken"],
    )
