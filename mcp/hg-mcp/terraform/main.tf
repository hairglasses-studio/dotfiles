terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "cr8"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

# Sydney region provider for archive buckets
provider "aws" {
  alias  = "sydney"
  region = "ap-southeast-2"

  default_tags {
    tags = {
      Project     = "aftrs"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

# CR8 Audio Analysis Lambda
module "cr8_analysis_lambda" {
  source = "./modules/cr8-analysis-lambda"

  environment   = var.environment
  batch_size    = 20
  schedule_rate = "rate(5 minutes)"
  enabled       = true
}
