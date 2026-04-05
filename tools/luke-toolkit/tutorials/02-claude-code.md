# Claude Code Introduction

Claude Code is an AI assistant that runs in your terminal. It can read files, write code, run commands, and help you build things.

## Starting Claude Code

```bash
# Open in any project folder
cd ~/aftrs-studio/luke-toolkit
claude
```

You'll see a prompt. Just type what you want to do in plain English.

## What Claude Code Can Do

### Read and Understand Code
```
"What does this script do?"
"Explain the video processing pipeline"
"Find where the NDI source is configured"
```

### Write Code
```
"Create a script that converts all .mov files to .mp4"
"Add a function to analyze BPM"
"Fix the error in setup.py"
```

### Run Commands
```
"Run the tests"
"Check if FFmpeg is installed"
"List all TouchDesigner projects"
```

### Edit Files
```
"Update the config to use port 8080"
"Add error handling to process_video()"
"Rename the function from 'foo' to 'processTrack'"
```

## Useful Commands

| Command | What it does |
|---------|--------------|
| `/help` | Show help |
| `/clear` | Clear conversation |
| `/compact` | Summarize and continue |
| `/cost` | Show token usage |

## Tips for Good Results

### Be Specific
Instead of:
> "Fix the video thing"

Try:
> "The video export is failing with an error. Check export_video.py and fix it"

### Give Context
> "I'm working on a DJ set for tonight. Create a script that reads the track list from tracks.txt and calculates total runtime"

### Ask to Explain
> "Explain what this does before running it"

## Example Session

```
You: Create a simple script that lists all video files in the current folder

Claude: I'll create a script for you...
[Creates list_videos.py]

You: Add the file size next to each filename

Claude: I'll update the script...
[Edits list_videos.py]

You: Run it

Claude: Running the script...
[Shows output]
```

## What Claude Code Has Access To

- **Your files** - Can read, write, and edit
- **Terminal** - Can run commands
- **MCP Tools** - Has access to aftrs-mcp tools (more on this next)
- **Web** - Can search and fetch documentation

## Safety

- Claude asks before making big changes
- You can always say "don't do that"
- Changes are tracked in git (undo with `git checkout`)

## Next Steps

- [AFTRS MCP Tools](03-aftrs-mcp.md) - Studio automation via Claude
