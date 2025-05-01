import boto3
import pytest
import json  # Added import
from botocore.exceptions import ClientError  # Added import

# Constants from Terraform/setup
ROLE_ARN = "arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025"
BUCKET_NAME = "s3-check-role-2025"
OBJECT_KEY = "testdata.txt"
OBJECT_ARN = f"arn:aws:s3:::{BUCKET_NAME}/{OBJECT_KEY}"
BUCKET_ARN = f"arn:aws:s3:::{BUCKET_NAME}/*"  # Policy might use wildcard


@pytest.fixture(scope="module")
def iam_client():
    """Provides a boto3 IAM client."""
    return boto3.client("iam")


@pytest.fixture(scope="module")
def s3_client():
    """Provides a boto3 S3 client."""
    return boto3.client("s3")


@pytest.fixture(scope="module")
def bucket_policy(s3_client):
    """Fetches the S3 bucket policy."""
    try:
        response = s3_client.get_bucket_policy(Bucket=BUCKET_NAME)
        return response.get("Policy")  # Returns None if no policy key
    except ClientError as e:
        if e.response["Error"]["Code"] == "NoSuchBucketPolicy":
            print(f"No bucket policy found for {BUCKET_NAME}.")
            return None  # No policy exists
        else:
            print(f"Error fetching bucket policy: {e}")
            raise  # Re-raise other errors


def test_readonly_role_can_read(iam_client, bucket_policy):
    """Simulate GetObject action for the role and verify it's allowed."""
    action = "s3:GetObject"
    print(f"\nSimulating {action} for {ROLE_ARN} on {OBJECT_ARN}")

    simulation_params = {
        "PolicySourceArn": ROLE_ARN,
        "ActionNames": [action],
        "ResourceArns": [OBJECT_ARN],
    }
    if bucket_policy:
        simulation_params["ResourcePolicy"] = bucket_policy

    response = iam_client.simulate_principal_policy(**simulation_params)

    assert len(response["EvaluationResults"]) == 1, "Expected one evaluation result"
    result = response["EvaluationResults"][0]
    print(f"Simulation result: {result['EvalDecision']}")

    assert (
        result["EvalDecision"] == "allowed"
    ), f"Expected {action} to be allowed, but got {result['EvalDecision']}"


def test_readonly_role_cannot_write(iam_client, bucket_policy):
    """Simulate PutObject action for the role and verify it's denied."""
    action = "s3:PutObject"
    print(f"\nSimulating {action} for {ROLE_ARN} on {OBJECT_ARN}")

    simulation_params = {
        "PolicySourceArn": ROLE_ARN,
        "ActionNames": [action],
        "ResourceArns": [OBJECT_ARN],
    }
    if bucket_policy:
        simulation_params["ResourcePolicy"] = bucket_policy

    response = iam_client.simulate_principal_policy(**simulation_params)

    assert len(response["EvaluationResults"]) == 1, "Expected one evaluation result"
    result = response["EvaluationResults"][0]
    print(f"Simulation result: {result['EvalDecision']}")

    assert (
        result["EvalDecision"] != "allowed"
    ), f"Expected {action} to be denied, but it was allowed"


def test_readonly_role_cannot_delete(iam_client, bucket_policy):
    """Simulate DeleteObject action for the role and verify it's denied."""
    action = "s3:DeleteObject"
    print(f"\nSimulating {action} for {ROLE_ARN} on {OBJECT_ARN}")

    simulation_params = {
        "PolicySourceArn": ROLE_ARN,
        "ActionNames": [action],
        "ResourceArns": [OBJECT_ARN],
    }
    if bucket_policy:
        simulation_params["ResourcePolicy"] = bucket_policy

    response = iam_client.simulate_principal_policy(**simulation_params)

    assert len(response["EvaluationResults"]) == 1, "Expected one evaluation result"
    result = response["EvaluationResults"][0]
    print(f"Simulation result: {result['EvalDecision']}")

    assert (
        result["EvalDecision"] != "allowed"
    ), f"Expected {action} to be denied, but it was allowed"
