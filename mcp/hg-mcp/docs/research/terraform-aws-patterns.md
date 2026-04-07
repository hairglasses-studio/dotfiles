# Terraform AWS Infrastructure Patterns for hg-mcp

> Comprehensive research document covering Terraform best practices for AWS infrastructure with DynamoDB, S3, Lambda, and API Gateway.

**Last Updated:** December 2025
**Terraform Version:** 1.11+
**AWS Provider Version:** 5.x+

---

## Table of Contents

1. [Terraform State Management](#1-terraform-state-management)
2. [Module Structure and Organization](#2-module-structure-and-organization)
3. [GitHub Actions CI/CD with OIDC](#3-github-actions-cicd-with-oidc)
4. [DynamoDB Table Design Patterns](#4-dynamodb-table-design-patterns)
5. [S3 Bucket Configurations](#5-s3-bucket-configurations)
6. [Lambda Deployment Patterns](#6-lambda-deployment-patterns)
7. [API Gateway with Lambda Integration](#7-api-gateway-with-lambda-integration)
8. [IAM Role and Policy Best Practices](#8-iam-role-and-policy-best-practices)
9. [Environment Separation](#9-environment-separation)
10. [Cost Optimization and Resource Tagging](#10-cost-optimization-and-resource-tagging)

---

## 1. Terraform State Management

### S3 Native Locking (Recommended for 2025)

As of Terraform 1.11.0, S3-native state locking is generally available (GA), eliminating the need for DynamoDB for state locking. This approach requires fewer resources and reduces IAM permissions complexity.

```hcl
# backend.tf - Modern S3 Backend with Native Locking
terraform {
  backend "s3" {
    bucket         = "hg-mcp-terraform-state"
    key            = "infrastructure/terraform.tfstate"
    region         = "ap-southeast-2"
    encrypt        = true
    use_lockfile   = true  # S3-native locking (Terraform 1.11+)

    # Optional: KMS encryption
    kms_key_id     = "alias/terraform-state"
  }
}
```

### Legacy DynamoDB Locking (Still Supported)

If using Terraform versions prior to 1.10 or migrating existing infrastructure:

```hcl
# backend.tf - Legacy DynamoDB Locking
terraform {
  backend "s3" {
    bucket         = "hg-mcp-terraform-state"
    key            = "infrastructure/terraform.tfstate"
    region         = "ap-southeast-2"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"  # Deprecated but supported
  }
}

# DynamoDB table for state locking (legacy)
resource "aws_dynamodb_table" "terraform_lock" {
  name           = "terraform-state-lock"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags = {
    Name        = "terraform-state-lock"
    Environment = "shared"
    ManagedBy   = "terraform"
  }
}
```

### State Bucket Configuration

```hcl
# state-bucket.tf
resource "aws_s3_bucket" "terraform_state" {
  bucket = "hg-mcp-terraform-state"

  lifecycle {
    prevent_destroy = true
  }

  tags = {
    Name        = "Terraform State"
    Environment = "shared"
    ManagedBy   = "terraform"
  }
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.terraform_state.arn
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
```

### References

- [HashiCorp S3 Backend Documentation](https://developer.hashicorp.com/terraform/language/backend/s3)
- [AWS Prescriptive Guidance - Backend Best Practices](https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/backend.html)
- [S3-Native State Locking Announcement](https://medium.com/aws-specialists/dynamodb-not-needed-for-terraform-state-locking-in-s3-anymore-29a8054fc0e9)

---

## 2. Module Structure and Organization

### Recommended Directory Structure

```
hg-mcp-infrastructure/
├── modules/                          # Reusable modules
│   ├── dynamodb/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   ├── lambda/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   ├── api-gateway/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   └── s3/
│       ├── main.tf
│       ├── variables.tf
│       ├── outputs.tf
│       └── README.md
├── environments/                     # Environment-specific configs
│   ├── dev/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── backend.tf
│   │   └── terraform.tfvars
│   ├── staging/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── backend.tf
│   │   └── terraform.tfvars
│   └── prod/
│       ├── main.tf
│       ├── variables.tf
│       ├── outputs.tf
│       ├── backend.tf
│       └── terraform.tfvars
├── shared/                           # Shared infrastructure
│   ├── iam/
│   ├── networking/
│   └── state-backend/
└── .github/
    └── workflows/
        └── terraform.yml
```

### Module Pattern Example

```hcl
# modules/lambda/main.tf
resource "aws_lambda_function" "this" {
  function_name = var.function_name
  description   = var.description
  role          = aws_iam_role.lambda_exec.arn
  handler       = var.handler
  runtime       = var.runtime
  timeout       = var.timeout
  memory_size   = var.memory_size

  filename         = var.filename
  source_code_hash = filebase64sha256(var.filename)

  environment {
    variables = var.environment_variables
  }

  dynamic "vpc_config" {
    for_each = var.vpc_config != null ? [var.vpc_config] : []
    content {
      subnet_ids         = vpc_config.value.subnet_ids
      security_group_ids = vpc_config.value.security_group_ids
    }
  }

  tags = var.tags
}

# modules/lambda/variables.tf
variable "function_name" {
  description = "Name of the Lambda function"
  type        = string
}

variable "description" {
  description = "Description of the Lambda function"
  type        = string
  default     = ""
}

variable "handler" {
  description = "Lambda function handler"
  type        = string
}

variable "runtime" {
  description = "Lambda runtime"
  type        = string
  default     = "python3.12"
}

variable "timeout" {
  description = "Function timeout in seconds"
  type        = number
  default     = 30
}

variable "memory_size" {
  description = "Memory allocation in MB"
  type        = number
  default     = 256
}

variable "filename" {
  description = "Path to the deployment package"
  type        = string
}

variable "environment_variables" {
  description = "Environment variables for the function"
  type        = map(string)
  default     = {}
}

variable "vpc_config" {
  description = "VPC configuration for the Lambda function"
  type = object({
    subnet_ids         = list(string)
    security_group_ids = list(string)
  })
  default = null
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# modules/lambda/outputs.tf
output "function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.this.arn
}

output "function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.this.function_name
}

output "invoke_arn" {
  description = "Invoke ARN for API Gateway integration"
  value       = aws_lambda_function.this.invoke_arn
}

output "role_arn" {
  description = "ARN of the Lambda execution role"
  value       = aws_iam_role.lambda_exec.arn
}
```

### Module Composition in Root Module

```hcl
# environments/prod/main.tf
module "mcp_api_lambda" {
  source = "../../modules/lambda"

  function_name         = "hg-mcp-api-${var.environment}"
  description           = "AFTRS MCP API Handler"
  handler               = "main.handler"
  runtime               = "python3.12"
  timeout               = 30
  memory_size           = 512
  filename              = "${path.module}/../../dist/lambda.zip"
  environment_variables = {
    ENVIRONMENT    = var.environment
    LOG_LEVEL      = "INFO"
    DYNAMODB_TABLE = module.mcp_data_table.table_name
  }

  tags = local.common_tags
}

module "mcp_data_table" {
  source = "../../modules/dynamodb"

  table_name   = "hg-mcp-data-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "PK"
  range_key    = "SK"

  tags = local.common_tags
}
```

### References

- [HashiCorp Standard Module Structure](https://developer.hashicorp.com/terraform/language/modules/develop/structure)
- [HashiCorp Module Composition](https://developer.hashicorp.com/terraform/language/modules/develop/composition)
- [AWS Prescriptive Guidance - Code Structure](https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/structure.html)

---

## 3. GitHub Actions CI/CD with OIDC

### OIDC Provider Setup in Terraform

```hcl
# shared/iam/github-oidc.tf

# GitHub OIDC Provider
resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1"]

  tags = {
    Name      = "github-actions-oidc"
    ManagedBy = "terraform"
  }
}

# IAM Role for Terraform in GitHub Actions
resource "aws_iam_role" "github_actions_terraform" {
  name = "github-actions-terraform"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = aws_iam_openid_connect_provider.github.arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
          StringLike = {
            "token.actions.githubusercontent.com:sub" = [
              "repo:hairglasses-studio/hg-mcp:ref:refs/heads/main",
              "repo:hairglasses-studio/hg-mcp:environment:production",
              "repo:hairglasses-studio/hg-mcp:environment:staging"
            ]
          }
        }
      }
    ]
  })

  tags = {
    Name      = "github-actions-terraform"
    ManagedBy = "terraform"
  }
}

# Policy for Terraform operations
resource "aws_iam_role_policy_attachment" "github_actions_terraform" {
  role       = aws_iam_role.github_actions_terraform.name
  policy_arn = aws_iam_policy.terraform_deploy.arn
}

resource "aws_iam_policy" "terraform_deploy" {
  name        = "terraform-deploy-policy"
  description = "Policy for Terraform deployments via GitHub Actions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "TerraformState"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::hg-mcp-terraform-state",
          "arn:aws:s3:::hg-mcp-terraform-state/*"
        ]
      },
      {
        Sid    = "DynamoDBManagement"
        Effect = "Allow"
        Action = [
          "dynamodb:CreateTable",
          "dynamodb:DeleteTable",
          "dynamodb:DescribeTable",
          "dynamodb:UpdateTable",
          "dynamodb:ListTagsOfResource",
          "dynamodb:TagResource",
          "dynamodb:UntagResource",
          "dynamodb:DescribeContinuousBackups",
          "dynamodb:UpdateContinuousBackups",
          "dynamodb:DescribeTimeToLive",
          "dynamodb:UpdateTimeToLive"
        ]
        Resource = "arn:aws:dynamodb:*:*:table/hg-mcp-*"
      },
      {
        Sid    = "LambdaManagement"
        Effect = "Allow"
        Action = [
          "lambda:CreateFunction",
          "lambda:DeleteFunction",
          "lambda:GetFunction",
          "lambda:GetFunctionConfiguration",
          "lambda:UpdateFunctionCode",
          "lambda:UpdateFunctionConfiguration",
          "lambda:ListVersionsByFunction",
          "lambda:PublishVersion",
          "lambda:CreateAlias",
          "lambda:DeleteAlias",
          "lambda:GetAlias",
          "lambda:UpdateAlias",
          "lambda:AddPermission",
          "lambda:RemovePermission",
          "lambda:GetPolicy",
          "lambda:TagResource",
          "lambda:UntagResource",
          "lambda:ListTags"
        ]
        Resource = "arn:aws:lambda:*:*:function:hg-mcp-*"
      },
      {
        Sid    = "APIGatewayManagement"
        Effect = "Allow"
        Action = [
          "apigateway:*"
        ]
        Resource = [
          "arn:aws:apigateway:*::/restapis/*",
          "arn:aws:apigateway:*::/apis/*"
        ]
      },
      {
        Sid    = "S3BucketManagement"
        Effect = "Allow"
        Action = [
          "s3:CreateBucket",
          "s3:DeleteBucket",
          "s3:GetBucketPolicy",
          "s3:PutBucketPolicy",
          "s3:GetBucketVersioning",
          "s3:PutBucketVersioning",
          "s3:GetBucketEncryption",
          "s3:PutBucketEncryption",
          "s3:GetBucketLifecycleConfiguration",
          "s3:PutBucketLifecycleConfiguration",
          "s3:GetBucketPublicAccessBlock",
          "s3:PutBucketPublicAccessBlock",
          "s3:GetBucketTagging",
          "s3:PutBucketTagging"
        ]
        Resource = "arn:aws:s3:::hg-mcp-*"
      },
      {
        Sid    = "IAMRoleManagement"
        Effect = "Allow"
        Action = [
          "iam:CreateRole",
          "iam:DeleteRole",
          "iam:GetRole",
          "iam:UpdateRole",
          "iam:PassRole",
          "iam:AttachRolePolicy",
          "iam:DetachRolePolicy",
          "iam:PutRolePolicy",
          "iam:DeleteRolePolicy",
          "iam:GetRolePolicy",
          "iam:ListRolePolicies",
          "iam:ListAttachedRolePolicies",
          "iam:TagRole",
          "iam:UntagRole"
        ]
        Resource = "arn:aws:iam::*:role/hg-mcp-*"
      }
    ]
  })

  tags = {
    Name      = "terraform-deploy-policy"
    ManagedBy = "terraform"
  }
}
```

### GitHub Actions Workflow

```yaml
# .github/workflows/terraform.yml
name: Terraform CI/CD

on:
  push:
    branches: [main]
    paths:
      - 'infrastructure/**'
      - '.github/workflows/terraform.yml'
  pull_request:
    branches: [main]
    paths:
      - 'infrastructure/**'
      - '.github/workflows/terraform.yml'

permissions:
  id-token: write   # Required for OIDC
  contents: read
  pull-requests: write

env:
  TF_VERSION: "1.11.0"
  AWS_REGION: "ap-southeast-2"

jobs:
  terraform-validate:
    name: Validate
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Format Check
        run: terraform fmt -check -recursive
        working-directory: infrastructure

      - name: Terraform Init (Validation Only)
        run: terraform init -backend=false
        working-directory: infrastructure/environments/dev

      - name: Terraform Validate
        run: terraform validate
        working-directory: infrastructure/environments/dev

  terraform-plan:
    name: Plan (${{ matrix.environment }})
    needs: terraform-validate
    runs-on: ubuntu-latest
    strategy:
      matrix:
        environment: [dev, staging, prod]
    environment: ${{ matrix.environment }}

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::${{ secrets.AWS_ACCOUNT_ID }}:role/github-actions-terraform
          role-session-name: terraform-${{ matrix.environment }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Init
        run: terraform init
        working-directory: infrastructure/environments/${{ matrix.environment }}

      - name: Terraform Plan
        id: plan
        run: |
          terraform plan -no-color -out=tfplan 2>&1 | tee plan_output.txt
          echo "plan_output<<EOF" >> $GITHUB_OUTPUT
          cat plan_output.txt >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT
        working-directory: infrastructure/environments/${{ matrix.environment }}

      - name: Upload Plan
        uses: actions/upload-artifact@v4
        with:
          name: tfplan-${{ matrix.environment }}
          path: infrastructure/environments/${{ matrix.environment }}/tfplan

      - name: Comment PR with Plan
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const output = `#### Terraform Plan - ${{ matrix.environment }}

            <details><summary>Show Plan</summary>

            \`\`\`terraform
            ${{ steps.plan.outputs.plan_output }}
            \`\`\`

            </details>

            *Pushed by: @${{ github.actor }}, Action: \`${{ github.event_name }}\`*`;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: output
            })

  terraform-apply:
    name: Apply (${{ matrix.environment }})
    needs: terraform-plan
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    runs-on: ubuntu-latest
    strategy:
      matrix:
        environment: [dev, staging, prod]
      max-parallel: 1  # Deploy sequentially
    environment: ${{ matrix.environment }}

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::${{ secrets.AWS_ACCOUNT_ID }}:role/github-actions-terraform
          role-session-name: terraform-${{ matrix.environment }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Download Plan
        uses: actions/download-artifact@v4
        with:
          name: tfplan-${{ matrix.environment }}
          path: infrastructure/environments/${{ matrix.environment }}

      - name: Terraform Init
        run: terraform init
        working-directory: infrastructure/environments/${{ matrix.environment }}

      - name: Terraform Apply
        run: terraform apply -auto-approve tfplan
        working-directory: infrastructure/environments/${{ matrix.environment }}
```

### Using the OIDC Module (Alternative)

```hcl
# Using the community module
module "github_oidc" {
  source  = "unfunco/oidc-github/aws"
  version = "~> 1.7"

  github_repositories = [
    "hairglasses-studio/hg-mcp:ref:refs/heads/main",
    "hairglasses-studio/hg-mcp:environment:production"
  ]

  attach_admin_policy = false

  iam_role_name        = "github-actions-terraform"
  iam_role_policy_arns = [aws_iam_policy.terraform_deploy.arn]

  tags = {
    ManagedBy = "terraform"
  }
}
```

### References

- [GitHub - configure-aws-credentials Action](https://github.com/aws-actions/configure-aws-credentials)
- [HashiCorp - Deploy with GitHub Actions](https://developer.hashicorp.com/terraform/tutorials/automation/github-actions)
- [Terraform AWS OIDC GitHub Module](https://github.com/unfunco/terraform-aws-oidc-github)
- [AWS - Configuring OIDC for GitHub](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html)

---

## 4. DynamoDB Table Design Patterns

### Single-Table Design Pattern

```hcl
# modules/dynamodb/main.tf

resource "aws_dynamodb_table" "this" {
  name           = var.table_name
  billing_mode   = var.billing_mode
  hash_key       = var.hash_key
  range_key      = var.range_key

  # Primary key attributes
  attribute {
    name = var.hash_key
    type = "S"
  }

  dynamic "attribute" {
    for_each = var.range_key != null ? [var.range_key] : []
    content {
      name = attribute.value
      type = "S"
    }
  }

  # GSI attributes
  dynamic "attribute" {
    for_each = var.global_secondary_indexes
    content {
      name = attribute.value.hash_key
      type = attribute.value.hash_key_type
    }
  }

  dynamic "attribute" {
    for_each = [for gsi in var.global_secondary_indexes : gsi if gsi.range_key != null]
    content {
      name = attribute.value.range_key
      type = attribute.value.range_key_type
    }
  }

  # Global Secondary Indexes
  dynamic "global_secondary_index" {
    for_each = var.global_secondary_indexes
    content {
      name               = global_secondary_index.value.name
      hash_key           = global_secondary_index.value.hash_key
      range_key          = global_secondary_index.value.range_key
      projection_type    = global_secondary_index.value.projection_type
      non_key_attributes = global_secondary_index.value.non_key_attributes

      # Only for PROVISIONED billing mode
      read_capacity  = var.billing_mode == "PROVISIONED" ? global_secondary_index.value.read_capacity : null
      write_capacity = var.billing_mode == "PROVISIONED" ? global_secondary_index.value.write_capacity : null
    }
  }

  # Local Secondary Indexes (must be defined at creation)
  dynamic "local_secondary_index" {
    for_each = var.local_secondary_indexes
    content {
      name               = local_secondary_index.value.name
      range_key          = local_secondary_index.value.range_key
      projection_type    = local_secondary_index.value.projection_type
      non_key_attributes = local_secondary_index.value.non_key_attributes
    }
  }

  # Time to Live
  dynamic "ttl" {
    for_each = var.ttl_attribute != null ? [var.ttl_attribute] : []
    content {
      attribute_name = ttl.value
      enabled        = true
    }
  }

  # Point-in-time recovery
  point_in_time_recovery {
    enabled = var.point_in_time_recovery_enabled
  }

  # Server-side encryption
  server_side_encryption {
    enabled     = true
    kms_key_arn = var.kms_key_arn
  }

  # Stream specification for CDC
  dynamic "stream_specification" {
    for_each = var.stream_enabled ? [1] : []
    content {
      stream_enabled   = true
      stream_view_type = var.stream_view_type
    }
  }

  tags = var.tags

  lifecycle {
    ignore_changes = [
      read_capacity,
      write_capacity,
    ]
  }
}

# Auto-scaling for PROVISIONED mode
resource "aws_appautoscaling_target" "read" {
  count              = var.billing_mode == "PROVISIONED" && var.autoscaling_enabled ? 1 : 0
  max_capacity       = var.autoscaling_read_max
  min_capacity       = var.autoscaling_read_min
  resource_id        = "table/${aws_dynamodb_table.this.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "read" {
  count              = var.billing_mode == "PROVISIONED" && var.autoscaling_enabled ? 1 : 0
  name               = "${var.table_name}-read-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.read[0].resource_id
  scalable_dimension = aws_appautoscaling_target.read[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.read[0].service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBReadCapacityUtilization"
    }
    target_value       = var.autoscaling_target_value
    scale_in_cooldown  = 60
    scale_out_cooldown = 60
  }
}
```

### Usage Example for hg-mcp

```hcl
# environments/prod/dynamodb.tf

module "mcp_main_table" {
  source = "../../modules/dynamodb"

  table_name   = "hg-mcp-main-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"  # Recommended for variable workloads
  hash_key     = "PK"
  range_key    = "SK"

  # GSI for querying by type
  global_secondary_indexes = [
    {
      name               = "GSI1"
      hash_key           = "GSI1PK"
      hash_key_type      = "S"
      range_key          = "GSI1SK"
      range_key_type     = "S"
      projection_type    = "ALL"
      non_key_attributes = null
      read_capacity      = null
      write_capacity     = null
    },
    {
      name               = "GSI2"
      hash_key           = "EntityType"
      hash_key_type      = "S"
      range_key          = "CreatedAt"
      range_key_type     = "S"
      projection_type    = "INCLUDE"
      non_key_attributes = ["Name", "Status"]
      read_capacity      = null
      write_capacity     = null
    }
  ]

  ttl_attribute                 = "ExpiresAt"
  point_in_time_recovery_enabled = true
  stream_enabled                = true
  stream_view_type              = "NEW_AND_OLD_IMAGES"

  tags = local.common_tags
}

# Session table with TTL
module "mcp_sessions_table" {
  source = "../../modules/dynamodb"

  table_name   = "hg-mcp-sessions-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "SessionId"
  range_key    = null

  global_secondary_indexes = []
  local_secondary_indexes  = []

  ttl_attribute                  = "ExpiresAt"
  point_in_time_recovery_enabled = false

  tags = local.common_tags
}
```

### DynamoDB Access Patterns Reference

| Access Pattern | PK | SK | GSI |
|---------------|----|----|-----|
| Get tool by ID | TOOL#<id> | TOOL#<id> | - |
| List tools by category | CATEGORY#<cat> | TOOL#<id> | - |
| Get user sessions | USER#<id> | SESSION#<timestamp> | - |
| List recent tools | - | - | GSI1: EntityType=TOOL, SK=CreatedAt |
| Get tool usage stats | TOOL#<id> | STATS#<date> | - |

### References

- [HashiCorp - Manage DynamoDB Scale](https://developer.hashicorp.com/terraform/tutorials/aws/aws-dynamodb-scale)
- [AWS - Global Secondary Indexes](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GSI.html)
- [Terraform Registry - DynamoDB Module](https://registry.terraform.io/modules/terraform-aws-modules/dynamodb-table/aws)
- [DynamoDB GSI Guide](https://dynobase.dev/dynamodb-gsi/)

---

## 5. S3 Bucket Configurations

### Secure S3 Bucket Module

```hcl
# modules/s3/main.tf

resource "aws_s3_bucket" "this" {
  bucket = var.bucket_name

  tags = var.tags
}

# Versioning
resource "aws_s3_bucket_versioning" "this" {
  bucket = aws_s3_bucket.this.id

  versioning_configuration {
    status     = var.versioning_enabled ? "Enabled" : "Suspended"
    mfa_delete = var.mfa_delete_enabled ? "Enabled" : "Disabled"
  }
}

# Encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = var.kms_key_arn != null ? "aws:kms" : "AES256"
      kms_master_key_id = var.kms_key_arn
    }
    bucket_key_enabled = var.kms_key_arn != null
  }
}

# Public access block (ALWAYS ENABLED)
resource "aws_s3_bucket_public_access_block" "this" {
  bucket = aws_s3_bucket.this.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Lifecycle configuration
resource "aws_s3_bucket_lifecycle_configuration" "this" {
  count  = length(var.lifecycle_rules) > 0 ? 1 : 0
  bucket = aws_s3_bucket.this.id

  dynamic "rule" {
    for_each = var.lifecycle_rules
    content {
      id     = rule.value.id
      status = rule.value.enabled ? "Enabled" : "Disabled"

      filter {
        prefix = rule.value.prefix
      }

      dynamic "transition" {
        for_each = rule.value.transitions
        content {
          days          = transition.value.days
          storage_class = transition.value.storage_class
        }
      }

      dynamic "noncurrent_version_transition" {
        for_each = rule.value.noncurrent_version_transitions
        content {
          noncurrent_days = noncurrent_version_transition.value.days
          storage_class   = noncurrent_version_transition.value.storage_class
        }
      }

      dynamic "expiration" {
        for_each = rule.value.expiration_days != null ? [rule.value.expiration_days] : []
        content {
          days = expiration.value
        }
      }

      dynamic "noncurrent_version_expiration" {
        for_each = rule.value.noncurrent_expiration_days != null ? [rule.value.noncurrent_expiration_days] : []
        content {
          noncurrent_days = noncurrent_version_expiration.value
        }
      }
    }
  }
}

# Logging
resource "aws_s3_bucket_logging" "this" {
  count  = var.logging_bucket != null ? 1 : 0
  bucket = aws_s3_bucket.this.id

  target_bucket = var.logging_bucket
  target_prefix = var.logging_prefix != null ? var.logging_prefix : "${var.bucket_name}/"
}

# CORS (for web access)
resource "aws_s3_bucket_cors_configuration" "this" {
  count  = length(var.cors_rules) > 0 ? 1 : 0
  bucket = aws_s3_bucket.this.id

  dynamic "cors_rule" {
    for_each = var.cors_rules
    content {
      allowed_headers = cors_rule.value.allowed_headers
      allowed_methods = cors_rule.value.allowed_methods
      allowed_origins = cors_rule.value.allowed_origins
      expose_headers  = cors_rule.value.expose_headers
      max_age_seconds = cors_rule.value.max_age_seconds
    }
  }
}

# Replication (for DR)
resource "aws_s3_bucket_replication_configuration" "this" {
  count  = var.replication_configuration != null ? 1 : 0
  bucket = aws_s3_bucket.this.id
  role   = var.replication_configuration.role_arn

  rule {
    id     = "replicate-all"
    status = "Enabled"

    destination {
      bucket        = var.replication_configuration.destination_bucket_arn
      storage_class = var.replication_configuration.storage_class
    }
  }

  depends_on = [aws_s3_bucket_versioning.this]
}
```

### Usage Examples

```hcl
# environments/prod/s3.tf

# Lambda deployment artifacts bucket
module "lambda_artifacts" {
  source = "../../modules/s3"

  bucket_name        = "hg-mcp-lambda-artifacts-${var.environment}"
  versioning_enabled = true

  lifecycle_rules = [
    {
      id      = "cleanup-old-versions"
      enabled = true
      prefix  = ""
      transitions = [
        {
          days          = 30
          storage_class = "STANDARD_IA"
        }
      ]
      noncurrent_version_transitions = [
        {
          days          = 30
          storage_class = "STANDARD_IA"
        }
      ]
      expiration_days             = null
      noncurrent_expiration_days  = 90
    }
  ]

  tags = local.common_tags
}

# Data storage bucket with encryption
module "data_bucket" {
  source = "../../modules/s3"

  bucket_name        = "hg-mcp-data-${var.environment}"
  versioning_enabled = true
  kms_key_arn        = aws_kms_key.data_encryption.arn

  lifecycle_rules = [
    {
      id      = "archive-old-data"
      enabled = true
      prefix  = "archives/"
      transitions = [
        {
          days          = 30
          storage_class = "STANDARD_IA"
        },
        {
          days          = 90
          storage_class = "GLACIER"
        }
      ]
      noncurrent_version_transitions = []
      expiration_days                = 365
      noncurrent_expiration_days     = 90
    }
  ]

  tags = local.common_tags
}

# Static assets with CORS
module "static_assets" {
  source = "../../modules/s3"

  bucket_name        = "hg-mcp-static-${var.environment}"
  versioning_enabled = false

  cors_rules = [
    {
      allowed_headers = ["*"]
      allowed_methods = ["GET", "HEAD"]
      allowed_origins = ["https://*.aftrs.edu.au"]
      expose_headers  = ["ETag"]
      max_age_seconds = 3600
    }
  ]

  tags = local.common_tags
}
```

### S3 Production Checklist

- [ ] All four public access blocks enabled
- [ ] Server-side encryption configured (AES256 or KMS)
- [ ] Versioning enabled for critical data
- [ ] Lifecycle policies attached
- [ ] Bucket policy restricts access
- [ ] MFA Delete enabled for critical buckets
- [ ] Logging configured to separate bucket
- [ ] Tags include Environment/Owner/DataClassification

### References

- [AWS - S3 Security Best Practices](https://docs.aws.amazon.com/AmazonS3/latest/userguide/security-best-practices.html)
- [Terraform - S3 Bucket Resource](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket)
- [Gruntwork - Private S3 Bucket Module](https://docs.gruntwork.io/reference/modules/terraform-aws-security/private-s3-bucket/)

---

## 6. Lambda Deployment Patterns

### ZIP File Deployment

```hcl
# modules/lambda/zip-deployment.tf

# Archive the Lambda code
data "archive_file" "lambda_zip" {
  count       = var.create_package ? 1 : 0
  type        = "zip"
  source_dir  = var.source_path
  output_path = "${path.module}/builds/${var.function_name}.zip"
  excludes    = var.package_excludes
}

resource "aws_lambda_function" "zip" {
  count = var.package_type == "Zip" ? 1 : 0

  function_name = var.function_name
  description   = var.description
  role          = aws_iam_role.lambda_exec.arn
  handler       = var.handler
  runtime       = var.runtime
  timeout       = var.timeout
  memory_size   = var.memory_size
  architectures = [var.architecture]

  # Use local zip or S3
  filename         = var.create_package ? data.archive_file.lambda_zip[0].output_path : var.local_existing_package
  source_code_hash = var.create_package ? data.archive_file.lambda_zip[0].output_base64sha256 : filebase64sha256(var.local_existing_package)

  # Or from S3
  # s3_bucket        = var.s3_bucket
  # s3_key           = var.s3_key
  # s3_object_version = var.s3_object_version

  environment {
    variables = var.environment_variables
  }

  # Layers
  layers = var.layers

  # VPC configuration
  dynamic "vpc_config" {
    for_each = var.vpc_config != null ? [var.vpc_config] : []
    content {
      subnet_ids         = vpc_config.value.subnet_ids
      security_group_ids = vpc_config.value.security_group_ids
    }
  }

  # Dead letter queue
  dynamic "dead_letter_config" {
    for_each = var.dead_letter_target_arn != null ? [var.dead_letter_target_arn] : []
    content {
      target_arn = dead_letter_config.value
    }
  }

  # Tracing
  tracing_config {
    mode = var.tracing_mode
  }

  # Logging
  logging_config {
    log_format = "JSON"
    log_group  = aws_cloudwatch_log_group.lambda.name
  }

  tags = var.tags

  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    aws_cloudwatch_log_group.lambda,
  ]
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.function_name}"
  retention_in_days = var.log_retention_days

  tags = var.tags
}
```

### Container Image Deployment

```hcl
# modules/lambda/container-deployment.tf

resource "aws_lambda_function" "container" {
  count = var.package_type == "Image" ? 1 : 0

  function_name = var.function_name
  description   = var.description
  role          = aws_iam_role.lambda_exec.arn
  package_type  = "Image"
  timeout       = var.timeout
  memory_size   = var.memory_size
  architectures = [var.architecture]

  image_uri = var.image_uri

  image_config {
    command           = var.image_config_command
    entry_point       = var.image_config_entry_point
    working_directory = var.image_config_working_directory
  }

  environment {
    variables = var.environment_variables
  }

  dynamic "vpc_config" {
    for_each = var.vpc_config != null ? [var.vpc_config] : []
    content {
      subnet_ids         = vpc_config.value.subnet_ids
      security_group_ids = vpc_config.value.security_group_ids
    }
  }

  tracing_config {
    mode = var.tracing_mode
  }

  tags = var.tags

  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    aws_cloudwatch_log_group.lambda,
  ]
}

# ECR Repository for Lambda images
resource "aws_ecr_repository" "lambda" {
  count = var.create_ecr_repository ? 1 : 0

  name                 = var.ecr_repository_name
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "KMS"
    kms_key         = var.kms_key_arn
  }

  tags = var.tags
}

# ECR lifecycle policy
resource "aws_ecr_lifecycle_policy" "lambda" {
  count      = var.create_ecr_repository ? 1 : 0
  repository = aws_ecr_repository.lambda[0].name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}
```

### Lambda Layers

```hcl
# modules/lambda-layer/main.tf

resource "aws_lambda_layer_version" "this" {
  layer_name          = var.layer_name
  description         = var.description
  compatible_runtimes = var.compatible_runtimes

  filename         = var.filename
  source_code_hash = filebase64sha256(var.filename)

  # OR from S3
  # s3_bucket = var.s3_bucket
  # s3_key    = var.s3_key
}

# Usage in Lambda
module "mcp_handler" {
  source = "../../modules/lambda"

  function_name = "hg-mcp-handler-${var.environment}"
  handler       = "main.handler"
  runtime       = "python3.12"

  layers = [
    module.shared_deps_layer.layer_arn,
    module.mcp_utils_layer.layer_arn
  ]

  # ... other configuration
}
```

### IAM Role for Lambda

```hcl
# modules/lambda/iam.tf

resource "aws_iam_role" "lambda_exec" {
  name = "${var.function_name}-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

# Basic execution policy (logs)
resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# VPC execution policy (if VPC enabled)
resource "aws_iam_role_policy_attachment" "lambda_vpc" {
  count      = var.vpc_config != null ? 1 : 0
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

# X-Ray tracing policy
resource "aws_iam_role_policy_attachment" "lambda_xray" {
  count      = var.tracing_mode == "Active" ? 1 : 0
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/AWSXRayDaemonWriteAccess"
}

# Custom policy for application permissions
resource "aws_iam_role_policy" "lambda_custom" {
  count  = var.custom_policy_json != null ? 1 : 0
  name   = "${var.function_name}-custom-policy"
  role   = aws_iam_role.lambda_exec.id
  policy = var.custom_policy_json
}
```

### References

- [Terraform Registry - Lambda Module](https://registry.terraform.io/modules/terraform-aws-modules/lambda/aws)
- [AWS - Lambda Layers](https://docs.aws.amazon.com/lambda/latest/dg/chapter-layers.html)
- [AWS - Lambda Container Images](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html)
- [HashiCorp - Lambda Tutorial](https://developer.hashicorp.com/terraform/tutorials/aws/lambda-api-gateway)

---

## 7. API Gateway with Lambda Integration

### REST API (API Gateway v1)

```hcl
# modules/api-gateway/rest-api.tf

resource "aws_api_gateway_rest_api" "this" {
  name        = var.api_name
  description = var.api_description

  endpoint_configuration {
    types = [var.endpoint_type]  # REGIONAL, EDGE, or PRIVATE
  }

  binary_media_types = var.binary_media_types

  tags = var.tags
}

# Proxy resource for Lambda integration
resource "aws_api_gateway_resource" "proxy" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_rest_api.this.root_resource_id
  path_part   = "{proxy+}"
}

# ANY method for proxy
resource "aws_api_gateway_method" "proxy" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.proxy.id
  http_method   = "ANY"
  authorization = var.authorization_type
  authorizer_id = var.authorizer_id

  request_parameters = var.request_parameters
}

# Lambda integration
resource "aws_api_gateway_integration" "lambda" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.proxy.id
  http_method = aws_api_gateway_method.proxy.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.lambda_invoke_arn
}

# Root path method
resource "aws_api_gateway_method" "root" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_rest_api.this.root_resource_id
  http_method   = "ANY"
  authorization = var.authorization_type
  authorizer_id = var.authorizer_id
}

resource "aws_api_gateway_integration" "root" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_rest_api.this.root_resource_id
  http_method = aws_api_gateway_method.root.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.lambda_invoke_arn
}

# Deployment
resource "aws_api_gateway_deployment" "this" {
  rest_api_id = aws_api_gateway_rest_api.this.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.proxy.id,
      aws_api_gateway_method.proxy.id,
      aws_api_gateway_integration.lambda.id,
      aws_api_gateway_method.root.id,
      aws_api_gateway_integration.root.id,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Stage
resource "aws_api_gateway_stage" "this" {
  deployment_id = aws_api_gateway_deployment.this.id
  rest_api_id   = aws_api_gateway_rest_api.this.id
  stage_name    = var.stage_name

  cache_cluster_enabled = var.cache_enabled
  cache_cluster_size    = var.cache_size

  xray_tracing_enabled = var.xray_tracing_enabled

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gw.arn
    format = jsonencode({
      requestId         = "$context.requestId"
      ip                = "$context.identity.sourceIp"
      requestTime       = "$context.requestTime"
      httpMethod        = "$context.httpMethod"
      resourcePath      = "$context.resourcePath"
      status            = "$context.status"
      responseLength    = "$context.responseLength"
      integrationLatency = "$context.integrationLatency"
    })
  }

  tags = var.tags
}

# CloudWatch Log Group for API Gateway
resource "aws_cloudwatch_log_group" "api_gw" {
  name              = "/aws/api-gateway/${var.api_name}"
  retention_in_days = var.log_retention_days

  tags = var.tags
}

# Lambda permission to be invoked by API Gateway
resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.this.execution_arn}/*/*"
}
```

### HTTP API (API Gateway v2) - Simpler and Cheaper

```hcl
# modules/api-gateway/http-api.tf

resource "aws_apigatewayv2_api" "this" {
  name          = var.api_name
  description   = var.api_description
  protocol_type = "HTTP"

  cors_configuration {
    allow_origins     = var.cors_allow_origins
    allow_methods     = var.cors_allow_methods
    allow_headers     = var.cors_allow_headers
    expose_headers    = var.cors_expose_headers
    max_age           = var.cors_max_age
    allow_credentials = var.cors_allow_credentials
  }

  tags = var.tags
}

# Lambda integration
resource "aws_apigatewayv2_integration" "lambda" {
  api_id           = aws_apigatewayv2_api.this.id
  integration_type = "AWS_PROXY"

  connection_type        = "INTERNET"
  integration_method     = "POST"
  integration_uri        = var.lambda_invoke_arn
  payload_format_version = "2.0"
}

# Default route
resource "aws_apigatewayv2_route" "default" {
  api_id    = aws_apigatewayv2_api.this.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

# Stage
resource "aws_apigatewayv2_stage" "this" {
  api_id      = aws_apigatewayv2_api.this.id
  name        = var.stage_name
  auto_deploy = var.auto_deploy

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gw.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      responseLength = "$context.responseLength"
    })
  }

  default_route_settings {
    throttling_burst_limit = var.throttling_burst_limit
    throttling_rate_limit  = var.throttling_rate_limit
  }

  tags = var.tags
}

# Lambda permission
resource "aws_lambda_permission" "api_gw_v2" {
  statement_id  = "AllowAPIGatewayV2Invoke"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.this.execution_arn}/*/*"
}

# Outputs
output "api_endpoint" {
  value = aws_apigatewayv2_stage.this.invoke_url
}
```

### Usage Example

```hcl
# environments/prod/api.tf

module "mcp_api" {
  source = "../../modules/api-gateway"

  api_name        = "hg-mcp-api-${var.environment}"
  api_description = "AFTRS MCP API"
  stage_name      = var.environment

  lambda_function_name = module.mcp_handler.function_name
  lambda_invoke_arn    = module.mcp_handler.invoke_arn

  # CORS configuration
  cors_allow_origins = var.environment == "prod" ? ["https://mcp.aftrs.edu.au"] : ["*"]
  cors_allow_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  cors_allow_headers = ["Content-Type", "Authorization", "X-Api-Key"]

  # Throttling
  throttling_burst_limit = var.environment == "prod" ? 1000 : 100
  throttling_rate_limit  = var.environment == "prod" ? 500 : 50

  log_retention_days = var.environment == "prod" ? 90 : 14

  tags = local.common_tags
}
```

### Custom Domain Setup

```hcl
# Custom domain for API Gateway
resource "aws_apigatewayv2_domain_name" "this" {
  domain_name = "api.${var.domain_name}"

  domain_name_configuration {
    certificate_arn = var.certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  tags = var.tags
}

resource "aws_apigatewayv2_api_mapping" "this" {
  api_id      = aws_apigatewayv2_api.this.id
  domain_name = aws_apigatewayv2_domain_name.this.id
  stage       = aws_apigatewayv2_stage.this.id
}

# Route53 record
resource "aws_route53_record" "api" {
  zone_id = var.route53_zone_id
  name    = aws_apigatewayv2_domain_name.this.domain_name
  type    = "A"

  alias {
    name                   = aws_apigatewayv2_domain_name.this.domain_name_configuration[0].target_domain_name
    zone_id                = aws_apigatewayv2_domain_name.this.domain_name_configuration[0].hosted_zone_id
    evaluate_target_health = false
  }
}
```

### References

- [HashiCorp - Lambda API Gateway Tutorial](https://developer.hashicorp.com/terraform/tutorials/aws/lambda-api-gateway)
- [AWS - API Gateway Proxy Integration](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-set-up-simple-proxy.html)
- [Terraform - API Gateway Resources](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/api_gateway_rest_api)
- [Serverless Land - API Gateway Patterns](https://serverlessland.com/patterns/apigw-lambda-dynamodb-terraform)

---

## 8. IAM Role and Policy Best Practices

### Least Privilege IAM Policies

```hcl
# modules/iam/lambda-role.tf

# Use aws_iam_policy_document for HCL-based policies
data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "lambda" {
  name               = "${var.prefix}-lambda-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = var.tags
}

# DynamoDB access policy - least privilege
data "aws_iam_policy_document" "dynamodb_access" {
  statement {
    sid    = "DynamoDBTableAccess"
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem",
      "dynamodb:Query",
      "dynamodb:Scan"
    ]
    resources = [
      var.dynamodb_table_arn,
      "${var.dynamodb_table_arn}/index/*"
    ]
  }

  # Conditional batch operations
  dynamic "statement" {
    for_each = var.enable_batch_operations ? [1] : []
    content {
      sid    = "DynamoDBBatchAccess"
      effect = "Allow"
      actions = [
        "dynamodb:BatchGetItem",
        "dynamodb:BatchWriteItem"
      ]
      resources = [var.dynamodb_table_arn]
    }
  }
}

resource "aws_iam_policy" "dynamodb_access" {
  name        = "${var.prefix}-dynamodb-access"
  description = "DynamoDB access for Lambda"
  policy      = data.aws_iam_policy_document.dynamodb_access.json

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "dynamodb_access" {
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.dynamodb_access.arn
}

# S3 access policy
data "aws_iam_policy_document" "s3_access" {
  statement {
    sid    = "S3BucketAccess"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject"
    ]
    resources = [
      "${var.s3_bucket_arn}/*"
    ]
  }

  statement {
    sid    = "S3BucketList"
    effect = "Allow"
    actions = [
      "s3:ListBucket"
    ]
    resources = [var.s3_bucket_arn]
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = var.s3_allowed_prefixes
    }
  }
}

