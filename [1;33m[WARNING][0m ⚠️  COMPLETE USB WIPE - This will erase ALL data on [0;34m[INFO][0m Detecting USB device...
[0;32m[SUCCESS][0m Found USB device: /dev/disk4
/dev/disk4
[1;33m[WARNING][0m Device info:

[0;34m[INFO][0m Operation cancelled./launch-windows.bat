@echo off
title Corridor OS Launcher
echo.
echo ================================================
echo   Corridor OS - Fresh Install
echo ================================================
echo.
echo Starting Corridor OS...
echo.

REM Try different browsers
if exist "C:\Program Files\Google\Chrome\Application\chrome.exe" (
    start "" "C:\Program Files\Google\Chrome\Application\chrome.exe" --new-window "%~dp0LAUNCH-CORRIDOR-OS.html"
) else if exist "C:\Program Files (x86)\Google\Chrome\Application\chrome.exe" (
    start "" "C:\Program Files (x86)\Google\Chrome\Application\chrome.exe" --new-window "%~dp0LAUNCH-CORRIDOR-OS.html"
) else if exist "C:\Program Files\Mozilla Firefox\firefox.exe" (
    start "" "C:\Program Files\Mozilla Firefox\firefox.exe" -new-window "%~dp0LAUNCH-CORRIDOR-OS.html"
) else (
    start "" "%~dp0LAUNCH-CORRIDOR-OS.html"
)

echo Corridor OS should now be loading...
pause
