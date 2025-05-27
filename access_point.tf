# Access Point Approach

locals {
  role_configs = {
    foo = { # Corresponds to the original "a_role"
      name   = "foo-via-access-point"
      prefix = var.prefix # Assuming the same prefix for now, can be customized per role if needed
    },
    bar = { # The new role
      name   = "bar-via-access-point"
      prefix = var.prefix # Assuming the same prefix for now
    }
  }
}

resource "aws_iam_role" "roles" {
  for_each           = local.role_configs
  name               = each.value.name
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy_doc.json
}

resource "aws_iam_role_policy" "roles_access_point_readonly" {
  for_each = local.role_configs
  name     = "${each.value.name}-ap-readonly"
  role     = aws_iam_role.roles[each.key].id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = "s3:*", # This policy grants broad S3 permissions to the role.
        # Access is then restricted by the S3 Access Point policy.
        Resource = "*"
      }
    ]
  })
}

resource "aws_s3_access_point" "secure_bucket_access_points" {
  for_each = local.role_configs
  name     = "${var.bucket_name}-ap-${each.key}" # e.g., mybucket-ap-foo, mybucket-ap-bar
  bucket   = aws_s3_bucket.secure_bucket.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid       = "DenyAllExceptConfiguredRole",
        Effect    = "Deny",
        Principal = { AWS = aws_iam_role.roles[each.key].arn },
        Action    = "s3:*",
        NotResource = [
          "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap-${each.key}",
          "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap-${each.key}/object/${each.value.prefix}*"
        ]
      },
      {
        Sid       = "DenyListBucketIfNotCorrectPrefixForRole",
        Effect    = "Deny",
        Principal = { AWS = aws_iam_role.roles[each.key].arn },
        Action    = "s3:ListBucket",
        Resource  = "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap-${each.key}",
        Condition = { StringNotLike = { "s3:prefix" = "${each.value.prefix}*" } }
      },
      {
        Sid       = "AllowListBucketIfCorrectPrefixForRole",
        Effect    = "Allow",
        Principal = { AWS = aws_iam_role.roles[each.key].arn },
        Action    = "s3:ListBucket",
        Resource  = "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap-${each.key}",
        Condition = { StringLike = { "s3:prefix" = "${each.value.prefix}*" } }
      },
      {
        Sid       = "AllowGetObjectIfCorrectPrefixForRole",
        Effect    = "Allow",
        Principal = { AWS = aws_iam_role.roles[each.key].arn },
        Action    = "s3:GetObject",
        Resource  = "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap-${each.key}/object/${each.value.prefix}*"
      }
    ]
  })
}

output "role_arns" {
  description = "The ARNs of the created IAM roles."
  value       = { for k, role in aws_iam_role.roles : k => role.arn }
}

output "secure_bucket_access_point_aliases" {
  description = "The aliases of the created S3 Access Points."
  value       = { for k, ap in aws_s3_access_point.secure_bucket_access_points : k => ap.alias }
}

output "secure_bucket_access_point_arns" {
  description = "The ARNs of the created S3 Access Points."
  value       = { for k, ap in aws_s3_access_point.secure_bucket_access_points : k => ap.arn }
}