resource "aws_iam_policy" "s3_access" {
  name        = "${var.prefix}-s3-access"
  description = "S3 access for Lambda"
  policy      = data.aws_iam_policy_document.s3_access.json

  tags = var.tags
}

# Secrets Manager access
data "aws_iam_policy_document" "secrets_access" {
  statement {
    sid    = "SecretsManagerAccess"
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue"
    ]
    resources = var.secret_arns
  }
}
```

### Service Role Patterns

```hcl
# modules/iam/service-roles.tf

# API Gateway logging role
resource "aws_iam_role" "api_gateway_cloudwatch" {
  name = "${var.prefix}-api-gw-cloudwatch"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "apigateway.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "api_gateway_cloudwatch" {
  role       = aws_iam_role.api_gateway_cloudwatch.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonAPIGatewayPushToCloudWatchLogs"
}

# DynamoDB Stream to Lambda role
data "aws_iam_policy_document" "dynamodb_stream" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetRecords",
      "dynamodb:GetShardIterator",
      "dynamodb:DescribeStream",
      "dynamodb:ListStreams"
    ]
    resources = [
      "${var.dynamodb_table_arn}/stream/*"
    ]
  }
}
```

### IAM Analyzer Integration

```hcl
# Enable IAM Access Analyzer
resource "aws_accessanalyzer_analyzer" "this" {
  analyzer_name = "${var.prefix}-analyzer"
  type          = "ACCOUNT"  # or "ORGANIZATION"

  tags = var.tags
}

