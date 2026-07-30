terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  # S3 backend - bucket/dynamodb_table/region is passed via -backend-config
  backend "s3" {
    key     = "terraform.tfstate"
    encrypt = true
  }
}

provider "aws" {
  region = local.aws_region

  default_tags {
    tags = {
      Project   = local.project_name
      ManagedBy = "terraform"
    }
  }
}

locals {
  project_name        = "abrp"
  aws_region          = "ap-northeast-1"
  cpu                 = "512"  # 0.5 vCPU
  memory              = "2048" # 2 GB (load test revealed 1 GB insufficient)
  api_stage_name      = "v1"
  log_retention_days  = 30
  schedule_expression = "cron(0 2 * * ? *)"

  # Fed to both the API Gateway preflight and the serve task, which each answer
  # a different half of a CORS exchange.
  cors_allow_origin = "*"
}

module "network" {
  source = "./modules/network"

  project_name = local.project_name
}

module "storage" {
  source = "./modules/storage"

  project_name = local.project_name
}

module "ecs" {
  source = "./modules/ecs"

  project_name          = local.project_name
  aws_region            = local.aws_region
  vpc_id                = module.network.vpc_id
  private_subnet_ids    = module.network.private_subnet_ids
  ecs_security_group_id = module.network.ecs_security_group_id
  ecr_repository_url    = module.storage.repository_url
  ecr_repository_arn    = module.storage.repository_arn
  cpu                   = local.cpu
  memory                = local.memory
  log_retention_days    = local.log_retention_days
  s3_bucket             = module.storage.cache_bucket_name
  cors_allow_origin     = local.cors_allow_origin
}

module "api_gateway" {
  source = "./modules/api_gateway"

  project_name       = local.project_name
  nlb_arn            = module.ecs.nlb_arn
  nlb_dns_name       = module.ecs.nlb_dns_name
  stage_name         = local.api_stage_name
  log_retention_days = local.log_retention_days
  cors_allow_origin  = local.cors_allow_origin
}

module "workflow" {
  source = "./modules/workflow"

  project_name          = local.project_name
  ecs_cluster_arn       = module.ecs.cluster_arn
  ecs_service_arn       = module.ecs.service_arn
  import_task_arn       = module.ecs.import_task_arn
  import_container_name = module.ecs.import_container_name
  cache_bucket_name     = module.storage.cache_bucket_name
  import_role_arns      = [module.ecs.execution_role_arn, module.ecs.import_task_role_arn]
  private_subnet_ids    = module.network.private_subnet_ids
  ecs_security_group_id = module.network.ecs_security_group_id
  log_retention_days    = local.log_retention_days
  schedule_expression   = local.schedule_expression
  aws_region            = local.aws_region
}

output "ecr_repository_url" {
  value = module.storage.repository_url
}

output "cache_bucket_name" {
  value = module.storage.cache_bucket_name
}

output "api_endpoint" {
  value = module.api_gateway.api_endpoint
}

output "api_key_value" {
  value     = module.api_gateway.api_key_value
  sensitive = true
}

output "ecs_cluster_name" {
  value = module.ecs.cluster_name
}

output "state_machine_arn" {
  value = module.workflow.state_machine_arn
}

output "scheduler_arn" {
  value = module.workflow.scheduler_arn
}
