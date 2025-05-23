PRODUCT = aptest

.PHONY: fmt plan apply destroy

.terraform:
	terraform init

apply: .terraform fmt
	terraform apply -auto-approve
	aws s3 ls --recursive s3://$(shell terraform output -raw bucket_name)/

fmt:
	terraform fmt

plan: .terraform
	terraform plan

destroy: .terraform
	terraform destroy

inspect-access:
	aws s3api get-bucket-policy --bucket s3-check-role-2025 --query Policy --output text | jq .
	aws s3control list-access-points --account-id $$(aws sts get-caller-identity --query Account --output text)
	aws s3control get-access-point-policy \
  		--account-id $$(aws sts get-caller-identity --query Account --output text) \
  		--name s3-check-role-2025-ap | jq -r .Policy | jq .

export-creds:
	$(eval CREDS := $(shell aws sts assume-role --role-arn arn:aws:iam::407461997746:role/foo-via-access-point --role-session-name test-session --duration-seconds 3600))
	@echo "[default]" > $(PRODUCT)-test-consume-credentials
	@echo "aws_access_key_id = $$(echo '$(CREDS)' | jq -r .Credentials.AccessKeyId)" >> $(PRODUCT)-test-consume-credentials
	@echo "aws_secret_access_key = $$(echo '$(CREDS)' | jq -r .Credentials.SecretAccessKey)" >> $(PRODUCT)-test-consume-credentials
	@echo "aws_session_token = $$(echo '$(CREDS)' | jq -r .Credentials.SessionToken)" >> $(PRODUCT)-test-consume-credentials
	@echo "Credentials written to $(PRODUCT)-test-consume-credentials"
	@echo "Use: export AWS_SHARED_CREDENTIALS_FILE=$$PWD/$(PRODUCT)-test-consume-credentials"
