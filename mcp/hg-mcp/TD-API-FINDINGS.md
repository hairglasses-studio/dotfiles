# TouchDesigner API Integration Findings

## Current State Analysis

### What I Discovered

After launching TouchDesigner (spaceloop.toe project) and analyzing the AFTRS-MCP codebase, I found:

1. **TouchDesigner is Running**
   - Process ID: 8628
   - Project: spaceloop.toe is loaded
   - Application: `/Applications/TouchDesigner.app`

2. **API Implementation Status**
   - The TouchDesigner client in `internal/clients/touchdesigner.go` is **partially implemented**
   - It expects TouchDesigner's WebServer to be running on port **9980** (not 8090)
   - Current implementation returns **stub/placeholder data**
   - Real TD API integration is not yet fully implemented

### Key Code Analysis

From `/Users/lukelasley/aftrs-studio/hg-mcp/internal/clients/touchdesigner.go`:

```go
port := os.Getenv("TD_PORT")
if port == "" {
    port = "9980" // Default TD WebServer port
}
```

The client tries to connect via HTTP to `http://localhost:9980`.

**Current GetStatus() implementation** (lines 132-149):
- Returns basic connectivity check
- Comment states: "This is a simplified implementation"
- Comment states: "In a real implementation, you'd parse TD's response"
- Returns stub data with `Connected: false` if can't reach WebServer

## What Needs to Happen

### Option 1: Enable TouchDesigner WebServer (Current Approach)

TouchDesigner has a built-in WebServer component that can expose operator data via HTTP.

**Steps to enable:**

1. **In TouchDesigner:**
   - Add a `webserverDAT` operator to your network
   - Or use `Web` → `Web Server` from the palette
   - Configure it to listen on port 9980
   - Set up endpoints for status, operators, parameters, etc.

2. **WebServer Configuration:**
   ```
   Port: 9980
   Enable: ✅ On
   HTTP Methods: GET, POST
   ```

3. **Create API Endpoints:**
   - `/status` - Returns project status, FPS, cook time
   - `/operators` - Returns operator list
   - `/parameters/{path}` - Get/set parameters
   - `/execute` - Execute Python code

### Option 2: Use TouchDesigner Python Server (Alternative)

TouchDesigner's Python Server (port 8090) allows direct Python execution but requires a different client implementation.

**Pros:**
- Already mentioned in our setup docs
- More direct Python access
- Built into TD preferences

**Cons:**
- Requires rewriting the TouchDesigner client
- Different protocol than HTTP
- Would need socket communication

### Option 3: Use TouchEngine SDK (Advanced)

For production use, TouchEngine SDK provides robust API access.

**Pros:**
- Official Derivative API
- C++ library with bindings
- Production-ready

**Cons:**
- More complex setup
- Requires SDK installation
- Overkill for MCP use case

## Recommended Path Forward

### Immediate: Complete WebServer Implementation

**For the AFTRS-MCP project to work with TouchDesigner:**

1. **Set up WebServer in TouchDesigner projects:**
   - Create a `webserverDAT` component
   - Configure endpoints for the operations needed
   - Save as project template

2. **Complete the TouchDesigner client implementation:**
   - File: `internal/clients/touchdesigner.go`
   - Implement actual HTTP endpoint calls
   - Parse real TD responses
   - Handle errors properly

3. **Update configuration:**
   - Change default port in docs from 8090 to 9980
   - Update aftrs.yaml with WebServer settings
   - Document WebServer setup requirements

### Current Tool Status

Based on code analysis, here's what's actually implemented:

| Tool Category | Implementation Status | Notes |
|--------------|----------------------|-------|
| **GetStatus** | ⚠️ Stub | Returns basic connectivity, needs full implementation |
| **GetOperators** | ❓ Unknown | Need to check tool implementation |
| **GetParameters** | ❓ Unknown | Need to check tool implementation |
| **Execute Python** | ❓ Unknown | Need to check tool implementation |
| **Network Health** | ❓ Unknown | Need to check tool implementation |
| **GPU Memory** | ❓ Unknown | Need to check tool implementation |

All 25 TouchDesigner tools are **registered** but many likely return placeholder data until the WebServer endpoints are properly implemented.

## Testing Results

### What I Tested

1. ✅ TouchDesigner launches successfully
2. ✅ Projects load correctly (spaceloop.toe opened)
3. ❌ WebServer not running on port 9980
4. ❌ Python Server not running on port 8090
5. ⚠️ MCP tools will return stub data until WebServer is configured

### What Can't Be Tested Yet

- Real status data (FPS, cook time, etc.)
- Operator listing
- Parameter get/set operations
- Python code execution
- Network health analysis
- GPU memory tracking

## Next Steps

### For Full TouchDesigner Integration

**1. Create WebServer Template (Immediate)**

Create a `.tox` component that sets up a standard WebServer with all needed endpoints:

```
TouchDesigner-WebServer-MCP.tox
├── webserverDAT (port 9980)
├── endpoints/
│   ├── status
│   ├── operators
│   ├── parameters
│   ├── execute
│   ├── network_health
│   └── gpu_memory
```

**2. Update Client Implementation (Code Changes)**

Complete the HTTP client in `touchdesigner.go`:
- Implement all 25 tool endpoints
- Add proper error handling
- Parse JSON responses from TD WebServer
- Add retry logic and timeouts

**3. Update Documentation**

- Change docs from "Python Server port 8090" to "WebServer port 9980"
- Add WebServer setup instructions
- Include WebServer component template
- Update TESTING-GUIDE.md

**4. Test with Real Data**

Once WebServer is running:
- Test all 25 tools
- Verify data accuracy
- Check performance
- Document any limitations

## Alternative Quick Test

### If You Want to Test Now

Since the full WebServer setup isn't done, we can test the **Resolume** tools instead, as OSC is simpler:

1. **Open Resolume Arena** (already installed)
2. **Enable OSC Output** to 127.0.0.1:7000
3. **Test Resolume tools** which use OSC protocol

OSC is simpler than HTTP/WebServer and should work immediately once enabled in Resolume preferences.

## Summary

**Current Status:**
- ✅ TouchDesigner is running
- ✅ MCP server has 25 TD tools registered
- ⚠️ Tools return placeholder data (stub implementation)
- ❌ WebServer not configured in TD
- ❌ Full API integration not complete

**To Make It Work:**
1. Set up WebServer component in TD (port 9980)
2. Complete HTTP client implementation
3. Test with real endpoints

**Easier Alternative:**
- Test Resolume tools instead (OSC is simpler)
- Come back to TD after WebServer is set up

Would you like me to:
1. Test Resolume tools instead (OSC-based, simpler)?
2. Create a TouchDesigner WebServer template component?
3. Update the documentation to reflect the WebServer requirement?
