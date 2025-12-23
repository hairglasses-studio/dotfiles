# aftrs-ai Org Breakdown (2025-09-14)

## Overview
Self-hosted LLM model hub and automation scripts for Cline AI and local development. Focused on privacy, GPU acceleration, and developer productivity.

## Main Project: Ollama Model Hub
- **Purpose:** Local LLM hosting for Cline AI, VSCode, and other tools
- **Models:** qwen3-coder:30b, wizardcoder:33b, wizard-agent, llama2, and more
- **Scripts:**
  - `start-ollama.sh`: Start Ollama container
  - `manage-models.sh`: Interactive model management
  - `test-cline-connection.sh`, `test-cline-llama2.sh`, `test-qwen-cline.sh`: Connectivity and model tests
- **Data:**
  - `ollama_data/`: Model storage, keys, manifests
- **API:**
  - Accessible at `localhost:11434` (codegen, chat, review)
- **VSCode Integration:**
  - Cline provider setup, custom model config via settings.json
- **Troubleshooting:**
  - Container status, API health, GPU memory, model persistence
- **Performance:**
  - GPU monitoring, concurrent model usage, quantization strategies
- **Advanced:**
  - Custom model install, backup/restore, Python/cURL integration
- **Monitoring:**
  - Health checks, logs, GPU stats
- **Documentation:**
  - `README.md`, `CLINE_SETUP.md` (detailed config guide)

## Recommendations
- Ensure all model data is backed up before upgrades
- Document any custom model installations or changes to Docker config
- Use `manage-models.sh` for all model management to avoid manual errors
- Monitor GPU usage and container health for optimal performance

---
*Last updated: 2025-09-14*
*Prepared by: GitHub Copilot (automated agent)*
