@echo off
setlocal enabledelayedexpansion

rem Determine repo root relative to this script
set SCRIPT_DIR=%~dp0
for %%I in ("%SCRIPT_DIR%\..") do set REPO_ROOT=%%~fI
set SRC=%REPO_ROOT%\diagrams
set OUT=%SRC%\_rendered

if not exist "%OUT%" mkdir "%OUT%"

rem Prefer Docker mermaid-cli
where docker >nul 2>&1
if %ERRORLEVEL%==0 (
    echo Using Docker (minlag/mermaid-cli) to render Mermaid diagrams...
    for %%F in ("%SRC%\*.mmd") do (
        echo Rendering %%~nxF ...
        docker run --rm -v "%SRC%:/work" minlag/mermaid-cli -i "%%~nxF" -o "_rendered/%%~nF.svg"
        if errorlevel 1 (
            echo Failed to render %%~nxF with Docker.
            exit /b 1
        )
    )
    echo Render complete. SVGs at %OUT%
    exit /b 0
)

rem Fall back to node mmdc if available
where mmdc >nul 2>&1
if %ERRORLEVEL%==0 (
    echo Using local mermaid-cli (mmdc) to render Mermaid diagrams...
    for %%F in ("%SRC%\*.mmd") do (
        echo Rendering %%~nxF ...
        mmdc -i "%%~fF" -o "%OUT%\%%~nF.svg"
        if errorlevel 1 (
            echo Failed to render %%~nxF with mmdc.
            exit /b 1
        )
    )
    echo Render complete. SVGs at %OUT%
    exit /b 0
)

echo:
echo No renderer found.
echo Install one of the following and re-run this script:
echo   1. Docker Desktop (recommended), then run scripts\render_diagrams.bat
echo   2. Node mermaid-cli: npm install -g @mermaid-js/mermaid-cli
exit /b 2
