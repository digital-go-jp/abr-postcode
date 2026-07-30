output "repository_url" {
  description = "ECR repository URL"
  value       = aws_ecr_repository.this.repository_url
}

output "repository_arn" {
  description = "ECR repository ARN"
  value       = aws_ecr_repository.this.arn
}

output "repository_name" {
  description = "ECR repository name"
  value       = aws_ecr_repository.this.name
}

output "cache_bucket_name" {
  description = "S3 cache bucket name"
  value       = aws_s3_bucket.cache.bucket
}

output "cache_bucket_arn" {
  description = "S3 cache bucket ARN"
  value       = aws_s3_bucket.cache.arn
}
