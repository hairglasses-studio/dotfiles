# TouchDesigner WebServer Setup Guide

This guide explains how to set up the WebServer DAT in TouchDesigner to enable control from hg-mcp tools.

## Prerequisites

- TouchDesigner 2022.28000 or later
- Network access on port 9980 (configurable)

## Quick Setup

### 1. Create the WebServer DAT

1. Open your TouchDesigner project
2. Press `Tab` to open the OP Create dialog
3. Type "webserver" and select **Web Server DAT**
4. Place it in your project (e.g., `/project1/webserver1`)

### 2. Create the Callbacks Text DAT

1. Press `Tab` again
2. Type "text" and select **Text DAT**
3. Name it `webserver_callbacks`
4. Copy the contents of `td_webserver.py` into this Text DAT

### 3. Configure the WebServer DAT

1. Select the WebServer DAT
2. Set the following parameters:
   - **Port**: `9980` (must match hg-mcp config)
   - **Callbacks DAT**: `webserver_callbacks` (the Text DAT you created)
   - **Active**: `On`

### 4. Verify Connection

Open a terminal and run:

```bash
curl http://localhost:9980/health
```

Expected response:
```json
{
  "status": "ok",
  "connected": true,
  "version": "2022.28000",
  "build": "2022.28000"
}
```

## Available Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Basic health check |
| GET | `/status` | Detailed project status |
| GET | `/operators?path=/project1` | List operators in a network |
| GET | `/parameters?op=/project1/geo1` | Get operator parameters |
| POST | `/parameters` | Set operator parameters |
| POST | `/execute` | Execute Python code |
| GET | `/performance` | Performance metrics |
| GET | `/errors` | List all errors in project |
| GET | `/project_info` | Project metadata |
| GET | `/network_health?path=/project1` | Health metrics for a network |
| POST | `/pulse` | Pulse a parameter |
| POST | `/cook` | Force cook an operator |

## Example Usage

### Get Project Status

```bash
curl http://localhost:9980/status
```

### List Operators

```bash
curl "http://localhost:9980/operators?path=/project1"
```

### Set a Parameter

```bash
curl -X POST http://localhost:9980/parameters \
  -H "Content-Type: application/json" \
  -d '{"op": "/project1/geo1", "parameters": {"tx": 5.0}}'
```

### Execute Python Code

```bash
curl -X POST http://localhost:9980/execute \
  -H "Content-Type: application/json" \
  -d '{"code": "result = len(root.children)"}'
```

## Integration with hg-mcp

Once the WebServer is running, the following hg-mcp tools will work:

- `aftrs_td_status` - Get TouchDesigner status
- `aftrs_td_operators_list` - List operators
- `aftrs_td_parameter_get` - Get parameter values
- `aftrs_td_parameter_set` - Set parameter values
- `aftrs_td_execute_python` - Execute Python code
- `aftrs_td_performance_metrics` - Get performance data
- `aftrs_td_errors_list` - List errors
- `aftrs_td_network_health` - Check network health

## Troubleshooting

### Port Already in Use

If port 9980 is in use, change the port in:
1. TouchDesigner WebServer DAT
2. hg-mcp config (`~/.config/aftrs/touchdesigner.yaml`)

### Connection Refused

1. Ensure the WebServer DAT is set to **Active: On**
2. Check Windows Firewall allows connections on port 9980
3. Verify TouchDesigner is running

### Callbacks Not Working

1. Ensure the Callbacks DAT parameter points to the correct Text DAT
2. Check the Textport (View > Panels > Textport) for Python errors
3. Verify the script loaded: "TD WebServer API loaded" should appear

### CORS Issues

If accessing from a browser, add CORS headers in the response:
```python
response['Access-Control-Allow-Origin'] = '*'
```

## Security Considerations

- The WebServer allows arbitrary Python execution via `/execute`
- Only run on trusted networks
- Consider adding authentication for production use
- Use firewall rules to restrict access to localhost only

## Performance Notes

- The WebServer runs on TouchDesigner's main thread
- Avoid long-running operations in request handlers
- Use async patterns for heavy computations
- Monitor FPS impact when under heavy API load