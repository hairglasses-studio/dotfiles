#!/usr/bin/env python3
"""Refresh the cached workspace health matrix from workspace/manifest.json."""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


@dataclass
class CommandSpec:
    argv: list[str]
    display: str
    mode: str
    env: dict[str, str] | None = None


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--local-dir", required=True)
    parser.add_argument("--repo", action="append", dest="repos", default=[])
    return parser.parse_args()


def load_manifest(local_dir: Path) -> list[dict[str, Any]]:
    manifest_path = local_dir / "workspace" / "manifest.json"
    data = json.loads(manifest_path.read_text(encoding="utf-8"))
    return data.get("repos", [])


def run_cmd(argv: list[str], cwd: Path | None = None, env: dict[str, str] | None = None) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        argv,
        cwd=str(cwd) if cwd else None,
        env=env,
        text=True,
        capture_output=True,
        timeout=1800,
        check=False,
    )


def git_output(repo_path: Path, *args: str) -> str:
    proc = run_cmd(["git", *args], cwd=repo_path)
    return proc.stdout.strip() if proc.returncode == 0 else ""


def repo_dirty(repo_path: Path) -> bool:
    proc = run_cmd(["git", "status", "--porcelain"], cwd=repo_path)
    if proc.returncode != 0:
        return False
    return bool(proc.stdout.strip())


def workflow_status(repo_path: Path, policy: str) -> str:
    if policy == "retired":
        return "retired"
    workflows = repo_path / ".github" / "workflows"
    if workflows.is_dir():
        return "present"
    if policy:
        return "missing"
    return ""


def mode_for_profile(profile: str) -> str:
    if profile == "go_test":
        return "standalone"
    if profile == "informational" or not profile:
        return "informational"
    return "repo_default"


def make_check_spec(repo_path: Path) -> CommandSpec | None:
    env = {"GOFLAGS": "-buildvcs=false", "GOWORK": "off"}
    for target in ("check", "pipeline-check"):
        probe_env = os.environ.copy()
        probe_env.update(env)
        probe = run_cmd(["make", "-n", target], cwd=repo_path, env=probe_env)
        if probe.returncode == 0:
            return CommandSpec(
                ["make", target],
                f"GOWORK=off GOFLAGS=-buildvcs=false make {target}",
                "repo_default",
                env,
            )
    return None


def command_for_profile(repo_path: Path, profile: str) -> CommandSpec | None:
    if profile == "make_check":
        return make_check_spec(repo_path)
    if profile == "go_test":
        return CommandSpec(
            ["go", "test", "-buildvcs=false", "./..."],
            "GOWORK=off go test -buildvcs=false ./...",
            "standalone",
            {"GOWORK": "off"},
        )
    if profile == "npm_test":
        return CommandSpec(["npm", "test"], "npm test", "repo_default")
    if profile == "python_pytest":
        return CommandSpec(["pytest"], "pytest", "repo_default")
    return None


def classify_failure(stderr: str, stdout: str) -> str:
    text = f"{stdout}\n{stderr}".lower()
    for marker in (
        "no rule to make target",
        "command not found",
        "executable file not found",
        "permission denied",
        "not a git repository",
        "must be run in a work tree",
        "error obtaining vcs status",
    ):
        if marker in text:
            return "runner_env"
    return "code_failure"


def snapshot_for_repo(local_dir: Path, policy: dict[str, Any]) -> dict[str, Any]:
    name = policy["name"]
    repo_path = local_dir / name
    profile = (policy.get("baseline_profile") or "").strip()
    workflow_policy = (policy.get("workflow_policy") or "").strip()
    workflow_family = (policy.get("workflow_family") or "").strip()
    checked_at = datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")

    row: dict[str, Any] = {
        "repo": name,
        "baseline_profile": profile,
        "workflow_policy": workflow_policy,
        "workflow_family": workflow_family,
        "workflow_status": "",
        "current_branch": "",
        "baseline_mode": mode_for_profile(profile),
        "local_baseline_status": "informational" if not profile or profile == "informational" else "unknown",
        "failure_class": "",
        "baseline_checked_at": checked_at,
        "baseline_commit": "",
        "baseline_command": "",
    }

    if not repo_path.exists() or not (repo_path / ".git").exists():
        row["local_baseline_status"] = "missing"
        row["failure_class"] = "runner_env"
        return row

    row["workflow_status"] = workflow_status(repo_path, workflow_policy)
    row["current_branch"] = git_output(repo_path, "rev-parse", "--abbrev-ref", "HEAD")
    row["baseline_commit"] = git_output(repo_path, "rev-parse", "HEAD")

    if row["local_baseline_status"] == "informational":
        return row
    if row["current_branch"] and row["current_branch"] != "main":
        row["local_baseline_status"] = "not_main"
        return row
    if repo_dirty(repo_path):
        row["local_baseline_status"] = "dirty"
        return row

    spec = command_for_profile(repo_path, profile)
    if spec is None:
        row["local_baseline_status"] = "informational"
        row["baseline_mode"] = "informational"
        return row

    row["baseline_mode"] = spec.mode
    row["baseline_command"] = spec.display

    env = os.environ.copy()
    if spec.env:
        env.update(spec.env)

    proc = run_cmd(spec.argv, cwd=repo_path, env=env)
    if proc.returncode == 0:
        row["local_baseline_status"] = "pass"
        return row

    row["local_baseline_status"] = "fail"
    row["failure_class"] = classify_failure(proc.stderr, proc.stdout)
    return row


def main() -> int:
    args = parse_args()
    local_dir = Path(args.local_dir).resolve()
    repos_filter = set(args.repos)

    rows = []
    for policy in load_manifest(local_dir):
        if repos_filter and policy["name"] not in repos_filter:
            continue
        rows.append(snapshot_for_repo(local_dir, policy))

    rows.sort(key=lambda row: row["repo"])
    out = {
        "generated_at": datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z"),
        "local_dir": str(local_dir),
        "repos": rows,
    }

    out_path = local_dir / "docs" / "agent-parity" / "workspace-health-matrix.json"
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(out, indent=2) + "\n", encoding="utf-8")
    print(str(out_path))
    return 0


if __name__ == "__main__":
    sys.exit(main())
