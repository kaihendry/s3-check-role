variable "aws_region" {
  description = "The AWS region to create resources in."
  type        = string
  default     = "eu-west-2"
}

variable "bucket_name" {
  description = "The exact name for the S3 bucket."
  type        = string
  default     = "s3-check-role-2025"
}

variable "aws_organization_id" {
  description = "The ID of your AWS Organization (e.g., o-xxxxxxxxxx)."
  type        = string
  default     = "o-nev2i5j9pw"
}
