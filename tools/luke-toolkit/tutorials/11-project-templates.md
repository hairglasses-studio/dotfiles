# Project Templates

Best practices for starting new projects with clean, maintainable structure.

## Why Structure Matters

- Easier to navigate
- Easier for Claude to understand
- Easier for others to contribute
- Professional appearance

---

## Universal Project Structure

```
my-project/
├── README.md           # What is this? How to use it?
├── LICENSE             # How can others use it?
├── .gitignore          # What NOT to commit
├── src/                # Source code
├── tests/              # Test files
├── docs/               # Documentation
└── scripts/            # Utility scripts
```

---

## README Template

Every project needs a README:

```markdown
# Project Name

One-line description of what this does.

## Quick Start

\`\`\`bash
# Installation
brew install something

# Run it
./my-script.sh
\`\`\`

## Features

- Feature one
- Feature two
- Feature three

## Requirements

- macOS 12+
- Python 3.11+
- Node.js 20+

## Usage

\`\`\`bash
# Basic usage
my-command input.txt

# With options
my-command -o output.txt input.txt
\`\`\`

## Configuration

Copy `config.example.yaml` to `config.yaml` and edit:

\`\`\`yaml
setting: value
\`\`\`

## Contributing

1. Fork the repo
2. Create a branch: `git checkout -b my-feature`
3. Make changes and commit
4. Push and create a Pull Request

## License

MIT License - see LICENSE file
```

---

## .gitignore Templates

### Python

```gitignore
# Python
__pycache__/
*.py[cod]
*.egg-info/
dist/
build/
venv/
.env

# IDE
.vscode/
.idea/
*.swp

# OS
.DS_Store
Thumbs.db
```

### Node.js

```gitignore
# Node
node_modules/
npm-debug.log
yarn-error.log

# Build
dist/
build/

# Environment
.env
.env.local

# IDE
.vscode/
.idea/

# OS
.DS_Store
```

### Go

```gitignore
# Binary
/bin/
*.exe

# Vendor (optional)
/vendor/

# IDE
.vscode/
.idea/

# OS
.DS_Store
```

### TouchDesigner

```gitignore
# TD backups
*.toe.bak
*.toe.*.toe
Backup/

# Python cache
__pycache__/
*.pyc

# OS
.DS_Store
```

---

## License Files

### MIT (Most Permissive)

```
MIT License

Copyright (c) 2024 Your Name

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

### Apache 2.0 (Patent Protection)

Good for larger projects where patent protection matters.

### GPL 3.0 (Copyleft)

Requires derivative works to also be open source.

---

## Language-Specific Templates

### Python Project

```
my-python-project/
├── README.md
├── LICENSE
├── .gitignore
├── pyproject.toml       # Modern Python config
├── requirements.txt     # Dependencies
├── src/
│   └── my_project/
│       ├── __init__.py
│       └── main.py
├── tests/
│   └── test_main.py
└── scripts/
    └── setup.sh
```

**pyproject.toml**:

```toml
[project]
name = "my-project"
version = "0.1.0"
description = "What it does"
readme = "README.md"
requires-python = ">=3.11"
dependencies = [
    "requests>=2.28.0",
]

[project.scripts]
my-command = "my_project.main:main"
```

### Node.js Project

```
my-node-project/
├── README.md
├── LICENSE
├── .gitignore
├── package.json
├── src/
│   └── index.js
├── tests/
│   └── index.test.js
└── scripts/
    └── build.sh
```

**package.json**:

```json
{
  "name": "my-project",
  "version": "1.0.0",
  "description": "What it does",
  "main": "src/index.js",
  "scripts": {
    "start": "node src/index.js",
    "test": "jest",
    "lint": "eslint src/"
  },
  "dependencies": {},
  "devDependencies": {
    "jest": "^29.0.0",
    "eslint": "^8.0.0"
  }
}
```

### Go Project

```
my-go-project/
├── README.md
├── LICENSE
├── .gitignore
├── go.mod
├── go.sum
├── main.go
├── internal/
│   └── handler/
│       └── handler.go
├── pkg/
│   └── utils/
│       └── utils.go
└── cmd/
    └── my-tool/
        └── main.go
```

**go.mod**:

```go
module github.com/username/my-project

go 1.21

require (
    github.com/some/package v1.0.0
)
```

---

## Configuration Files

### YAML Config (Recommended)

```yaml
# config.yaml
app:
  name: "My App"
  debug: false

database:
  host: localhost
  port: 5432

logging:
  level: info
  file: app.log
```

### JSON Config

```json
{
  "app": {
    "name": "My App",
    "debug": false
  },
  "database": {
    "host": "localhost",
    "port": 5432
  }
}
```

### Environment Variables (.env)

```bash
# .env (NEVER commit this!)
DATABASE_URL=postgres://localhost:5432/mydb
API_KEY=your-secret-key
DEBUG=false
```

Create `.env.example` to show what's needed:

```bash
# .env.example (commit this)
DATABASE_URL=postgres://localhost:5432/mydb
API_KEY=your-api-key-here
DEBUG=false
```

---

## Documentation Structure

```
docs/
├── README.md           # Docs overview
├── getting-started.md  # Quick start guide
├── configuration.md    # Config options
├── api/                # API documentation
│   └── endpoints.md
└── examples/           # Usage examples
    └── basic.md
```

---

## Using Claude to Create Projects

```
"Create a new Python project called 'setlist-manager' with:
- CLI using argparse
- YAML configuration
- Basic tests with pytest
- GitHub Actions for testing"
```

Claude will generate:
- Proper directory structure
- All configuration files
- README with usage instructions
- GitHub workflow file
- Test scaffolding

---

## Quick Start Commands

### Python

```bash
mkdir my-project && cd my-project && git init && echo "# My Project" > README.md && echo "__pycache__/\n.env\n*.pyc" > .gitignore && mkdir src tests && touch src/__init__.py
```

### Node.js

```bash
mkdir my-project && cd my-project && npm init -y && git init && echo "node_modules/\n.env" > .gitignore && mkdir src tests
```

### Go

```bash
mkdir my-project && cd my-project && go mod init my-project && git init && echo "bin/\n*.exe" > .gitignore && mkdir internal pkg cmd
```

---

## Checklist for New Projects

- [ ] README.md with description and usage
- [ ] LICENSE file
- [ ] .gitignore appropriate for language
- [ ] Source code in src/ or appropriate directory
- [ ] Tests directory
- [ ] Configuration example file
- [ ] GitHub Actions workflow (optional)

---

## Next Steps

- [Contributing with Claude](12-contributing-with-claude.md) - Use Claude to build your projects
- [YouTube Resources](13-youtube-resources.md) - Video tutorials
