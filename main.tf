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
      Repo = "https://github.com/kaihendry/s3-bucket-policy-hell/"
    }
  }
}

resource "aws_s3_bucket" "secure_bucket" {
  bucket        = var.bucket_name
  force_destroy = true
}

data "aws_caller_identity" "current" {}

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

resource "aws_s3_object" "test_file_foo" {
  bucket  = aws_s3_bucket.secure_bucket.id
  key     = "${var.prefix}test.txt"
  content = "This is a test file in foo prefix - ${formatdate("YYYY-MM-DD", timestamp())}"

  tags = {
    Description = "Test file in foo prefix for access validation"
  }
}

resource "aws_s3_object" "test_file_bar" {
  bucket  = aws_s3_bucket.secure_bucket.id
  key     = "bar/test.txt"
  content = "This is a test file in bar prefix - ${formatdate("YYYY-MM-DD", timestamp())}"

  tags = {
    Description = "Test file in bar prefix for access validation"
  }
}

output "bucket_name" {
  value = aws_s3_bucket.secure_bucket.bucket
}
