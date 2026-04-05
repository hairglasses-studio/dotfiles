# IAM Role for GitHub Actions with OIDC authentication
# Provides permissions for Terraform state management and infrastructure provisioning

locals {
  github_org  = "aftrs-studio"
  github_repo = "hg-mcp"
}

# Trust policy allowing GitHub Actions to assume this role via OIDC
data "aws_iam_policy_document" "github_actions_assume_role" {
  statement {
    sid     = "GitHubActionsOIDC"
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github_actions.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    # Restrict to specific repository and branches
    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values = [
        "repo:${local.github_org}/${local.github_repo}:ref:refs/heads/main",
        "repo:${local.github_org}/${local.github_repo}:pull_request"
      ]
    }
  }
}

# IAM role that GitHub Actions will assume
resource "aws_iam_role" "github_actions" {
  name               = "github-actions-hg-mcp"
  assume_role_policy = data.aws_iam_policy_document.github_actions_assume_role.json

  tags = {
    Name      = "github-actions-hg-mcp"
    Project   = "hg-mcp"
    ManagedBy = "terraform"
  }
}

# Policy for Terraform state management (S3 + DynamoDB)
data "aws_iam_policy_document" "terraform_state" {
  # S3 bucket permissions for state storage
  statement {
    sid    = "S3StateAccess"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
      "s3:ListBucket",
      "s3:GetBucketVersioning"
    ]
    resources = [
      "arn:aws:s3:::cr8-terraform-state-804005416684",
      "arn:aws:s3:::cr8-terraform-state-804005416684/*"
    ]
  }

  # DynamoDB permissions for state locking
  statement {
    sid    = "DynamoDBStateLocking"
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:DeleteItem",
      "dynamodb:DescribeTable"
    ]
    resources = [
      "arn:aws:dynamodb:us-east-1:*:table/cr8-terraform-locks"
    ]
  }
}

resource "aws_iam_policy" "terraform_state" {
  name        = "terraform-state-hg-mcp"
  description = "Permissions for Terraform state management in S3 and DynamoDB"
  policy      = data.aws_iam_policy_document.terraform_state.json

  tags = {
    Name      = "terraform-state-hg-mcp"
    Project   = "hg-mcp"
    ManagedBy = "terraform"
  }
}

resource "aws_iam_role_policy_attachment" "terraform_state" {
  role       = aws_iam_role.github_actions.name
  policy_arn = aws_iam_policy.terraform_state.arn
}

# Output the role ARN for use in GitHub Actions secrets
output "github_actions_role_arn" {
  description = "ARN of the IAM role for GitHub Actions (set as AWS_ROLE_ARN secret)"
  value       = aws_iam_role.github_actions.arn
}
