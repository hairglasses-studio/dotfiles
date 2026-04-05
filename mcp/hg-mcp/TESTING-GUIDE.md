# AFTRS-MCP Testing Guide
## TouchDesigner & Resolume Tools Testing

This guide will help you test the AFTRS MCP tools with your local TouchDesigner and Resolume projects.

## Prerequisites

### TouchDesigner Setup

1. **Open TouchDesigner**
   - Launch: `/Applications/TouchDesigner.app`
   - Open one of your projects from `~/aftrs-studio/visual-projects/touchdesigner/projects/`
   - Recommended test project: `Watercolor_Splats.toe` or `spaceloop.toe`

2. **Enable Python Server**
   - Open Preferences: `Edit → Preferences` (or `Cmd+,`)
   - Navigate to: `Network → Python Server`
   - Check: ✅ Enable Server
   - Port: `8090` (must match TD_PORT environment variable)
   - Click `Apply`

3. **Start Python Bridge**
   - Open a new terminal window
   - Run: `cd ~/aftrs-studio/hg-mcp/td-python-bridge && ./start.sh`
   - Bridge will start on port 9980 and connect to TD Python Server on port 8090
   - Keep both TouchDesigner and the bridge running

4. **Verify Connection**
   - The Python server should show "Listening on port 8090"
   - The bridge should show "Connected to TouchDesigner Python Server"
   - Keep TouchDesigner running with a project open

### Resolume Setup

1. **Open Resolume Arena**
   - Launch: `/Applications/Resolume Arena`
   - Open one of your compositions from `~/aftrs-studio/visual-projects/resolume/compositions/`
   - Recommended test composition: `LUKE_2.21.avc` or `Example.avc`

2. **Enable OSC Output**
   - Open Preferences: `Cmd+,`
   - Navigate to: `OSC` section
   - In `OSC Output` tab:
     - Check: ✅ Enable OSC Output
     - Target Host: `127.0.0.1`
     - Target Port: `7000` (must match aftrs.yaml config)
   - Click `OK`

3. **Verify OSC is Active**
   - You should see OSC output enabled in the status bar
   - Keep Resolume running with a composition loaded

## Testing with AFTRS MCP Tools

### Method 1: Using the CLI Tool Directly

Open a new terminal and run these commands:

#### TouchDesigner Tests

```bash
# Test 1: Check TouchDesigner connection and status
aftrs-cli tool touchdesigner status

# Test 2: List available TD operators
aftrs-cli tool touchdesigner operators /

# Test 3: Get TouchDesigner performance metrics
aftrs-cli tool touchdesigner performance

# Test 4: Check for TD errors/warnings
aftrs-cli tool touchdesigner errors

# Test 5: Get GPU memory usage
aftrs-cli tool touchdesigner gpu_memory

# Test 6: Execute simple Python command
aftrs-cli tool touchdesigner execute "print('Hello from AFTRS MCP!')"
```

#### Resolume Tests

```bash
# Test 1: Check Resolume connection status
aftrs-cli tool resolume status

# Test 2: List all clips in composition
aftrs-cli tool resolume clips

# Test 3: Get current layer information
aftrs-cli tool resolume layers

# Test 4: Get composition info
aftrs-cli tool resolume composition

# Test 5: Check BPM settings
aftrs-cli tool resolume bpm

# Test 6: Trigger a clip (change layer/clip numbers as needed)
aftrs-cli tool resolume trigger --layer 1 --clip 1
```

### Method 2: Using Claude Code with MCP

Once you've verified the applications are running and configured:

1. **Start the MCP Server**
   ```bash
   aftrs-start
   ```

2. **Restart your MCP client** (if already running)
   - The MCP server needs to be running before the client starts
   - Codex or Claude Code will automatically connect to the stdio MCP server

3. **Test via Natural Language**

   Ask Claude Code:

   **For TouchDesigner:**
   ```
   "Check TouchDesigner performance and show FPS"
   "What operators are in the TouchDesigner root network?"
   "Show me any TouchDesigner errors or warnings"
   "Execute this Python in TouchDesigner: print(op('/').path)"
   "Check TouchDesigner GPU memory usage"
   ```

   **For Resolume:**
   ```
   "Check Resolume connection status"
   "List all clips in Resolume layer 1"
   "Show me the current Resolume composition info"
   "What's the current BPM in Resolume?"
   "Trigger clip 2 on layer 1 in Resolume"
   ```

## Expected Test Results

### TouchDesigner Success Indicators

✅ **Status Check** should return:
- Project name
- FPS (should be ~60 for good performance)
- Cook time in milliseconds
- Window resolution
- Build number

✅ **Operators List** should return:
- List of operators in the network
- Operator types and paths

✅ **Performance** should return:
- CPU usage percentage
- Frame time
- Cook time breakdown
- Memory usage

✅ **GPU Memory** should return:
- Texture memory usage
- GPU memory available
- Memory warnings if any

### Resolume Success Indicators

✅ **Status Check** should return:
- Connection: "Connected"
- Composition name
- Active layers
- Current BPM

✅ **Clips List** should return:
- Array of clips per layer
- Clip names
- Clip types (video/image)

✅ **Layers** should return:
- Layer numbers
- Active layers
- Layer properties

✅ **Trigger Clip** should:
- Return success message
- Actually trigger the clip in Resolume (visual confirmation)

