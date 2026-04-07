# TouchDesigner Python Bridge

HTTP bridge server that connects AFTRS-MCP to TouchDesigner via the Python Server.

## Architecture

```
AFTRS-MCP (Go)  →  HTTP (port 9980)  →  Python Bridge (Flask)  →  Socket (port 8090)  →  TouchDesigner
```

The bridge translates HTTP REST requests into Python commands that execute inside TouchDesigner's Python environment.

## Quick Start

### 1. Enable TouchDesigner Python Server

**In TouchDesigner:**
1. Open Preferences: `Edit → Preferences` (or `Cmd+,` on macOS)
2. Navigate to: `Network → Python Server`
3. Configure:
   - ✅ **Enable Server** (check this box)
   - **Port:** `8090` (default, must match bridge config)
   - **Allow Connections From:** `localhost` (recommended for security)
4. Click `Apply`
5. Verify in textport: Should see "Python Server listening on port 8090"

**Important:** The Python Server must be enabled before starting the bridge.

### 2. Start the Bridge

**Option A: Using the startup script (recommended)**
```bash
cd ~/hairglasses-studio/hg-mcp/td-python-bridge
./start.sh
```

**Option B: Manual start**
```bash
cd ~/hairglasses-studio/hg-mcp/td-python-bridge
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python bridge.py
```

**Expected Output:**
```
========================================
TouchDesigner Python Bridge
========================================
TouchDesigner Python Server: localhost:8090
Bridge HTTP Server: http://localhost:9980
========================================

Connecting to TouchDesigner...
✓ Connected to TouchDesigner Python Server

Starting Flask server...
 * Running on http://localhost:9980

Endpoints available:
  GET  /health           - Health check
  GET  /status           - Project status
  GET  /operators        - List operators
  GET  /parameters       - Get parameters
  POST /parameters       - Set parameter
  POST /execute          - Execute Python code
  GET  /network_health   - Network health analysis
  GET  /errors           - List errors/warnings
  GET  /performance      - Performance metrics
  GET  /project_info     - Project information

Bridge ready!
```

### 3. Test the Bridge

**Health check:**
```bash
curl http://localhost:9980/health
```

**Expected:** `{"status":"ok","connected":true}`

**Get project status:**
```bash
curl http://localhost:9980/status
```

**Expected:** JSON with project name, FPS, cook time, etc.

