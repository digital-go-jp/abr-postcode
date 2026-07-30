# Workflow Module - Step Functions + EventBridge for automated data updates
#
# Workflow (ECS-only, no CodeBuild):
#   EventBridge Scheduler -> Step Functions
#   -> UpdateData (ECS: abrp import, upload CSV to S3;
#      import skips when already up to date)
#   -> RefreshService (ECS: ForceNewDeployment to roll new data to serve tasks)

data "aws_caller_identity" "current" {}

# ============================================================================
# IAM Role for Step Functions
# ============================================================================

resource "aws_iam_role" "step_functions" {
  name = "${var.project_name}-sfn-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "states.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-sfn-role"
  }
}

resource "aws_iam_role_policy" "step_functions" {
  name = "${var.project_name}-sfn-policy"
  role = aws_iam_role.step_functions.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecs:RunTask",
          "ecs:StopTask",
          "ecs:DescribeTasks"
        ]
        Resource = "*"
        Condition = {
          ArnEquals = {
            "ecs:cluster" = var.ecs_cluster_arn
          }
        }
      },
      {
        Effect   = "Allow"
        Action   = "iam:PassRole"
        Resource = var.import_role_arns
        Condition = {
          StringLike = {
            "iam:PassedToService" = "ecs-tasks.amazonaws.com"
          }
        }
      },
      {
        Effect   = "Allow"
        Action   = "ecs:UpdateService"
        Resource = var.ecs_service_arn
      },
      {
        Effect = "Allow"
        Action = [
          "events:PutTargets",
          "events:PutRule",
          "events:DescribeRule"
        ]
        Resource = "arn:aws:events:${var.aws_region}:${data.aws_caller_identity.current.account_id}:rule/StepFunctionsGetEventsForECSTaskRule"
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogDelivery",
          "logs:GetLogDelivery",
          "logs:UpdateLogDelivery",
          "logs:DeleteLogDelivery",
          "logs:ListLogDeliveries",
          "logs:PutResourcePolicy",
          "logs:DescribeResourcePolicies",
          "logs:DescribeLogGroups"
        ]
        Resource = "*"
      }
    ]
  })
}

# ============================================================================
# CloudWatch Log Group for Step Functions
# ============================================================================

resource "aws_cloudwatch_log_group" "step_functions" {
  name              = "/aws/stepfunctions/${var.project_name}-data-update"
  retention_in_days = var.log_retention_days

  tags = {
    Name = "${var.project_name}-sfn-logs"
  }
}

# ============================================================================
# Step Functions State Machine
# ============================================================================

