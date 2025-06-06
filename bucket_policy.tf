# Bucket policy to only allow admin or access point access
resource "aws_s3_bucket_policy" "secure_bucket_policy" {
  bucket = aws_s3_bucket.secure_bucket.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid       = "RestrictedDataAccess",
        Effect    = "Deny",
        Principal = "*",
        Action    = "s3:*",
        Resource = [
          aws_s3_bucket.secure_bucket.arn,
          "${aws_s3_bucket.secure_bucket.arn}/*"
        ],
        Condition = {
          StringNotLike = {
            "aws:PrincipalArn" = [
              "arn:aws:iam::407461997746:role/github-actions-Role-56IHHM969DKJ",
              "arn:aws:iam::407461997746:role/aws-reserved/sso.amazonaws.com/ap-southeast-1/AWSReservedSSO_AdministratorAccess_*"
            ]
          }
          StringNotEqualsIfExists = {
            "s3:DataAccessPointAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

