# Access Point Approach

resource "aws_iam_role" "a_role" {
  name               = "foo-via-access-point"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy_doc.json
}

data "aws_iam_policy_document" "a_role_access_point_readonly" {
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
      "s3:GetObject"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "a_role_access_point_readonly" {
  name   = "access-point-readonly"
  role   = aws_iam_role.a_role.id
  policy = data.aws_iam_policy_document.a_role_access_point_readonly.json
}

data "aws_iam_policy_document" "secure_bucket_access_point_policy" {
  statement {
    sid     = "DenyAllExceptAllowedRoles"
    effect  = "Deny"
    actions = ["s3:*"]
    resources = [
      "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap",
      "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap/object/*"
    ]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    condition {
      test     = "StringNotLike"
      variable = "aws:PrincipalArn"
      values   = local.effective_allowed_role_arns
    }
  }
  statement {
    sid    = "DenyWriteActionsForAllowedRoles"
    effect = "Deny"
    actions = [
      "s3:PutObject",
      "s3:PutObjectAcl",
      "s3:DeleteObject",
      "s3:DeleteObjectVersion",
      "s3:PutObjectTagging",
      "s3:PutObjectVersionAcl",
      "s3:PutObjectVersionTagging",
      "s3:AbortMultipartUpload",
      "s3:RestoreObject"
    ]
    resources = [
      "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap/object/*"
    ]
    principals {
      type        = "AWS"
      identifiers = local.effective_allowed_role_arns
    }
  }
  statement {
    sid     = "DenyListBucketOutsidePrefix"
    effect  = "Deny"
    actions = ["s3:ListBucket"]
    resources = [
      "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap"
    ]
    principals {
      type        = "AWS"
      identifiers = local.effective_allowed_role_arns
    }
    condition {
      test     = "StringNotLike"
      variable = "s3:prefix"
      values   = ["${var.prefix}*"]
    }
  }
  statement {
    sid     = "DenyGetObjectOutsidePrefix"
    effect  = "Deny"
    actions = ["s3:GetObject"]
    not_resources = [
      "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap/object/${var.prefix}*"
    ]
    principals {
      type        = "AWS"
      identifiers = local.effective_allowed_role_arns
    }
  }
  statement {
    sid     = "AllowListingPrefix"
    effect  = "Allow"
    actions = ["s3:ListBucket"]
    resources = [
      "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap"
    ]
    principals {
      type        = "AWS"
      identifiers = local.effective_allowed_role_arns
    }
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = ["${var.prefix}*"]
    }
  }
  statement {
    sid     = "AllowGetObject"
    effect  = "Allow"
    actions = ["s3:GetObject"]
    resources = [
      "arn:aws:s3:${var.aws_region}:${data.aws_caller_identity.current.account_id}:accesspoint/${var.bucket_name}-ap/object/${var.prefix}*"
    ]
    principals {
      type        = "AWS"
      identifiers = local.effective_allowed_role_arns
    }
  }
}

resource "aws_s3_access_point" "secure_bucket_access_point" {
  name   = "${var.bucket_name}-ap"
  bucket = aws_s3_bucket.secure_bucket.id
  policy = data.aws_iam_policy_document.secure_bucket_access_point_policy.json
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
