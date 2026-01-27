@echo off
echo Starting Guardian Test Environment...
echo.

echo Checking if required dependencies are installed...
cd /d %~dp0\web
if not exist "node_modules" (
    echo Installing dependencies...
    npm install
    npm install antd @ant-design/icons
)

echo.
echo Checking if backend is already running...
tasklist | findstr /I "main.exe" >nul
if "%ERRORLEVEL%"=="0" (
    echo Backend server is already running.
) else (
    echo Starting backend server on port 8888...
    start cmd /k "cd /d %~dp0 && go run main.go"
)

timeout /t 5 /nobreak >nul

echo.
tasklist | findstr /I "node.exe" | findstr /I "vite" >nul
if "%ERRORLEVEL%"=="0" (
    echo Frontend server is already running.
    echo Opening Guardian Test Page in browser...
    timeout /t 3 /nobreak >nul
    start http://localhost:5174/guardian-test
) else (
    echo Starting frontend server on port 5174...
    start cmd /k "cd /d %~dp0\web && npx vite"
    echo Opening Guardian Test Page in browser...
    timeout /t 10 /nobreak >nul
    start http://localhost:5174/guardian-test
)

echo.
echo Guardian Test Environment is now running!
echo Backend: http://localhost:8888
echo Frontend: http://localhost:5174
echo Guardian Test Page: http://localhost:5174/guardian-test
pause