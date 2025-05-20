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