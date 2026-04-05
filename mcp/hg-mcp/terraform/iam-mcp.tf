# MCP Server IAM Roles for ECS
# These roles are used by the cr8-cli MCP server running on ECS

locals {
  mcp_tags = {
    Project     = "cr8-cli"
    Component   = "mcp-server"
    Environment = "production"
    ManagedBy   = "terraform"
  }
}

# ECS Task Execution Role - pulls images, writes logs
resource "aws_iam_role" "mcp_execution" {
  name = "cr8-mcp-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })

  tags = local.mcp_tags
}

resource "aws_iam_role_policy_attachment" "mcp_execution_policy" {
  role       = aws_iam_role.mcp_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# ECS Task Role - runtime permissions for the MCP server
resource "aws_iam_role" "mcp_task" {
  name = "cr8-mcp-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })

  tags = local.mcp_tags
}

# S3 access policy for MCP task role
resource "aws_iam_role_policy" "mcp_task_s3" {
  name = "s3-access"
  role = aws_iam_role.mcp_task.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "MusicStorageRead"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.music_storage.arn,
          "${aws_s3_bucket.music_storage.arn}/*"
        ]
      },
      {
        Sid    = "MediaBucketsFullAccess"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket",
          "s3:GetObjectVersion",
          "s3:ListBucketVersions"
        ]
        Resource = [
          aws_s3_bucket.resolume_plugins.arn,
          "${aws_s3_bucket.resolume_plugins.arn}/*",
          aws_s3_bucket.vj_clips.arn,
          "${aws_s3_bucket.vj_clips.arn}/*",
          aws_s3_bucket.touchdesigner.arn,
          "${aws_s3_bucket.touchdesigner.arn}/*",
          aws_s3_bucket.obs.arn,
          "${aws_s3_bucket.obs.arn}/*",
          aws_s3_bucket.grandma3.arn,
          "${aws_s3_bucket.grandma3.arn}/*",
          aws_s3_bucket.ledfx.arn,
          "${aws_s3_bucket.ledfx.arn}/*"
        ]
      }
    ]
  })
}

# Inventory system access policy
resource "aws_iam_role_policy" "mcp_task_inventory" {
  name = "inventory-access"
  role = aws_iam_role.mcp_task.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "InventoryDynamoDB"
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:BatchGetItem",
          "dynamodb:BatchWriteItem"
        ]
        Resource = [
          aws_dynamodb_table.inventory.arn,
          "${aws_dynamodb_table.inventory.arn}/index/*"
        ]
      },
      {
        Sid    = "InventoryS3"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket",
          "s3:GetObjectVersion",
          "s3:ListBucketVersions"
        ]
        Resource = [
          aws_s3_bucket.inventory_assets.arn,
          "${aws_s3_bucket.inventory_assets.arn}/*"
        ]
      },
      {
        Sid    = "InventorySecrets"
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = [
          aws_secretsmanager_secret.ebay_credentials.arn,
          aws_secretsmanager_secret.gmail_credentials.arn
        ]
      }
    ]
  })
}

# Outputs
output "mcp_execution_role_arn" {
  description = "ARN of the MCP ECS task execution role"
  value       = aws_iam_role.mcp_execution.arn
}

output "mcp_task_role_arn" {
  description = "ARN of the MCP ECS task role"
  value       = aws_iam_role.mcp_task.arn
}
