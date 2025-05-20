import boto3
import pytest
from typing import Callable, Any
from dataclasses import dataclass
from botocore.exceptions import ClientError


# Define constants for ARNs, bucket name, and access point alias
BUCKET_NAME = "s3-check-role-2025"
ACCESS_POINT_ALIAS = "s3-check-role-2025-a-ns86askpr5cwp5kqkmjrmznbmpjaaeuw2b-s3alias"
ROLE_ARN_BUCKET_POLICY = "arn:aws:iam::407461997746:role/foo-via-bucket-policy"
ROLE_ARN_ACCESS_POINT = "arn:aws:iam::407461997746:role/foo-via-access-point"


@dataclass
class S3Test:
    name: str
    role_arn: str
    operation: Callable[[Any], None]
    should_succeed: bool


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


@pytest.fixture(scope="module")
def s3_clients():
    return {}


@pytest.mark.parametrize(
    "test_case",
    [
        # Tests for foo-via-bucket-policy role
        S3Test(
            name="BucketPolicyRole - List top level Bucket Contents should fail",
            role_arn=ROLE_ARN_BUCKET_POLICY,
            operation=lambda client: client.list_objects_v2(Bucket=BUCKET_NAME),
            should_succeed=False,
        ),
        S3Test(
            name="BucketPolicyRole - Listing /foo/ should succeed",
            role_arn=ROLE_ARN_BUCKET_POLICY,
            operation=lambda client: client.list_objects_v2(
                Bucket=BUCKET_NAME, Prefix="foo/"
            ),
            should_succeed=True,
        ),
        S3Test(
            name="BucketPolicyRole - Get /foo/test.txt should succeed",
            role_arn=ROLE_ARN_BUCKET_POLICY,
            operation=lambda client: client.get_object(
                Bucket=BUCKET_NAME, Key="foo/test.txt"
            ),
            should_succeed=True,
        ),
        # Tests for foo-via-access-point role (direct bucket access - should fail)
        S3Test(
            name="AccessPointRole - Direct List top level Bucket Contents should fail",
            role_arn=ROLE_ARN_ACCESS_POINT,
            operation=lambda client: client.list_objects_v2(Bucket=BUCKET_NAME),
            should_succeed=False,
        ),
        S3Test(
            name="AccessPointRole - Direct Listing /foo/ should fail",
            role_arn=ROLE_ARN_ACCESS_POINT,
            operation=lambda client: client.list_objects_v2(
                Bucket=BUCKET_NAME, Prefix="foo/"
            ),
            should_succeed=False,
        ),
        S3Test(
            name="AccessPointRole - Direct Get /foo/test.txt should fail",
            role_arn=ROLE_ARN_ACCESS_POINT,
            operation=lambda client: client.get_object(
                Bucket=BUCKET_NAME, Key="foo/test.txt"
            ),
            should_succeed=False,
        ),
        # Tests for foo-via-access-point role (via access point - should succeed)
        S3Test(
            name="AccessPointRole - List via access point should succeed",
            role_arn=ROLE_ARN_ACCESS_POINT,
            operation=lambda client: client.list_objects_v2(Bucket=ACCESS_POINT_ALIAS),
            should_succeed=True,
        ),
        S3Test(
            name="AccessPointRole - Get /foo/test.txt via access point should succeed",
            role_arn=ROLE_ARN_ACCESS_POINT,
            operation=lambda client: client.get_object(
                Bucket=ACCESS_POINT_ALIAS, Key="foo/test.txt"
            ),
            should_succeed=True,
        ),
    ],
)
def test_s3_access(test_case: S3Test, s3_clients: dict):
    # Reuse client for same role
    if test_case.role_arn not in s3_clients:
        s3_clients[test_case.role_arn] = get_s3_client_for_role(test_case.role_arn)

    client = s3_clients[test_case.role_arn]

    try:
        test_case.operation(client)
        if not test_case.should_succeed:
            pytest.fail("Expected operation to fail but it succeeded")
    except ClientError as e:
        if test_case.should_succeed:
            pytest.fail(f"Expected operation to succeed but got error: {e}")
