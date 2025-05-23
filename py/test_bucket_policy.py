import pytest
from botocore.exceptions import ClientError
from s3_test_utils import (
    S3Test,
    get_s3_client_for_role,
    BUCKET_NAME,
    ROLE_ARN_BUCKET_POLICY,
)


@pytest.fixture(scope="module")
def s3_client():
    return get_s3_client_for_role(ROLE_ARN_BUCKET_POLICY)


@pytest.mark.parametrize(
    "test_case",
    [
        S3Test(
            name="List top level Bucket Contents should fail",
            operation=lambda client: client.list_objects_v2(Bucket=BUCKET_NAME),
            expect_access_err=True,
        ),
        S3Test(
            name="Listing foo/ should succeed",
            operation=lambda client: client.list_objects_v2(
                Bucket=BUCKET_NAME, Prefix="foo/"
            ),
            expect_access_err=False,
        ),
        S3Test(
            name="Get foo/test.txt should succeed",
            operation=lambda client: client.get_object(
                Bucket=BUCKET_NAME, Key="foo/test.txt"
            ),
            expect_access_err=False,
        ),
        S3Test(
            name="List bar/ should fail",
            operation=lambda client: client.list_objects_v2(
                Bucket=BUCKET_NAME, Prefix="bar/"
            ),
            expect_access_err=True,
        ),
        S3Test(
            name="Get object outside /foo should fail",
            operation=lambda client: client.get_object(
                Bucket=BUCKET_NAME, Key="bar/test.txt"
            ),
            expect_access_err=True,
        ),
        S3Test(
            name="List with parent path fo should fail",
            operation=lambda client: client.list_objects_v2(
                Bucket=BUCKET_NAME, Prefix="fo"
            ),
            expect_access_err=True,
        ),
        S3Test(
            name="List with similar path foobar/ should fail",
            operation=lambda client: client.list_objects_v2(
                Bucket=BUCKET_NAME, Prefix="foobar/"
            ),
            expect_access_err=True,
        ),
    ],
)
def test_bucket_policy_s3_access(test_case: S3Test, s3_client):
    try:
        test_case.operation(s3_client)
        if test_case.expect_access_err:
            pytest.fail("Expected access denied error but operation succeeded")
    except ClientError as e:
        if not test_case.expect_access_err:
            pytest.fail(f"Expected operation to succeed but got error: {e}")
        error_code = e.response["Error"]["Code"]
        if test_case.expect_access_err and error_code != "AccessDenied":
            pytest.fail(f"Expected AccessDenied error but got: {error_code}")
