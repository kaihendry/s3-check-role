# Bucket policy to only allow admin or access point access
data "aws_iam_policy_document" "secure_bucket_policy" {
  statement {
    sid     = "RestrictedDataAccess"
    effect  = "Deny"
    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.secure_bucket.arn,
      "${aws_s3_bucket.secure_bucket.arn}/*"
    ]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    condition {
      test     = "ArnNotLike"
      variable = "aws:PrincipalArn"
      values = [
        "arn:aws:iam::407461997746:role/github-actions-Role-56IHHM969DKJ",
        "arn:aws:iam::407461997746:role/aws-reserved/sso.amazonaws.com/ap-southeast-1/AWSReservedSSO_AdministratorAccess_*"
      ]
    }
    condition {
      test     = "StringNotEquals"
      variable = "s3:DataAccessPointAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}

resource "aws_s3_bucket_policy" "secure_bucket_policy" {
  bucket = aws_s3_bucket.secure_bucket.id
  policy = data.aws_iam_policy_document.secure_bucket_policy.json
}