# Archive findings rule (optional)
resource "aws_accessanalyzer_archive_rule" "internal_s3" {
  analyzer_name = aws_accessanalyzer_analyzer.this.analyzer_name
  rule_name     = "internal-s3-access"

  filter {
    criteria = "resourceType"
    eq       = ["AWS::S3::Bucket"]
  }

  filter {
    criteria = "isPublic"
    eq       = ["false"]
  }
}
```

### IAM Best Practices Checklist

- [ ] Use IAM roles instead of IAM users
- [ ] Avoid static/long-lived access keys
- [ ] Define granular permissions at operation level
- [ ] Use IAM Access Analyzer to find unused permissions
- [ ] Create separate roles for different functions
- [ ] Leverage Terraform modules for IAM patterns
- [ ] Use `aws_iam_policy_document` for HCL-based policies
- [ ] Implement regular audits and monitoring
- [ ] Store Terraform state securely with encryption

### References

- [AWS Prescriptive Guidance - Security Best Practices](https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/security.html)
- [HashiCorp - Create IAM Policies](https://developer.hashicorp.com/terraform/tutorials/aws/aws-iam-policy)
- [AWS - IAM Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)

---

## 9. Environment Separation

### Approach 1: Directory-Based Separation (Recommended)

```
infrastructure/
├── modules/                 # Shared modules
├── environments/
│   ├── dev/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── backend.tf      # dev-specific backend
│   │   └── terraform.tfvars
│   ├── staging/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── backend.tf
│   │   └── terraform.tfvars
│   └── prod/
│       ├── main.tf
│       ├── variables.tf
│       ├── outputs.tf
│       ├── backend.tf
│       └── terraform.tfvars
```

### Environment-Specific Configuration

```hcl
# environments/dev/terraform.tfvars
environment = "dev"
aws_region  = "ap-southeast-2"

