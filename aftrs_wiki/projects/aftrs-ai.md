# aftrs-ai Org Breakdown (2025-09-14)

## Overview
The aftrs-ai org hosts self-hosted AI/LLM infrastructure centered on Ollama for the Cline provider in VS Code. It includes scripts, model data, and environment setup for GPU-accelerated local inference.

## Inventory
See `aftrs-ai-inventory.tsv` in this folder for a machine-readable list of repos with status/metadata.

## Key Projects
- ollama: Dockerized Ollama server with GPU support (CUDA / NVIDIA Container Toolkit). Exposes API on 11434 for VS Code Cline provider.
- ollama_data: Persistent model storage and manifests; includes quantized models and custom configurations.

## Scripts & Ops
- start-ollama.sh: Start service with proper GPU access; set model cache and persistence.
- manage-models.sh: Install, list, prune models; handle large model downloads and quantization.
- health checks: curl API probe; nvidia-smi monitoring; container logs.

## VS Code Integration
- Configure Cline to point to Ollama base URL; set model per-task (e.g., qwen3-coder:32b, wizardcoder:33b).
- Provide project-level .vscode/settings.json snippets for quick adoption.

## Troubleshooting
- Container cannot see GPU: verify nvidia-container-toolkit and --gpus all.
- API timeout: check port mapping and container logs.
- Model not found: ensure download completed; check disk quota.

## Next Steps
- Add model inventory markdown with sizes, quantization levels, and recommended use-cases.
- Add backup/restore guidance for ollama_data.
- Create minimal examples for Cline tasks (codegen, review, refactor) with model suggestions.
