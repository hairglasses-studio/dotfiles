# CR8 Audio Analysis Lambda
# Container-based Lambda for BPM/key analysis using librosa

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

variable "environment" {
  description = "Environment name (dev, prod)"
  type        = string
  default     = "prod"
}

variable "batch_size" {
  description = "Number of tracks to process per invocation"
  type        = number
  default     = 20
}

variable "schedule_rate" {
  description = "How often to run the analysis worker"
  type        = string
  default     = "rate(5 minutes)"
}

variable "enabled" {
  description = "Whether the scheduled trigger is enabled"
  type        = bool
  default     = true
}

locals {
  name_prefix = "cr8-analysis"
  tags = {
    Project     = "cr8"
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

# ECR Repository for Lambda container image
resource "aws_ecr_repository" "analysis_lambda" {
  name                 = "${local.name_prefix}-lambda"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = local.tags
}

resource "aws_ecr_lifecycle_policy" "analysis_lambda" {
  repository = aws_ecr_repository.analysis_lambda.name

  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep last 5 images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = 5
      }
      action = {
        type = "expire"
      }
    }]
  })
}

# IAM Role for Lambda
resource "aws_iam_role" "analysis_lambda" {
  name = "${local.name_prefix}-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })

  tags = local.tags
}

# IAM Policy for Lambda
resource "aws_iam_role_policy" "analysis_lambda" {
  name = "${local.name_prefix}-lambda-policy"
  role = aws_iam_role.analysis_lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:*"
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:Scan",
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:PutItem"
        ]
        Resource = [
          "arn:aws:dynamodb:us-east-1:*:table/cr8_tracks",
          "arn:aws:dynamodb:us-east-1:*:table/cr8_sync_queue"
        ]
      },
      {
        Effect   = "Allow"
        Action   = ["s3:GetObject"]
        Resource = "arn:aws:s3:::cr8-music-storage/*"
      }
    ]
  })
}

# Lambda Function (container image)
resource "aws_lambda_function" "analysis" {
  function_name = "${local.name_prefix}-worker"
  role          = aws_iam_role.analysis_lambda.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.analysis_lambda.repository_url}:latest"
  timeout       = 900  # 15 minutes max
  memory_size   = 3008 # 3GB for librosa

  environment {
    variables = {
      BATCH_SIZE        = tostring(var.batch_size)
      S3_BUCKET         = "cr8-music-storage"
      TRACKS_TABLE      = "cr8_tracks"
      QUEUE_TABLE       = "cr8_sync_queue"
      ANALYSIS_DURATION = "60"
      NUMBA_CACHE_DIR   = "/tmp"
    }
  }

  tags = local.tags

  depends_on = [aws_ecr_repository.analysis_lambda]

  lifecycle {
    ignore_changes = [image_uri]
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "analysis_lambda" {
  name              = "/aws/lambda/${aws_lambda_function.analysis.function_name}"
  retention_in_days = 14
  tags              = local.tags
}

# EventBridge Scheduler to trigger Lambda periodically
resource "aws_cloudwatch_event_rule" "analysis_schedule" {
  name                = "${local.name_prefix}-schedule"
  description         = "Trigger CR8 analysis worker on schedule"
  schedule_expression = var.schedule_rate
  state               = var.enabled ? "ENABLED" : "DISABLED"
  tags                = local.tags
}

resource "aws_cloudwatch_event_target" "analysis_lambda" {
  rule      = aws_cloudwatch_event_rule.analysis_schedule.name
  target_id = "analysis-lambda"
  arn       = aws_lambda_function.analysis.arn

  input = jsonencode({
    batch_size = var.batch_size
    source     = "scheduled"
  })
}

resource "aws_lambda_permission" "eventbridge" {
  statement_id  = "AllowEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.analysis.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.analysis_schedule.arn
}

# Outputs
output "ecr_repository_url" {
  description = "ECR repository URL for pushing container images"
  value       = aws_ecr_repository.analysis_lambda.repository_url
}

output "lambda_function_name" {
  description = "Lambda function name"
  value       = aws_lambda_function.analysis.function_name
}

output "lambda_function_arn" {
  description = "Lambda function ARN"
  value       = aws_lambda_function.analysis.arn
}