# Smaller resources for dev
lambda_memory_size = 256
lambda_timeout     = 30
dynamodb_billing_mode = "PAY_PER_REQUEST"

# Less retention
log_retention_days = 7

# Relaxed throttling
api_throttling_burst_limit = 50
api_throttling_rate_limit  = 25

# Tags
project     = "hg-mcp"
cost_center = "CC-DEV"
```

```hcl
# environments/prod/terraform.tfvars
environment = "prod"
aws_region  = "ap-southeast-2"

# Production-sized resources
lambda_memory_size = 1024
lambda_timeout     = 60
dynamodb_billing_mode = "PAY_PER_REQUEST"  # or PROVISIONED with autoscaling

# Longer retention
log_retention_days = 90

# Production throttling
api_throttling_burst_limit = 1000
api_throttling_rate_limit  = 500

# Tags
project     = "hg-mcp"
cost_center = "CC-PROD"
```

### Common Variables Pattern

```hcl
# environments/dev/variables.tf

variable "environment" {
  description = "Environment name"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-southeast-2"
}

locals {
  common_tags = {
    Environment = var.environment
    Project     = var.project
    ManagedBy   = "terraform"
    CostCenter  = var.cost_center
  }

  # Environment-specific settings
  is_production = var.environment == "prod"

  # Resource naming
  name_prefix = "hg-mcp-${var.environment}"
}
```

### Backend Per Environment

```hcl
# environments/dev/backend.tf
terraform {
  backend "s3" {
    bucket       = "hg-mcp-terraform-state"
    key          = "environments/dev/terraform.tfstate"
    region       = "ap-southeast-2"
    encrypt      = true
    use_lockfile = true
  }
}

