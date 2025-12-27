@echo off
chcp 65001 >nul 2>&1
title Fix Python PATH
color 0E

echo.
echo  ============================================================
echo       Python PATH Diagnostic and Fix Tool
echo  ============================================================
echo.

:: Check current PATH
echo [1] Checking if Python is in PATH...
where python 2>nul
if errorlevel 1 (
    echo     Python NOT found in PATH
) else (
    echo     Python found in PATH:
    where python
    python --version 2>nul
    if errorlevel 1 (
        echo     But it's Windows Store alias, not real Python
    ) else (
        echo.
        echo     Python is working! Try running install.bat again.
        goto :end
    )
)

echo.
echo [2] Searching for Python installation...
echo.

:: Common Python installation paths
set FOUND=0

:: Check Python 3.11
if exist "C:\Python311\python.exe" (
    echo     Found: C:\Python311\python.exe
    set PYTHON_PATH=C:\Python311
    set FOUND=1
)

:: Check Python 3.10
if exist "C:\Python310\python.exe" (
    echo     Found: C:\Python310\python.exe
    set PYTHON_PATH=C:\Python310
    set FOUND=1
)

:: Check AppData Local
for /d %%i in ("%LOCALAPPDATA%\Programs\Python\Python*") do (
    if exist "%%i\python.exe" (
        echo     Found: %%i\python.exe
        set PYTHON_PATH=%%i
        set FOUND=1
    )
)

:: Check Program Files
for /d %%i in ("%ProgramFiles%\Python*") do (
    if exist "%%i\python.exe" (
        echo     Found: %%i\python.exe
        set PYTHON_PATH=%%i
        set FOUND=1
    )
)

:: Check user profile
if exist "%USERPROFILE%\AppData\Local\Programs\Python\Python311\python.exe" (
    echo     Found: %USERPROFILE%\AppData\Local\Programs\Python\Python311\python.exe
    set PYTHON_PATH=%USERPROFILE%\AppData\Local\Programs\Python\Python311
    set FOUND=1
)

if exist "%USERPROFILE%\AppData\Local\Programs\Python\Python312\python.exe" (
    echo     Found: %USERPROFILE%\AppData\Local\Programs\Python\Python312\python.exe
    set PYTHON_PATH=%USERPROFILE%\AppData\Local\Programs\Python\Python312
    set FOUND=1
)

echo.

if "%FOUND%"=="0" (
    echo [X] Python installation not found!
    echo.
    echo     Please download and install Python:
    echo     https://mirrors.huaweicloud.com/python/3.11.9/python-3.11.9-amd64.exe
    echo.
    echo     During installation, CHECK: [x] Add Python to PATH
    echo.
    goto :end
)

echo [3] Python found at: %PYTHON_PATH%
echo.
echo     Current PATH does not include this location.
echo.
echo [4] Solutions:
echo.
echo     OPTION A: Reinstall Python (Recommended)
echo     -----------------------------------------
echo     1. Uninstall current Python from Control Panel
echo     2. Download: https://mirrors.huaweicloud.com/python/3.11.9/python-3.11.9-amd64.exe
echo     3. Run installer, CHECK [x] Add Python to PATH
echo     4. Restart computer
echo     5. Run install.bat again
echo.
echo     OPTION B: Add to PATH manually
echo     -----------------------------------------
echo     1. Press Win+R, type: sysdm.cpl
echo     2. Click "Advanced" tab
echo     3. Click "Environment Variables"
echo     4. In "User variables", find "Path", click "Edit"
echo     5. Click "New", add: %PYTHON_PATH%
echo     6. Click "New", add: %PYTHON_PATH%\Scripts
echo     7. Click OK, OK, OK
echo     8. CLOSE this window and open a NEW command prompt
echo     9. Run install.bat again
echo.
echo     OPTION C: Use full path (Quick fix)
echo     -----------------------------------------
echo     Run this command in the scripts folder:
echo.
echo     "%PYTHON_PATH%\python.exe" -m venv ..\backend\venv
echo.

:end
echo.
echo ============================================================
pause

