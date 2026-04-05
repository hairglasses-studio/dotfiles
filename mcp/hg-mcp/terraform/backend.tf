terraform {
  backend "s3" {
    bucket         = "cr8-terraform-state-REDACTED_AWS_ACCOUNT"
    key            = "cr8/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "cr8-terraform-locks"
  }
}
