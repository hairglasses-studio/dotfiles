"""Local LLM MCP server — routes coding tasks to ollama models on this workstation.

Exposes 4 tools backed by benchmarked local models:
  - local_code_fast:     DeepSeek-Coder-V2 16B Lite (~228 tok/s, simple tasks)
  - local_code_generate: Qwen3-Coder 30B-A3B       (~137 tok/s, best overall, 256K ctx)
  - local_code_review:   DeepSeek-R1-Distill-32B   ( ~10 tok/s WITH thinking, hard bugs)
  - local_code_heavy:    Qwen3-Coder-Next 51GB     ( ~11 tok/s, max quality, partial offload)

All models run via the local ollama service at http://localhost:11434.
"""
from __future__ import annotations

import os

import httpx
from fastmcp import FastMCP

OLLAMA_URL = os.environ.get("OLLAMA_URL", "http://localhost:11434")

MODELS = {
    "fast":     "deepseek-coder-v2:16b-lite-instruct-q4_0",
    "generate": "qwen3-coder:30b-a3b-q4_K_M",
    "review":   "deepseek-r1:32b-qwen-distill-q4_K_M",
    "heavy":    "qwen3-coder-next:latest",
}

mcp = FastMCP("ollama-local")


def _generate(model: str, prompt: str, *, think: bool = False,
              num_predict: int = 2000, num_ctx: int = 16384,
              temperature: float = 0.2) -> dict:
    """Call ollama /api/generate and return a structured result."""
    payload = {
        "model": model,
        "prompt": prompt,
        "stream": False,
        "think": think,
        "options": {
            "temperature": temperature,
            "num_predict": num_predict,
            "num_ctx": num_ctx,
        },
    }
    with httpx.Client(timeout=600.0) as client:
        r = client.post(f"{OLLAMA_URL}/api/generate", json=payload)
        r.raise_for_status()
        data = r.json()
    eval_dur_s = max(data.get("eval_duration", 1) / 1e9, 1e-9)
    return {
        "model": model,
        "response": data.get("response", ""),
        "thinking": data.get("thinking", ""),
        "tokens": data.get("eval_count", 0),
        "tok_per_sec": round(data.get("eval_count", 0) / eval_dur_s, 1),
        "total_seconds": round(data.get("total_duration", 0) / 1e9, 2),
    }


@mcp.tool
def local_code_fast(prompt: str, num_predict: int = 1500) -> dict:
    """Generate code with the FASTEST local model (~228 tok/s).

    Use for: boilerplate, simple refactors, single-function generation,
    shell commands, quick snippets. Quality is solid (4/4 on standard
    bench) but smaller model means less nuance on complex tasks.

    Backing: DeepSeek-Coder-V2 16B Lite (MoE, 2.4B active, pure GPU).
    """
    return _generate(MODELS["fast"], prompt, num_predict=num_predict)


@mcp.tool
def local_code_generate(prompt: str, num_predict: int = 2000,
                        num_ctx: int = 32768) -> dict:
    """Generate code with the BEST overall local model (~137 tok/s).

    Use for: default coding tasks, agentic workflows, anything needing
    quality + speed + long context. 256K native context window available.

    Backing: Qwen3-Coder 30B-A3B (MoE, 3B active, pure GPU, Apache 2.0).
    """
    return _generate(MODELS["generate"], prompt,
                     num_predict=num_predict, num_ctx=num_ctx)


@mcp.tool
def local_code_review(prompt: str, num_predict: int = 4000) -> dict:
    """Review code with chain-of-thought reasoning (~10 tok/s, think=on).

    Use for: hard bugs, security review, algorithmic correctness, tricky
    edge-case analysis. SLOW but deep reasoning makes it worth it for
    problems where wrong answers are expensive.

    Backing: DeepSeek-R1-Distill-Qwen-32B (32B dense, thinking enabled).
    Includes separate `thinking` field in response with the reasoning trace.
    """
    return _generate(MODELS["review"], prompt, think=True,
                     num_predict=num_predict)


@mcp.tool
def local_code_heavy(prompt: str, num_predict: int = 3000) -> dict:
    """Max-quality generation with a larger local MoE (~11 tok/s).

    Use for: complex multi-file generation, architecture decisions, tasks
    where 30B-A3B quality is not enough. Partial CPU offload so slower
    than `local_code_generate` but higher quality ceiling.

    Backing: Qwen3-Coder-Next (80B total / 3B active, 51GB partial offload).
    """
    return _generate(MODELS["heavy"], prompt, num_predict=num_predict)


@mcp.tool
def local_list_models() -> dict:
    """List all ollama models available locally with sizes."""
    with httpx.Client(timeout=10.0) as client:
        r = client.get(f"{OLLAMA_URL}/api/tags")
        r.raise_for_status()
        data = r.json()
    return {
        "mcp_mapped": MODELS,
        "available": [
            {
                "name": m["name"],
                "size_gb": round(m.get("size", 0) / 1024**3, 1),
                "modified": m.get("modified_at", ""),
            }
            for m in data.get("models", [])
        ],
    }


if __name__ == "__main__":
    mcp.run()
