.PHONY: fmt plan apply destroy

.terraform:
	terraform init

apply: .terraform fmt
	terraform apply -auto-approve

fmt:
	terraform fmt

plan: .terraform
	terraform plan

destroy: .terraform
	terraform destroy
