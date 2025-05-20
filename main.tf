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

# on secure_bucket resource policy allow foo-via-bucket-policy role to access foo/*
resource "aws_s3_bucket_policy" "secure_bucket_policy" {
  bucket = aws_s3_bucket.secure_bucket.id

  policy = data.aws_iam_policy_document.bucket_policy.json
}

data "aws_iam_policy_document" "bucket_policy" {
  # allow list bucket only for foo/* prefix
  statement {
    actions = ["s3:ListBucket"]
    resources = [
      aws_s3_bucket.secure_bucket.arn
    ]
    principals {
      type        = "AWS"
      identifiers = [aws_iam_role.b_role.arn]
    }
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = ["foo/*"]
    }
  }
  statement {
    actions = ["s3:GetObject"]
    resources = [
      "${aws_s3_bucket.secure_bucket.arn}/foo/*"
    ]
    principals {
      type        = "AWS"
      identifiers = [aws_iam_role.b_role.arn]
    }
  }
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
# Role that can access foo/test.txt via S3 Access Point
resource "aws_iam_role" "a_role" {
  name               = "foo-via-access-point"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy_doc.json
}

# Role that can access foo/test.txt via bucket policy
resource "aws_iam_role" "b_role" {
  name               = "foo-via-bucket-policy"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy_doc.json
}


# Test objects for validation
resource "aws_s3_object" "test_file_foo" {
  bucket  = aws_s3_bucket.secure_bucket.id
  key     = "foo/test.txt"
  content = "This is a test file in foo prefix - ${formatdate("YYYY-MM-DD", timestamp())}"

  tags = {
    Description = "Test file in foo prefix for access validation"
  }
}

output "bucket_name" {
  value = aws_s3_bucket.secure_bucket.bucket
}

output "a_role_arn" {
  value = aws_iam_role.a_role.arn
}

output "b_role_arn" {
  value = aws_iam_role.b_role.arn
}
