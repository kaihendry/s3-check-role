import pytest


# --- Add command-line options ---
def pytest_addoption(parser):
    parser.addoption(
        "--role-arn", action="store", required=True, help="ARN of the IAM role to test"
    )
    parser.addoption(
        "--bucket-name", action="store", required=True, help="Name of the S3 bucket"
    )


# --- Fixtures to get values from options ---
@pytest.fixture(scope="session")
def role_arn(request):
    return request.config.getoption("--role-arn")


@pytest.fixture(scope="session")
def bucket_name(request):
    return request.config.getoption("--bucket-name")


# --- Fixture for the wildcard resource ARN ---
@pytest.fixture(scope="session")
def bucket_resource_arn(bucket_name):
    return f"arn:aws:s3:::{bucket_name}/*"  # Target all objects
