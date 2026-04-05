# S3 Outputs
output "music_storage_bucket_name" {
  description = "Name of the S3 bucket for music storage"
  value       = aws_s3_bucket.music_storage.id
}

output "music_storage_bucket_arn" {
  description = "ARN of the S3 bucket for music storage"
  value       = aws_s3_bucket.music_storage.arn
}

output "terraform_state_bucket_name" {
  description = "Name of the Terraform state bucket (managed externally)"
  value       = "cr8-terraform-state-REDACTED_AWS_ACCOUNT"
}

# DynamoDB Table ARNs
output "dynamodb_table_arns" {
  description = "ARNs of all DynamoDB tables"
  value = {
    tracks           = aws_dynamodb_table.tracks.arn
    playlists        = aws_dynamodb_table.playlists.arn
    playlist_tracks  = aws_dynamodb_table.playlist_tracks.arn
    sync_queue       = aws_dynamodb_table.sync_queue.arn
    audio_analysis   = aws_dynamodb_table.audio_analysis.arn
    sync_history     = aws_dynamodb_table.sync_history.arn
    user_preferences = aws_dynamodb_table.user_preferences.arn
    playlist_state   = aws_dynamodb_table.playlist_state.arn
    track_history    = aws_dynamodb_table.track_history.arn
  }
}

# DynamoDB Table Names
output "dynamodb_table_names" {
  description = "Names of all DynamoDB tables"
  value = {
    tracks           = aws_dynamodb_table.tracks.name
    playlists        = aws_dynamodb_table.playlists.name
    playlist_tracks  = aws_dynamodb_table.playlist_tracks.name
    sync_queue       = aws_dynamodb_table.sync_queue.name
    audio_analysis   = aws_dynamodb_table.audio_analysis.name
    sync_history     = aws_dynamodb_table.sync_history.name
    user_preferences = aws_dynamodb_table.user_preferences.name
    playlist_state   = aws_dynamodb_table.playlist_state.name
    track_history    = aws_dynamodb_table.track_history.name
  }
}

# GSI Names for reference
output "dynamodb_gsi_names" {
  description = "GSI names for each table"
  value = {
    tracks       = ["url-index", "bpm_bucket-index", "artist-index"]
    playlists    = ["user_id-index", "status-index"]
    sync_queue   = ["status-index", "priority-index"]
    sync_history = ["playlist_id-index"]
  }
}

# Region
output "aws_region" {
  description = "AWS region where resources are deployed"
  value       = var.aws_region
}
