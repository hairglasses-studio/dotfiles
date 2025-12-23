@echo off
setlocal
if exist "%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe" (
  powershell -ExecutionPolicy Bypass -File ".\scripts\bootstrap.ps1"
  exit /b %errorlevel%
) else (
  echo PowerShell not found. Please run scripts\bootstrap.ps1 manually or install PowerShell.
  exit /b 1
)
