output "api_endpoint" {
  description = "API Gateway endpoint URL"
  value       = "${aws_api_gateway_stage.main.invoke_url}/"
}

output "api_id" {
  description = "API Gateway REST API ID"
  value       = aws_api_gateway_rest_api.main.id
}

output "api_key_id" {
  description = "API Key ID"
  value       = aws_api_gateway_api_key.main.id
}

output "api_key_value" {
  description = "API Key value"
  value       = aws_api_gateway_api_key.main.value
  sensitive   = true
}

output "api_key_secret_arn" {
  description = "Secrets Manager ARN for the API key"
  value       = aws_secretsmanager_secret.api_key.arn
}
