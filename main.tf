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

resource "aws_s3_bucket_policy" "bucket_policy" {
  bucket = aws_s3_bucket.secure_bucket.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowDPBarConsumerRead"
        Effect = "Allow"
        Principal = {
          AWS = aws_iam_role.dp_bar_consumer_role.arn
        }
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "${aws_s3_bucket.secure_bucket.arn}/bar/*",
          aws_s3_bucket.secure_bucket.arn
        ]
      }
    ]
  })
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

# New role for bar consumer
resource "aws_iam_role" "dp_bar_consumer_role" {
  name = "dp-bar-consumer-rp"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Condition = {
          StringEquals = {
            "aws:PrincipalOrgID" = var.aws_organization_id
          }
        }
      }
    ]
  })

  tags = {
    Description = "Role for read-only access to ${aws_s3_bucket.secure_bucket.id}/bar prefix via resource policy"
  }
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

resource "aws_s3_object" "test_file_bar" {
  bucket  = aws_s3_bucket.secure_bucket.id
  key     = "bar/test.txt"
  content = "This is a test file in bar prefix - ${formatdate("YYYY-MM-DD", timestamp())}"

  tags = {
    Description = "Test file in bar prefix for access validation"
  }
}

output "bucket_name" {
  description = "The name of the created S3 bucket."
  value       = aws_s3_bucket.secure_bucket.bucket
}

output "role_arn" {
  description = "The ARN of the created IAM role."
  value       = aws_iam_role.s3_read_only_role.arn
}

output "dp_bar_consumer_role_arn" {
  description = "The ARN of the bar consumer role (with access via resource policy)"
  value       = aws_iam_role.dp_bar_consumer_role.arn
}
