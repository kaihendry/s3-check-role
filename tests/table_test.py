import pytest
import boto3

# Replace with your actual AWS region
AWS_REGION = "eu-west-2"
iam_client = boto3.client("iam", region_name=AWS_REGION)


def simulate_permission(
    role_arn: str,
    principal_type: str,
    action: str,
    resource_arn: str,
    context: dict = None,
) -> str:
    """
    Simulates IAM permissions for a given role, action, and resource.
    Returns the IAM evaluation decision: 'allowed', 'explicitDeny', or 'implicitDeny'
    """
    try:
        response = iam_client.simulate_principal_policy(
            PolicySourceArn=role_arn,
            ActionNames=[action],
            ResourceArns=[resource_arn],
        )
        if response["EvaluationResults"]:
            return response["EvaluationResults"][0]["EvalDecision"]  # Remove .lower()
        return "implicitDeny"
    except Exception as e:
        print(f"Error simulating policy: {e}")
        return "implicitDeny"


@pytest.mark.parametrize(
    "test_case_name, action, resource_arn, expected_decision",
    [
        (
            "ReadOnlyRoleCanGetObject",
            "s3:GetObject",
            "arn:aws:s3:::s3-check-role-2025/example.txt",
            "allowed",
        ),
        (
            "ReadOnlyRoleCanListBucket",
            "s3:ListBucket",
            "arn:aws:s3:::s3-check-role-2025",
            "allowed",
        ),
        (
            "ReadOnlyRoleCannotPutObject",
            "s3:PutObject",
            "arn:aws:s3:::s3-check-role-2025/example.txt",
            "implicitDeny",
        ),
        (
            "ReadOnlyRoleCannotDeleteObject",
            "s3:DeleteObject",
            "arn:aws:s3:::s3-check-role-2025/example.txt",
            "implicitDeny",
        ),
    ],
)
def test_s3_permissions(
    test_case_name,
    action,
    resource_arn,
    expected_decision,
    role_arn,
    permission_results,
    s3_client,  # Add this fixture
):
    """
    Test S3 permissions against the read-only role
    """
    # Set role information
    permission_results.set_role(role_arn)

    # Try to get bucket policy
    try:
        policy = s3_client.get_bucket_policy(
            Bucket=resource_arn.split(":")[5].split("/")[0]
        )
        permission_results.set_resource_policy(policy.get("Policy"))
    except:
        pass  # No bucket policy exists

    actual_decision = simulate_permission(
        role_arn=role_arn,
        principal_type="Role",
        action=action,
        resource_arn=resource_arn,
    )

    permission_results.add_result(
        test_case_name,
        action,
        resource_arn,  # Now storing full resource ARN
        expected_decision,
        actual_decision,
    )

    assert (
        actual_decision == expected_decision
    ), f"Test Case: {test_case_name} - Expected decision to be {expected_decision}, got {actual_decision}"
