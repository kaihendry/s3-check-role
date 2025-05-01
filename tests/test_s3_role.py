import boto3
import pytest
import json
from botocore.exceptions import ClientError


@pytest.fixture(scope="module")
def iam_client():
    """Provides a boto3 IAM client."""
    return boto3.client("iam")


@pytest.fixture(scope="module")
def s3_client():
    """Provides a boto3 S3 client."""
    return boto3.client("s3")


@pytest.fixture(scope="module")
def bucket_policy(
    s3_client, bucket_name
):  # bucket_name fixture is now sourced from conftest.py
    """Fetches the S3 bucket policy."""
    try:
        response = s3_client.get_bucket_policy(Bucket=bucket_name)
        # Ensure policy is returned as a string, not dict/None
        policy_str = response.get("Policy")
        return policy_str if policy_str else None
    except ClientError as e:
        if e.response["Error"]["Code"] == "NoSuchBucketPolicy":
            print(f"No bucket policy found for {bucket_name}.")
            return None  # No policy exists
        else:
            print(f"Error fetching bucket policy: {e}")
            raise  # Re-raise other errors


@pytest.mark.parametrize(
    "action, expected_decision",
    [
        ("s3:GetObject", "allowed"),
        ("s3:PutObject", "denied"),  # Implicitly denied is okay
        ("s3:DeleteObject", "denied"),  # Implicitly denied is okay
    ],
)
def test_role_permissions(
    iam_client, bucket_policy, role_arn, bucket_resource_arn, action, expected_decision
):
    """Simulate actions for the role and verify expected outcome."""

    simulation_params = {
        "PolicySourceArn": role_arn,
        "ActionNames": [action],
        "ResourceArns": [bucket_resource_arn],
    }
    # Only add ResourcePolicy if it exists and is not None
    if bucket_policy:
        simulation_params["ResourcePolicy"] = bucket_policy

    response = iam_client.simulate_principal_policy(**simulation_params)

    assert len(response["EvaluationResults"]) == 1, "Expected one evaluation result"
    result = response["EvaluationResults"][0]
    actual_decision = result["EvalDecision"]

    if expected_decision == "allowed":
        assert (
            actual_decision == "allowed"
        ), f"Expected {action} to be allowed, but got {actual_decision}"
    else:  # expected_decision == "denied"
        # We check it's NOT allowed (could be explicitly or implicitly denied)
        assert (
            actual_decision != "allowed"
        ), f"Expected {action} to be denied, but it was allowed"
