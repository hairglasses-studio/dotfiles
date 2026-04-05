# CR8 Audio Analysis Lambda

Container-based AWS Lambda for processing the audio analysis queue. Uses librosa for BPM/key detection.

## Architecture

```
EventBridge Scheduler (every 5 min)
        │
        ▼
    Lambda (container)
        │
        ├──► DynamoDB (cr8_sync_queue) - claim items
        ├──► S3 (cr8-music-storage) - download audio
        ├──► librosa - analyze BPM/key
        └──► DynamoDB (cr8_tracks) - update results
```

## Deployment

### 1. Build and Push Container Image

```bash
# Set your AWS account ID
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export AWS_REGION=us-east-1

# Login to ECR
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

# Build the image
cd terraform/modules/cr8-analysis-lambda/lambda
docker build --platform linux/amd64 -t cr8-analysis-lambda .

# Tag and push
docker tag cr8-analysis-lambda:latest $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/cr8-analysis-lambda:latest
docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/cr8-analysis-lambda:latest
```

### 2. Deploy with Terraform

```bash
cd terraform
terraform init
terraform apply -target=module.cr8_analysis_lambda
```

### 3. Update Lambda to Use New Image

After pushing a new container image:

```bash
aws lambda update-function-code \
  --function-name cr8-analysis-worker \
  --image-uri $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/cr8-analysis-lambda:latest
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `batch_size` | 20 | Tracks per invocation |
| `schedule_rate` | 5 min | How often to run |
| `enabled` | true | Enable/disable scheduler |

## Manual Invocation

```bash
# Invoke directly (for testing)
aws lambda invoke \
  --function-name cr8-analysis-worker \
  --payload '{"batch_size": 10}' \
  response.json

cat response.json
```

## Monitoring

```bash
# View recent logs
aws logs tail /aws/lambda/cr8-analysis-worker --follow

# Check queue depth
aws dynamodb scan \
  --table-name cr8_sync_queue \
  --filter-expression "queue_type = :qt AND #status = :s" \
  --expression-attribute-names '{"#status":"status"}' \
  --expression-attribute-values '{":qt":{"S":"analysis"},":s":{"S":"pending"}}' \
  --select COUNT
```

## Cost Estimate

With 5,737 tracks in queue and 20 tracks per invocation:
- ~287 invocations needed
- At 3GB memory, ~2-3 min per invocation
- Estimated cost: ~$1-2 total

Ongoing (after queue cleared):
- EventBridge: Free tier (14M/month)
- Lambda: ~$0.50/month (assuming ~100 new tracks/day)
