# Studio Automation

Scripts and tools for automating studio operations.

## Session Management

### Start a Session

```bash
./session.sh start "Friday Live Stream"
```

This logs the start time and creates a session file.

### Log Events

```bash
./session.sh log "Started main visuals"
./session.sh log "Switched to camera 2"
```

### End Session

```bash
./session.sh end
```

Saves the complete log to the vault.

## Equipment Status

### Check All Systems

```bash
./status.sh all
```

Shows:
- TouchDesigner FPS
- NDI sources
- Stream health
- Recording status

### Individual Checks

```bash
./status.sh td      # TouchDesigner
./status.sh ndi     # NDI sources
./status.sh stream  # Stream status
```

## Using with Claude

```
"Start a new session called 'Friday Night'"
"Log that we're switching to the backup camera"
"What's the status of all systems?"
"End the session and save notes"
"Check if everything is ready for streaming"
```

## Scheduling

### Set Up Recurring Tasks

Edit `~/.luke-toolkit/schedule.yaml`:

```yaml
tasks:
  - name: "Pre-show check"
    time: "18:00"
    command: "./status.sh all"

  - name: "Start recording"
    time: "19:00"
    command: "./session.sh start 'Evening Stream'"
```

## Quick Reference

| Task | Command |
|------|---------|
| Start session | `./session.sh start "Name"` |
| Log event | `./session.sh log "What happened"` |
| End session | `./session.sh end` |
| Check status | `./status.sh all` |
| Check TD | `./status.sh td` |
| Check NDI | `./status.sh ndi` |
