# Discord Bot with AWS Lambda: Comprehensive Research

> Research document for implementing a serverless Discord bot for the AFTRS MCP project with Claude AI integration.

**Date:** December 2024
**Status:** Research Complete

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Discord HTTP Interactions vs WebSocket Gateway](#discord-http-interactions-vs-websocket-gateway)
4. [AWS Lambda + API Gateway Implementation](#aws-lambda--api-gateway-implementation)
5. [Signature Verification](#signature-verification)
6. [DynamoDB for Conversation State](#dynamodb-for-conversation-state)
7. [Claude API Integration](#claude-api-integration)
8. [Rate Limiting](#rate-limiting)
9. [Cold Start Mitigation](#cold-start-mitigation)
10. [Security Best Practices](#security-best-practices)
11. [Cost Analysis](#cost-analysis)
12. [Step-by-Step Implementation Guide](#step-by-step-implementation-guide)
13. [Example Implementations](#example-implementations)
14. [References](#references)

---

## Executive Summary

This document analyzes the feasibility and implementation patterns for deploying a Discord bot on AWS Lambda that integrates with Claude AI for the AFTRS MCP project. The serverless approach offers significant advantages for a 2-user server:

**Key Benefits:**
- Near-zero cost within AWS Free Tier
- No server maintenance required
- Pay-per-use model scales with demand
- Perfect for slash commands and AI interactions

**Key Challenges:**
- 3-second response deadline requires deferred response pattern
- Cold starts can cause timeouts without mitigation
- Cannot listen for real-time events (message streams, presence, voice)

**Recommendation:** Use HTTP Interactions Endpoint with a multi-Lambda architecture that defers AI responses and stores conversation history in DynamoDB.

---

## Architecture Overview

### High-Level Architecture Diagram

```
                                    AWS Cloud
+------------------+           +------------------------------------------+
|                  |           |                                          |
|  Discord Server  |  HTTPS    |  +----------------+    +---------------+ |
|  (2 users)       +---------->|  | API Gateway    |--->| Proxy Lambda  | |
|                  |           |  | (HTTP API)     |    | (Verify + ACK)| |
|  /ask command    |           |  +----------------+    +-------+-------+ |
|  @bot mentions   |           |                                |         |
|                  |           |                    SNS/SQS     v         |
+------------------+           |                        +-------+-------+ |
        ^                      |                        | Worker Lambda | |
        |                      |                        | (Claude API)  | |
        |  Webhook             |                        +-------+-------+ |
        |  (followup)          |                                |         |
        |                      |    +----------------+          |         |
        +----------------------|----+ DynamoDB       |<---------+         |
                               |    | (Sessions)     |                    |
                               |    +----------------+                    |
                               |                                          |
                               |    +----------------+                    |
                               |    | Secrets Manager|                    |
                               |    | (Tokens/Keys)  |                    |
                               |    +----------------+                    |
                               +------------------------------------------+
```

### Request Flow Sequence

```
User          Discord        API Gateway    Proxy Lambda    Worker Lambda    Claude API
 |               |                |              |               |               |
 | /ask "hello"  |                |              |               |               |
 |-------------->|                |              |               |               |
 |               | POST request   |              |               |               |
 |               | + signature    |              |               |               |
 |               |--------------->|              |               |               |
 |               |                | Forward      |               |               |
 |               |                |------------->|               |               |
 |               |                |              | 1. Verify sig |               |
 |               |                |              | 2. ACK defer  |               |
 |               |                |<-------------|    (type: 5)  |               |
 |               |<---------------|              |               |               |
 | "Bot thinking"|                |              |               |               |
 |<--------------|                |              | 3. Invoke     |               |
 |               |                |              |   async       |               |
 |               |                |              |-------------->|               |
 |               |                |              |               | 4. Call Claude|
 |               |                |              |               |-------------->|
 |               |                |              |               |<--------------|
 |               |                |              |               | 5. Store in DB|
 |               | PATCH webhook  |              |               |               |
 |               |<---------------|--------------|---------------|               |
 | Final answer  |                |              |               |               |
 |<--------------|                |              |               |               |
```

---

## Discord HTTP Interactions vs WebSocket Gateway

### Comparison Table

| Feature | WebSocket Gateway | HTTP Interactions |
|---------|------------------|-------------------|
| Connection Type | Persistent WebSocket | Stateless HTTP POST |
| Serverless Compatible | No | **Yes** |
| Slash Commands | Yes | **Yes** |
| Buttons/Modals | Yes | **Yes** |
| Message Events | **Yes** | No |
| Presence/Voice Events | **Yes** | No |
| Hosting Cost | 24/7 server required | **Pay-per-request** |
| Complexity | Higher | **Lower** |
| Sharding Required | Yes (>2500 guilds) | **No** |

### Why HTTP Interactions for AFTRS

For a 2-user Discord server focused on slash commands and AI chat:

1. **Slash commands work perfectly** - `/ask`, `/studio-status`, etc.
2. **Button interactions work** - For confirmation dialogs, pagination
3. **Modal forms work** - For longer input collection
4. **Cost: Essentially free** - Within AWS Free Tier
5. **No 24/7 server needed** - Functions only run when invoked

### What You Cannot Do

- Listen for all messages in a channel (no @mention detection without Gateway)
- Detect when users join/leave voice channels
- Show "online" presence status
- React to message reactions in real-time

### Hybrid Alternative

If you need both slash commands AND message listening:

```
+-------------------+     +----------------------+
| VPS/EC2 Instance  |     | AWS Lambda           |
| ($5-10/month)     |     | (Pay-per-use)        |
|                   |     |                      |
| - WebSocket       |     | - Heavy compute      |
| - Message events  |---->| - AI processing      |
| - Presence        |     | - Async tasks        |
+-------------------+     +----------------------+
```

---

## AWS Lambda + API Gateway Implementation

### Lambda Architecture Patterns

#### Pattern 1: Single Lambda (Simple, Cold Start Risk)

```
API Gateway --> Lambda (verify + process + respond)
```

**Pros:** Simple, fewer resources
**Cons:** Cold start can exceed 3-second deadline

#### Pattern 2: Dual Lambda (Recommended)

```
API Gateway --> Proxy Lambda (verify + defer) --> Worker Lambda (async)
                     |                                    |
                     v                                    v
              Return type:5                     PATCH followup webhook
```

**Pros:** Reliable 3-second compliance, async AI processing
**Cons:** More complexity, additional Lambda

### API Gateway Configuration

Use **HTTP API** (not REST API) for cost savings:

| Feature | REST API | HTTP API |
|---------|----------|----------|
| Price | $3.50/million | **$1.00/million** |
| Latency | Higher | **30% faster** |
| Features | Full | Essential |

### Proxy Lambda Implementation (Python)

```python
import json
import os
import boto3
from nacl.signing import VerifyKey
from nacl.exceptions import BadSignatureError

# Initialize outside handler for reuse
lambda_client = boto3.client('lambda')
PUBLIC_KEY = os.environ.get('DISCORD_PUBLIC_KEY')

def verify_signature(event):
    """Verify Discord request signature using Ed25519."""
    signature = event['headers'].get('x-signature-ed25519', '')
    timestamp = event['headers'].get('x-signature-timestamp', '')
    body = event.get('body', '')

    message = timestamp + body
    verify_key = VerifyKey(bytes.fromhex(PUBLIC_KEY))

    try:
        verify_key.verify(message.encode(), bytes.fromhex(signature))
        return True
    except BadSignatureError:
        return False

def lambda_handler(event, context):
    # Verify signature
    if not verify_signature(event):
        return {'statusCode': 401, 'body': 'Invalid signature'}

    body = json.loads(event.get('body', '{}'))
    interaction_type = body.get('type')

    # Handle PING (required for endpoint verification)
    if interaction_type == 1:
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'type': 1})
        }

    # Handle Application Command
    if interaction_type == 2:
        # Invoke worker Lambda asynchronously
        lambda_client.invoke(
            FunctionName=os.environ['WORKER_LAMBDA_ARN'],
            InvocationType='Event',  # Async
            Payload=json.dumps({
                'interaction': body,
                'token': body['token'],
                'application_id': body['application_id']
            })
        )

        # Return deferred response (type 5 = "thinking...")
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'type': 5})
        }

    return {'statusCode': 400, 'body': 'Unknown interaction type'}
```

### Worker Lambda Implementation (Python)

```python
import json
import os
import boto3
import httpx
from anthropic import Anthropic

dynamodb = boto3.resource('dynamodb')
table = dynamodb.Table(os.environ['SESSIONS_TABLE'])
anthropic = Anthropic(api_key=os.environ['ANTHROPIC_API_KEY'])

DISCORD_API_BASE = 'https://discord.com/api/v10'

def get_conversation_history(user_id: str, limit: int = 10) -> list:
    """Retrieve conversation history from DynamoDB."""
    response = table.query(
        KeyConditionExpression='user_id = :uid',
        ExpressionAttributeValues={':uid': user_id},
        ScanIndexForward=False,
        Limit=limit
    )
    messages = []
    for item in reversed(response.get('Items', [])):
        messages.append({'role': 'user', 'content': item['user_message']})
        messages.append({'role': 'assistant', 'content': item['assistant_response']})
    return messages

def save_conversation(user_id: str, user_message: str, assistant_response: str):
    """Save conversation turn to DynamoDB."""
    import time
    table.put_item(Item={
        'user_id': user_id,
        'timestamp': int(time.time() * 1000),
        'user_message': user_message,
        'assistant_response': assistant_response,
        'ttl': int(time.time()) + (90 * 24 * 60 * 60)  # 90 days TTL
    })

def send_followup(application_id: str, token: str, content: str):
    """Send followup message to Discord."""
    url = f"{DISCORD_API_BASE}/webhooks/{application_id}/{token}/messages/@original"
    httpx.patch(url, json={'content': content}, timeout=10)

def lambda_handler(event, context):
    interaction = event['interaction']
    token = event['token']
    application_id = event['application_id']

    user_id = interaction['member']['user']['id']
    user_input = interaction['data']['options'][0]['value']

    try:
        # Get conversation history
        history = get_conversation_history(user_id)

        # Build messages for Claude
        messages = history + [{'role': 'user', 'content': user_input}]

        # Call Claude API
        response = anthropic.messages.create(
            model='claude-sonnet-4-20250514',
            max_tokens=1024,
            system='You are a helpful AI assistant for the AFTRS studio Discord server.',
            messages=messages
        )

        assistant_message = response.content[0].text

        # Save to conversation history
        save_conversation(user_id, user_input, assistant_message)

        # Send response to Discord
        send_followup(application_id, token, assistant_message[:2000])

    except Exception as e:
        send_followup(application_id, token, f'Error: {str(e)[:1900]}')

    return {'statusCode': 200}
```

---

## Signature Verification

Discord uses Ed25519 signatures to verify that requests originate from Discord.

### Requirements

1. **Install PyNaCl** (required cryptography library)
2. **Get your Public Key** from Discord Developer Portal
3. **Verify EVERY request** before processing

### Verification Flow

```
Discord sends:
  - Header: x-signature-ed25519 (signature)
  - Header: x-signature-timestamp (timestamp)
  - Body: JSON payload

Your code:
  1. Concatenate: timestamp + body
  2. Verify signature using PUBLIC_KEY
  3. If invalid, return 401
  4. If valid, process request
```

### Lambda Layer for PyNaCl

Create a Lambda Layer with PyNaCl pre-installed:

```bash
mkdir -p python && pip install pynacl -t python/ && zip -r pynacl-layer.zip python
```

Then upload as a Lambda Layer and attach to your function.

### Important Notes

- **Header names are lowercase** in API Gateway (not `X-Signature-Ed25519`)
- Discord sends **invalid signatures intentionally** to test your verification
- Failure to verify will prevent saving your Interactions URL

---

## DynamoDB for Conversation State

### Table Design

```
Table Name: aftrs-discord-sessions

Partition Key: user_id (String)
Sort Key: timestamp (Number)

Attributes:
  - user_message (String)
  - assistant_response (String)
  - ttl (Number) - Unix timestamp for automatic deletion

Global Secondary Index (optional):
  - session_id-index for grouping conversations into sessions
```

### DynamoDB Schema (CloudFormation/CDK)

```yaml
SessionsTable:
  Type: AWS::DynamoDB::Table
  Properties:
    TableName: aftrs-discord-sessions
    BillingMode: PAY_PER_REQUEST
    AttributeDefinitions:
      - AttributeName: user_id
        AttributeType: S
      - AttributeName: timestamp
        AttributeType: N
    KeySchema:
      - AttributeName: user_id
        KeyType: HASH
      - AttributeName: timestamp
        KeyType: RANGE
    TimeToLiveSpecification:
      AttributeName: ttl
      Enabled: true
```

### Access Patterns

| Pattern | Query |
|---------|-------|
| Get user's recent messages | `user_id = :uid` (limit 10, desc) |
| Get all messages in timeframe | `user_id = :uid AND timestamp BETWEEN :start AND :end` |
| Delete user's data | `user_id = :uid` (for GDPR compliance) |

### Item Size Warning

DynamoDB has a **400 KB item limit**. For long conversations:
- Store each message turn as a separate item (not array in single item)
- Implement pagination for history retrieval
- Use TTL to automatically clean old messages

---

## Claude API Integration

### Conversation Context Pattern

```python
def build_claude_messages(history: list, current_message: str) -> list:
    """Build message array for Claude API with conversation history."""
    messages = []

    # Add history (limited to prevent token overflow)
    for turn in history[-10:]:  # Last 10 turns
        messages.append({'role': 'user', 'content': turn['user_message']})
        messages.append({'role': 'assistant', 'content': turn['assistant_response']})

    # Add current message
    messages.append({'role': 'user', 'content': current_message})

    return messages
```

### System Prompt for AFTRS

```python
SYSTEM_PROMPT = """You are the AFTRS Studio AI assistant, helping users with:
- Studio equipment status and troubleshooting
- TouchDesigner project questions
- Streaming (OBS) configuration
- General creative technology questions

Keep responses concise for Discord (under 2000 characters).
Be friendly but professional. If you don't know something about the specific
studio setup, say so and offer general guidance."""
```

### Token Management

| Model | Context Window | Input Price | Output Price |
|-------|---------------|-------------|--------------|
| Claude Sonnet 4 | 200K tokens | $3/1M | $15/1M |
| Claude Haiku | 200K tokens | $0.25/1M | $1.25/1M |

**Recommendation for AFTRS:** Use Claude Haiku for cost efficiency on a 2-user server. Claude Sonnet for complex queries only.

### Response Length Handling

Discord message limit is **2000 characters**. Handle long responses:

```python
def format_discord_response(response: str, max_length: int = 2000) -> str:
    """Truncate response for Discord with indicator."""
    if len(response) <= max_length:
        return response
    return response[:max_length - 20] + '\n\n*[truncated]*'
```

---

## Rate Limiting

### Discord Rate Limits

| Limit Type | Value |
|------------|-------|
| Global | 50 requests/second |
| Interaction Response | 1 response per interaction |
| Followup Messages | 5 per interaction token |
| Webhook | 30 requests/second per webhook |

### Claude API Rate Limits

Anthropic uses tiered rate limits based on usage:

| Tier | Requests/min | Tokens/min |
|------|--------------|------------|
| Tier 1 (new) | 50 | 40,000 |
| Tier 2 | 1,000 | 80,000 |
| Tier 3 | 2,000 | 160,000 |

### Implementing User Rate Limits

```python
import time
from collections import defaultdict

# In-memory for single Lambda (use DynamoDB for distributed)
user_requests = defaultdict(list)
RATE_LIMIT = 10  # requests per minute

def check_rate_limit(user_id: str) -> bool:
    """Check if user has exceeded rate limit."""
    now = time.time()
    minute_ago = now - 60

    # Clean old requests
    user_requests[user_id] = [t for t in user_requests[user_id] if t > minute_ago]

    if len(user_requests[user_id]) >= RATE_LIMIT:
        return False

    user_requests[user_id].append(now)
    return True
```

---

## Cold Start Mitigation

### Cold Start Impact

| Scenario | Cold Start Time |
|----------|-----------------|
| Python + minimal deps | 100-300ms |
| Python + PyNaCl | 500-800ms |
| Python + boto3 + anthropic | 800-1500ms |
| Python + VPC | +1000ms |

**Problem:** Discord requires response within **3 seconds**. Cold starts can cause timeouts.

### Mitigation Strategies

#### 1. Use Lightweight Proxy Lambda

```
Proxy Lambda (Python, minimal deps) --> Worker Lambda (full deps)
        |
        v
    Returns type:5 quickly
```

The proxy only verifies signature and returns deferred response. Heavy work happens async.

#### 2. Scheduled Warming (Cost-Effective)

```python
# CloudWatch Event: rate(5 minutes)
def warmer_handler(event, context):
    if event.get('source') == 'aws.events':
        return {'statusCode': 200, 'body': 'Warmed'}
    # Normal handling...
```

**Cost:** ~$0.02/month for keeping Lambda warm during business hours.

#### 3. Lambda SnapStart (Python 3.12+)

SnapStart creates a snapshot of initialized Lambda environment:

```yaml
# SAM template
ProxyFunction:
  Type: AWS::Serverless::Function
  Properties:
    Runtime: python3.12
    SnapStart:
      ApplyOn: PublishedVersions
```

**Benefit:** 90% reduction in cold start time
**Limitation:** Not compatible with >512MB ephemeral storage or provisioned concurrency

#### 4. Provisioned Concurrency (Premium)

```yaml
ProvisionedConcurrencyConfig:
  ProvisionedConcurrentExecutions: 1
```

**Cost:** ~$15-20/month for 1 instance
**Use case:** If you absolutely cannot tolerate cold starts

### Recommended Strategy for AFTRS

1. **Primary:** SnapStart on Python 3.12 for proxy Lambda
2. **Secondary:** CloudWatch scheduled warming every 5 minutes during active hours
3. **Architecture:** Dual-Lambda pattern to guarantee 3-second compliance

---

## Security Best Practices

### Token Storage

**DO NOT** use plain environment variables for sensitive tokens.

#### Option 1: AWS Secrets Manager (Recommended)

```python
import boto3
from botocore.exceptions import ClientError

def get_secret(secret_name: str) -> dict:
    """Retrieve secret from AWS Secrets Manager."""
    client = boto3.client('secretsmanager')
    response = client.get_secret_value(SecretId=secret_name)
    return json.loads(response['SecretString'])

# Usage (cache at init for performance)
secrets = get_secret('aftrs/discord-bot')
DISCORD_TOKEN = secrets['discord_token']
ANTHROPIC_KEY = secrets['anthropic_key']
```

**Cost:** $0.40/secret/month + $0.05/10,000 API calls

#### Option 2: AWS Parameter Store with Encryption

```python
import boto3

ssm = boto3.client('ssm')

def get_parameter(name: str) -> str:
    """Get encrypted parameter from SSM."""
    response = ssm.get_parameter(Name=name, WithDecryption=True)
    return response['Parameter']['Value']
```

**Cost:** Free for standard parameters, $0.05/10,000 calls for advanced

#### Option 3: Lambda Environment Variables with KMS

```python
# Set via AWS Console or CLI with KMS encryption
import os
DISCORD_TOKEN = os.environ['DISCORD_TOKEN']  # Encrypted at rest
```

**Note:** Still visible to anyone with `lambda:GetFunctionConfiguration` permission.

### Request Validation Checklist

- [ ] Verify Ed25519 signature on EVERY request
- [ ] Validate interaction type before processing
- [ ] Sanitize user input before passing to Claude
- [ ] Rate limit per user
- [ ] Log suspicious activity (invalid signatures, rate limit hits)

### IAM Least Privilege

```yaml
LambdaExecutionRole:
  Type: AWS::IAM::Role
  Properties:
    Policies:
      - PolicyName: MinimalAccess
        PolicyDocument:
          Statement:
            - Effect: Allow
              Action:
                - dynamodb:Query
                - dynamodb:PutItem
              Resource: !GetAtt SessionsTable.Arn
            - Effect: Allow
              Action:
                - secretsmanager:GetSecretValue
              Resource: !Sub 'arn:aws:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:aftrs/*'
            - Effect: Allow
              Action:
                - logs:CreateLogGroup
                - logs:CreateLogStream
                - logs:PutLogEvents
              Resource: '*'
```

---

## Cost Analysis

### 2-User Server Estimate

Assumptions:
- 50 interactions/day average
- 30 days/month
- 1500 interactions/month total
- Average Lambda execution: 2 seconds
- 512MB memory allocation

#### AWS Costs (Monthly)

| Service | Usage | Cost |
|---------|-------|------|
| **API Gateway (HTTP)** | 1,500 requests | **Free** (1M free/month) |
| **Lambda (Proxy)** | 1,500 x 100ms x 128MB | **Free** (400K GB-s free) |
| **Lambda (Worker)** | 1,500 x 2s x 512MB | **Free** (400K GB-s free) |
| **DynamoDB** | ~1,500 writes, ~3,000 reads | **Free** (25 WCU/RCU free) |
| **Secrets Manager** | 2 secrets | **$0.80** |
| **CloudWatch Logs** | ~50MB | **Free** (5GB free) |
| **Data Transfer** | ~100MB | **Free** (100GB free) |

**Total AWS Cost: ~$0.80/month** (Secrets Manager only)

#### Claude API Costs (Monthly)

Using Claude Haiku:
- 1,500 interactions
- ~500 input tokens/request (with history)
- ~300 output tokens/request

| Item | Calculation | Cost |
|------|-------------|------|
| Input tokens | 1,500 x 500 = 750K | $0.19 |
| Output tokens | 1,500 x 300 = 450K | $0.56 |

**Total Claude Cost: ~$0.75/month**

### Total Monthly Cost

| Component | Cost |
|-----------|------|
| AWS Infrastructure | $0.80 |
| Claude API | $0.75 |
| **Total** | **$1.55/month** |

### Cost Optimization Tips

1. **Use HTTP API** instead of REST API (70% cheaper)
2. **Use Claude Haiku** for simple queries (5x cheaper than Sonnet)
3. **Limit conversation history** to 5-10 turns
4. **Enable DynamoDB TTL** to auto-delete old data
5. **Use scheduled warming** instead of provisioned concurrency

---

## Step-by-Step Implementation Guide

### Phase 1: Discord Application Setup (15 minutes)

1. **Create Discord Application**
   ```
   1. Go to https://discord.com/developers/applications
   2. Click "New Application"
   3. Name: "AFTRS Studio Bot"
   4. Copy APPLICATION ID (needed for API calls)
   5. Copy PUBLIC KEY (needed for signature verification)
   ```

2. **Create Bot User**
   ```
   1. Go to "Bot" section
   2. Click "Add Bot"
   3. Copy TOKEN (store securely!)
   4. Disable "Public Bot" (private server only)
   ```

3. **Enable Interactions**
   ```
   1. Go to "General Information"
   2. Leave INTERACTIONS ENDPOINT URL blank (add later)
   ```

4. **Generate Invite URL**
   ```
   1. Go to OAuth2 > URL Generator
   2. Scopes: bot, applications.commands
   3. Permissions: Send Messages, Use Slash Commands
   4. Copy URL and invite bot to your server
   ```

### Phase 2: AWS Infrastructure (30 minutes)

1. **Create Secrets in Secrets Manager**
   ```bash
   aws secretsmanager create-secret --name aftrs/discord-bot --secret-string '{"discord_token":"YOUR_TOKEN","discord_public_key":"YOUR_PUBLIC_KEY","anthropic_key":"YOUR_ANTHROPIC_KEY"}'
   ```

2. **Create DynamoDB Table**
   ```bash
   aws dynamodb create-table --table-name aftrs-discord-sessions --attribute-definitions AttributeName=user_id,AttributeType=S AttributeName=timestamp,AttributeType=N --key-schema AttributeName=user_id,KeyType=HASH AttributeName=timestamp,KeyType=RANGE --billing-mode PAY_PER_REQUEST
   ```

3. **Create Lambda Layer for Dependencies**
   ```bash
   mkdir -p layer/python && pip install pynacl anthropic httpx -t layer/python/ && cd layer && zip -r ../lambda-layer.zip . && cd .. && aws lambda publish-layer-version --layer-name aftrs-discord-deps --zip-file fileb://lambda-layer.zip --compatible-runtimes python3.12
   ```

4. **Create Proxy Lambda**
   ```bash
   # Create proxy_lambda.py with verification code
   zip proxy.zip proxy_lambda.py
   aws lambda create-function --function-name aftrs-discord-proxy --runtime python3.12 --handler proxy_lambda.lambda_handler --zip-file fileb://proxy.zip --role YOUR_ROLE_ARN --layers YOUR_LAYER_ARN --environment Variables={DISCORD_PUBLIC_KEY=YOUR_KEY,WORKER_LAMBDA_ARN=arn:...}
   ```

5. **Create Worker Lambda**
   ```bash
   # Create worker_lambda.py with Claude integration
   zip worker.zip worker_lambda.py
   aws lambda create-function --function-name aftrs-discord-worker --runtime python3.12 --handler worker_lambda.lambda_handler --zip-file fileb://worker.zip --role YOUR_ROLE_ARN --layers YOUR_LAYER_ARN --timeout 30 --environment Variables={SESSIONS_TABLE=aftrs-discord-sessions}
   ```

6. **Create HTTP API Gateway**
   ```bash
   aws apigatewayv2 create-api --name aftrs-discord-api --protocol-type HTTP --target arn:aws:lambda:REGION:ACCOUNT:function:aftrs-discord-proxy
   ```

### Phase 3: Connect Discord to AWS (10 minutes)

1. **Get API Gateway URL**
   ```bash
   aws apigatewayv2 get-apis --query 'Items[?Name==`aftrs-discord-api`].ApiEndpoint'
   ```

2. **Set Interactions Endpoint URL**
   ```
   1. Go to Discord Developer Portal
   2. Your Application > General Information
   3. INTERACTIONS ENDPOINT URL: https://YOUR_API_ID.execute-api.REGION.amazonaws.com/
   4. Click "Save Changes"
   5. Discord will verify your endpoint (must handle PING)
   ```

### Phase 4: Register Slash Commands (10 minutes)

```python
# register_commands.py
import httpx
import os

APPLICATION_ID = os.environ['DISCORD_APPLICATION_ID']
BOT_TOKEN = os.environ['DISCORD_BOT_TOKEN']
GUILD_ID = os.environ['DISCORD_GUILD_ID']  # Optional: for guild-specific commands

commands = [
    {
        'name': 'ask',
        'description': 'Ask the AI assistant a question',
        'options': [{
            'name': 'question',
            'description': 'Your question',
            'type': 3,  # STRING
            'required': True
        }]
    },
    {
        'name': 'studio-status',
        'description': 'Get current studio system status'
    },
    {
        'name': 'help',
        'description': 'Show available commands'
    }
]

# Register guild commands (instant) or global commands (up to 1 hour)
url = f'https://discord.com/api/v10/applications/{APPLICATION_ID}/guilds/{GUILD_ID}/commands'
headers = {'Authorization': f'Bot {BOT_TOKEN}'}

for cmd in commands:
    response = httpx.post(url, headers=headers, json=cmd)
    print(f"Registered {cmd['name']}: {response.status_code}")
```

Run: `python register_commands.py`

### Phase 5: Testing and Verification (15 minutes)

1. **Test PING Response**
   ```bash
   curl -X POST https://YOUR_API.execute-api.REGION.amazonaws.com/ -H "Content-Type: application/json" -d '{"type": 1}'
   # Should return {"type": 1}
   ```

2. **Test in Discord**
   ```
   1. Go to your Discord server
   2. Type /ask and select the command
   3. Enter a question
   4. Verify "thinking..." appears
   5. Verify response is returned
   ```

3. **Check CloudWatch Logs**
   ```bash
   aws logs tail /aws/lambda/aftrs-discord-proxy --follow
   aws logs tail /aws/lambda/aftrs-discord-worker --follow
   ```

---

## Example Implementations

### Reference Repositories

1. **pixegami/discord-bot-lambda**
   - https://github.com/pixegami/discord-bot-lambda
   - Python + Flask + Docker + CDK
   - Good starting template

2. **ker0olos/aws-lambda-discord-bot**
   - https://github.com/ker0olos/aws-lambda-discord-bot
   - Python-focused, simpler setup
   - Good signature verification example

3. **ytausch/serverless-discord-bot**
   - https://github.com/ytausch/serverless-discord-bot
   - TypeScript + slash-create library
   - AWS SAM deployment

4. **aws-samples/anthropic-on-aws**
   - https://github.com/aws-samples/anthropic-on-aws
   - Official AWS + Claude integration patterns
   - Tools/function calling examples

### Full Working Example

See the `examples/discord-lambda/` directory in this repository for a complete working implementation.

---

## References

### Official Documentation

- [Discord Interactions Overview](https://discord.com/developers/docs/interactions/overview)
- [Discord Receiving and Responding to Interactions](https://discord.com/developers/docs/interactions/receiving-and-responding)
- [Discord Rate Limits](https://discord.com/developers/docs/topics/rate-limits)
- [AWS Lambda Documentation](https://docs.aws.amazon.com/lambda/latest/dg/)
- [AWS API Gateway HTTP APIs](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api.html)
- [AWS DynamoDB Best Practices](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)
- [Claude API Documentation](https://docs.anthropic.com/en/api)

### Tutorials and Guides

- [Serverless Discord Bot on AWS in 5 Steps](https://medium.com/better-programming/serverless-discord-bot-on-aws-in-5-steps-956dca04d899) - jakjus
- [Building Discord Bot with AWS Serverless - Part 1](https://www.guptaakashdeep.com/building-discord-bot-with-aws-serverless/) - Akash Gupta
- [Building Discord Bot with AWS Serverless - Part 2](https://www.guptaakashdeep.com/building-discord-bot-with-aws-serverless-part-2/) - Akash Gupta
- [Crafting a Serverless Discord Bot with AWS Lambda](https://anisimow.com/crafting-a-serverless-discord-bot-with-aws-lambda/) - Anisimow
- [AWS Lambda Cold Start Optimization in 2025](https://zircon.tech/blog/aws-lambda-cold-start-optimization-in-2025-what-actually-works/) - Zircon Tech

### Cost Calculators

- [AWS Pricing Calculator](https://calculator.aws/)
- [Bref Serverless Costs Calculator](https://cost-calculator.bref.sh/)
- [Claude API Pricing](https://www.anthropic.com/pricing)

### Security Resources

- [AWS Secrets Manager Best Practices](https://docs.aws.amazon.com/secretsmanager/latest/userguide/best-practices.html)
- [Ultimate Guide to Secrets in Lambda](https://aaronstuyvenberg.com/posts/ultimate-lambda-secrets-guide) - AJ Stuyvenberg
- [Securing Lambda Environment Variables](https://docs.aws.amazon.com/lambda/latest/dg/configuration-envvars-encryption.html)

---

## Appendix: SAM Template

Complete AWS SAM template for deploying the solution:

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31
Description: AFTRS Discord Bot - Serverless

Globals:
  Function:
    Runtime: python3.12
    Timeout: 30
    MemorySize: 512

Parameters:
  DiscordPublicKey:
    Type: String
    NoEcho: true

Resources:
  # DynamoDB Table
  SessionsTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: aftrs-discord-sessions
      BillingMode: PAY_PER_REQUEST
      AttributeDefinitions:
        - AttributeName: user_id
          AttributeType: S
        - AttributeName: timestamp
          AttributeType: N
      KeySchema:
        - AttributeName: user_id
          KeyType: HASH
        - AttributeName: timestamp
          KeyType: RANGE
      TimeToLiveSpecification:
        AttributeName: ttl
        Enabled: true

  # Lambda Layer
  DependenciesLayer:
    Type: AWS::Serverless::LayerVersion
    Properties:
      LayerName: aftrs-discord-deps
      ContentUri: layer/
      CompatibleRuntimes:
        - python3.12

  # Proxy Lambda
  ProxyFunction:
    Type: AWS::Serverless::Function
    Properties:
      FunctionName: aftrs-discord-proxy
      Handler: proxy.lambda_handler
      CodeUri: src/proxy/
      MemorySize: 128
      Timeout: 3
      Layers:
        - !Ref DependenciesLayer
      Environment:
        Variables:
          DISCORD_PUBLIC_KEY: !Ref DiscordPublicKey
          WORKER_LAMBDA_ARN: !GetAtt WorkerFunction.Arn
      Policies:
        - LambdaInvokePolicy:
            FunctionName: !Ref WorkerFunction
      Events:
        Api:
          Type: HttpApi
          Properties:
            Path: /
            Method: POST

  # Worker Lambda
  WorkerFunction:
    Type: AWS::Serverless::Function
    Properties:
      FunctionName: aftrs-discord-worker
      Handler: worker.lambda_handler
      CodeUri: src/worker/
      Timeout: 30
      Layers:
        - !Ref DependenciesLayer
      Environment:
        Variables:
          SESSIONS_TABLE: !Ref SessionsTable
          SECRETS_NAME: aftrs/discord-bot
      Policies:
        - DynamoDBCrudPolicy:
            TableName: !Ref SessionsTable
        - Version: '2012-10-17'
          Statement:
            - Effect: Allow
              Action: secretsmanager:GetSecretValue
              Resource: !Sub 'arn:aws:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:aftrs/*'

Outputs:
  ApiUrl:
    Description: API Gateway endpoint URL
    Value: !Sub 'https://${ServerlessHttpApi}.execute-api.${AWS::Region}.amazonaws.com/'
```

Deploy with: `sam build && sam deploy --guided`
