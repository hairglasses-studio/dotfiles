$ErrorActionPreference = "Stop"
    $repoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
    Set-Location $repoRoot

    function Say-Ok($msg)   { Write-Host "[OK]   $msg" -ForegroundColor Green }
    function Say-Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
    function Say-Err($msg)  { Write-Host "[ERR]  $msg" -ForegroundColor Red }

    $warn = $false
    $err = $false

    Write-Host "==> Validating DMX patch"
    try {
      python3 tools/validate_patch.py | Write-Host
      if ($LASTEXITCODE -eq 0) { Say-Ok "DMX patch valid" } else { Say-Err "DMX patch validation failed"; $err = $true }
    } catch {
      Say-Warn "Python3 not found or script failed; skipping DMX validation"; $warn = $true
    }

    Write-Host "==> Checking Mermaid diagram freshness"
    try {
      & ./scripts/check_diagrams_fresh.sh | Write-Host
      switch ($LASTEXITCODE) {
        0 { Say-Ok "Mermaid outputs are fresh" }
        3 { Say-Warn "No Mermaid renderer (docker/podman/mmdc). Skipping freshness check"; $warn = $true }
        4 { Say-Err "Mermaid outputs are stale. Run: ./scripts/render_diagrams.sh; then commit updates."; $err = $true }
        default { Say-Err "Unexpected error from check_diagrams_fresh.sh (exit $LASTEXITCODE)"; $err = $true }
      }
    } catch {
      Say-Warn "Could not run freshness check script"; $warn = $true
    }

    Write-Host "==> Pinging Art-Net nodes (optional)"
    $cfg = "configs/artnet_plan.yaml"
    $ips = @()
    try {
      python3 - << 'PY' $cfg 2>$null | ForEach-Object { $ips += $_ }
import sys, yaml
p = sys.argv[1]
y = yaml.safe_load(open(p))
for _, node in (y.get("nodes", {}) or {}).items():
    ip = node.get("ip")
    if ip: print(ip)
PY
    } catch { }
    if ($ips.Count -eq 0 -and (Test-Path $cfg)) {
      $ips = Select-String -Path $cfg -Pattern 'ip:\s*([0-9.]+)' | ForEach-Object { $_.Matches[0].Groups[1].Value }
    }
    foreach ($ip in $ips) {
      if (-not $ip) { continue }
      if (Test-Connection -Quiet -Count 1 -TimeoutSeconds 1 $ip) {
        Say-Ok "Reachable: $ip"
      } else {
        Say-Warn "Unreachable: $ip"
      }
    }

    Write-Host "==> Summary"
    if ($err) { Write-Host "One or more errors detected."; exit 2 }
    elseif ($warn) { Write-Host "Completed with warnings."; exit 1 }
    else { Write-Host "All checks passed."; exit 0 }
