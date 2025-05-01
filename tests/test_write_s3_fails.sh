#!/bin/bash

ROLE_ARN="arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025"
S3_BUCKET="s3://s3-check-role-2025"
SESSION_NAME="S3WriteCheckSession"
DUMMY_FILE="/tmp/dummy_upload.txt"

echo "Creating dummy file for upload attempt..."
touch "${DUMMY_FILE}"

echo "Attempting to assume role: ${ROLE_ARN}"

# Assume the role
CREDENTIALS=$(aws sts assume-role --role-arn "${ROLE_ARN}" --role-session-name "${SESSION_NAME}" --query 'Credentials' --output json)

if [ $? -ne 0 ]; then
    echo "Failed to assume role ${ROLE_ARN}. Exiting."
    rm "${DUMMY_FILE}"
    exit 1
fi

# Extract credentials
AWS_ACCESS_KEY_ID=$(echo "${CREDENTIALS}" | jq -r '.AccessKeyId')
AWS_SECRET_ACCESS_KEY=$(echo "${CREDENTIALS}" | jq -r '.SecretAccessKey')
AWS_SESSION_TOKEN=$(echo "${CREDENTIALS}" | jq -r '.SessionToken')

if [ -z "$AWS_ACCESS_KEY_ID" ] || [ -z "$AWS_SECRET_ACCESS_KEY" ] || [ -z "$AWS_SESSION_TOKEN" ]; then
    echo "Failed to extract credentials from assume-role output."
    rm "${DUMMY_FILE}"
    exit 1
fi

# Export credentials as environment variables
export AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY
export AWS_SESSION_TOKEN

echo "Successfully assumed role. Temporary credentials exported."
echo "Attempting to write ${DUMMY_FILE} to ${S3_BUCKET}/ using assumed role (this should fail)..."

# Attempt to write the dummy file to the S3 bucket
aws s3 cp "${DUMMY_FILE}" "${S3_BUCKET}/dummy_upload_test.txt" --quiet

# Check the exit code - we expect failure (non-zero exit code)
if [ $? -ne 0 ]; then
    echo "SUCCESS: Write operation failed as expected for role ${ROLE_ARN}."
    EXIT_CODE=0
else
    echo "FAILURE: Write operation unexpectedly succeeded for role ${ROLE_ARN}."
    # Clean up the unexpectedly uploaded file
    aws s3 rm "${S3_BUCKET}/dummy_upload_test.txt" --quiet
    EXIT_CODE=1
fi

# Clean up local dummy file
rm "${DUMMY_FILE}"

# Unset temporary credentials
unset AWS_ACCESS_KEY_ID
unset AWS_SECRET_ACCESS_KEY
unset AWS_SESSION_TOKEN

echo "Temporary credentials unset."

exit ${EXIT_CODE}
