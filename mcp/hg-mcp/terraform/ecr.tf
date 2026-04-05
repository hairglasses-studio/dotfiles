# ECR Repositories for CR8 containers

# MCP Server container repository
resource "aws_ecr_repository" "mcp_server" {
  name                 = "cr8-cli/mcp"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Project   = "cr8-cli"
    Component = "mcp-server"
    ManagedBy = "terraform"
  }
}

resource "aws_ecr_lifecycle_policy" "mcp_server" {
  repository = aws_ecr_repository.mcp_server.name

  policy = jsonencode({
    rules = [{
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
    }]
  })
}

# Outputs
output "mcp_server_repository_url" {
  description = "ECR repository URL for MCP server container"
  value       = aws_ecr_repository.mcp_server.repository_url
}
