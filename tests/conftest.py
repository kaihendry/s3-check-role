import pytest
import boto3
import json


# --- Add command-line options ---
def pytest_addoption(parser):
    parser.addoption(
        "--role-arn", action="store", required=True, help="ARN of the IAM role to test"
    )
    parser.addoption(
        "--bucket-name", action="store", required=True, help="Name of the S3 bucket"
    )
    parser.addoption(
        "--role-type",
        action="store",
        default="standard",
        choices=["input", "control", "output"],
        help="Type of role being tested (input, control, output)",
    )


# --- Fixtures to get values from options ---
@pytest.fixture(scope="session")
def role_arn(request):
    return request.config.getoption("--role-arn")


@pytest.fixture(scope="session")
def bucket_name(request):
    return request.config.getoption("--bucket-name")


@pytest.fixture(scope="session")
def role_type(request):
    return request.config.getoption("--role-type")


# --- Fixture for the wildcard resource ARN ---
@pytest.fixture(scope="session")
def bucket_resource_arn(bucket_name):
    return f"arn:aws:s3:::{bucket_name}/*"  # Target all objects


class PermissionTestResults:
    def __init__(self):
        self.results = []
        self.role_arn = None
        self.resource_policy = None

    def set_role(self, role_arn):
        self.role_arn = role_arn

    def set_resource_policy(self, policy):
        self.resource_policy = policy

    def add_result(self, test_case, action, resource, expected, actual):
        self.results.append(
            {
                "test_case": test_case,
                "action": action,
                "resource": resource,
                "expected": expected,
                "actual": actual,
                "passed": expected == actual,
            }
        )


@pytest.fixture(scope="session")
def permission_results(request):
    """Create and attach results collector to the session"""
    results = PermissionTestResults()
    request.session.permission_results = results  # Attach to session
    return results


@pytest.fixture(scope="session")
def s3_client():
    """Provides a boto3 S3 client."""
    return boto3.client("s3")


def pytest_sessionfinish(session, exitstatus):
    """Print detailed permission test results in a table format"""
    if hasattr(session, "permission_results"):
        results = session.permission_results
        if results and results.results:
            print("\n=== S3 Permission Test Results ===")
            role_name = results.role_arn.split("/")[-1]
            print(f"Role Name: {role_name}")
            print(f"Role ARN:  {results.role_arn}")

            if results.resource_policy:
                print("\nBucket Policy:")
                print(json.dumps(json.loads(results.resource_policy), indent=2))

            print("\nPermission Test Results:")
            print("-" * 120)
            print(
                f"{'Action':<15} {'Resource ARN':<65} {'Expected':<15} {'Result':<15}"
            )
            print("-" * 120)

            for r in results.results:
                action = r["action"].split(":")[1]
                print(
                    f"{action:<15} "
                    f"{r['resource']:<65} "
                    f"{r['expected']:<15} "
                    f"{'✓ ' + r['actual'] if r['passed'] else '✗ ' + r['actual']}"
                )

            print("-" * 120)
            passed = sum(1 for r in results.results if r["passed"])
            total = len(results.results)
            print(f"Results: {passed}/{total} tests passed")


# Register custom markers for role types
def pytest_configure(config):
    """Register custom markers."""
    config.addinivalue_line(
        "markers", "input_role: mark test as requiring input role permissions"
    )
    config.addinivalue_line(
        "markers", "control_role: mark test as requiring control role permissions"
    )
    config.addinivalue_line(
        "markers", "output_role: mark test as requiring output role permissions"
    )


# Helper function for permission simulation - move from table_test.py to conftest.py
def simulate_permission(
    role_arn: str,
    action: str,
    resource_arn: str,
) -> str:
    """
    Simulates IAM permissions for a given role, action, and resource.
    Returns the IAM evaluation decision: 'allowed', 'explicitDeny', or 'implicitDeny'
    """
    import boto3

    iam_client = boto3.client("iam")

    try:
        response = iam_client.simulate_principal_policy(
            PolicySourceArn=role_arn,
            ActionNames=[action],
            ResourceArns=[resource_arn],
        )
        if response["EvaluationResults"]:
            return response["EvaluationResults"][0]["EvalDecision"]
        return "implicitDeny"
    except Exception as e:
        print(f"Error simulating policy: {e}")
        return "implicitDeny"
