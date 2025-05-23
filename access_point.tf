# Access Point Approach

resource "aws_iam_role" "a_role" {
  name               = "foo-via-access-point"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy_doc.json
}

resource "aws_iam_role_policy" "a_role_access_point_readonly" {
  name = "access-point-readonly"
  role = aws_iam_role.a_role.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect   = "Allow",
        Action   = "s3:*",
        Resource = "*"
      }
    ]
  })
}

resource "aws_s3_access_point" "secure_bucket_access_point" {
  name   = "${var.bucket_name}-ap"
  bucket = aws_s3_bucket.secure_bucket.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid       = "DenyAllExceptFoo",
        Effect    = "Deny",
        Principal = { AWS = aws_iam_role.a_role.arn },
        Action    = "s3:*",
        NotResource = [
          "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap",
          "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap/object/foo/*"
        ]
      },
      {
        Sid       = "DenyListNotFoo",
        Effect    = "Deny",
        Principal = { AWS = aws_iam_role.a_role.arn },
        Action    = "s3:ListBucket",
        Resource  = "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap",
        Condition = { StringNotLike = { "s3:prefix" = "foo/*" } }
      },
      {
        Sid       = "AllowListFoo",
        Effect    = "Allow",
        Principal = { AWS = aws_iam_role.a_role.arn },
        Action    = "s3:ListBucket",
        Resource  = "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap",
        Condition = { StringLike = { "s3:prefix" = "foo/*" } }
      },
      {
        Sid       = "AllowGetFoo",
        Effect    = "Allow",
        Principal = { AWS = aws_iam_role.a_role.arn },
        Action    = "s3:GetObject",
        Resource  = "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap/object/foo/*"
      }
    ]
  })
}

output "a_role_arn" {
  value = aws_iam_role.a_role.arn
}

output "secure_bucket_access_point_alias" {
  value = aws_s3_access_point.secure_bucket_access_point.alias
}
