terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }

  backend "s3" {
    bucket = "hendry-terraform-state"
    key    = "s3-check-role/terraform.tfstate" # Path within the bucket for the state file
    region = "ap-southeast-1"                  # Region where the state bucket exists
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Repo = "https://github.com/kaihendry/s3-check-role"
    }
  }
}

resource "aws_s3_bucket" "secure_bucket" {
  bucket = var.bucket_name
}

data "aws_iam_policy_document" "s3_read_only_policy_doc" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.secure_bucket.arn,
      "${aws_s3_bucket.secure_bucket.arn}/foo/*"
    ]
    effect = "Allow"
  }
}

resource "aws_iam_policy" "s3_read_only_policy" {
  name        = "S3ReadOnlyAccessPolicy-${aws_s3_bucket.secure_bucket.id}"
  description = "Policy that grants read-only access to a specific S3 bucket"
  policy      = data.aws_iam_policy_document.s3_read_only_policy_doc.json
}

data "aws_iam_policy_document" "assume_role_policy_doc" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "AWS"
      identifiers = ["*"] # Use wildcard principal with OrgID condition
    }
    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"
      values   = [var.aws_organization_id]
    }
    effect = "Allow"
  }
}

resource "aws_iam_role" "s3_read_only_role" {
  name               = "S3ReadOnlyRole-${aws_s3_bucket.secure_bucket.id}"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy_doc.json

  tags = {
    Description = "Role for read-only access to ${aws_s3_bucket.secure_bucket.id}"
  }
}

resource "aws_iam_role_policy_attachment" "s3_read_only_attach" {
  role       = aws_iam_role.s3_read_only_role.name
  policy_arn = aws_iam_policy.s3_read_only_policy.arn
}

output "bucket_name" {
  description = "The name of the created S3 bucket."
  value       = aws_s3_bucket.secure_bucket.bucket
}

output "role_arn" {
  description = "The ARN of the created IAM role."
  value       = aws_iam_role.s3_read_only_role.arn
}
