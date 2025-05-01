ROLE_ARN ?= "arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025" # Provide defaults or ensure they are set
BUCKET_NAME ?= "s3-check-role-2025"

test:
	uv run pytest -v --role-arn=$(ROLE_ARN) --bucket-name=$(BUCKET_NAME)