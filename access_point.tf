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
        Sid : "DenyAllExceptAllowedRoles",
        Effect : "Deny",
        Principal : { "AWS" : local.effective_allowed_role_arns },
        Action : "s3:*",
        NotResource : [
          "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap",
          "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap/object/${var.prefix}*"
        ],
      },
      {
        Sid    = "RestrictListToAllowedPrefix",
        Effect = "Deny",
        Principal : { "AWS" : local.effective_allowed_role_arns },
        Action    = "s3:ListBucket",
        Resource  = "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap",
        Condition = { StringNotLike = { "s3:prefix" = "${var.prefix}*" } }
      },
      {
        Sid       = "AllowListingPrefix"
        Effect    = "Allow"
        Principal = { "AWS" : local.effective_allowed_role_arns }
        Action    = "s3:ListBucket"
        Resource  = "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap"
        Condition = { StringLike = { "s3:prefix" = "${var.prefix}*" } }
      },
      {
        Sid       = "AllowGetObject"
        Effect    = "Allow"
        Principal = { "AWS" : local.effective_allowed_role_arns }
        Action    = "s3:GetObject"
        Resource  = "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap/object/${var.prefix}*"
      }
    ]
  })
}

locals {
  effective_allowed_role_arns = setunion(
    var.allowed_role_arns,
    [aws_iam_role.a_role.arn]
  )
}

output "a_role_arn" {
  value = aws_iam_role.a_role.arn
}

output "secure_bucket_access_point_alias" {
  value = aws_s3_access_point.secure_bucket_access_point.alias
}
