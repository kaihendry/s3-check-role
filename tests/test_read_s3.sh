#!/bin/bash

ROLE_ARN="arn:aws:iam::407461997746:role/S3ReadOnlyRole-s3-check-role-2025"
S3_PATH="s3://s3-check-role-2025/testdata.txt"
SESSION_NAME="S3ReadCheckSession"
OUTPUT_FILE="/tmp/testdata_downloaded.txt"

echo "Attempting to assume role: ${ROLE_ARN}"

# Assume the role
CREDENTIALS=$(aws sts assume-role --role-arn "${ROLE_ARN}" --role-session-name "${SESSION_NAME}" --query 'Credentials' --output json)

if [ $? -ne 0 ]; then
    echo "Failed to assume role ${ROLE_ARN}. Exiting."
    exit 1
fi

# Extract credentials
AWS_ACCESS_KEY_ID=$(echo "${CREDENTIALS}" | jq -r '.AccessKeyId')
AWS_SECRET_ACCESS_KEY=$(echo "${CREDENTIALS}" | jq -r '.SecretAccessKey')
AWS_SESSION_TOKEN=$(echo "${CREDENTIALS}" | jq -r '.SessionToken')

if [ -z "$AWS_ACCESS_KEY_ID" ] || [ -z "$AWS_SECRET_ACCESS_KEY" ] || [ -z "$AWS_SESSION_TOKEN" ]; then
    echo "Failed to extract credentials from assume-role output."
    exit 1
fi

# Export credentials as environment variables
export AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY
export AWS_SESSION_TOKEN

echo "Successfully assumed role. Temporary credentials exported."
echo "Attempting to read ${S3_PATH} using assumed role..."

# Attempt to read the S3 object
aws s3 cp "${S3_PATH}" "${OUTPUT_FILE}" --quiet

# Check the exit code
if [ $? -eq 0 ]; then
    echo "SUCCESS: Read operation successful for ${S3_PATH} with role ${ROLE_ARN}."
    rm "${OUTPUT_FILE}" # Clean up downloaded file
    EXIT_CODE=0
else
    echo "FAILURE: Read operation failed for ${S3_PATH} with role ${ROLE_ARN}."
    EXIT_CODE=1
fi

# Unset temporary credentials
unset AWS_ACCESS_KEY_ID
unset AWS_SECRET_ACCESS_KEY
unset AWS_SESSION_TOKEN

echo "Temporary credentials unset."

exit ${EXIT_CODE}
