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
	aws s3api get-bucket-policy --bucket s3-check-role-2025 | jq
	aws iam list-attached-role-policies --role-name S3ReadOnlyRole-s3-check-role-2025
	aws iam get-policy-version --policy-arn $$(aws iam list-attached-role-policies --role-name S3ReadOnlyRole-s3-check-role-2025 --query 'AttachedPolicies[0].PolicyArn' --output text) --version-id $$(aws iam get-policy --policy-arn $$(aws iam list-attached-role-policies --role-name S3ReadOnlyRole-s3-check-role-2025 --query 'AttachedPolicies[0].PolicyArn' --output text) --query 'Policy.DefaultVersionId' --output text) | jq