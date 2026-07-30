# Workflow Module Variables

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "ecs_cluster_arn" {
  description = "ECS cluster ARN"
  type        = string
}

variable "import_task_arn" {
  description = "ECS task definition ARN for import (provided by ecs module)"
  type        = string
}

variable "import_container_name" {
  description = "Container name in the import task definition (target of the change-check command override)"
  type        = string
}

variable "cache_bucket_name" {
  description = "S3 bucket holding data_modified.txt, used by the change-check step"
  type        = string
}

variable "import_role_arns" {
  description = "IAM role ARNs Step Functions may pass to ECS to run the import task (execution + task roles)"
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "Private subnet IDs for ECS tasks"
  type        = list(string)
}

variable "ecs_security_group_id" {
  description = "Security group ID for ECS tasks"
  type        = string
}

variable "schedule_expression" {
  description = "EventBridge schedule expression (cron or rate)"
  type        = string
  default     = "cron(0 2 * * ? *)"
}

variable "log_retention_days" {
  type        = number
  default     = 30
  description = "CloudWatch Logs retention in days"
}

variable "aws_region" {
  type        = string
  description = "AWS region for resource creation"
  default     = "ap-northeast-1"
}

variable "enable_schedule" {
  type        = bool
  description = "Enable EventBridge Scheduler for daily data update"
  default     = true
}

variable "ecs_service_arn" {
  description = "ECS service ARN to force-redeploy after a data update"
  type        = string
}