resource "aws_sfn_state_machine" "data_update" {
  name     = "${var.project_name}-data-update"
  role_arn = aws_iam_role.step_functions.arn

  logging_configuration {
    level                  = "ERROR"
    include_execution_data = true
    log_destination        = "${aws_cloudwatch_log_group.step_functions.arn}:*"
  }

  definition = jsonencode({
    Comment = "abrp data update workflow: check changes -> import data -> refresh service"
    StartAt = "CheckChanges"
    States = {
      # Compare the DCAT feed against the last-imported timestamp. The task
      # needs the previously stored data_modified.txt from S3; without it the
      # dry-run always reports an update on a fresh container.
      # - exit 0: up to date -> end workflow (no redeploy)
      # - exit 1: update pending -> caught as a task failure -> import
      CheckChanges = {
        Type           = "Task"
        Resource       = "arn:aws:states:::ecs:runTask.sync"
        TimeoutSeconds = 600
        Parameters = {
          Cluster        = var.ecs_cluster_arn
          TaskDefinition = var.import_task_arn
          LaunchType     = "FARGATE"
          NetworkConfiguration = {
            AwsvpcConfiguration = {
              Subnets        = var.private_subnet_ids
              SecurityGroups = [var.ecs_security_group_id]
              AssignPublicIp = "DISABLED"
            }
          }
          Overrides = {
            ContainerOverrides = [
              {
                Name    = var.import_container_name
                Command = ["aws s3 cp --only-show-errors s3://${var.cache_bucket_name}/data_modified.txt /app/data/data_modified.txt || true; /app/abrp import --dry-run"]
              }
            ]
          }
        }
        Next = "NoChanges"
        # exit 1 and genuine errors both surface as States.TaskFailed, so the
        # exit code decides which one it was
        Catch = [
          {
            ErrorEquals = ["States.TaskFailed"]
            ResultPath  = "$.error"
            Next        = "ParseCheckFailure"
          }
        ]
      }

      # The task result arrives as a JSON string, so it has to be parsed before
      # the exit code can be read.
      ParseCheckFailure = {
        Type = "Pass"
        Parameters = {
          "cause.$" = "States.StringToJson($.error.Cause)"
        }
        ResultPath = "$.parsed"
        Next       = "ClassifyCheckFailure"
      }

      # Only exit 1 means an update is pending. Anything else (a crashed
      # container, an ECS API rejection) must fail loudly instead of being
      # mistaken for new data and redeploying the service.
      ClassifyCheckFailure = {
        Type = "Choice"
        Choices = [
          {
            And = [
              {
                Variable  = "$.parsed.cause.Containers[0].ExitCode"
                IsPresent = true
              },
              {
                Variable      = "$.parsed.cause.Containers[0].ExitCode"
                NumericEquals = 1
              }
            ]
            Next = "UpdateData"
          }
        ]
        Default = "CheckChangesFailed"
      }

      CheckChangesFailed = {
        Type  = "Fail"
        Error = "CheckChangesFailed"
        Cause = "abrp import --dry-run did not report a change-detection result"
      }

      NoChanges = {
        Type    = "Succeed"
        Comment = "Data is up to date, nothing to redeploy"
      }

      # Update data (download, convert, upload to S3). abrp import compares
      # data_modified.txt against the DCAT feed and exits cleanly without
      # doing anything when already up to date.
      UpdateData = {
        Type           = "Task"
        Resource       = "arn:aws:states:::ecs:runTask.sync"
        TimeoutSeconds = 3600
        Parameters = {
          Cluster        = var.ecs_cluster_arn
          TaskDefinition = var.import_task_arn
          LaunchType     = "FARGATE"
          NetworkConfiguration = {
            AwsvpcConfiguration = {
              Subnets        = var.private_subnet_ids
              SecurityGroups = [var.ecs_security_group_id]
              AssignPublicIp = "DISABLED"
            }
          }
        }
        Next = "RefreshService"
      }

      # Rolling-redeploy the serve tasks so every task picks up the latest
      # CSVs from S3, converging any stale/fresh data mix from scale-out.
      RefreshService = {
        Type           = "Task"
        Resource       = "arn:aws:states:::aws-sdk:ecs:updateService"
        TimeoutSeconds = 300
        Parameters = {
          Cluster            = var.ecs_cluster_arn
          Service            = var.ecs_service_arn
          ForceNewDeployment = true
        }
        End = true
      }
    }
  })

  tags = {
    Name = "${var.project_name}-data-update"
  }
}

# ============================================================================
# IAM Role for EventBridge Scheduler
# ============================================================================

resource "aws_iam_role" "eventbridge" {
  count = var.enable_schedule ? 1 : 0
  name  = "${var.project_name}-eventbridge-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "scheduler.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-eventbridge-role"
  }
}

resource "aws_iam_role_policy" "eventbridge" {
  count = var.enable_schedule ? 1 : 0
  name  = "${var.project_name}-eventbridge-policy"
  role  = aws_iam_role.eventbridge[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "states:StartExecution"
        Resource = aws_sfn_state_machine.data_update.arn
      }
    ]
  })
}

# ============================================================================
# EventBridge Scheduler
# ============================================================================

resource "aws_scheduler_schedule" "daily_update" {
  count = var.enable_schedule ? 1 : 0
  name  = "${var.project_name}-daily-update"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression          = var.schedule_expression
  schedule_expression_timezone = "Asia/Tokyo"

  target {
    arn      = aws_sfn_state_machine.data_update.arn
    role_arn = aws_iam_role.eventbridge[0].arn
  }

  state = "ENABLED"
}

