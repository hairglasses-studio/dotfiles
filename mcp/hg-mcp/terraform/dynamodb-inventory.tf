# aftrs_inventory - Electronics inventory management (single-table design)
resource "aws_dynamodb_table" "inventory" {
  name         = "aftrs_inventory"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "PK"
  range_key    = "SK"

  # Primary key attributes
  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  # GSI1: Category + Price (browse by category, sorted by price)
  attribute {
    name = "GSI1PK"
    type = "S"
  }

  attribute {
    name = "GSI1SK"
    type = "S"
  }

  # GSI2: Status + Date (filter by listing status)
  attribute {
    name = "GSI2PK"
    type = "S"
  }

  attribute {
    name = "GSI2SK"
    type = "S"
  }

  # GSI3: Location + Category (physical organization)
  attribute {
    name = "GSI3PK"
    type = "S"
  }

  attribute {
    name = "GSI3SK"
    type = "S"
  }

  global_secondary_index {
    name            = "GSI1-category-price"
    hash_key        = "GSI1PK"
    range_key       = "GSI1SK"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "GSI2-status-date"
    hash_key        = "GSI2PK"
    range_key       = "GSI2SK"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "GSI3-location-category"
    hash_key        = "GSI3PK"
    range_key       = "GSI3SK"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  tags = {
    Project = "aftrs"
    Service = "inventory"
  }

  lifecycle {
    prevent_destroy = true
  }
}

# Output the table name and ARN for use in application config
output "inventory_table_name" {
  description = "Name of the inventory DynamoDB table"
  value       = aws_dynamodb_table.inventory.name
}

output "inventory_table_arn" {
  description = "ARN of the inventory DynamoDB table"
  value       = aws_dynamodb_table.inventory.arn
}
