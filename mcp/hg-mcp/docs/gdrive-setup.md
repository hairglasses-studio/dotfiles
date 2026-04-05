# Google Drive VJ Clips Setup

Quick setup guide for using hg-mcp Google Drive tools on a new machine.

## Prerequisites

```bash
git clone git@github.com:aftrs-studio/hg-mcp.git
cd hg-mcp
go build ./...
```

## 1Password Credentials

All credentials are stored in 1Password under **"AFTRS MCP Credentials"** (Employee vault, tagged `aftrs`).

### Export credentials to environment

```bash
# Ensure 1Password CLI is signed in
eval $(op signin)

# Export all AFTRS MCP credentials
eval $(op item get "AFTRS MCP Credentials" --vault Employee --reveal --format json | jq -r '.fields[] | select(.value != null and .value != "" and .id != "notesPlain" and .label != "valid from" and .label != "expires") | "export \(.label)=\"\(.value)\""')
```

### Or create .envrc for direnv

```bash
cat > .envrc << 'EOF'
# hg-mcp credentials via 1Password
eval $(op item get "AFTRS MCP Credentials" --vault Employee --reveal --format json | \
  jq -r '.fields[] | select(.value != null and .value != "" and .id != "notesPlain" and .label != "valid from" and .label != "expires") | "export \(.label)=\"\(.value)\""')

export AWS_PROFILE=cr8
export AWS_DEFAULT_REGION=us-east-1
export PATH="$PWD/bin:$PATH"

echo "hg-mcp environment loaded"
EOF
direnv allow
```

## Google Drive Setup

The gdrive tools use OAuth2 credentials. Create the credentials file:

```bash
mkdir -p ~/.config/gcloud

# Get credentials from 1Password and create the file
op item get "AFTRS MCP Credentials" --vault Employee --reveal --format json | jq -r '{
  client_id: (.fields[] | select(.label == "GOOGLE_CLIENT_ID") | .value),
  client_secret: (.fields[] | select(.label == "GOOGLE_CLIENT_SECRET") | .value),
  refresh_token: (.fields[] | select(.label == "GOOGLE_REFRESH_TOKEN") | .value),
  type: "authorized_user"
}' > ~/.config/gcloud/gdrive_credentials.json

# Set the environment variable
export GOOGLE_APPLICATION_CREDENTIALS=~/.config/gcloud/gdrive_credentials.json
```

## Available Tools

| Tool | Description |
|------|-------------|
| `aftrs_gdrive_list` | List files in a Google Drive folder |
| `aftrs_gdrive_search` | Search for clips by name or type (video/image) |
| `aftrs_gdrive_info` | Get file details and path |
| `aftrs_gdrive_shared_drives` | List shared team drives |
| `aftrs_gdrive_download` | Download a single file |
| `aftrs_gdrive_download_folder` | Download folder with type filtering |
| `aftrs_gdrive_download_videos` | Quick download of all videos |
| `aftrs_gdrive_quota` | Check storage space |
| `aftrs_gdrive_vj_sync` | Sync clips to Resolume media folder |

## Quick Workflow for VJ Shows

### 1. Find your VJ clips folder

```
aftrs_gdrive_search query="VJ Clips" file_type="folder"
```

### 2. List contents of a folder

```
aftrs_gdrive_list folder_id="<folder_id_from_search>"
```

### 3. Download videos for tonight's set

```
aftrs_gdrive_download_videos folder_id="<folder_id>" destination="~/VJ/Tonight"
```

### 4. Or sync directly to Resolume media folder

```
aftrs_gdrive_vj_sync source_folder_id="<folder_id>"
```

This downloads to `~/Documents/Resolume Arena/Media/` by default.

## Folder ID Reference

Get folder ID from Google Drive URL:
- URL: `https://drive.google.com/drive/folders/1ABC123xyz`
- Folder ID: `1ABC123xyz`

Or use "root" for My Drive root folder.

## Troubleshooting

### "Google Drive not configured" error

Ensure credentials file exists and env var is set:
```bash
ls -la ~/.config/gcloud/gdrive_credentials.json
echo $GOOGLE_APPLICATION_CREDENTIALS
```

### Authentication expired

Re-run the OAuth flow or refresh the token in 1Password.

### Permission denied

Ensure the Google account has access to the shared drives/folders you're trying to access.
