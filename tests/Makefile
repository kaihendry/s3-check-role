ROLE_ARN ?= "arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025" # Provide defaults or ensure they are set
BUCKET_NAME ?= "s3-check-role-2025"

# Run all tests
cd test:
	uv run pytest -v --junitxml=report.xml --role-arn=$(ROLE_ARN) --bucket-name=$(BUCKET_NAME)

# Test specific role types using markers
test-input:
	uv run pytest -v -m input_role --junitxml=report-input.xml --role-arn=$(ROLE_ARN) --bucket-name=$(BUCKET_NAME)

test-control:
	uv run pytest -v -m control_role --junitxml=report-control.xml --role-arn=$(ROLE_ARN) --bucket-name=$(BUCKET_NAME)

test-output:
	uv run pytest -v -m output_role --junitxml=report-output.xml --role-arn=$(ROLE_ARN) --bucket-name=$(BUCKET_NAME)