# cr8_tracks - Main track metadata table
resource "aws_dynamodb_table" "tracks" {
  name         = "cr8_tracks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "url"
    type = "S"
  }

  attribute {
    name = "bpm_bucket"
    type = "S"
  }

  attribute {
    name = "artist"
    type = "S"
  }

  attribute {
    name = "label"
    type = "S"
  }

  global_secondary_index {
    name            = "url-index"
    hash_key        = "url"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "bpm_bucket-index"
    hash_key        = "bpm_bucket"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "artist-index"
    hash_key        = "artist"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "label-index"
    hash_key        = "label"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  lifecycle {
    prevent_destroy = true
  }
}

# cr8_playlists - Playlist metadata
resource "aws_dynamodb_table" "playlists" {
  name         = "cr8_playlists"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "user_id"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  global_secondary_index {
    name            = "user_id-index"
    hash_key        = "user_id"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "status-index"
    hash_key        = "status"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  lifecycle {
    prevent_destroy = true
  }
}

# cr8_playlist_tracks - Playlist-track relationship (composite key)
resource "aws_dynamodb_table" "playlist_tracks" {
  name         = "cr8_playlist_tracks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "playlist_id"
  range_key    = "track_id"

  attribute {
    name = "playlist_id"
    type = "S"
  }

  attribute {
    name = "track_id"
    type = "S"
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  lifecycle {
    prevent_destroy = true
  }
}

# cr8_sync_queue - Background sync processing queue
resource "aws_dynamodb_table" "sync_queue" {
  name         = "cr8_sync_queue"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  attribute {
    name = "priority"
    type = "N"
  }

  global_secondary_index {
    name            = "status-index"
    hash_key        = "status"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "priority-index"
    hash_key        = "priority"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  lifecycle {
    prevent_destroy = true
  }
}

# cr8_audio_analysis - Audio analysis results
resource "aws_dynamodb_table" "audio_analysis" {
  name         = "cr8_audio_analysis"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "track_id"

  attribute {
    name = "track_id"
    type = "S"
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  lifecycle {
    prevent_destroy = true
  }
}

# cr8_sync_history - Historical sync operations
resource "aws_dynamodb_table" "sync_history" {
  name         = "cr8_sync_history"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "playlist_id"
    type = "S"
  }

  global_secondary_index {
    name            = "playlist_id-index"
    hash_key        = "playlist_id"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  lifecycle {
    prevent_destroy = true
  }
}

# cr8_user_preferences - User settings and preferences
resource "aws_dynamodb_table" "user_preferences" {
  name         = "cr8_user_preferences"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "user_id"

  attribute {
    name = "user_id"
    type = "S"
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  lifecycle {
    prevent_destroy = true
  }
}

# cr8_playlist_state - PlaylistTracker state (existing)
resource "aws_dynamodb_table" "playlist_state" {
  name         = "cr8_playlist_state"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  lifecycle {
    prevent_destroy = true
  }
}

# cr8_track_history - PlaylistTracker history (existing)
resource "aws_dynamodb_table" "track_history" {
  name         = "cr8_track_history"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  lifecycle {
    prevent_destroy = true
  }
}

# cr8_sync_state - Sync state per service/user/playlist
resource "aws_dynamodb_table" "sync_state" {
  name         = "cr8_sync_state"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "playlist_id"

  attribute {
    name = "playlist_id"
    type = "S"
  }

  attribute {
    name = "service"
    type = "S"
  }

  attribute {
    name = "user"
    type = "S"
  }

  global_secondary_index {
    name            = "service-index"
    hash_key        = "service"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "user-index"
    hash_key        = "user"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  lifecycle {
    prevent_destroy = true
  }
}
