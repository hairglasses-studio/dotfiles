# bench_cli

*Project plan consolidated from standalone repository*

---

# bench_cli
a wrapper CLI to check LLM performance leaderboards, bulk manage models, and optimize models for specific GPUs 

## Usage (via agentctl)

```
# Refresh leaderboards, pull best-fit models, run benchmark
agentctl bench run

# View stored results
agentctl bench results

# Serve Prometheus metrics on :9337 (scraped by llm-viewer)
agentctl bench metrics

# Generate PNG bar-chart + embed in README
agentctl bench chart
```

Leaderboards downloaded:

* BigCode Models Leaderboard
* LLM-Perf Leaderboard
* Open-LLM Leaderboard
* ToolBench Leaderboard

Results are cached in `~/.agentctl_cache/benchmarks.db`. A Prometheus exporter (`bench-metrics` service in `docker-compose.yml`) exposes metrics, and `llm-viewer` plots them automatically.

Web dashboards are reachable via Traefik path prefixes (default proxy on :8088). Example:

```
http://localhost:8088/viewer      # LLM-Viewer
http://localhost:8088/prometheus  # Prometheus metrics
```

For knowledge-base sync with Open-WebUI:

```
agentctl openwebui kb-sync ./docs
```

---

*Repository archived on 2025-09-23 - originally located at `bench_cli/`*
