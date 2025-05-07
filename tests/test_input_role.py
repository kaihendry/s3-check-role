import pytest
from conftest import simulate_permission


@pytest.mark.input_role
@pytest.mark.parametrize(
    "test_case_name, action, resource_arn, expected_decision",
    [
        (
            "InputRoleCanGetObject",
            "s3:GetObject",
            "arn:aws:s3:::s3-check-role-2025/example.txt",
            "allowed",
        ),
        (
            "InputRoleCanListBucket",
            "s3:ListBucket",
            "arn:aws:s3:::s3-check-role-2025",
            "allowed",
        ),
        (
            "InputRoleCannotPutObject",
            "s3:PutObject",
            "arn:aws:s3:::s3-check-role-2025/example.txt",
            "implicitDeny",
        ),
    ],
)
def test_input_role_permissions(
    test_case_name,
    action,
    resource_arn,
    expected_decision,
    role_arn,
    permission_results,
    s3_client,
):
    """Test input role permissions (should allow read but not write)"""
    # Set role information in results
    permission_results.set_role(role_arn)

    # Get bucket policy if available
    bucket_name = resource_arn.split(":")[5].split("/")[0]
    try:
        policy = s3_client.get_bucket_policy(Bucket=bucket_name)
        permission_results.set_resource_policy(policy.get("Policy"))
    except:
        pass  # No bucket policy

    # Simulate the permission
    actual_decision = simulate_permission(
        role_arn=role_arn,
        action=action,
        resource_arn=resource_arn,
    )

    # Record result
    permission_results.add_result(
        test_case_name, action, resource_arn, expected_decision, actual_decision
    )

    # Assert
    assert (
        actual_decision == expected_decision
    ), f"Test Case: {test_case_name} - Expected decision to be {expected_decision}, got {actual_decision}"