## API Endpoints

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "connected": true
}
```

### GET /status
Get TouchDesigner project status.

**Response:**
```json
{
  "connected": true,
  "project_name": "myproject.toe",
  "fps": 60.0,
  "realtime_fps": 59.94,
  "cook_time_ms": 12.5,
  "error_count": 0,
  "warning_count": 2,
  "version": "2023.11760"
}
```

### GET /operators?path=/
List operators in a network.

**Parameters:**
- `path` (optional, default: `/`) - Network path to query

**Response:**
```json
{
  "operators": [
    {
      "name": "null1",
      "path": "/null1",
      "type": "null",
      "family": "TOP",
      "has_errors": false,
      "cook_time_ms": 0.05
    }
  ]
}
```

### GET /parameters?path=/geo1
Get operator parameters.

**Parameters:**
- `path` (required) - Operator path

**Response:**
```json
{
  "tx": 0.0,
  "ty": 0.0,
  "tz": 0.0,
  "rx": 0.0,
  "ry": 0.0,
  "rz": 0.0
}
```

### POST /parameters
Set operator parameter value.

**Request Body:**
```json
{
  "path": "/null1",
  "param": "opacity",
  "value": 0.5
}
```

**Response:**
```json
{
  "success": true
}
```

### POST /execute
Execute arbitrary Python code in TouchDesigner.

**Request Body:**
```json
{
  "code": "print('Hello from AFTRS-MCP!')\nprint(app.version)"
}
```

**Response:**
```json
{
  "success": true,
  "output": "Hello from AFTRS-MCP!\n2023.11760"
}
```

### GET /network_health?path=/
Analyze network health.

**Parameters:**
- `path` (optional, default: `/`) - Network path to analyze

**Response:**
```json
{
  "score": 95,
  "status": "healthy",
  "total_operators": 42,
  "error_count": 0,
  "warning_count": 2,
  "slow_operators": [
    {
      "path": "/moviefilein1",
      "cook_time_ms": 15.2
    }
  ]
}
```

**Health Score Calculation:**
- Start at 100
- Subtract 10 per error
- Subtract 2 per warning
- Minimum score: 0

**Status Values:**
- `healthy` - No errors
- `degraded` - Has errors

### GET /errors
List all errors and warnings in the project.

**Response:**
```json
{
  "errors": [
    {
      "path": "/moviefilein1",
      "error": "File not found: video.mov"
    }
  ],
  "warnings": [
    {
      "path": "/null1",
      "warning": "No input connected"
    }
  ]
}
```

### GET /performance
Get performance metrics.

**Response:**
```json
{
  "fps": 60.0,
  "cook_time_ms": 12.5,
  "frame": 1234,
  "seconds": 20.567
}
```

### GET /project_info
Get project information.

**Response:**
```json
{
  "name": "myproject.toe",
  "folder": "~/projects",
  "save_file": "~/projects/myproject.toe"
}
```

## Configuration

**Environment Variables:**

- `TD_HOST` - TouchDesigner Python Server host (default: `localhost`)
- `TD_PORT` - TouchDesigner Python Server port (default: `8090`)
- `BRIDGE_HOST` - Bridge HTTP server host (default: `0.0.0.0`)
- `BRIDGE_PORT` - Bridge HTTP server port (default: `9980`)

**Example:**
```bash
export TD_PORT=8091
export BRIDGE_PORT=9981
python bridge.py
```

## Troubleshooting

### Bridge can't connect to TouchDesigner

**Error:** `Connection refused` or `Failed to connect to TouchDesigner`

**Solutions:**

1. **Verify Python Server is enabled:**
   - TouchDesigner → Preferences → Network → Python Server
   - Must be checked and showing "listening on port 8090"

2. **Check port number:**
   - Bridge expects port 8090 by default
   - Verify TD Python Server port matches: `echo $TD_PORT`

3. **Verify TouchDesigner is running:**
   - Bridge needs a running TD instance with a project open

4. **Check firewall:**
   - macOS may block localhost connections
   - System Preferences → Security & Privacy → Firewall → Firewall Options
   - Ensure TouchDesigner is allowed

5. **Check for port conflicts:**
   ```bash
   lsof -i :8090  # Check if port is in use by another process
   ```

### Bridge starts but returns errors

**Error:** `500 Internal Server Error` on API calls

**Solutions:**

1. **Check TouchDesigner textport for errors:**
   - Open textport in TD (Alt+T)
   - Look for Python execution errors

2. **Verify Python code syntax:**
   - Bridge executes Python in TD's environment
   - Code must be valid Python 3.9+ (TD 2023+)

3. **Check bridge logs:**
   - Bridge prints detailed error messages to console
   - Look for `ERROR:` lines in output

4. **Test simple execute:**
   ```bash
   curl -X POST http://localhost:9980/execute \
     -H "Content-Type: application/json" \
     -d '{"code":"print(123)"}'
   ```
   Should return: `{"success":true,"output":"123"}`

### Specific endpoint returns empty data

**Issue:** API call succeeds but returns empty arrays or null values

**Causes:**

1. **Network path doesn't exist:**
   - `/operators?path=/foo` returns empty if `/foo` doesn't exist
   - Verify path exists in TouchDesigner

2. **Operator path doesn't exist:**
   - `/parameters?path=/null999` returns error if operator missing
   - Check operator name and path

3. **No operators in network:**
   - `/operators?path=/` returns empty if network is empty
   - Normal for empty networks

### Bridge crashes or disconnects

**Error:** `Connection lost` or bridge exits unexpectedly

**Solutions:**

1. **Restart both TD and bridge:**
   - Stop bridge (Ctrl+C)
   - Close and reopen TouchDesigner project
   - Re-enable Python Server
   - Restart bridge

2. **Check for TD crashes:**
   - If TD crashes, bridge will lose connection
   - Monitor TD stability (check Activity Monitor)

3. **Check for socket timeouts:**
   - Bridge has 5-second timeout on socket operations
   - Very slow TD operations may timeout
   - Check for cook performance issues in TD

### Python Server won't enable

**Issue:** Can't check "Enable Server" in TD Preferences

**Solutions:**

1. **Port already in use:**
   ```bash
   lsof -i :8090
   ```
   If something else is using port 8090, change TD to different port:
   - Change port in TD Preferences
   - Set `TD_PORT=8091` environment variable
   - Restart bridge

2. **Permissions issue:**
   - macOS may require admin password to enable network services
   - Try running TD with elevated permissions (not recommended)
   - Or use port > 1024 (no admin needed)

3. **TD installation issue:**
   - Reinstall TouchDesigner
   - Check for license issues

## Testing

**Unit tests:**
```bash
cd ~/hairglasses-studio/hg-mcp/td-python-bridge
python -m pytest test_bridge.py -v
```

**Integration test with live TouchDesigner:**

1. Start TouchDesigner with Python Server enabled
2. Open a project (e.g., spaceloop.toe)
3. Start bridge: `./start.sh`
4. Run test script:
   ```bash
   cd ~/hairglasses-studio/hg-mcp
   ./test/cli_test.sh
   ```

## Development

**Adding new endpoints:**

1. Add route to `bridge.py`:
   ```python
   @app.route('/my_endpoint', methods=['GET'])
   def my_endpoint():
       code = """
   # Your TD Python code here
   result = op('/').name
   print(json.dumps({'result': result}))
   """
       result = td_conn.execute(code)
       return jsonify(result.get('result'))
   ```

2. Update this README with endpoint documentation

3. Update AFTRS-MCP Go client to call new endpoint

**Debugging:**

Enable Flask debug mode:
```python
# At bottom of bridge.py
if __name__ == '__main__':
    app.run(host='0.0.0.0', port=9980, debug=True)
