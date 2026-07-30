# Storage Module - ECR + S3 cache bucket

# Get current AWS account ID for globally unique bucket names
data "aws_caller_identity" "current" {}

locals {
  cache_bucket = "${var.project_name}-cache-${data.aws_caller_identity.current.account_id}"
}

# ============================================================================
# S3 Bucket for CSV cache
# ============================================================================

resource "aws_s3_bucket" "cache" {
  bucket        = local.cache_bucket
  force_destroy = true

  tags = {
    Name = local.cache_bucket
  }
}

resource "aws_s3_bucket_versioning" "cache" {
  bucket = aws_s3_bucket.cache.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "cache" {
  bucket = aws_s3_bucket.cache.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "cache" {
  bucket = aws_s3_bucket.cache.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Require SSL/TLS for all S3 access
resource "aws_s3_bucket_policy" "cache" {
  bucket = aws_s3_bucket.cache.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "RequireSSL"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.cache.arn,
          "${aws_s3_bucket.cache.arn}/*"
        ]
        Condition = {
          Bool = {
            "aws:SecureTransport" = "false"
          }
        }
      }
    ]
  })

  depends_on = [aws_s3_bucket_public_access_block.cache]
}

# Lifecycle: cleanup old versions & incomplete uploads
resource "aws_s3_bucket_lifecycle_configuration" "cache" {
  bucket = aws_s3_bucket.cache.id

  rule {
    id     = "cleanup-old-versions"
    status = "Enabled"

    filter {}

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }

  rule {
    id     = "abort-incomplete-uploads"
    status = "Enabled"

    filter {}

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

# ============================================================================
# ECR Repository
# ============================================================================

resource "aws_ecr_repository" "this" {
  name                 = var.project_name
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_lifecycle_policy" "this" {
  repository = aws_ecr_repository.this.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 3 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 3
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}
