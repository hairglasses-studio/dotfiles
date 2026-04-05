# S3 Buckets for AFTRS Media and Plugin Storage
# Created for Resolume plugins, VJ clips, and future creative tools

locals {
  media_buckets = {
    resolume_plugins = {
      name        = "aftrs-resolume-plugins"
      description = "Resolume FFGL plugins and ISF shaders"
    }
    vj_clips = {
      name        = "aftrs-vj-clips"
      description = "VJ video clips for Resolume"
    }
    touchdesigner = {
      name        = "aftrs-touchdesigner-assets"
      description = "TouchDesigner TOX components and assets"
    }
    obs = {
      name        = "aftrs-obs-assets"
      description = "OBS scenes and overlays"
    }
    grandma3 = {
      name        = "aftrs-grandma3-assets"
      description = "GrandMA3 lighting show files"
    }
    ledfx = {
      name        = "aftrs-ledfx-assets"
      description = "LedFX effects and configurations"
    }
  }

  media_tags = {
    Project   = "aftrs-studio"
    Component = "media-storage"
    ManagedBy = "terraform"
  }
}

# Resolume Plugins Bucket
resource "aws_s3_bucket" "resolume_plugins" {
  bucket = local.media_buckets.resolume_plugins.name

  tags = merge(local.media_tags, {
    Description = local.media_buckets.resolume_plugins.description
  })
}

resource "aws_s3_bucket_versioning" "resolume_plugins" {
  bucket = aws_s3_bucket.resolume_plugins.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "resolume_plugins" {
  bucket = aws_s3_bucket.resolume_plugins.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "resolume_plugins" {
  bucket                  = aws_s3_bucket.resolume_plugins.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "resolume_plugins" {
  bucket = aws_s3_bucket.resolume_plugins.id

  rule {
    id     = "cleanup-incomplete-uploads"
    status = "Enabled"
    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }

  rule {
    id     = "transition-old-versions"
    status = "Enabled"
    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }
    noncurrent_version_expiration {
      noncurrent_days = 180
    }
  }
}

# VJ Clips Bucket
resource "aws_s3_bucket" "vj_clips" {
  bucket = local.media_buckets.vj_clips.name

  tags = merge(local.media_tags, {
    Description = local.media_buckets.vj_clips.description
  })
}

resource "aws_s3_bucket_versioning" "vj_clips" {
  bucket = aws_s3_bucket.vj_clips.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "vj_clips" {
  bucket = aws_s3_bucket.vj_clips.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "vj_clips" {
  bucket                  = aws_s3_bucket.vj_clips.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "vj_clips" {
  bucket = aws_s3_bucket.vj_clips.id

  rule {
    id     = "cleanup-incomplete-uploads"
    status = "Enabled"
    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }

  rule {
    id     = "transition-old-versions"
    status = "Enabled"
    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }
    noncurrent_version_expiration {
      noncurrent_days = 180
    }
  }

  rule {
    id     = "intelligent-tiering"
    status = "Enabled"
    transition {
      days          = 90
      storage_class = "INTELLIGENT_TIERING"
    }
  }
}

# TouchDesigner Assets Bucket
resource "aws_s3_bucket" "touchdesigner" {
  bucket = local.media_buckets.touchdesigner.name

  tags = merge(local.media_tags, {
    Description = local.media_buckets.touchdesigner.description
  })
}

resource "aws_s3_bucket_versioning" "touchdesigner" {
  bucket = aws_s3_bucket.touchdesigner.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "touchdesigner" {
  bucket = aws_s3_bucket.touchdesigner.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "touchdesigner" {
  bucket                  = aws_s3_bucket.touchdesigner.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# OBS Assets Bucket
resource "aws_s3_bucket" "obs" {
  bucket = local.media_buckets.obs.name

  tags = merge(local.media_tags, {
    Description = local.media_buckets.obs.description
  })
}

resource "aws_s3_bucket_versioning" "obs" {
  bucket = aws_s3_bucket.obs.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "obs" {
  bucket = aws_s3_bucket.obs.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "obs" {
  bucket                  = aws_s3_bucket.obs.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# GrandMA3 Assets Bucket
resource "aws_s3_bucket" "grandma3" {
  bucket = local.media_buckets.grandma3.name

  tags = merge(local.media_tags, {
    Description = local.media_buckets.grandma3.description
  })
}

resource "aws_s3_bucket_versioning" "grandma3" {
  bucket = aws_s3_bucket.grandma3.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "grandma3" {
  bucket = aws_s3_bucket.grandma3.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "grandma3" {
  bucket                  = aws_s3_bucket.grandma3.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# LedFX Assets Bucket
resource "aws_s3_bucket" "ledfx" {
  bucket = local.media_buckets.ledfx.name

  tags = merge(local.media_tags, {
    Description = local.media_buckets.ledfx.description
  })
}

resource "aws_s3_bucket_versioning" "ledfx" {
  bucket = aws_s3_bucket.ledfx.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "ledfx" {
  bucket = aws_s3_bucket.ledfx.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "ledfx" {
  bucket                  = aws_s3_bucket.ledfx.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Outputs
output "resolume_plugins_bucket_name" {
  description = "Resolume plugins S3 bucket name"
  value       = aws_s3_bucket.resolume_plugins.id
}

output "resolume_plugins_bucket_arn" {
  description = "Resolume plugins S3 bucket ARN"
  value       = aws_s3_bucket.resolume_plugins.arn
}

output "vj_clips_bucket_name" {
  description = "VJ clips S3 bucket name"
  value       = aws_s3_bucket.vj_clips.id
}

output "vj_clips_bucket_arn" {
  description = "VJ clips S3 bucket ARN"
  value       = aws_s3_bucket.vj_clips.arn
}

output "touchdesigner_bucket_name" {
  description = "TouchDesigner assets S3 bucket name"
  value       = aws_s3_bucket.touchdesigner.id
}

output "obs_bucket_name" {
  description = "OBS assets S3 bucket name"
  value       = aws_s3_bucket.obs.id
}

output "grandma3_bucket_name" {
  description = "GrandMA3 assets S3 bucket name"
  value       = aws_s3_bucket.grandma3.id
}

output "ledfx_bucket_name" {
  description = "LedFX assets S3 bucket name"
  value       = aws_s3_bucket.ledfx.id
}
