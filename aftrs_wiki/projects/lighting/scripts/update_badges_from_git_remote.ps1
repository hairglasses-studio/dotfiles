$url = git config --get remote.origin.url
if (-not $url) {
  Write-Error "No remote.origin.url found; cannot update badges."
  exit 1
}
if ($url.StartsWith("git@github.com:")) {
  $slug = $url.Substring(15)
} elseif ($url.StartsWith("https://github.com/")) {
  $slug = $url.Substring(19)
} else {
  Write-Error "Unsupported remote URL: $url"
  exit 2
}
$slug = $slug -replace '\.git$',''
$parts = $slug.Split('/')
$owner = $parts[0]
$repo  = $parts[1]
(Get-Content README.md) -replace 'github.com/OWNER/REPO', "github.com/$owner/$repo" | Set-Content README.md
Write-Host "Updated badges to github.com/$owner/$repo"
