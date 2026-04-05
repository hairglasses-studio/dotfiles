# Archive S3 Buckets for DJ and VJ content
# Located in ap-southeast-2 (Sydney) for local access

# DJ Archive - DJ set recordings and mixes
resource "aws_s3_bucket" "dj_archive" {
  provider = aws.sydney
  bucket   = "aftrs-dj-archive"

  lifecycle {
    prevent_destroy = true
  }

  tags = {
    Project   = "aftrs"
    Component = "dj-archive"
    ManagedBy = "terraform"
  }
}

resource "aws_s3_bucket_versioning" "dj_archive" {
  provider = aws.sydney
  bucket   = aws_s3_bucket.dj_archive.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "dj_archive" {
  provider = aws.sydney
  bucket   = aws_s3_bucket.dj_archive.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "dj_archive" {
  provider = aws.sydney
  bucket   = aws_s3_bucket.dj_archive.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# VJ Archive - Visual content and recordings
resource "aws_s3_bucket" "vj_archive" {
  provider = aws.sydney
  bucket   = "aftrs-vj-archive"

  lifecycle {
    prevent_destroy = true
  }

  tags = {
    Project   = "aftrs"
    Component = "vj-archive"
    ManagedBy = "terraform"
  }
}

resource "aws_s3_bucket_versioning" "vj_archive" {
  provider = aws.sydney
  bucket   = aws_s3_bucket.vj_archive.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "vj_archive" {
  provider = aws.sydney
  bucket   = aws_s3_bucket.vj_archive.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "vj_archive" {
  provider = aws.sydney
  bucket   = aws_s3_bucket.vj_archive.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Outputs
output "dj_archive_bucket_name" {
  description = "Name of the DJ archive S3 bucket (ap-southeast-2)"
  value       = aws_s3_bucket.dj_archive.id
}

output "vj_archive_bucket_name" {
  description = "Name of the VJ archive S3 bucket (ap-southeast-2)"
  value       = aws_s3_bucket.vj_archive.id
}
