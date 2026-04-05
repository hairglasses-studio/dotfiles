#!/bin/bash
# Build and push CR8 Analysis Lambda container image

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AWS_REGION="${AWS_REGION:-us-east-1}"
AWS_PROFILE="${AWS_PROFILE:-cr8}"
IMAGE_NAME="cr8-analysis-lambda"

echo "=== CR8 Analysis Lambda Build & Push ==="

# Get AWS account ID
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --profile "$AWS_PROFILE" --query Account --output text)
echo "AWS Account: $AWS_ACCOUNT_ID"
echo "Region: $AWS_REGION"

ECR_URI="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$IMAGE_NAME"

# Login to ECR
echo "Logging in to ECR..."
aws ecr get-login-password --region "$AWS_REGION" --profile "$AWS_PROFILE" | docker login --username AWS --password-stdin "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"

# Build the image (--provenance=false required for Lambda compatibility)
echo "Building Docker image..."
cd "$SCRIPT_DIR/lambda"
docker build --platform linux/amd64 --provenance=false -t "$IMAGE_NAME" .

# Tag and push
echo "Pushing to ECR..."
docker tag "$IMAGE_NAME:latest" "$ECR_URI:latest"
docker push "$ECR_URI:latest"

echo "=== Done ==="
echo "Image URI: $ECR_URI:latest"
echo ""
echo "To update Lambda function:"
echo "  aws lambda update-function-code --function-name cr8-analysis-worker --image-uri $ECR_URI:latest --profile $AWS_PROFILE"