# environments/prod/backend.tf
terraform {
  backend "s3" {
    bucket       = "hg-mcp-terraform-state"
    key          = "environments/prod/terraform.tfstate"
    region       = "ap-southeast-2"
    encrypt      = true
    use_lockfile = true
    # Additional security for prod
    kms_key_id   = "alias/terraform-state-prod"
  }
}
```

### Approach 2: Workspaces (For Similar Environments)

```hcl
# main.tf - Using workspaces
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = terraform.workspace
      Project     = "hg-mcp"
      ManagedBy   = "terraform"
    }
  }
}

# Workspace-aware resource configuration
locals {
  environment_config = {
    dev = {
      lambda_memory_size = 256
      log_retention      = 7
      instance_type      = "t3.micro"
    }
    staging = {
      lambda_memory_size = 512
      log_retention      = 30
      instance_type      = "t3.small"
    }
    prod = {
      lambda_memory_size = 1024
      log_retention      = 90
      instance_type      = "t3.medium"
    }
  }

  config = local.environment_config[terraform.workspace]
}

# Using workspace-aware config
resource "aws_lambda_function" "this" {
  function_name = "hg-mcp-${terraform.workspace}"
  memory_size   = local.config.lambda_memory_size
  # ...
}
```

### Workspace Commands

```bash
# Create and switch workspaces
terraform workspace new dev
terraform workspace new staging
terraform workspace new prod

