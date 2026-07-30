output "cluster_arn" {
  value       = aws_ecs_cluster.main.arn
  description = "ECS Cluster ARN"
}

output "cluster_name" {
  value       = aws_ecs_cluster.main.name
  description = "ECS Cluster name"
}

output "nlb_arn" {
  value       = aws_lb.nlb.arn
  description = "NLB ARN"
}

output "nlb_dns_name" {
  value       = aws_lb.nlb.dns_name
  description = "NLB DNS name"
}

output "service_name" {
  value       = aws_ecs_service.abrp.name
  description = "ECS Service name"
}

output "abrp_task_definition_arn" {
  value       = aws_ecs_task_definition.abrp.arn
  description = "Task definition ARN for abrp"
}

output "import_task_arn" {
  value       = aws_ecs_task_definition.import.arn
  description = "Task definition ARN for import (used by workflow)"
}

output "service_arn" {
  value       = aws_ecs_service.abrp.id
  description = "ECS Service ARN (used by workflow for ForceNewDeployment)"
}

output "execution_role_arn" {
  value       = aws_iam_role.ecs_task_execution.arn
  description = "ECS execution role ARN (passed by Step Functions to run the import task)"
}

output "import_task_role_arn" {
  value       = aws_iam_role.import_task.arn
  description = "Import task role ARN (passed by Step Functions to run the import task)"
}

output "import_container_name" {
  value       = "${var.project_name}-import"
  description = "Container name in the import task definition (needed to target container overrides)"
}
