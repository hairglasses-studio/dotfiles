# AWS Secrets Manager for Inventory API credentials

# eBay API Credentials
resource "aws_secretsmanager_secret" "ebay_credentials" {
  name        = "aftrs/inventory/ebay-credentials"
  description = "eBay API credentials for inventory listing management"

  tags = {
    Project = "aftrs"
    Service = "inventory"
  }
}

# Note: Secret value should be set manually or via CLI:
# aws secretsmanager put-secret-value --secret-id aftrs/inventory/ebay-credentials --secret-string '{
#   "app_id": "MitchMit-priceche-PRD-f71b3bcb7-8624c0f6",
#   "dev_id": "4b99c522-1882-450d-9950-1adf6207852a",
#   "cert_id": "YOUR_CERT_ID",
#   "sandbox_app_id": "MitchMit-priceche-SBX-f71b3bcb7-9abe6097",
#   "oauth_refresh_token": "YOUR_REFRESH_TOKEN",
#   "environment": "production"
# }'

# Gmail API Credentials (for order email parsing)
resource "aws_secretsmanager_secret" "gmail_credentials" {
  name        = "aftrs/inventory/gmail-credentials"
  description = "Gmail OAuth credentials for order email parsing"

  tags = {
    Project = "aftrs"
    Service = "inventory"
  }
}

# Note: Secret value should include:
# - client_id, client_secret (from Google Cloud Console)
# - refresh_token (from OAuth flow)

# Output the secret ARNs
output "ebay_credentials_secret_arn" {
  description = "ARN of the eBay credentials secret"
  value       = aws_secretsmanager_secret.ebay_credentials.arn
}

output "gmail_credentials_secret_arn" {
  description = "ARN of the Gmail credentials secret"
  value       = aws_secretsmanager_secret.gmail_credentials.arn
}