# Switch workspace
terraform workspace select prod

# List workspaces
terraform workspace list

# Apply to specific workspace
terraform workspace select dev && terraform apply -var-file="env/dev.tfvars"
```

### Multi-Account Strategy (Advanced)

```hcl
# For separate AWS accounts per environment
provider "aws" {
  alias  = "dev"
  region = var.aws_region

  assume_role {
    role_arn = "arn:aws:iam::DEV_ACCOUNT_ID:role/TerraformRole"
  }
}

provider "aws" {
  alias  = "prod"
  region = var.aws_region

  assume_role {
    role_arn = "arn:aws:iam::PROD_ACCOUNT_ID:role/TerraformRole"
  }
}

module "dev_infrastructure" {
  source = "./modules/infrastructure"
  providers = {
    aws = aws.dev
  }
  environment = "dev"
}

module "prod_infrastructure" {
  source = "./modules/infrastructure"
  providers = {
    aws = aws.prod
  }
  environment = "prod"
}
```

### References

- [HashiCorp - Organize Configuration](https://developer.hashicorp.com/terraform/tutorials/modules/organize-configuration)
- [Gruntwork - Terraform Workspaces](https://www.gruntwork.io/blog/how-to-manage-multiple-environments-with-terraform-using-workspaces)
- [HashiCorp - Recommended Practices](https://developer.hashicorp.com/terraform/cloud-docs/recommended-practices/part1)

---

## 10. Cost Optimization and Resource Tagging

### Default Tags Configuration

```hcl
# providers.tf

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment     = var.environment
      Project         = "hg-mcp"
      Owner           = "platform-team"
      CostCenter      = var.cost_center
      ManagedBy       = "terraform"
      Repository      = "hairglasses-studio/hg-mcp"
      DataClassification = "internal"
    }
  }
}
```

### Tagging Strategy Module

```hcl
# modules/tagging/main.tf

