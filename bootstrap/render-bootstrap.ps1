$ErrorActionPreference = "Stop"

$requiredVars = @(
  "BACKEND_IMAGE",
  "FRONTEND_IMAGE",
  "DATABASE_URL",
  "JWT_SECRET",
  "DEFAULT_DOMAIN",
  "DEFAULT_ADMIN_EMAIL",
  "DEFAULT_ADMIN_PASSWORD",
  "BACKEND_HOST",
  "FRONTEND_HOST",
  "CORS_ALLOWED",
  "LOAD_BALANCER_IP",
  "SERVER_IP"
)

foreach ($name in $requiredVars) {
  if ([string]::IsNullOrWhiteSpace([Environment]::GetEnvironmentVariable($name))) {
    throw "Missing required environment variable: $name"
  }
}

if ([string]::IsNullOrWhiteSpace($env:DEFAULT_ADMIN_USERNAME)) {
  $env:DEFAULT_ADMIN_USERNAME = "admin"
}
if ([string]::IsNullOrWhiteSpace($env:DEFAULT_ADMIN_NAME)) {
  $env:DEFAULT_ADMIN_NAME = "Default Admin"
}
if ([string]::IsNullOrWhiteSpace($env:TCP_PROXY_HOST)) {
  $env:TCP_PROXY_HOST = "proxy.$env:DEFAULT_DOMAIN"
}
if ([string]::IsNullOrWhiteSpace($env:TCP_PROXY_PORT_START)) {
  $env:TCP_PROXY_PORT_START = "24000"
}
if ([string]::IsNullOrWhiteSpace($env:TCP_PROXY_PORT_END)) {
  $env:TCP_PROXY_PORT_END = "24999"
}

$templatePath = Join-Path $PSScriptRoot "kubesa-bootstrap.yaml.tpl"
$rendered = Get-Content $templatePath -Raw

$matches = [regex]::Matches($rendered, "__([A-Z0-9_]+)__")
if ($matches.Count -gt 0) {
  $placeholders = $matches | ForEach-Object { $_.Groups[1].Value } | Sort-Object -Unique
  $missing = @()

  foreach ($name in $placeholders) {
    $value = [Environment]::GetEnvironmentVariable($name)
    if ($null -eq $value) {
      $missing += $name
      continue
    }

    $escaped = $value.Replace("\", "\\").Replace('"', '\"').Replace("`n", "\n")
    $rendered = $rendered.Replace("__$name`__", $escaped)
  }

  if ($missing.Count -gt 0) {
    throw "Unresolved template placeholders: $($missing -join ', ')"
  }
}

$rendered
