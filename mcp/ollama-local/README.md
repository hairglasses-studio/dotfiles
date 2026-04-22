# ollama-local MCP Server

Routes coding tasks to local models on this workstation instead of burning Claude tokens. Backed by ollama running at `http://localhost:11434`.

## Hardware target

- NVIDIA RTX 3090 FE 24GB (Ampere, SM 8.6, 936 GB/s VRAM)
- AMD Ryzen 9 7950X + 96GB DDR5-6000 CL30 (~88 GB/s)
- ollama systemd service with flash-attention + Q8 KV cache enabled

## Tools

| Tool | Backing model | Speed | Use for |
|------|--------------|-------|---------|
| `local_code_fast` | DeepSeek-Coder-V2 16B Lite (MoE, 2.4B active) | ~228 tok/s | Boilerplate, commit messages, ROADMAP writeback, quick snippets |
| `local_code_generate` | Qwen3-Coder 30B-A3B (MoE, 3B active) | ~137 tok/s | Default coding; agentic tool use; long context (256K) |
| `local_code_review` | DeepSeek-R1-Distill-Qwen-32B (32B dense, think=on) | ~10 tok/s | Hard bugs, security review, deep reasoning with chain-of-thought |
| `local_code_heavy` | Qwen3-Coder-Next 51GB (MoE, partial offload) | ~11 tok/s | Max quality multi-file gen; when 30B-A3B isn't enough |
| `local_list_models` | — | instant | Availability check / inventory |

Every tool returns `{model, response, thinking, tokens, tok_per_sec, total_seconds}`.

## Quality benchmarks (heapq implementation, 4/4 rubric)

| Model | Quality | Notes |
|-------|--------:|-------|
| DeepSeek-Coder-V2 16B Lite | 4/4 | Sweet spot of speed + correctness |
| Qwen3-Coder 30B-A3B | 4/4 | Clean types, proper edge cases |
| DeepSeek-R1-Distill-Qwen-32B | 4/4 | Think=on needed for reasoning benefit |
| Qwen3-Coder-Next 51GB | 4/4 | Most thorough (Optional types, 1-indexed docs) |

## Integration with `/fleet-autopilot`

Wired in via `~/.claude/skills/fleet-autopilot/SKILL.md` (sync'd to `.agents/` + `.codex/` mirrors). Offload points:

| Phase | Step | Tool | Savings |
|-------|------|------|--------:|
| 0 | Triage parse/score | `local_code_fast` | ~5K tokens |
| 1 | Trivial impl (S-size) | `local_code_generate` | ~10K tokens |
| 2 | Diff review checklist | `local_code_review` | ~8K tokens |
| 3 | Commit message | `local_code_fast` | ~2K tokens |
| 3 | ROADMAP writeback | `local_code_fast` | ~3K tokens |

Fully offloaded iteration saves ~28K Claude tokens (~$0.21). See SKILL.md "Local Model Offload" section for exact flow.

## Files

- `server.py` — FastMCP Python server (one file, ~100 lines)
- `run-ollama-local-mcp.sh` — launcher (stdio transport)
- `.venv/` — isolated Python 3.12 env (fastmcp + httpx)

## Registration

Registered in `~/.claude.json` under the `/home/hg/hairglasses-studio` workspace as `ollama-local`. Exposes tools as `mcp__ollama-local__local_*`. Reconnect MCP (`/reconnect`) after config changes.

## Ollama model inventory (required for all tools to work)

```bash
ollama pull deepseek-coder-v2:16b-lite-instruct-q4_0   # fast
ollama pull qwen3-coder:30b-a3b-q4_K_M                  # generate
ollama pull deepseek-r1:32b-qwen-distill-q4_K_M         # review
ollama pull qwen3-coder-next                            # heavy (51GB, already present as code-long alias)
```

## Known limitations

- Qwen3 base models (not coder) use a separate `thinking` field — the server passes `think=false` for code tools so the full answer appears in `response`. For `local_code_review`, `think=true` is set intentionally to surface reasoning.
- Single-GPU, no concurrent serving — ollama loads one model at a time. Switching models costs ~2-5s reload. For concurrent serving, see the vLLM systemd unit in the overclock plan.
- FP8 unsupported on Ampere → all models run at GGUF Q4_K_M.

## Testing

```bash
# Smoke test (without reconnecting MCP)
cd /home/hg/hairglasses-studio/dotfiles/mcp/ollama-local
./run-ollama-local-mcp.sh <<< '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}'

# Actual tool call (after MCP reconnect in Claude Code)
"Use local_code_fast to write a one-liner that returns 1..5 product"
```
