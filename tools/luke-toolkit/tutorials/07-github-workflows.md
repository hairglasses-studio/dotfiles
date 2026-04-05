# GitHub Actions & CI/CD

Automate testing, building, and deployment with GitHub Actions.

## What is CI/CD?

- **CI** (Continuous Integration): Automatically test code when pushed
- **CD** (Continuous Deployment): Automatically deploy when tests pass

---

## Basic Workflow Structure

Workflows live in `.github/workflows/`:

```
my-project/
├── .github/
│   └── workflows/
│       ├── test.yml       # Run tests
│       ├── lint.yml       # Check code style
│       └── deploy.yml     # Deploy to production
└── src/
```

---

## Your First Workflow

Create `.github/workflows/test.yml`:

```yaml
name: Tests

# When to run
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

# What to do
jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      # Get the code
      - uses: actions/checkout@v4

      # Set up language
      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'

      # Install dependencies
      - name: Install dependencies
        run: |
          pip install -r requirements.txt

      # Run tests
      - name: Run tests
        run: |
          pytest tests/
```

---

## Common Triggers

### On Push

```yaml
on:
  push:
    branches: [main, develop]
```

### On Pull Request

```yaml
on:
  pull_request:
    branches: [main]
```

### On Schedule (Cron)

```yaml
on:
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight
```

### Manual Trigger

```yaml
on:
  workflow_dispatch:
```

---

## Language-Specific Examples

### Python

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - run: pip install -r requirements.txt
      - run: pytest
```

### Node.js

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: npm install
      - run: npm test
```

### Go

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test ./...
```

---

## Linting Workflow

Create `.github/workflows/lint.yml`:

```yaml
name: Lint

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'

      - name: Install linters
        run: |
          pip install flake8 black isort

      - name: Check formatting (black)
        run: black --check .

      - name: Check imports (isort)
        run: isort --check-only .

      - name: Lint (flake8)
        run: flake8 .
```

---

## Using Secrets

### Adding Secrets

1. Go to your repo on GitHub
2. Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Add name and value

### Using in Workflow

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy
        env:
          API_KEY: ${{ secrets.API_KEY }}
          SERVER_PASSWORD: ${{ secrets.SERVER_PASSWORD }}
        run: |
          ./deploy.sh
```

---

## Build and Release

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'  # Trigger on version tags

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build
        run: |
          make build

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            dist/*
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## Matrix Builds

Test across multiple versions:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: ['3.9', '3.10', '3.11']

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: ${{ matrix.python-version }}
      - run: pytest
```

---

## Caching Dependencies

Speed up workflows:

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Cache pip packages
        uses: actions/cache@v3
        with:
          path: ~/.cache/pip
          key: ${{ runner.os }}-pip-${{ hashFiles('requirements.txt') }}

      - run: pip install -r requirements.txt
```

---

## Complete Example

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - run: pip install flake8 black
      - run: black --check .
      - run: flake8 .

  test:
    runs-on: ubuntu-latest
    needs: lint  # Run after lint passes
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - run: pip install -r requirements.txt
      - run: pytest --cov=src tests/

  deploy:
    runs-on: ubuntu-latest
    needs: test  # Run after test passes
    if: github.ref == 'refs/heads/main'  # Only on main branch
    steps:
      - uses: actions/checkout@v4
      - name: Deploy
        env:
          DEPLOY_KEY: ${{ secrets.DEPLOY_KEY }}
        run: ./deploy.sh
```

---

## Debugging Workflows

### View Logs

1. Go to repo → Actions tab
2. Click on workflow run
3. Click on job
4. Expand steps to see output

### Add Debug Output

```yaml
- name: Debug info
  run: |
    echo "Branch: ${{ github.ref }}"
    echo "Event: ${{ github.event_name }}"
    pwd
    ls -la
```

### Re-run Failed Jobs

1. Go to failed workflow
2. Click "Re-run jobs"

---

## Quick Reference

| Task | Syntax |
|------|--------|
| Checkout code | `uses: actions/checkout@v4` |
| Setup Python | `uses: actions/setup-python@v5` |
| Setup Node | `uses: actions/setup-node@v4` |
| Setup Go | `uses: actions/setup-go@v5` |
| Cache | `uses: actions/cache@v3` |
| Run command | `run: command here` |
| Use secret | `${{ secrets.NAME }}` |
| Conditional | `if: github.ref == 'refs/heads/main'` |

---

## Next Steps

- [Discord Bots](08-discord-bots.md) - Build and deploy bots
- [Project Templates](11-project-templates.md) - Standard project setup
