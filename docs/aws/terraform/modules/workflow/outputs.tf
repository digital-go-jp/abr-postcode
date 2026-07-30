output "state_machine_arn" {
  description = "Step Functions State Machine ARN"
  value       = aws_sfn_state_machine.data_update.arn
}

output "state_machine_name" {
  description = "Step Functions State Machine name"
  value       = aws_sfn_state_machine.data_update.name
}

output "scheduler_arn" {
  description = "EventBridge Scheduler ARN"
  value       = try(aws_scheduler_schedule.daily_update[0].arn, null)
}
