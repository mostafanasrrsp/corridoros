@echo off
title Corridor OS Launcher
echo.
echo ================================================
echo   Corridor OS - Hybrid Quantum-Photonic OS
echo ================================================
echo.
echo Starting Corridor OS...
echo.

REM Try different browsers
if exist "C:\Program Files\Google\Chrome\Application\chrome.exe" (
    start "" "C:\Program Files\Google\Chrome\Application\chrome.exe" --new-window "%~dp0START-CORRIDOR-OS.html"
) else if exist "C:\Program Files (x86)\Google\Chrome\Application\chrome.exe" (
    start "" "C:\Program Files (x86)\Google\Chrome\Application\chrome.exe" --new-window "%~dp0START-CORRIDOR-OS.html"
) else if exist "C:\Program Files\Mozilla Firefox\firefox.exe" (
    start "" "C:\Program Files\Mozilla Firefox\firefox.exe" -new-window "%~dp0START-CORRIDOR-OS.html"
) else (
    start "" "%~dp0START-CORRIDOR-OS.html"
)

echo Corridor OS should now be loading...
pause