variable "environment" {
  type = string
}

variable "project" {
  type    = string
  default = "hg-mcp"
}

variable "owner" {
  type = string
}

variable "cost_center" {
  type = string
}

variable "additional_tags" {
  type    = map(string)
  default = {}
}

locals {
  required_tags = {
    Environment        = var.environment
    Project            = var.project
    Owner              = var.owner
    CostCenter         = var.cost_center
    ManagedBy          = "terraform"
    DataClassification = "internal"
    CreatedAt          = timestamp()
  }

  all_tags = merge(local.required_tags, var.additional_tags)
}

output "tags" {
  value = local.all_tags
}
```

### Cost Allocation Tags Setup

```hcl
# Enable cost allocation tags in AWS
resource "aws_ce_cost_allocation_tag" "environment" {
  tag_key = "Environment"
  status  = "Active"
}

resource "aws_ce_cost_allocation_tag" "project" {
  tag_key = "Project"
  status  = "Active"
}

resource "aws_ce_cost_allocation_tag" "cost_center" {
  tag_key = "CostCenter"
  status  = "Active"
}

resource "aws_ce_cost_allocation_tag" "owner" {
  tag_key = "Owner"
  status  = "Active"
}
```

### Cost-Optimized Resource Configurations

```hcl
# Cost-optimized Lambda
resource "aws_lambda_function" "optimized" {
  function_name = "${local.name_prefix}-handler"

  # Right-size memory (CPU scales with memory)
  memory_size = 256  # Start small, increase based on metrics

  # Appropriate timeout
  timeout = 30  # Don't over-provision

  # ARM64 for cost savings (up to 34% cheaper)
  architectures = ["arm64"]

  # Provisioned concurrency only if needed
  # reserved_concurrent_executions = var.reserved_concurrency

  tags = local.common_tags
}

