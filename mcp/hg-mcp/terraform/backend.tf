terraform {
  backend "s3" {
    bucket         = "cr8-terraform-state-804005416684"
    key            = "cr8/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "cr8-terraform-locks"
  }
}
