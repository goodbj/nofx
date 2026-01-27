@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    本地Ollama服务器测试脚本
echo ========================================
echo.

echo 正在测试Ollama服务器连接...
echo.

REM 测试Ollama服务器是否运行
echo [1/3] 检查Ollama服务器状态...
curl -s -o nul -w "HTTP Status: %%{http_code}\n" http://127.0.0.1:11434/api/tags
if %errorlevel% equ 0 (
    echo ✓ Ollama服务器响应正常
) else (
    echo ✗ Ollama服务器未响应，请确认Ollama服务已启动
    goto:end
)
echo.

REM 获取模型列表
echo [2/3] 检查模型列表...
curl -s http://127.0.0.1:11434/api/tags | findstr "deepseek-r1:8b" >nul
if %errorlevel% equ 0 (
    echo ✓ 找到 deepseek-r1:8b 模型
) else (
    echo ✗ 未找到 deepseek-r1:8b 模型，请确认模型已安装
    echo 可用模型：
    curl -s http://127.0.0.1:11434/api/tags
)
echo.

REM 测试模型推理
echo [3/3] 测试模型推理功能...
echo 发送测试请求到 deepseek-r1:8b 模型...
curl -s -X POST http://127.0.0.1:11434/api/generate ^
-H "Content-Type: application/json" ^
-d "{ \"model\": \"deepseek-r1:8b\", \"prompt\": \"你好，请简单介绍一下自己，用一句话回答\", \"stream\": false, \"options\": { \"temperature\": 0.7, \"num_predict\": 50 } }" > test_response.json

if exist test_response.json (
    echo ✓ 模型推理测试完成
    echo.
    echo 模型响应示例：
    type test_response.json | findstr "response"
    del test_response.json
) else (
    echo ✗ 模型推理测试失败
)

echo.
echo ========================================
echo    测试完成！
echo ========================================
echo.

echo 提示：
echo - 如果测试全部通过，您的Ollama服务器和deepseek-r1:8b模型已准备好使用
echo - 在NOFX系统中配置时，请使用 http://host.docker.internal:11434 作为API端点
echo - 本地模型推理可能需要一些时间，请耐心等待
echo.

pause
exit /b 0

:end
echo.
echo ========================================
echo    测试失败！
echo ========================================
pause