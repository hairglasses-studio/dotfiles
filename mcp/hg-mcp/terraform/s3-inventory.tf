# S3 Bucket for Inventory Assets (images, receipts, imports)
resource "aws_s3_bucket" "inventory_assets" {
  bucket = "aftrs-inventory-assets-${data.aws_caller_identity.current.account_id}"

  tags = {
    Project = "aftrs"
    Service = "inventory"
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_versioning" "inventory_assets" {
  bucket = aws_s3_bucket.inventory_assets.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "inventory_assets" {
  bucket = aws_s3_bucket.inventory_assets.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "inventory_assets" {
  bucket = aws_s3_bucket.inventory_assets.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "inventory_assets" {
  bucket = aws_s3_bucket.inventory_assets.id

  # Cleanup incomplete multipart uploads
  rule {
    id     = "cleanup-incomplete-uploads"
    status = "Enabled"

    filter {}

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }

  # Archive old import files after 90 days, delete after 1 year
  rule {
    id     = "archive-imports"
    status = "Enabled"

    filter {
      prefix = "imports/"
    }

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    expiration {
      days = 365
    }
  }

  # Transition old receipt versions to IA, keep indefinitely
  rule {
    id     = "receipts-lifecycle"
    status = "Enabled"

    filter {
      prefix = "receipts/"
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }
  }

  # Transition old image versions to IA
  rule {
    id     = "images-lifecycle"
    status = "Enabled"

    filter {
      prefix = "images/"
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }

    noncurrent_version_expiration {
      noncurrent_days = 180
    }
  }
}

# CORS configuration for presigned URL uploads
resource "aws_s3_bucket_cors_configuration" "inventory_assets" {
  bucket = aws_s3_bucket.inventory_assets.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST"]
    allowed_origins = ["*"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3600
  }
}

# Output the bucket name and ARN
output "inventory_bucket_name" {
  description = "Name of the inventory assets S3 bucket"
  value       = aws_s3_bucket.inventory_assets.id
}

output "inventory_bucket_arn" {
  description = "ARN of the inventory assets S3 bucket"
  value       = aws_s3_bucket.inventory_assets.arn
}
