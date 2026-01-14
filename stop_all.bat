@echo off
chcp 65001 >nul
setlocal

echo 停止 NOFX 相关进程...
echo ==============================

echo 正在查找并终止 Go 进程...
taskkill /f /im go.exe 2>nul
if %errorlevel% equ 0 (echo 已终止 Go 进程) else (echo 未找到 Go 进程)

echo.
echo 正在查找并终止 Go build 进程...
taskkill /f /im gcc.exe 2>nul
if %errorlevel% equ 0 (echo 已终止 gcc 进程) else (echo 未找到 gcc 进程)

echo.
echo 正在查找并终止 Node.js 进程...
taskkill /f /im node.exe 2>nul
if %errorlevel% equ 0 (echo 已终止 Node.js 进程) else (echo 未找到 Node.js 进程)

echo.
echo 正在查找并终止 npm 进程...
taskkill /f /im npm.exe 2>nul
if %errorlevel% equ 0 (echo 已终止 npm 进程) else (echo 未找到 npm 进程)

echo.
echo 正在查找并终止 Vite 进程...
taskkill /f /im vite.exe 2>nul
if %errorlevel% equ 0 (echo 已终止 Vite 进程) else (echo 未找到 Vite 进程)

echo.
echo 所有相关进程已尝试终止。
pause