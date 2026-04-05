# AWS S3 + DynamoDB Data Patterns Research

> Research document for cr8-cli Supabase to AWS migration
> Last updated: December 2025

## Table of Contents

1. [DynamoDB Table Design Patterns](#1-dynamodb-table-design-patterns)
2. [Global Secondary Index (GSI) Best Practices](#2-global-secondary-index-gsi-best-practices)
3. [S3 Storage Classes and Lifecycle Policies](#3-s3-storage-classes-and-lifecycle-policies)
4. [Data Migration Strategies](#4-data-migration-strategies)
5. [Cost Optimization for Music/Audio Storage](#5-cost-optimization-for-musicaudio-storage)
6. [DynamoDB Streams for Event Processing](#6-dynamodb-streams-for-event-processing)
7. [Point-in-Time Recovery and Backup Strategies](#7-point-in-time-recovery-and-backup-strategies)
8. [Query Patterns and Access Optimization](#8-query-patterns-and-access-optimization)
9. [S3 Transfer Acceleration](#9-s3-transfer-acceleration)
10. [Multipart Upload Patterns](#10-multipart-upload-patterns)

---

## 1. DynamoDB Table Design Patterns

### Single-Table vs Multi-Table Design

#### Single-Table Design (STD)

Store multiple entity types (Users, Tracks, Playlists, etc.) in a single table using generic attribute names (PK, SK).

**Benefits:**
- Fetch related entities in one query (no joins in DynamoDB)
- Reduced operational overhead
- Pre-joined data via item collections
- Efficient data locality

**Drawbacks:**
- Increased cognitive overhead
- Requires upfront access pattern design
- Adding new access patterns may require migrations

**When to Use STD:**
- Applications frequently querying multiple entity types together
- Need to maintain relationships between data types
- Benefit from data locality

#### Multi-Table Design

**When to Use:**
- Access patterns don't require cross-entity queries
- Teams more familiar with relational patterns
- Simpler data models

### Schema Design Example for cr8-cli

```python
# Single-Table Design for Music Library
# PK (Partition Key) and SK (Sort Key) patterns

SCHEMA_PATTERNS = {
    # User entity
    "user": {
        "PK": "USER#<user_id>",
        "SK": "PROFILE#<user_id>",
    },

    # Track entity
    "track": {
        "PK": "USER#<user_id>",
        "SK": "TRACK#<track_id>",
    },

    # Playlist entity
    "playlist": {
        "PK": "USER#<user_id>",
        "SK": "PLAYLIST#<playlist_id>",
    },

    # Playlist tracks (for ordering)
    "playlist_track": {
        "PK": "PLAYLIST#<playlist_id>",
        "SK": "TRACK#<position>#<track_id>",
    },

    # Track metadata for search
    "track_by_genre": {
        "PK": "GENRE#<genre>",
        "SK": "TRACK#<bpm>#<track_id>",
    },

    # Session/crate
    "session": {
        "PK": "USER#<user_id>",
        "SK": "SESSION#<timestamp>#<session_id>",
    },
}
```

### DynamoDB Table Definition (CDK/CloudFormation)

```python
import aws_cdk as cdk
from aws_cdk import aws_dynamodb as dynamodb

class Cr8Table(cdk.Stack):
    def __init__(self, scope, construct_id, **kwargs):
        super().__init__(scope, construct_id, **kwargs)

        # Main table with single-table design
        self.table = dynamodb.Table(
            self, "Cr8MusicLibrary",
            table_name="cr8-music-library",
            partition_key=dynamodb.Attribute(
                name="PK",
                type=dynamodb.AttributeType.STRING
            ),
            sort_key=dynamodb.Attribute(
                name="SK",
                type=dynamodb.AttributeType.STRING
            ),
            billing_mode=dynamodb.BillingMode.PAY_PER_REQUEST,
            removal_policy=cdk.RemovalPolicy.RETAIN,
            point_in_time_recovery=True,
            stream=dynamodb.StreamViewType.NEW_AND_OLD_IMAGES,
        )
```

### 2025 Update: Multi-Attribute Composite Keys

As of November 2025, DynamoDB supports up to 4 attributes each for partition and sort keys in GSIs. This eliminates the need for synthetic concatenated keys.

**References:**
- [AWS Blog: Single-Table Design](https://aws.amazon.com/blogs/compute/creating-a-single-table-design-with-amazon-dynamodb/)
- [DynamoDB Data Modeling Foundations](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/data-modeling-foundations.html)
- [Alex DeBrie: Single-Table Design](https://www.alexdebrie.com/posts/dynamodb-single-table/)

---

## 2. Global Secondary Index (GSI) Best Practices

### GSI Overview

GSIs allow querying on attributes other than the primary key, with their own partition and sort keys.

### Key Design Patterns

#### Pattern 1: GSI Key Overloading

```python
# GSI with overloaded keys for multiple access patterns
GSI_OVERLOADING = {
    "GSI1": {
        "partition_key": "GSI1PK",  # Overloaded
        "sort_key": "GSI1SK",       # Overloaded
    }
}

# Example usage patterns
GSI_PATTERNS = {
    # Find all tracks by BPM range
    "tracks_by_bpm": {
        "GSI1PK": "BPM#<bpm_bucket>",  # e.g., BPM#120-130
        "GSI1SK": "TRACK#<track_id>",
    },

    # Find all tracks by key (musical key)
    "tracks_by_key": {
        "GSI1PK": "KEY#<musical_key>",  # e.g., KEY#Am
        "GSI1SK": "ENERGY#<energy>#<track_id>",
    },

    # Find user's recent activity
    "user_activity": {
        "GSI1PK": "USER#<user_id>",
        "GSI1SK": "ACTIVITY#<timestamp>",
    },
}
```

#### Pattern 2: Sparse Index

```python
# Only items with 'featured' attribute appear in index
SPARSE_INDEX = {
    "GSI_Featured": {
        "partition_key": "featured_category",
        "sort_key": "featured_rank",
        # Only tracks marked as featured are indexed
    }
}
```

### CDK GSI Definition

```python
# Add GSIs to table
self.table.add_global_secondary_index(
    index_name="GSI1",
    partition_key=dynamodb.Attribute(
        name="GSI1PK",
        type=dynamodb.AttributeType.STRING
    ),
    sort_key=dynamodb.Attribute(
        name="GSI1SK",
        type=dynamodb.AttributeType.STRING
    ),
    projection_type=dynamodb.ProjectionType.ALL,
)

# GSI for genre/BPM queries
self.table.add_global_secondary_index(
    index_name="GenreBpmIndex",
    partition_key=dynamodb.Attribute(
        name="genre",
        type=dynamodb.AttributeType.STRING
    ),
    sort_key=dynamodb.Attribute(
        name="bpm",
        type=dynamodb.AttributeType.NUMBER
    ),
    projection_type=dynamodb.ProjectionType.INCLUDE,
    non_key_attributes=["title", "artist", "duration", "s3_key"],
)
```

### Best Practices

1. **Design for query patterns first** - GSIs should match your access patterns
2. **Keep indexes minimal** - Max 20 GSIs per table; unused indexes waste resources
3. **Project only required attributes** - Reduces storage and improves performance
4. **Avoid hot partitions** - Distribute data evenly across partitions
5. **Consider write amplification** - Each GSI doubles write operations

### Limits

- Maximum 20 GSIs per table
- Cannot use GetItem/BatchGetItem on GSIs
- GSI keys don't need to be unique

**References:**
- [AWS: Using Global Secondary Indexes](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GSI.html)
- [Dynobase: GSI Guide](https://dynobase.dev/dynamodb-gsi/)
- [AWS: Multi-Attribute Keys Pattern](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GSI.DesignPattern.MultiAttributeKeys.html)

---

## 3. S3 Storage Classes and Lifecycle Policies

### Storage Classes Overview

| Class | Use Case | Price (US East) | Retrieval |
|-------|----------|-----------------|-----------|
| **S3 Standard** | Frequently accessed | $0.023/GB | Immediate |
| **S3 Intelligent-Tiering** | Unknown/variable access | $0.023/GB + monitoring | Automatic |
| **S3 Standard-IA** | Infrequent access (30+ days) | $0.0125/GB | Immediate |
| **S3 One Zone-IA** | Non-critical infrequent | $0.01/GB | Immediate |
| **S3 Glacier Instant** | Archive with instant access | $0.004/GB | Milliseconds |
| **S3 Glacier Flexible** | Archive | $0.0036/GB | Minutes-hours |
| **S3 Glacier Deep Archive** | Long-term archive | $0.00099/GB | 12 hours |
| **S3 Express One Zone** | Ultra-low latency | Higher | <10ms |

### Recommended Lifecycle Policy for Audio Files

```json
{
  "Rules": [
    {
      "ID": "AudioLifecycleRule",
      "Status": "Enabled",
      "Filter": {
        "Prefix": "audio/"
      },
      "Transitions": [
        {
          "Days": 30,
          "StorageClass": "STANDARD_IA"
        },
        {
          "Days": 90,
          "StorageClass": "GLACIER_IR"
        },
        {
          "Days": 365,
          "StorageClass": "DEEP_ARCHIVE"
        }
      ]
    },
    {
      "ID": "WaveformCacheRule",
      "Status": "Enabled",
      "Filter": {
        "Prefix": "waveforms/"
      },
      "Transitions": [
        {
          "Days": 7,
          "StorageClass": "STANDARD_IA"
        }
      ],
      "Expiration": {
        "Days": 90
      }
    },
    {
      "ID": "TempUploadsCleanup",
      "Status": "Enabled",
      "Filter": {
        "Prefix": "temp/"
      },
      "Expiration": {
        "Days": 1
      },
      "AbortIncompleteMultipartUpload": {
        "DaysAfterInitiation": 1
      }
    }
  ]
}
```

### CDK S3 Bucket Configuration

```python
from aws_cdk import aws_s3 as s3, Duration

class Cr8StorageBucket(cdk.Stack):
    def __init__(self, scope, construct_id, **kwargs):
        super().__init__(scope, construct_id, **kwargs)

        self.bucket = s3.Bucket(
            self, "Cr8AudioBucket",
            bucket_name="cr8-audio-library",
            versioned=True,
            encryption=s3.BucketEncryption.S3_MANAGED,
            block_public_access=s3.BlockPublicAccess.BLOCK_ALL,
            transfer_acceleration=True,
            intelligent_tiering_configurations=[
                s3.IntelligentTieringConfiguration(
                    name="AudioTiering",
                    prefix="audio/",
                    archive_access_tier_time=Duration.days(90),
                    deep_archive_access_tier_time=Duration.days(180),
                )
            ],
            lifecycle_rules=[
                s3.LifecycleRule(
                    id="TransitionToIA",
                    prefix="audio/",
                    transitions=[
                        s3.Transition(
                            storage_class=s3.StorageClass.INFREQUENT_ACCESS,
                            transition_after=Duration.days(30)
                        ),
                    ]
                ),
                s3.LifecycleRule(
                    id="CleanupIncompleteUploads",
                    abort_incomplete_multipart_upload_after=Duration.days(1)
                ),
            ],
        )
```

**References:**
- [AWS: S3 Storage Classes](https://aws.amazon.com/s3/storage-classes/)
- [AWS: Lifecycle Configuration Examples](https://docs.aws.amazon.com/AmazonS3/latest/userguide/lifecycle-configuration-examples.html)
- [AWS: Managing Object Lifecycle](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html)

---

## 4. Data Migration Strategies

### Migration Approaches

#### Option 1: AWS Database Migration Service (DMS)

Best for: Large datasets, continuous replication during cutover

```bash
# DMS supports PostgreSQL (Supabase) to DynamoDB
# Requires object mapping rules for data transformation
```

#### Option 2: Custom Migration Script (Recommended for cr8-cli)

```python
import asyncio
import boto3
from typing import AsyncGenerator
from supabase import create_client
from boto3.dynamodb.conditions import Key

class Cr8Migration:
    """Migrate cr8-cli data from Supabase to DynamoDB."""

    def __init__(self):
        self.supabase = create_client(SUPABASE_URL, SUPABASE_KEY)
        self.dynamodb = boto3.resource('dynamodb')
        self.table = self.dynamodb.Table('cr8-music-library')
        self.s3 = boto3.client('s3')

    async def migrate_tracks(self, batch_size: int = 25):
        """Migrate tracks with parallel batch writes."""
        offset = 0

        while True:
            # Fetch batch from Supabase
            response = self.supabase.table('tracks').select('*').range(offset, offset + batch_size - 1).execute()

            if not response.data:
                break

            # Transform to DynamoDB format
            items = [self._transform_track(track) for track in response.data]

            # Batch write with exponential backoff
            await self._batch_write_with_retry(items)

            offset += batch_size
            print(f"Migrated {offset} tracks...")

    def _transform_track(self, supabase_track: dict) -> dict:
        """Transform Supabase track to DynamoDB single-table format."""
        user_id = supabase_track['user_id']
        track_id = supabase_track['id']

        return {
            'PK': f"USER#{user_id}",
            'SK': f"TRACK#{track_id}",
            'GSI1PK': f"GENRE#{supabase_track.get('genre', 'unknown')}",
            'GSI1SK': f"BPM#{supabase_track.get('bpm', 0):03d}#{track_id}",
            'entity_type': 'track',
            'track_id': track_id,
            'title': supabase_track['title'],
            'artist': supabase_track.get('artist'),
            'bpm': supabase_track.get('bpm'),
            'key': supabase_track.get('musical_key'),
            'duration': supabase_track.get('duration'),
            's3_key': f"audio/{user_id}/{track_id}.mp3",
            'created_at': supabase_track['created_at'],
            'updated_at': supabase_track.get('updated_at'),
        }

    async def _batch_write_with_retry(self, items: list, max_retries: int = 5):
        """Batch write with exponential backoff for unprocessed items."""
        with self.table.batch_writer() as batch:
            for item in items:
                batch.put_item(Item=item)
```

#### Option 3: Step Functions Distributed Map

For massive parallel processing of large datasets:

```python
# AWS CDK Step Functions for parallel migration
from aws_cdk import aws_stepfunctions as sfn, aws_stepfunctions_tasks as tasks

class MigrationStateMachine(cdk.Stack):
    def __init__(self, scope, construct_id, **kwargs):
        super().__init__(scope, construct_id, **kwargs)

        # Distributed Map for parallel processing
        distributed_map = sfn.DistributedMap(
            self, "ParallelMigration",
            items_path="$.batches",
            max_concurrency=100,
            item_reader=sfn.S3JsonItemReader(
                bucket=migration_bucket,
                key="migration-batches/"
            ),
        )
```

### Parallel Processing Pattern

```python
import asyncio
from concurrent.futures import ThreadPoolExecutor

async def parallel_migrate(source_data: list, workers: int = 10):
    """Process migration in parallel batches."""

    semaphore = asyncio.Semaphore(workers)

    async def process_batch(batch):
        async with semaphore:
            # Randomize partition keys to avoid hot partitions
            random.shuffle(batch)
            await batch_write(batch)

    batches = [source_data[i:i+25] for i in range(0, len(source_data), 25)]
    await asyncio.gather(*[process_batch(b) for b in batches])
```

### Pre-Migration Checklist

1. **Enable On-Demand capacity** - Handles burst traffic during migration
2. **Pre-warm tables** - If using provisioned capacity, scale up before migration
3. **Disable unnecessary indexes** - Add GSIs after base data migration
4. **Plan for rollback** - Keep Supabase running until cutover verified

**References:**
- [AWS: Migrating to DynamoDB](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/migration-guide.html)
- [AWS DMS with DynamoDB Target](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.DynamoDB.html)
- [Dynobase: Mass Migrations](https://github.com/dynobase/dynamodb-mass-migrations)

---

## 5. Cost Optimization for Music/Audio Storage

### S3 Storage Cost Analysis

#### Scenario: 100,000 audio files (avg 10MB each = 1TB total)

| Storage Class | Monthly Cost | Annual Cost |
|---------------|--------------|-------------|
| S3 Standard | $23.00 | $276.00 |
| S3 Intelligent-Tiering | $23.00 + $2.50 monitoring | $306.00 |
| S3 Standard-IA | $12.50 | $150.00 |
| S3 Glacier Instant | $4.00 | $48.00 |
| S3 Glacier Flexible | $3.60 | $43.20 |
| S3 Glacier Deep Archive | $0.99 | $11.88 |

### Recommended Strategy for Music Library

```python
STORAGE_STRATEGY = {
    # Hot tier: Recently uploaded/played tracks
    "hot": {
        "storage_class": "STANDARD",
        "criteria": "last_accessed < 30 days",
        "estimated_percentage": "20%",
    },

    # Warm tier: Occasional access
    "warm": {
        "storage_class": "INTELLIGENT_TIERING",
        "criteria": "30 days < last_accessed < 90 days",
        "estimated_percentage": "30%",
    },

    # Cold tier: Rarely accessed catalog
    "cold": {
        "storage_class": "GLACIER_INSTANT_RETRIEVAL",
        "criteria": "last_accessed > 90 days",
        "estimated_percentage": "50%",
    },
}

# Estimated monthly cost for 1TB library
# Hot (200GB): $4.60
# Warm (300GB): $6.90 + monitoring
# Cold (500GB): $2.00
# Total: ~$14/month vs $23 for all-Standard
# Savings: ~40%
```

### DynamoDB Cost Optimization

```python
DYNAMODB_COST_TIPS = {
    "use_on_demand": "Pay per request for unpredictable workloads",
    "optimize_item_size": "Keep items under 1KB when possible",
    "project_less_in_gsis": "Only project needed attributes",
    "batch_operations": "Use BatchGetItem/BatchWriteItem to reduce requests",
    "use_sparse_indexes": "Only index items that need querying",
    "compress_large_attributes": "Compress JSON/text before storing",
}

# Example: Compress track metadata
import gzip
import json

def compress_metadata(metadata: dict) -> bytes:
    """Compress large metadata to reduce storage costs."""
    return gzip.compress(json.dumps(metadata).encode())

def decompress_metadata(compressed: bytes) -> dict:
    """Decompress metadata on read."""
    return json.loads(gzip.decompress(compressed))
```

### Request Cost Optimization

```python
# Batch reads to reduce request costs
def get_tracks_batch(track_ids: list) -> list:
    """Batch get up to 100 tracks in single request."""
    response = dynamodb.batch_get_item(
        RequestItems={
            'cr8-music-library': {
                'Keys': [
                    {'PK': f'USER#{user_id}', 'SK': f'TRACK#{tid}'}
                    for tid in track_ids
                ],
                'ProjectionExpression': 'title, artist, bpm, s3_key',
            }
        }
    )
    return response['Responses']['cr8-music-library']
```

**References:**
- [AWS S3 Pricing](https://aws.amazon.com/s3/pricing/)
- [AWS S3 Cost Optimization](https://aws.amazon.com/s3/cost-optimization/)
- [nOps: S3 Storage Costs Guide](https://www.nops.io/blog/how-much-do-aws-s3-storage-classes-cost/)

---

## 6. DynamoDB Streams for Event Processing

### Overview

DynamoDB Streams captures every change (INSERT, MODIFY, DELETE) and enables event-driven architectures.

### Use Cases for cr8-cli

```python
STREAM_USE_CASES = {
    "waveform_generation": "Generate waveform when new track uploaded",
    "search_indexing": "Update OpenSearch when track metadata changes",
    "analytics": "Track usage patterns and popular tracks",
    "notifications": "Notify users of playlist updates",
    "audit_log": "Maintain change history for compliance",
    "cache_invalidation": "Invalidate CDN cache on track update",
}
```

### Lambda Stream Processor

```python
import json
import boto3
from aws_lambda_powertools import Logger

logger = Logger()
s3 = boto3.client('s3')
sqs = boto3.client('sqs')

def handler(event, context):
    """Process DynamoDB Stream events."""

    for record in event['Records']:
        event_name = record['eventName']

        if event_name == 'INSERT':
            handle_new_track(record['dynamodb']['NewImage'])
        elif event_name == 'MODIFY':
            handle_track_update(
                record['dynamodb']['OldImage'],
                record['dynamodb']['NewImage']
            )
        elif event_name == 'REMOVE':
            handle_track_delete(record['dynamodb']['OldImage'])

    return {'statusCode': 200}

def handle_new_track(new_image: dict):
    """Trigger waveform generation for new track."""
    if new_image.get('entity_type', {}).get('S') == 'track':
        s3_key = new_image['s3_key']['S']

        # Queue waveform generation job
        sqs.send_message(
            QueueUrl=WAVEFORM_QUEUE_URL,
            MessageBody=json.dumps({
                'action': 'generate_waveform',
                's3_key': s3_key,
                'track_id': new_image['track_id']['S'],
            })
        )
```

### CDK Stream Configuration

```python
from aws_cdk import aws_lambda as lambda_, aws_lambda_event_sources as event_sources

class StreamProcessor(cdk.Stack):
    def __init__(self, scope, construct_id, table, **kwargs):
        super().__init__(scope, construct_id, **kwargs)

        # Lambda function for stream processing
        processor = lambda_.Function(
            self, "StreamProcessor",
            runtime=lambda_.Runtime.PYTHON_3_11,
            handler="stream_handler.handler",
            code=lambda_.Code.from_asset("lambda/stream"),
            timeout=Duration.seconds(30),
        )

        # Add DynamoDB Stream as event source
        processor.add_event_source(
            event_sources.DynamoEventSource(
                table,
                starting_position=lambda_.StartingPosition.TRIM_HORIZON,
                batch_size=100,
                bisect_batch_on_error=True,
                retry_attempts=3,
                parallelization_factor=10,  # Process shard with 10 concurrent batches
                filters=[
                    # Only process track events
                    lambda_.FilterCriteria.filter({
                        "dynamodb": {
                            "NewImage": {
                                "entity_type": {"S": ["track"]}
                            }
                        }
                    })
                ],
            )
        )
```

### Best Practices

1. **Keep Lambda short-lived** - Avoid complex logic in stream processor
2. **Use event filtering** - Reduce Lambda invocations with filters
3. **Implement DLQ** - Capture failed events for retry
4. **Max 2 Lambda consumers per stream** - Avoid throttling
5. **Use parallelization factor** - Up to 10 concurrent batches per shard

**References:**
- [AWS: DynamoDB Streams and Lambda](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.Lambda.html)
- [AWS: Process DynamoDB Records with Lambda](https://docs.aws.amazon.com/lambda/latest/dg/services-dynamodb-eventsourcemapping.html)
- [Dynobase: DynamoDB Streams Guide](https://dynobase.dev/dynamodb-streams/)

---

## 7. Point-in-Time Recovery and Backup Strategies

### Backup Options

| Method | RPO | Retention | Cost |
|--------|-----|-----------|------|
| PITR | 1 second | 1-35 days | ~20% of table storage |
| On-Demand Backup | Last backup | Unlimited | Storage cost only |
| DynamoDB Streams + S3 | Real-time | Unlimited | Stream + Lambda + S3 |

### Enable PITR (Recommended)

```python
# CDK
table = dynamodb.Table(
    self, "Cr8Table",
    point_in_time_recovery=True,
    # ...
)

# CLI
aws dynamodb update-continuous-backups --table-name cr8-music-library --point-in-time-recovery-specification PointInTimeRecoveryEnabled=true
```

### Restore from PITR

```bash
# Restore to a specific point in time
aws dynamodb restore-table-to-point-in-time \
    --source-table-name cr8-music-library \
    --target-table-name cr8-music-library-restored \
    --restore-date-time "2025-12-15T10:30:00Z"
```

### Backup Strategy for cr8-cli

```python
BACKUP_STRATEGY = {
    "pitr": {
        "enabled": True,
        "recovery_period_days": 35,
        "use_case": "Accidental deletes, data corruption",
    },

    "scheduled_backups": {
        "frequency": "weekly",
        "retention": "90 days",
        "use_case": "Long-term archival, compliance",
    },

    "cross_region_backup": {
        "enabled": True,
        "target_region": "us-west-2",
        "use_case": "Disaster recovery",
    },
}
```

### Automated Backup with EventBridge

```python
from aws_cdk import aws_events as events, aws_events_targets as targets

# Weekly backup rule
backup_rule = events.Rule(
    self, "WeeklyBackup",
    schedule=events.Schedule.cron(
        minute="0",
        hour="3",
        week_day="SUN"
    ),
)

backup_rule.add_target(targets.LambdaFunction(backup_lambda))
```

**References:**
- [AWS: Point-in-Time Recovery](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Point-in-time-recovery.html)
- [AWS: Enable PITR](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/PointInTimeRecovery_Howitworks.html)
- [N2W: DynamoDB Backup Best Practices](https://n2ws.com/blog/aws-cloud/dynamodb-backup)

---

## 8. Query Patterns and Access Optimization

### Partition Key Design

```python
# GOOD: High cardinality, even distribution
GOOD_PARTITION_KEYS = [
    "USER#<user_id>",          # Unique per user
    "SESSION#<session_id>",    # Unique per session
    "TRACK#<track_id>",        # Unique per track
]

# BAD: Low cardinality, hot partitions
BAD_PARTITION_KEYS = [
    "GENRE#electronic",        # Too many tracks in one genre
    "DATE#2025-12-29",         # All daily activity in one partition
    "STATUS#active",           # Most items are active
]
```

### Avoid Hot Partitions

```python
# Partition key sharding for high-traffic keys
def get_sharded_pk(base_key: str, num_shards: int = 10) -> str:
    """Add shard suffix to distribute load."""
    shard = hash(base_key) % num_shards
    return f"{base_key}#SHARD{shard}"

# Example: Popular playlist with many concurrent readers
pk = get_sharded_pk("PLAYLIST#top100", num_shards=10)
# Results in: PLAYLIST#top100#SHARD7
```

### Query Examples for cr8-cli

```python
from boto3.dynamodb.conditions import Key, Attr

# Get all tracks for a user
def get_user_tracks(user_id: str) -> list:
    response = table.query(
        KeyConditionExpression=Key('PK').eq(f'USER#{user_id}') &
                               Key('SK').begins_with('TRACK#')
    )
    return response['Items']

# Get tracks by BPM range using GSI
def get_tracks_by_bpm(genre: str, min_bpm: int, max_bpm: int) -> list:
    response = table.query(
        IndexName='GenreBpmIndex',
        KeyConditionExpression=Key('genre').eq(genre) &
                               Key('bpm').between(min_bpm, max_bpm)
    )
    return response['Items']

# Get user's recent sessions (sorted by timestamp)
def get_recent_sessions(user_id: str, limit: int = 10) -> list:
    response = table.query(
        KeyConditionExpression=Key('PK').eq(f'USER#{user_id}') &
                               Key('SK').begins_with('SESSION#'),
        ScanIndexForward=False,  # Descending order
        Limit=limit
    )
    return response['Items']

# Get playlist tracks in order
def get_playlist_tracks(playlist_id: str) -> list:
    response = table.query(
        KeyConditionExpression=Key('PK').eq(f'PLAYLIST#{playlist_id}') &
                               Key('SK').begins_with('TRACK#'),
        ScanIndexForward=True,  # Ascending for position order
    )
    return response['Items']
```

### Performance Tips

1. **Always query with partition key** - Never use Scan for production queries
2. **Use begins_with for sort key** - Enable hierarchical queries
3. **Project only needed attributes** - Reduce read capacity usage
4. **Use pagination** - Limit results and use LastEvaluatedKey
5. **Avoid filters on large datasets** - Filters apply after read, wasting RCU

**References:**
- [AWS: Choosing the Right Partition Key](https://aws.amazon.com/blogs/database/choosing-the-right-dynamodb-partition-key/)
- [AWS: Best Practices for DynamoDB](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)
- [Dynobase: Partition Key Guide](https://dynobase.dev/dynamodb-partition-key/)

---

## 9. S3 Transfer Acceleration

### Overview

Transfer Acceleration uses CloudFront edge locations to speed up uploads by 50-500% for distant clients.

### Enable Transfer Acceleration

```python
# CDK
bucket = s3.Bucket(
    self, "Cr8AudioBucket",
    transfer_acceleration=True,
)

# CLI
aws s3api put-bucket-accelerate-configuration \
    --bucket cr8-audio-library \
    --accelerate-configuration Status=Enabled
```

### Using Accelerated Endpoint

```python
import boto3

# Create S3 client with Transfer Acceleration
s3_accelerated = boto3.client(
    's3',
    config=boto3.session.Config(
        s3={'use_accelerate_endpoint': True}
    )
)

# Upload using accelerated endpoint
s3_accelerated.upload_file(
    'local_track.mp3',
    'cr8-audio-library',
    'audio/user123/track456.mp3'
)
```

### When to Use

```python
TRANSFER_ACCELERATION_DECISION = {
    "use_acceleration": [
        "Users uploading from distant geographic locations",
        "Large files (100MB+) over internet",
        "Variable network conditions",
        "Cross-continent uploads",
    ],
    "skip_acceleration": [
        "Same region uploads (use VPC endpoint)",
        "Small files (<10MB)",
        "Intra-AWS transfers (use S3 endpoint)",
    ],
}
```

### Cost

- Only charged when acceleration actually improves transfer speed
- $0.04/GB for accelerated uploads (US/EU)
- $0.08/GB for accelerated uploads (other regions)

**References:**
- [AWS: S3 Transfer Acceleration](https://aws.amazon.com/s3/transfer-acceleration/)
- [AWS: Configure Transfer Acceleration](https://docs.aws.amazon.com/AmazonS3/latest/userguide/transfer-acceleration.html)
- [AWS: Transfer Acceleration Speed Test](https://s3-accelerate-speedtest.s3-accelerate.amazonaws.com/en/accelerate-speed-comparsion.html)

---

## 10. Multipart Upload Patterns

### When to Use Multipart Upload

- **Required:** Files > 5GB
- **Recommended:** Files > 100MB
- **Supported:** Files 5MB - 5TB

### Basic Multipart Upload

```python
import boto3
from boto3.s3.transfer import TransferConfig

s3 = boto3.client('s3')

# Configure multipart upload
config = TransferConfig(
    multipart_threshold=100 * 1024 * 1024,  # 100MB
    max_concurrency=10,
    multipart_chunksize=50 * 1024 * 1024,   # 50MB chunks
    use_threads=True,
)

# Upload large file
s3.upload_file(
    'large_audio_file.wav',
    'cr8-audio-library',
    'audio/user123/large_track.wav',
    Config=config,
)
```

### Manual Multipart Upload (for presigned URLs)

```python
import hashlib
import boto3

class MultipartUploader:
    """Manual multipart upload with progress tracking."""

    def __init__(self, bucket: str, key: str):
        self.s3 = boto3.client('s3')
        self.bucket = bucket
        self.key = key
        self.upload_id = None
        self.parts = []

    def initiate(self) -> str:
        """Start multipart upload."""
        response = self.s3.create_multipart_upload(
            Bucket=self.bucket,
            Key=self.key,
            ContentType='audio/mpeg',
        )
        self.upload_id = response['UploadId']
        return self.upload_id

    def upload_part(self, part_number: int, data: bytes) -> dict:
        """Upload a single part."""
        response = self.s3.upload_part(
            Bucket=self.bucket,
            Key=self.key,
            UploadId=self.upload_id,
            PartNumber=part_number,
            Body=data,
        )

        part_info = {
            'PartNumber': part_number,
            'ETag': response['ETag'],
        }
        self.parts.append(part_info)
        return part_info

    def complete(self) -> dict:
        """Complete the multipart upload."""
        response = self.s3.complete_multipart_upload(
            Bucket=self.bucket,
            Key=self.key,
            UploadId=self.upload_id,
            MultipartUpload={'Parts': sorted(self.parts, key=lambda x: x['PartNumber'])},
        )
        return response

    def abort(self):
        """Abort the multipart upload."""
        self.s3.abort_multipart_upload(
            Bucket=self.bucket,
            Key=self.key,
            UploadId=self.upload_id,
        )

# Usage
async def upload_large_file(file_path: str, bucket: str, key: str):
    """Upload large file with progress tracking."""
    CHUNK_SIZE = 50 * 1024 * 1024  # 50MB

    uploader = MultipartUploader(bucket, key)
    uploader.initiate()

    try:
        part_number = 1
        with open(file_path, 'rb') as f:
            while chunk := f.read(CHUNK_SIZE):
                uploader.upload_part(part_number, chunk)
                print(f"Uploaded part {part_number}")
                part_number += 1

        result = uploader.complete()
        print(f"Upload complete: {result['Location']}")
        return result

    except Exception as e:
        uploader.abort()
        raise e
```

### Presigned URLs for Browser Uploads

```python
def generate_presigned_multipart_urls(bucket: str, key: str, num_parts: int) -> dict:
    """Generate presigned URLs for browser-based multipart upload."""
    s3 = boto3.client('s3')

    # Create multipart upload
    response = s3.create_multipart_upload(Bucket=bucket, Key=key)
    upload_id = response['UploadId']

    # Generate presigned URL for each part
    part_urls = []
    for part_number in range(1, num_parts + 1):
        url = s3.generate_presigned_url(
            'upload_part',
            Params={
                'Bucket': bucket,
                'Key': key,
                'UploadId': upload_id,
                'PartNumber': part_number,
            },
            ExpiresIn=3600,  # 1 hour
        )
        part_urls.append({'part_number': part_number, 'url': url})

    return {
        'upload_id': upload_id,
        'key': key,
        'part_urls': part_urls,
    }
```

### Best Practices

1. **Use 5-50MB parts** - Balance between parallelism and overhead
2. **Implement retry logic** - Retry individual failed parts
3. **Track upload progress** - Store part ETags for resume capability
4. **Clean up incomplete uploads** - Set lifecycle policy to abort stale uploads
5. **Verify with checksums** - Use MD5/SHA256 for data integrity

**References:**
- [AWS: Multipart Upload Overview](https://docs.aws.amazon.com/AmazonS3/latest/userguide/mpuoverview.html)
- [AWS: Multipart Upload Limits](https://docs.aws.amazon.com/AmazonS3/latest/userguide/qfacts.html)
- [AWS Blog: Multipart Upload with Transfer Acceleration](https://aws.amazon.com/blogs/compute/uploading-large-objects-to-amazon-s3-using-multipart-upload-and-transfer-acceleration/)

---

## Migration Script Template

Complete migration script template combining all patterns:

```python
#!/usr/bin/env python3
"""cr8-cli Supabase to AWS Migration Script."""

import asyncio
import os
import boto3
from dataclasses import dataclass
from typing import Optional
from supabase import create_client

@dataclass
class MigrationConfig:
    supabase_url: str
    supabase_key: str
    dynamodb_table: str
    s3_bucket: str
    batch_size: int = 25
    parallel_workers: int = 10

class Cr8Migrator:
    def __init__(self, config: MigrationConfig):
        self.config = config
        self.supabase = create_client(config.supabase_url, config.supabase_key)
        self.dynamodb = boto3.resource('dynamodb')
        self.table = self.dynamodb.Table(config.dynamodb_table)
        self.s3 = boto3.client('s3', config=boto3.session.Config(s3={'use_accelerate_endpoint': True}))

    async def run_migration(self):
        """Run full migration."""
        print("Starting cr8-cli migration...")

        # Migrate in order of dependencies
        await self.migrate_users()
        await self.migrate_tracks()
        await self.migrate_playlists()
        await self.migrate_sessions()

        print("Migration complete!")

    async def migrate_tracks(self):
        """Migrate tracks with audio files."""
        offset = 0
        total = 0

        while True:
            response = self.supabase.table('tracks').select('*').range(offset, offset + self.config.batch_size - 1).execute()

            if not response.data:
                break

            items = []
            for track in response.data:
                # Transform track data
                item = self._transform_track(track)
                items.append(item)

                # Queue audio file migration (if needed)
                if track.get('storage_path'):
                    await self._migrate_audio_file(track)

            # Batch write to DynamoDB
            with self.table.batch_writer() as batch:
                for item in items:
                    batch.put_item(Item=item)

            offset += self.config.batch_size
            total += len(items)
            print(f"Migrated {total} tracks...")

    def _transform_track(self, track: dict) -> dict:
        """Transform Supabase track to DynamoDB format."""
        user_id = track['user_id']
        track_id = track['id']

        return {
            'PK': f"USER#{user_id}",
            'SK': f"TRACK#{track_id}",
            'GSI1PK': f"GENRE#{track.get('genre', 'unknown')}",
            'GSI1SK': f"BPM#{track.get('bpm', 0):03d}#{track_id}",
            'entity_type': 'track',
            'track_id': track_id,
            'title': track['title'],
            'artist': track.get('artist'),
            'bpm': track.get('bpm'),
            'key': track.get('musical_key'),
            'duration': track.get('duration'),
            's3_key': f"audio/{user_id}/{track_id}.mp3",
            'metadata': track.get('metadata', {}),
            'created_at': track['created_at'],
        }

if __name__ == '__main__':
    config = MigrationConfig(
        supabase_url=os.environ['SUPABASE_URL'],
        supabase_key=os.environ['SUPABASE_KEY'],
        dynamodb_table='cr8-music-library',
        s3_bucket='cr8-audio-library',
    )

    migrator = Cr8Migrator(config)
    asyncio.run(migrator.run_migration())
```

---

## Additional Resources

### AWS Documentation
- [DynamoDB Developer Guide](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/)
- [S3 User Guide](https://docs.aws.amazon.com/AmazonS3/latest/userguide/)
- [DynamoDB Best Practices](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)

### Tools
- [NoSQL Workbench](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/workbench.html) - Data modeling tool
- [Dynobase](https://dynobase.dev/) - DynamoDB GUI client
- [AWS CLI](https://aws.amazon.com/cli/) - Command-line interface

### Community Resources
- [Alex DeBrie's DynamoDB Book](https://www.dynamodbbook.com/)
- [Serverless Life DynamoDB Patterns](https://www.serverlesslife.com/DynamoDB_Design_Patterns_for_Single_Table_Design.html)
- [AWS Database Blog](https://aws.amazon.com/blogs/database/)

---

*Document generated for cr8-cli AWS migration planning - December 2025*
