import unittest
import boto3
from dataclasses import dataclass
from typing import Callable
from botocore.exceptions import ClientError
import os


# Test prefixes for objects
FOO_PREFIX = "megalake/TANK/BLAH/foo/"  # MUST have trailing slash!!
BAR_PREFIX = "bar/"

# Use AP_ALIAS from environment if set, else default to the current value
AP_ALIAS = os.environ.get("AP_ALIAS", "s3-check-role-2025-a-woxf6459ho7bn61ydx5f74iwbk4qoeuw2b-s3alias")


@dataclass
class S3TestCase:
    """Dataclass representing a test case for S3 access testing."""

    name: str
    role_arn: str
    bucket: str
    item_key_or_prefix: str
    operation: Callable[[boto3.client, str], None]
    expect_access_err: bool


class TestS3BucketPolicy(unittest.TestCase):
    """Test S3 bucket policy access using different roles and access points."""

    def get_s3_client_for_role(self, role_arn: str) -> boto3.client:
        """Create an S3 client using assumed role credentials."""
        sts_client = boto3.client("sts")

        # Assume the role
        response = sts_client.assume_role(
            RoleArn=role_arn, RoleSessionName="s3-test-session"
        )

        credentials = response["Credentials"]

        # Create S3 client with assumed role credentials
        s3_client = boto3.client(
            "s3",
            aws_access_key_id=credentials["AccessKeyId"],
            aws_secret_access_key=credentials["SecretAccessKey"],
            aws_session_token=credentials["SessionToken"],
        )

        return s3_client

    def operation_list_bucket_root(self, client: boto3.client, bucket: str) -> None:
        """List objects in bucket root."""
        client.list_objects_v2(Bucket=bucket)

    def operation_list_foo_prefix_via_bucket(
        self, client: boto3.client, bucket: str
    ) -> None:
        """List objects with foo/ prefix via parent bucket."""
        client.list_objects_v2(Bucket=bucket, Prefix="foo/")

    def operation_list_foo_prefix_via_access_point(
        self, client: boto3.client, bucket: str
    ) -> None:
        """List objects with foo/ prefix via access point."""
        client.list_objects_v2(Bucket=bucket, Prefix=FOO_PREFIX)

    def operation_get_foo_test_file(self, client: boto3.client, bucket: str) -> None:
        """Get foo/test.txt file via access point."""
        client.get_object(Bucket=bucket, Key=FOO_PREFIX + "test.txt")

    def operation_list_bar_prefix_via_access_point(
        self, client: boto3.client, bucket: str
    ) -> None:
        """List objects with bar/ prefix via access point (should fail)."""
        client.list_objects_v2(Bucket=bucket, Prefix=BAR_PREFIX)

    def operation_get_bar_test_file(self, client: boto3.client, bucket: str) -> None:
        """Get bar/test.txt file via access point (should fail)."""
        client.get_object(Bucket=bucket, Key=BAR_PREFIX + "test.txt")

    def get_test_cases(self) -> list[S3TestCase]:
        """Return list of test cases."""
        return [
            S3TestCase(
                name="List parent bucket should not succeed",
                role_arn="arn:aws:iam::407461997746:role/foo-via-access-point",
                bucket="s3-check-role-2025",
                item_key_or_prefix="",  # Listing bucket root, no specific prefix/key
                operation=self.operation_list_bucket_root,
                expect_access_err=True,  # Changed to True - direct bucket access should be denied
            ),
            S3TestCase(
                name="List foo/ via parent bucket should not succeed",
                role_arn="arn:aws:iam::407461997746:role/foo-via-access-point",
                bucket="s3-check-role-2025",
                item_key_or_prefix="foo/",
                operation=self.operation_list_foo_prefix_via_bucket,
                expect_access_err=True,  # Changed to True - direct bucket access should be denied
            ),
            S3TestCase(
                name="List foo/ via access point should succeed",
                role_arn="arn:aws:iam::407461997746:role/foo-via-access-point",
                bucket=AP_ALIAS,
                item_key_or_prefix=FOO_PREFIX,
                operation=self.operation_list_foo_prefix_via_access_point,
                expect_access_err=False,
            ),
            S3TestCase(
                name="Get foo/test.txt via access point should succeed",
                role_arn="arn:aws:iam::407461997746:role/foo-via-access-point",
                bucket=AP_ALIAS,
                item_key_or_prefix=FOO_PREFIX + "test.txt",
                operation=self.operation_get_foo_test_file,
                expect_access_err=False,
            ),
            S3TestCase(
                name="List bar/ via access point should fail",
                role_arn="arn:aws:iam::407461997746:role/foo-via-access-point",
                bucket=AP_ALIAS,
                item_key_or_prefix=BAR_PREFIX,
                operation=self.operation_list_bar_prefix_via_access_point,
                expect_access_err=True,
            ),
            S3TestCase(
                name="Get bar/test.txt via access point should fail",
                role_arn="arn:aws:iam::407461997746:role/foo-via-access-point",
                bucket=AP_ALIAS,
                item_key_or_prefix=BAR_PREFIX + "test.txt",
                operation=self.operation_get_bar_test_file,
                expect_access_err=True,
            ),
        ]

    def _run_test_case(self, test_case: S3TestCase):
        """Helper method to run a single test case."""
        print(f"\nTesting: {test_case.name}")
        print(f"Assumed role ARN: {test_case.role_arn}")
        print(f"Testing S3 Bucket: {test_case.bucket}")

        if test_case.item_key_or_prefix:
            print(f"Target Item (Key/Prefix): {test_case.item_key_or_prefix}")
        else:
            print("Target Item (Key/Prefix): <bucket root>")

        # Get S3 client for the role
        try:
            client = self.get_s3_client_for_role(test_case.role_arn)
        except Exception as e:
            self.fail(
                f"Failed to create S3 client for role ARN {test_case.role_arn}: {e}"
            )

        # Execute the operation
        try:
            test_case.operation(client, test_case.bucket)
            operation_error = None
        except ClientError as e:
            operation_error = e
        except Exception as e:
            operation_error = e

        # Assert based on expected outcome
        if test_case.expect_access_err:
            self.assertIsNotNone(
                operation_error,
                f"Expected access denied error but operation succeeded for test: {test_case.name}",
            )
            if operation_error:
                error_code = (
                    getattr(operation_error, "response", {})
                    .get("Error", {})
                    .get("Code", "")
                )
                error_message = str(operation_error)
                self.assertTrue(
                    "AccessDenied" in error_code or "AccessDenied" in error_message,
                    f"Expected AccessDenied error but got: {operation_error} for test: {test_case.name}",
                )
        else:
            self.assertIsNone(
                operation_error,
                f"Expected operation to succeed but got error: {operation_error} for test: {test_case.name}",
            )

    def test_list_parent_bucket_should_not_succeed(self):
        """Test: List parent bucket should not succeed."""
        test_cases = self.get_test_cases()
        self._run_test_case(test_cases[0])

    def test_list_foo_via_parent_bucket_should_not_succeed(self):
        """Test: List foo/ via parent bucket should not succeed."""
        test_cases = self.get_test_cases()
        self._run_test_case(test_cases[1])

    def test_list_foo_via_access_point_should_succeed(self):
        """Test: List foo/ via access point should succeed."""
        test_cases = self.get_test_cases()
        self._run_test_case(test_cases[2])

    def test_get_foo_test_txt_via_access_point_should_succeed(self):
        """Test: Get foo/test.txt via access point should succeed."""
        test_cases = self.get_test_cases()
        self._run_test_case(test_cases[3])

    def test_list_bar_via_access_point_should_fail(self):
        """Test: List bar/ via access point should fail."""
        test_cases = self.get_test_cases()
        self._run_test_case(test_cases[4])

    def test_get_bar_test_txt_via_access_point_should_fail(self):
        """Test: Get bar/test.txt via access point should fail."""
        test_cases = self.get_test_cases()
        self._run_test_case(test_cases[5])


if __name__ == "__main__":
    unittest.main(verbosity=2)
