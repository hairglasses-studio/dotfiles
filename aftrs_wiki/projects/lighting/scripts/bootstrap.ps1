$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $repoRoot

Write-Host "==> Enabling Git hooks (.githooks/)"
./scripts/setup_hooks.ps1

Write-Host "==> Updating README badges from git remote (if any)"
try {
  ./scripts/update_badges_from_git_remote.ps1
  Write-Host "   - Badges patched"
} catch {
  Write-Host "   - Skipped (no remote or unsupported URL)"
}

Write-Host "==> Rendering Mermaid diagrams"
$rendered = $false
if (Get-Command docker -ErrorAction SilentlyContinue) {
  & docker --version | Out-Null
  & scripts/render_diagrams.bat
  $rendered = $true
} elseif (Get-Command mmdc -ErrorAction SilentlyContinue) {
  if (!(Test-Path diagrams/_rendered)) { New-Item -ItemType Directory -Force diagrams/_rendered | Out-Null }
  Get-ChildItem diagrams -Filter *.mmd | ForEach-Object {
    $base = [System.IO.Path]::GetFileNameWithoutExtension($_.Name)
    mmdc -i $_.FullName -o ("diagrams/_rendered/{0}.svg" -f $base)
  }
  $rendered = $true
} else {
  Write-Host "   - No Docker or mermaid-cli found."
  Write-Host "   - Install Docker Desktop or run: npm install -g @mermaid-js/mermaid-cli"
}

if (-not $rendered) { exit 2 }
Write-Host "   - Diagrams rendered to diagrams/_rendered"

Write-Host "==> Bootstrap complete."
