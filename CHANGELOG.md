# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- CodeQL security scanning workflow
- CODEOWNERS, FUNDING.yml, issue/PR templates
- Bluetooth MAC address env var parameterization

### Changed
- Replaced hardcoded device identifiers with environment variables

## [1.0.0] - 2026-04-04

### Added
- 138+ GLSL terminal shaders with central TOML manifest (`shaders.toml`)
- Shader pipeline: preprocessor, compile tester, playlist engine, auto-rotate
- 5 procgen wallpaper shaders (cyber-rain, neon-grid, plasma-flow, fractal-pulse, particle-aurora)
- Compositor abstraction layer (`scripts/lib/compositor.sh`) supporting Hyprland and AeroSpace
- eww bar configuration for dual monitors (2560x1440 + 5120x1440 ultrawide)
- swaync notification center with Snazzy theme
- Ghostty terminal config with atomic-write update scripts
- Starship prompt with custom Snazzy palette
- Shell framework (`scripts/lib/hg-core.sh`) with consistent CLI utilities
- Atomic config write library (`scripts/lib/config.sh`) for safe concurrent updates
- Claude Code hooks: PostToolUse auto-reload, PreToolUse validation, prompt capture
- MCP server integration via dotfiles-mcp (90 tools across 15 modules)
- Desktop management tools: Hyprland IPC, Bluetooth, MIDI, input devices, shader pipeline
- GitHub org lifecycle tools: transfer, fork, clone, sync, audit
- Fleet management: health checks, dependency audits, workflow sync
- rEFInd bootloader with Matrix cyberpunk theme
- Plymouth boot splash (proxzima cyberpunk theme)
- metapac declarative package management (12 groups)
- Comprehensive install script with symlink management

### Infrastructure
- MIT License (2024-2026 hairglasses-studio)
- CI workflows: Go build/test, shell linting, CodeQL scanning
- Dependabot for Go modules and GitHub Actions
- .editorconfig for consistent formatting

[Unreleased]: https://github.com/hairglasses-studio/dotfiles/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/hairglasses-studio/dotfiles/releases/tag/v1.0.0
