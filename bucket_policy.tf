# Bucket policy to only allow admin or access point access
resource "aws_s3_bucket_policy" "secure_bucket_policy" {
  bucket = aws_s3_bucket.secure_bucket.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid       = "DenyAllOtherAccess",
        Effect    = "Deny",
        Principal = "*",
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ],
        Resource = [
          aws_s3_bucket.secure_bucket.arn,
          "${aws_s3_bucket.secure_bucket.arn}/*"
        ],
        Condition = {
          StringNotLike = {
            "aws:PrincipalArn" = [
              "arn:aws:iam::407461997746:user/hendry",
              "arn:aws:iam::407461997746:role/github-actions-Role-56IHHM969DKJ",
              "arn:aws:iam::*:role/AWSReservedSSO_AdministratorAccess_*"
            ]
          },
          Null = {
            "s3:DataAccessPointArn" = "true"
          }
        }
      }
    ]
  })
}

