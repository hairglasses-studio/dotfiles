# Luke's Toolkit

Personal toolkit for studio work - DJ/music, video processing, TouchDesigner, and automation.

---

## 📚 Tutorials

### Getting Started
| # | Tutorial | Description |
|---|----------|-------------|
| 00 | [Terminal Basics](tutorials/00-terminal-basics.md) | iTerm2, shell commands, Homebrew |
| 01 | [GitHub Basics](tutorials/01-github-basics.md) | Clone repos, make changes, push updates |
| 02 | [Claude Code Introduction](tutorials/02-claude-code.md) | AI-powered coding assistant |
| 03 | [AFTRS MCP Tools](tutorials/03-aftrs-mcp.md) | Studio automation via AI |

### Intermediate
| # | Tutorial | Description |
|---|----------|-------------|
| 04 | [Advanced Claude Code](tutorials/04-claude-code-advanced.md) | Multi-file operations, search, debugging |
| 05 | [Understanding MCP Servers](tutorials/05-mcp-servers.md) | How AI tools work |
| 06 | [Building MCP Modules](tutorials/06-building-mcp-modules.md) | Create your own AI tools |
| 07 | [GitHub Workflows](tutorials/07-github-workflows.md) | CI/CD and automation |

### Creative Tools
| # | Tutorial | Description |
|---|----------|-------------|
| 08 | [Discord Bots](tutorials/08-discord-bots.md) | Build and deploy bots |
| 09 | [TouchDesigner Development](tutorials/09-touchdesigner-plugins.md) | Python scripting in TD |
| 10 | [Resolume & FFGL](tutorials/10-resolume-ffgl.md) | VJ software and plugins |

### Project Management
| # | Tutorial | Description |
|---|----------|-------------|
| 11 | [Project Templates](tutorials/11-project-templates.md) | Structure and best practices |
| 12 | [Contributing with Claude](tutorials/12-contributing-with-claude.md) | Build your own projects |
| 13 | [YouTube Resources](tutorials/13-youtube-resources.md) | Curated video tutorials |
| 14 | [Committing Existing Projects](tutorials/14-committing-existing-projects.md) | Upload work to GitHub |

### Collaboration & A/V
| # | Tutorial | Description |
|---|----------|-------------|
| 15 | [Collaborative Git Workflow](tutorials/15-collaborative-git-workflow.md) | Working together on aftrs-studio projects |
| 16 | [OSC & MIDI Communication](tutorials/16-osc-midi-communication.md) | Connect TD, Resolume, Ableton |
| 17 | [Real-time A/V Sync](tutorials/17-realtime-av-sync.md) | Sync visuals to audio |

---

## 🚀 Quick Start

**Brand new to terminal/coding?** Start here:

1. **[Terminal Basics](tutorials/00-terminal-basics.md)** - Install iTerm2, learn shell commands
2. **[GitHub Basics](tutorials/01-github-basics.md)** - Set up Git, clone this repo
3. **[Claude Code Introduction](tutorials/02-claude-code.md)** - Your AI coding assistant

**Ready to collaborate?** Essential reading:
- **[Collaborative Git Workflow](tutorials/15-collaborative-git-workflow.md)** - Branching, PRs, code review

**Building audiovisual projects?**
- **[OSC & MIDI Communication](tutorials/16-osc-midi-communication.md)** - Connect your apps
- **[Real-time A/V Sync](tutorials/17-realtime-av-sync.md)** - Sync visuals to audio

**Have existing TD plugins?** Jump to:
- **[Committing Existing Projects](tutorials/14-committing-existing-projects.md)**

---

## 📁 What's Here

```
luke-toolkit/
├── tutorials/          # Step-by-step guides (18 tutorials)
├── tools/
│   ├── music/          # DJ and music production
│   ├── video/          # Video processing pipelines
│   ├── touchdesigner/  # TD components and projects
│   └── automation/     # Studio automation scripts
├── scripts/            # Utility scripts
├── configs/            # OSC/MIDI configuration templates
│   ├── osc-defaults.yaml
│   └── midi-defaults.yaml
└── docs/               # Reference documentation
```

---

## 🛠 Tools Overview

### Music/DJ
- Setlist management
- BPM analysis
- Track organization

### Video Processing
- AI-powered processing via video-ai-toolkit
- Batch conversion scripts
- Export presets

### TouchDesigner
- Reusable components
- Project templates
- Performance monitoring

### Studio Automation
- Session logging
- Equipment control
- NDI/streaming helpers

---

## 🤝 Collaboration

Working on aftrs-studio projects together:

1. **Read [CONTRIBUTING.md](CONTRIBUTING.md)** - Guidelines and conventions
2. **Use feature branches** - `luke/feature-name`
3. **Create pull requests** - For code review before merging
4. **Follow commit conventions** - `feat:`, `fix:`, `docs:`

```bash
# Quick workflow
git checkout -b luke/new-feature
# ... make changes ...
git add . && git commit -m "feat: add new feature"
git push -u origin luke/new-feature
gh pr create
```

---

## 💡 Getting Help

Ask Claude Code! Just describe what you need:

```
"How do I send OSC from TouchDesigner to Resolume?"
"Set up beat detection for audio-reactive visuals"
"Create a pull request for my changes"
"Help me sync my visuals to Ableton's tempo"
```

Or use the tutorials above to learn at your own pace.

---

## 🎓 Learning Path

### Week 1: Foundations
- Complete tutorials 00-02 (Terminal, Git, Claude Code)
- Install iTerm2, Homebrew, and essential tools
- Clone this repo to your Mac

### Week 2: Git & Collaboration
- Complete tutorials 03, 15 (AFTRS MCP, Collaborative Workflow)
- Practice git workflow daily
- Create your first pull request

### Week 3-4: Building Skills
- Complete tutorials 04-07
- Practice with small projects
- Commit your work regularly

### Week 5+: Audiovisual Projects
- Complete tutorials 16-17 (OSC/MIDI, A/V Sync)
- Build audio-reactive visuals
- Collaborate on aftrs-studio projects

---

## 🔗 Related Repositories

Other aftrs-studio projects you'll work with:

- **aftrs-mcp** - MCP tools for studio automation
- **video-ai-toolkit** - AI video processing
- Your TD plugins repo (coming soon!)
