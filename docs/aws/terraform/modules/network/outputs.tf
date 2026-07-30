output "vpc_id" {
  value       = aws_vpc.main.id
  description = "VPC ID"
}

output "private_subnet_ids" {
  value       = aws_subnet.private[*].id
  description = "Private subnet IDs"
}

output "ecs_security_group_id" {
  value       = aws_security_group.ecs.id
  description = "ECS security group ID"
}
