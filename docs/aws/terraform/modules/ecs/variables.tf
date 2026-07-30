variable "project_name" {
  type        = string
  description = "Project name"
}

variable "aws_region" {
  type        = string
  description = "AWS region"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID"
}

variable "private_subnet_ids" {
  type        = list(string)
  description = "Private subnet IDs for ECS"
}

variable "ecs_security_group_id" {
  type        = string
  description = "Security group ID for ECS"
}

variable "ecr_repository_url" {
  type        = string
  description = "ECR repository URL for abrp server"
}

variable "ecr_repository_arn" {
  type        = string
  description = "ECR repository ARN"
}

variable "cpu" {
  type        = string
  description = "CPU units (256, 512, 1024, 2048, 4096)"
  default     = "256"
}

variable "memory" {
  type        = string
  description = "Memory in MB (512, 1024, 2048, ...)"
  default     = "512"
}

variable "log_retention_days" {
  type        = number
  description = "CloudWatch Logs retention in days"
  default     = 30
}

variable "s3_bucket" {
  type        = string
  description = "S3 bucket name for CSV data caching"
}

variable "cpu_target" {
  type        = number
  description = "Target CPU utilization (%) for auto scaling"
  default     = 70
}

variable "memory_target" {
  type        = number
  description = "Target memory utilization (%) for auto scaling"
  default     = 80
}

variable "cors_allow_origin" {
  # Must match the value given to the API Gateway module: preflight is answered
  # there while the response to the actual GET comes from this service.
  type        = string
  description = "CORS Access-Control-Allow-Origin value the serve task returns"
  default     = "*"
}