# Cost-optimized DynamoDB
resource "aws_dynamodb_table" "optimized" {
  name         = "${local.name_prefix}-data"
  billing_mode = "PAY_PER_REQUEST"  # Best for variable workloads

  # Enable TTL to auto-delete old items
  ttl {
    attribute_name = "ExpiresAt"
    enabled        = true
  }

  tags = local.common_tags
}

# Cost-optimized S3 with lifecycle rules
resource "aws_s3_bucket_lifecycle_configuration" "optimized" {
  bucket = aws_s3_bucket.data.id

  rule {
    id     = "transition-to-ia"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"  # 45% cheaper
    }

    transition {
      days          = 90
      storage_class = "GLACIER_IR"  # 68% cheaper
    }

    noncurrent_version_expiration {
      noncurrent_days = 90
    }
  }

  rule {
    id     = "expire-incomplete-uploads"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}
```

### Budget Alerts

```hcl
# AWS Budget for cost monitoring
resource "aws_budgets_budget" "monthly" {
  name         = "hg-mcp-monthly-budget"
  budget_type  = "COST"
  limit_amount = var.monthly_budget_limit
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  cost_filter {
    name = "TagKeyValue"
    values = [
      "user:Project$hg-mcp"
    ]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type          = "FORECASTED"
    subscriber_email_addresses = var.budget_alert_emails
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 100
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = var.budget_alert_emails
  }
}
```

### CloudWatch Cost Dashboard

```hcl
# Cost monitoring dashboard
resource "aws_cloudwatch_dashboard" "cost" {
  dashboard_name = "hg-mcp-costs"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        properties = {
          title  = "Lambda Invocations"
          region = var.aws_region
          metrics = [
            ["AWS/Lambda", "Invocations", "FunctionName", "${local.name_prefix}-handler"]
          ]
          stat   = "Sum"
          period = 86400
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        properties = {
          title  = "DynamoDB Consumed Capacity"
          region = var.aws_region
          metrics = [
            ["AWS/DynamoDB", "ConsumedReadCapacityUnits", "TableName", "${local.name_prefix}-data"],
            [".", "ConsumedWriteCapacityUnits", ".", "."]
          ]
          stat   = "Sum"
          period = 86400
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 12
        height = 6
        properties = {
          title  = "API Gateway Requests"
          region = var.aws_region
          metrics = [
            ["AWS/ApiGateway", "Count", "ApiName", "${local.name_prefix}-api"]
          ]
          stat   = "Sum"
          period = 86400
        }
      }
    ]
  })
}
```

### Cost Optimization Checklist

- [ ] Use PAY_PER_REQUEST for variable DynamoDB workloads
- [ ] Enable S3 lifecycle policies to transition to cheaper storage
- [ ] Use ARM64 Lambda architecture (34% cheaper)
- [ ] Right-size Lambda memory allocation
- [ ] Set appropriate Lambda timeouts
- [ ] Enable DynamoDB TTL for auto-cleanup
- [ ] Use S3 Intelligent-Tiering for unpredictable access
- [ ] Enable cost allocation tags
- [ ] Set up AWS Budgets with alerts
- [ ] Review unused resources regularly

### References

- [HashiCorp - Configure Default Tags](https://developer.hashicorp.com/terraform/tutorials/aws/aws-default-tags)
- [AWS - Cost Allocation Tags](https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/cost-alloc-tags.html)
- [Terraform - Resource Tagging Guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/resource-tagging)
- [AWS - Lambda Cost Optimization](https://docs.aws.amazon.com/lambda/latest/operatorguide/profile-costs.html)

---

## Appendix A: Complete Project Structure

```
hg-mcp-infrastructure/
├── .github/
│   └── workflows/
│       ├── terraform.yml           # Main CI/CD pipeline
│       ├── terraform-pr-check.yml  # PR validation
│       └── terraform-drift.yml     # Drift detection
├── modules/
│   ├── dynamodb/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   ├── lambda/
│   │   ├── main.tf
│   │   ├── iam.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   ├── api-gateway/
│   │   ├── rest-api.tf
│   │   ├── http-api.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   ├── s3/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   └── iam/
│       ├── lambda-role.tf
│       ├── github-oidc.tf
│       ├── variables.tf
│       └── outputs.tf
├── environments/
│   ├── dev/
│   │   ├── main.tf
│   │   ├── providers.tf
│   │   ├── backend.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── terraform.tfvars
│   ├── staging/
│   │   └── ... (same structure)
│   └── prod/
│       └── ... (same structure)
├── shared/
│   ├── state-backend/
│   │   ├── main.tf           # S3 bucket + KMS for state
│   │   └── outputs.tf
│   └── iam/
│       └── github-oidc.tf    # GitHub Actions OIDC setup
├── scripts/
│   ├── init-backend.sh       # Bootstrap state bucket
│   └── validate-all.sh       # Validate all environments
├── .terraform-version        # tfenv version file
├── .pre-commit-config.yaml   # Pre-commit hooks
└── README.md
```

---

## Appendix B: Quick Start Commands

```bash
# Initialize new environment
cd environments/dev && terraform init

# Format check
terraform fmt -check -recursive

# Validate configuration
terraform validate

# Plan changes
terraform plan -out=tfplan

# Apply changes
terraform apply tfplan

# Destroy (with confirmation)
terraform destroy

# Import existing resource
terraform import aws_dynamodb_table.main hg-mcp-data-dev

# State management
terraform state list
terraform state show aws_lambda_function.main
terraform state mv aws_lambda_function.old aws_lambda_function.new

# Workspace commands
terraform workspace list
terraform workspace new staging
terraform workspace select prod
```

---

## Appendix C: Additional Resources

### Official Documentation
- [Terraform AWS Provider Documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [AWS Prescriptive Guidance - Terraform Best Practices](https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/welcome.html)
- [HashiCorp Learn - AWS Tutorials](https://developer.hashicorp.com/terraform/tutorials/aws)

### Community Modules
- [terraform-aws-modules/lambda](https://registry.terraform.io/modules/terraform-aws-modules/lambda/aws)
- [terraform-aws-modules/dynamodb-table](https://registry.terraform.io/modules/terraform-aws-modules/dynamodb-table/aws)
- [terraform-aws-modules/apigateway-v2](https://registry.terraform.io/modules/terraform-aws-modules/apigateway-v2/aws)

### Tools
- [tfsec](https://github.com/aquasecurity/tfsec) - Security scanner
- [checkov](https://github.com/bridgecrewio/checkov) - Policy-as-code
- [infracost](https://github.com/infracost/infracost) - Cost estimation
- [terraform-docs](https://github.com/terraform-docs/terraform-docs) - Documentation generator
- [pre-commit-terraform](https://github.com/antonbabenko/pre-commit-terraform) - Pre-commit hooks

---

*This document was researched and compiled in December 2025. Terraform and AWS services evolve rapidly; always verify against current documentation before implementing in production.*