## Troubleshooting

### TouchDesigner Connection Fails

**Error**: "Cannot connect to TouchDesigner"

**Solutions**:
1. Verify TouchDesigner is running
2. Check Python Server is enabled:
   - `Edit → Preferences → Network → Python Server`
   - Must be checked and on port 8090
3. Check firewall isn't blocking port 8090
4. Try restarting TouchDesigner
5. Verify no other app is using port 8090:
   ```bash
   lsof -i :8090
   ```

### Resolume Connection Fails

**Error**: "Cannot connect to Resolume" or "OSC not responding"

**Solutions**:
1. Verify Resolume Arena is running
2. Check OSC Output is enabled:
   - `Preferences → OSC → OSC Output`
   - Must be enabled, targeting 127.0.0.1:7000
3. Check firewall isn't blocking port 7000
4. Try restarting Resolume
5. Verify no other app is using port 7000:
   ```bash
   lsof -i :7000
   ```

### MCP Server Not Starting

**Error**: "aftrs-start: command not found"

**Solutions**:
1. Reload shell configuration:
   ```bash
   source ~/.zshrc
   ```
2. Verify aliases are loaded:
   ```bash
   alias | grep aftrs
   ```

### Tools Return No Data

**Issue**: Commands run but return empty results

**Solutions**:
1. Make sure a project/composition is actually open
2. Verify the project has content (operators in TD, clips in Resolume)
3. Check the applications aren't in a loading/busy state
4. Try a simple test first (status check)

## Test Checklist

Before reporting results, verify:

**TouchDesigner:**
- [ ] TouchDesigner is running
- [ ] Python Server is enabled on port 8090
- [ ] A project file (.toe) is open
- [ ] Project has operators in the network
- [ ] Status check returns project info
- [ ] Can list operators successfully
- [ ] Performance metrics are returned
- [ ] Python execution works

**Resolume:**
- [ ] Resolume Arena is running
- [ ] OSC Output is enabled (127.0.0.1:7000)
- [ ] A composition (.avc) is loaded
- [ ] Composition has clips on layers
- [ ] Status check returns composition info
- [ ] Can list clips successfully
- [ ] Can get layer information
- [ ] Can trigger clips (visible in Resolume)

## Advanced Testing

### Test All TouchDesigner Tools (25 tools)

```bash
# Status and monitoring
aftrs-cli tool touchdesigner status
aftrs-cli tool touchdesigner performance
aftrs-cli tool touchdesigner network_health
aftrs-cli tool touchdesigner errors
aftrs-cli tool touchdesigner gpu_memory
aftrs-cli tool touchdesigner textures

# Network operations
aftrs-cli tool touchdesigner operators /
aftrs-cli tool touchdesigner operators /project1
aftrs-cli tool touchdesigner connections /

# Parameters
aftrs-cli tool touchdesigner parameters /null1
aftrs-cli tool touchdesigner set_parameter --path /null1 --param opacity --value 0.5

# CHOPs and DATs
aftrs-cli tool touchdesigner chop_channels /
aftrs-cli tool touchdesigner dat_content /table1

# Execution
aftrs-cli tool touchdesigner execute "print(app.version)"
aftrs-cli tool touchdesigner variables
```

### Test All Resolume Tools (20 tools)

```bash
# Status and info
aftrs-cli tool resolume status
aftrs-cli tool resolume composition
aftrs-cli tool resolume layers
aftrs-cli tool resolume clips

# Clip control
aftrs-cli tool resolume trigger --layer 1 --clip 1
aftrs-cli tool resolume clear_layer --layer 1

# Deck control
aftrs-cli tool resolume deck_active --deck A
aftrs-cli tool resolume crossfade --position 0.5

# Effects
aftrs-cli tool resolume effects --layer 1
aftrs-cli tool resolume toggle_effect --layer 1 --effect 1

# BPM and tempo
aftrs-cli tool resolume bpm
aftrs-cli tool resolume set_bpm --bpm 128

# Output
aftrs-cli tool resolume master_output
aftrs-cli tool resolume autopilot --enable true
```

## Reporting Results

When testing is complete, note:

1. **What worked**: List successful commands/tools
2. **What failed**: List any errors or connection issues
3. **Performance**: Note response times and reliability
4. **Use cases**: Which tools would be most useful for your workflow

Create a new file with results:
```bash
touch ~/aftrs-studio/hg-mcp/TEST-RESULTS.md
```

## Next Steps After Successful Testing

Once both TouchDesigner and Resolume tools are working:

1. **Create Custom Workflows**: Combine tools into sequences
2. **Set Up Automated Checks**: Use health monitoring tools
3. **Integrate with Discord**: Get notifications for errors
4. **Build Show Control**: Create startup/shutdown sequences
5. **Session Logging**: Document your sessions to the vault

## Quick Reference

| Application | Port | Config Location |
|-------------|------|----------------|
| TouchDesigner | 8090 | Preferences → Network → Python Server |
| Resolume | 7000 | Preferences → OSC → OSC Output |
| MCP Server | stdio | ~/.config/claude/mcp_settings.json |

**Test Projects:**
- TD: `~/aftrs-studio/visual-projects/touchdesigner/projects/`
- Resolume: `~/aftrs-studio/visual-projects/resolume/compositions/`

**Logs:**
```bash
aftrs-logs
```
