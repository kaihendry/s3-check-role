# Bucket Policy Approach

resource "aws_iam_role" "b_role" {
  name               = "foo-via-bucket-policy"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy_doc.json
}

# IAM policy for b_role (no Principal block)
data "aws_iam_policy_document" "b_role_readonly_policy" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.secure_bucket.arn]
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = ["foo/*"]
    }
  }
  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.secure_bucket.arn}/foo/*"]
  }
}

resource "aws_iam_role_policy" "b_role_bucket_policy_readonly" {
  name   = "bucket-policy-readonly"
  role   = aws_iam_role.b_role.id
  policy = data.aws_iam_policy_document.b_role_readonly_policy.json
}

resource "aws_s3_bucket_policy" "secure_bucket_policy" {
  bucket = aws_s3_bucket.secure_bucket.id
  policy = data.aws_iam_policy_document.bucket_policy.json
}

data "aws_iam_policy_document" "bucket_policy" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.secure_bucket.arn]
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
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.secure_bucket.arn}/foo/*"]
    principals {
      type        = "AWS"
      identifiers = [aws_iam_role.b_role.arn]
    }
  }
}

output "b_role_arn" {
  value = aws_iam_role.b_role.arn
}