```

**Important:** Never use debug mode in production (security risk)

## Architecture Notes

**Why a bridge instead of direct connection?**

1. **No TD project modification:** Python Server is just a preference setting
2. **Works with all projects:** No need to add WebServer DAT to every project
3. **Centralized updates:** Update bridge once, affects all projects
4. **REST API simplicity:** Go HTTP client easier than Python socket protocol
5. **Future extensibility:** Can add caching, request queuing, etc.

**Performance Considerations:**

- Each API call requires: HTTP request → Socket send → TD execute → Socket receive → HTTP response
- Typical latency: 10-50ms for simple operations
- Heavy operations (network analysis) may take 100-500ms
- Bridge is single-threaded (Flask default)
- For production, consider using gunicorn with multiple workers:
  ```bash
  pip install gunicorn
  gunicorn -w 4 -b 0.0.0.0:9980 bridge:app
  ```

**Security:**

- Bridge listens on `0.0.0.0` (all interfaces) by default
- **Recommendation:** Only use on localhost or trusted networks
- No authentication implemented (assumes trusted local environment)
- Python code execution endpoint can run arbitrary code in TD
- For production: Add API key authentication, rate limiting, input validation

## Integration with AFTRS-MCP

The AFTRS-MCP Go client connects to this bridge:

**File:** `internal/clients/touchdesigner.go`

```go
func NewTouchDesignerClient(baseURL string) *TouchDesignerClient {
    return &TouchDesignerClient{
        baseURL: "http://localhost:9980", // Python bridge
        client:  &http.Client{Timeout: 10 * time.Second},
    }
}
```

All 27 TouchDesigner MCP tools call bridge endpoints.

## Requirements

- **Python:** 3.9+ (3.11+ recommended)
- **TouchDesigner:** 2022.24xxx+ (tested on 2023.11760)
- **Operating System:** macOS, Windows, Linux
- **Network:** Localhost connection (no internet required)

## License

Part of the AFTRS-MCP project.

## Support

Issues and questions: https://github.com/hairglasses-studio/hg-mcp/issues
