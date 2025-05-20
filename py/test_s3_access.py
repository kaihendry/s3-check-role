import boto3
import pytest
from botocore.config import Config
from typing import Callable, Any
import os
from dataclasses import dataclass
from botocore.exceptions import ClientError


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

    return boto3.client(
        "s3",
        aws_access_key_id=credentials["AccessKeyId"],
        aws_secret_access_key=credentials["SecretAccessKey"],
        aws_session_token=credentials["SessionToken"],
        config=Config(s3={"addressing_style": "path"}),
    )


@pytest.fixture(scope="module")
def s3_clients():
    return {}


@pytest.mark.parametrize(
    "test_case",
    [
        S3Test(
            name="S3ReadOnlyRole - List Bucket Contents",
            role_arn="arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
            operation=lambda client: client.list_objects_v2(
                Bucket="s3-check-role-2025"
            ),
            should_succeed=True,
        ),
        S3Test(
            name="S3ReadOnlyRole - Get Object from foo/",
            role_arn="arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
            operation=lambda client: client.get_object(
                Bucket="s3-check-role-2025", Key="foo/test.txt"
            ),
            should_succeed=True,
        ),
        S3Test(
            name="S3ReadOnlyRole - Get Object from bar/",
            role_arn="arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
            operation=lambda client: client.get_object(
                Bucket="s3-check-role-2025", Key="bar/test.txt"
            ),
            should_succeed=False,
        ),
        S3Test(
            name="S3ReadOnlyRole - Put Object Attempt to foo/",
            role_arn="arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025",
            operation=lambda client: client.put_object(
                Bucket="s3-check-role-2025", Key="foo/test.txt", Body=b"test"
            ),
            should_succeed=False,
        ),
        S3Test(
            name="BarConsumerRole - List Bucket Contents",
            role_arn="arn:aws:iam::407461997746:role/dp-bar-consumer-rp",
            operation=lambda client: client.list_objects_v2(
                Bucket="s3-check-role-2025"
            ),
            should_succeed=True,
        ),
        S3Test(
            name="BarConsumerRole - Get Object from bar/",
            role_arn="arn:aws:iam::407461997746:role/dp-bar-consumer-rp",
            operation=lambda client: client.get_object(
                Bucket="s3-check-role-2025", Key="bar/test.txt"
            ),
            should_succeed=True,
        ),
        S3Test(
            name="BarConsumerRole - Get Object from foo/",
            role_arn="arn:aws:iam::407461997746:role/dp-bar-consumer-rp",
            operation=lambda client: client.get_object(
                Bucket="s3-check-role-2025", Key="foo/test.txt"
            ),
            should_succeed=False,
        ),
        S3Test(
            name="BarConsumerRole - Put Object Attempt to bar/",
            role_arn="arn:aws:iam::407461997746:role/dp-bar-consumer-rp",
            operation=lambda client: client.put_object(
                Bucket="s3-check-role-2025", Key="bar/test.txt", Body=b"test"
            ),
            should_succeed=False,
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
            pytest.fail(f"Expected operation to fail but it succeeded")
    except ClientError as e:
        if test_case.should_succeed:
            pytest.fail(f"Expected operation to succeed but got error: {e}")
