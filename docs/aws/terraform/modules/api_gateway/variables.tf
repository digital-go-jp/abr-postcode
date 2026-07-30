variable "project_name" {
  description = "Project name"
  type        = string
}

variable "nlb_arn" {
  description = "NLB ARN"
  type        = string
}

variable "nlb_dns_name" {
  description = "NLB DNS name"
  type        = string
}

variable "stage_name" {
  description = "API Gateway stage name"
  type        = string
  default     = "v1"
}

variable "log_retention_days" {
  description = "CloudWatch Logs retention period in days for API Gateway access logs"
  type        = number
  default     = 30
}

variable "cors_allow_origin" {
  # Preflight is answered here, while the response to the actual GET carries
  # whatever the backend returns, so both have to be fed the same value. Only a
  # single origin can be expressed: the MOCK integration returns one fixed
  # string and cannot reflect the request's Origin.
  description = "CORS Access-Control-Allow-Origin value, unquoted (e.g. \"*\" or \"https://example.com\"). Must match the value given to the ECS service."
  type        = string
  default     = "*"
}
